// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dbratus/loghub/rnglock"
	"github.com/dbratus/loghub/trace"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var logManagerTrace = trace.New("LogManager")

var logFileTimeout = time.Second * 30

type readLogCmd struct {
	query   *LogQuery
	entries chan *LogEntry
}

type truncateLogCmd struct {
	source string
	limit  int64
}

type defaultLogManager struct {
	home         string
	writeChan    chan *LogEntry
	readChan     chan readLogCmd
	sizeChan     chan chan int64
	truncateChan chan truncateLogCmd
	closeChan    chan chan bool
}

func NewDefaultLogManager(home string) LogManager {
	var mg = &defaultLogManager{
		home,
		make(chan *LogEntry),
		make(chan readLogCmd),
		make(chan chan int64),
		make(chan truncateLogCmd),
		make(chan chan bool),
	}

	go mg.run()

	return mg
}

func (mg *defaultLogManager) WriteLog(ent *LogEntry) {
	mg.writeChan <- ent
}

func (mg *defaultLogManager) ReadLog(q *LogQuery, entries chan *LogEntry) {
	mg.readChan <- readLogCmd{q, entries}
}

func (mg *defaultLogManager) Truncate(source string, limit int64) {
	mg.truncateChan <- truncateLogCmd{source, limit}
}

func (mg *defaultLogManager) Size() int64 {
	result := make(chan int64)
	mg.sizeChan <- result
	return <-result
}

func (mg *defaultLogManager) Close() {
	close(mg.writeChan)
	close(mg.readChan)
	close(mg.sizeChan)
	close(mg.truncateChan)

	ack := make(chan bool)
	mg.closeChan <- ack
	<-ack

	close(mg.closeChan)
}

func getSourceDirName(source string) string {
	return base64.URLEncoding.EncodeToString([]byte(source))
}

func getSourceNameByDir(dir string) (string, error) {
	if name, err := base64.URLEncoding.DecodeString(dir); err == nil {
		return string(name), nil
	} else {
		return "", err
	}
}

func getFileNameForTimestamp(timestamp int64) string {
	t := time.Unix(0, timestamp)
	return fmt.Sprintf("%.4d.%.2d.%.2d.%.2d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func getRangeByFileName(fileName string) (start int64, end int64) {
	if st, err := time.Parse("2006.01.02.15", fileName); err != nil {
		start = minTimestamp
		end = maxTimestamp

	} else {
		start = st.UnixNano()
		end = st.Add(time.Hour - time.Millisecond).UnixNano()
	}

	return
}

func getLogFileNameForEntry(entry *LogEntry) string {
	return getSourceDirName(entry.Source) + "/" + getFileNameForTimestamp(entry.Timestamp)
}

func parseLogFileName(fileName string) (source string, ts string) {
	slashIdx := strings.Index(fileName, "/")
	srcDir := fileName[:slashIdx]

	if s, err := getSourceNameByDir(srcDir); err == nil {
		source = s
	} else {
		source = ""
	}

	ts = fileName[slashIdx+1:]
	return
}

func initLogManager(home string) (logSources map[string]*rnglock.RangeLock, size int64, initialized bool) {
	initialized = true
	logSources = make(map[string]*rnglock.RangeLock)
	size = 0

	if homeDir, err := os.Open(home); err == nil {
		defer homeDir.Close()

		if dirnames, err := homeDir.Readdirnames(0); err == nil {
			for _, dir := range dirnames {
				if src, err := getSourceNameByDir(dir); err == nil {
					logSources[src] = rnglock.New()
				}

				if srcDir, err := os.Open(home + "/" + dir); err == nil {
					defer srcDir.Close()

					if logFiles, err := srcDir.Readdir(0); err == nil {
						for _, logFile := range logFiles {
							size += logFile.Size()
						}
					} else {
						logManagerTrace.Errorf("Initialization failed: %s.", err.Error())
						initialized = false
						return
					}

				} else {
					logManagerTrace.Errorf("Initialization failed: %s.", err.Error())
					initialized = false
					return
				}
			}
		} else {
			logManagerTrace.Errorf("Initialization failed: %s.", err.Error())
			initialized = false
		}
	} else {
		logManagerTrace.Errorf("Initialization failed: %s.", err.Error())
		initialized = false
	}

	return
}

func getLogFileNamesForRange(sources []string, minTimestamp int64, maxTimestamp int64, fileNames chan string) {
	minHour := time.Unix(0, minTimestamp).Truncate(time.Hour).UnixNano()
	maxHour := time.Unix(0, maxTimestamp).Truncate(time.Hour).UnixNano()

	for _, src := range sources {
		for curTs := minHour; curTs <= maxHour; curTs += int64(time.Hour) {
			fileNames <- getSourceDirName(src) + "/" + getFileNameForTimestamp(curTs)
		}
	}

	close(fileNames)
}

func getMaxOpenFiles() uint64 {
	var lim syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	return (lim.Cur / 4) * 3
}

func (mg *defaultLogManager) run() {
	logSources, closedSize, initialized := initLogManager(mg.home)
	openLogFiles := make(map[string]LogManager)
	logFileTimeouts := make(map[string]*int64)
	closeLogFileChan := make(chan string)

	maxOpenFiles := getMaxOpenFiles()

	opCnt := int64(0)

	logManagerTrace.Debugf("Max. open files %d.", maxOpenFiles)

	filterLogSources := func(srcFilter string) []string {
		sources := make([]string, 0, 100)

		if srcFilter != "" {
			if re, err := regexp.Compile(srcFilter); err == nil {
				for src, _ := range logSources {
					if re.MatchString(src) {
						sources = append(sources, src)
					}
				}
			}
		} else {
			for src, _ := range logSources {
				sources = append(sources, src)
			}
		}

		return sources
	}

	waitAndCloseFile := func(fileName string, timeout *int64) {
		for tmUnix := atomic.LoadInt64(timeout); ; {
			now := time.Now()
			tm := time.Unix(0, tmUnix)

			if tm.After(now) {
				<-time.After(tm.Sub(now))
			} else {
				break
			}
		}

		closeLogFileChan <- fileName
	}

	onClose := func(logFileToClose string) {
		if logFile, found := openLogFiles[logFileToClose]; found {
			//Tracking the closed files' size.
			closedSize += logFile.Size()

			logFile.Close()

			delete(openLogFiles, logFileToClose)
			delete(logFileTimeouts, logFileToClose)
		}
	}

	getLogFile := func(fileName string, create bool) (LogManager, error) {
		if logFile, found := openLogFiles[fileName]; !found {
			srcDir := mg.home + "/" + fileName[0:strings.LastIndex(fileName, "/")]

			//Creating source directory if not exists.
			if srcDirStat, err := os.Stat(srcDir); err != nil {
				if os.IsNotExist(err) && create {
					if err = os.Mkdir(srcDir, 0777); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				if !srcDirStat.IsDir() {
					return nil, errors.New("Source directory " + srcDir + " is not a directory")
				}
			}

			//If the resource limit on open files is hit,
			//trying to close one of them.
			if uint64(len(openLogFiles)) == maxOpenFiles {
				closestTimeout := int64(0)
				logFileToClose := ""

				//Selecting the file with the closest timeout.
				for fileName, timeout := range logFileTimeouts {
					if logFileToClose == "" || *timeout < closestTimeout {
						logFileToClose = fileName
						closestTimeout = *timeout
					}
				}

				src, fileName := parseLogFileName(logFileToClose)

				//Locking the file's range and closing the file.
				if slock, found := logSources[src]; found {
					start, end := getRangeByFileName(fileName)

					//logManagerTrace.Debugf("Locking range %s %d-%d for write.", src, start, end)
					lck := slock.Lock(opCnt, start, end, false)

					onClose(logFileToClose)

					//logManagerTrace.Debugf("Unlocking range %s %d-%d.", src, start, end)
					slock.Unlock(lck)
				}
			}

			//Opening the file.
			if logFile, err := OpenLogFile(mg.home+"/"+fileName, create); err == nil {
				openLogFiles[fileName] = logFile

				//Tracking the closed files' size.
				closedSize -= logFile.Size()

				timeout := new(int64)
				*timeout = time.Now().Add(logFileTimeout).UnixNano()
				logFileTimeouts[fileName] = timeout

				go waitAndCloseFile(fileName, timeout)

				return logFile, nil
			} else {
				return nil, err
			}
		} else {
			if timeout, found := logFileTimeouts[fileName]; found {
				atomic.StoreInt64(timeout, time.Now().Add(logFileTimeout).UnixNano())
			}

			return logFile, nil
		}
	}

	queryLogSources := func(sources []string, cmd readLogCmd) {
		fileNames := make(chan string)
		locksHold := make(map[string]rnglock.LockId)

		for _, src := range sources {
			if slock, found := logSources[src]; found {
				//logManagerTrace.Debugf("Locking range %s %d-%d for read.", src, cmd.query.From, cmd.query.To)
				locksHold[src] = slock.Lock(opCnt, cmd.query.From, cmd.query.To, true)
			}
		}

		unlockAll := func() {
			for src, lck := range locksHold {
				if slock, found := logSources[src]; found {
					//logManagerTrace.Debugf("Unlocking range %s %d-%d.", src, cmd.query.From, cmd.query.To)
					slock.Unlock(lck)
				}
			}
		}

		go getLogFileNamesForRange(sources, cmd.query.From, cmd.query.To, fileNames)

		var results chan *LogEntry = nil

		for fileName := range fileNames {
			if logFile, err := getLogFile(fileName, false); err == nil {
				res := make(chan *LogEntry)

				logFile.ReadLog(cmd.query, res)

				if results == nil {
					results = res
				} else {
					merged := make(chan *LogEntry)
					go MergeLogs(results, res, merged)
					results = merged
				}
			} else {
				if !os.IsNotExist(err) {
					logManagerTrace.Errorf("Failed to get log file: %s.", err.Error())
				}
			}
		}

		if results != nil {
			go func() {
				ForwardLog(results, cmd.entries)
				unlockAll()
			}()
		} else {
			close(cmd.entries)
			unlockAll()
		}
	}

	onWrite := func(ent *LogEntry) {
		if initialized {
			if ent.Timestamp == 0 {
				ent.Timestamp = time.Now().UnixNano()
			}

			logFileToWrite := getLogFileNameForEntry(ent)

			if _, found := logSources[ent.Source]; !found {
				logSources[ent.Source] = rnglock.New()
			}

			if logFile, err := getLogFile(logFileToWrite, true); err == nil {
				logFile.WriteLog(ent)
			} else {
				logManagerTrace.Errorf("Failed to obtain log file: %s.", err.Error())
			}
		}
	}

	onRead := func(cmd readLogCmd) {
		if initialized {
			sources := filterLogSources(cmd.query.Source)

			if len(sources) > 0 {
				queryLogSources(sources, cmd)
			} else {
				close(cmd.entries)
			}
		} else {
			close(cmd.entries)
		}
	}

	onSize := func(sz chan int64) {
		size := closedSize

		for _, logFile := range openLogFiles {
			size += logFile.Size()
		}

		sz <- size
		close(sz)
	}

	truncateLogSource := func(slock *rnglock.RangeLock, lck rnglock.LockId, src string, limit int64) {
		defer slock.Unlock(lck)

		srcDirName := mg.home + "/" + getSourceDirName(src)
		limitFName := getFileNameForTimestamp(limit)

		if dir, err := os.Open(srcDirName); err == nil {
			if fnames, err := dir.Readdirnames(0); err == nil {

				for _, name := range fnames {
					if name <= limitFName {
						os.Remove(srcDirName + "/" + name)
					}
				}

			} else {
				logManagerTrace.Errorf("Failed to read names in the source dir. %s: %s.", srcDirName, err.Error())
			}
		} else {
			logManagerTrace.Errorf("Failed to open source directory %s: %s.", srcDirName, err.Error())
		}
	}

	onTruncate := func(cmd truncateLogCmd) {
		sources := filterLogSources(cmd.source)

		if len(sources) > 0 {
			limitFName := getFileNameForTimestamp(cmd.limit)

			for _, src := range sources {
				if slock, found := logSources[src]; found {
					lck := slock.Lock(opCnt, minTimestamp, cmd.limit, false)

					for fileName, _ := range openLogFiles {
						fileSrc, fileTs := parseLogFileName(fileName)

						if fileSrc == src && fileTs <= limitFName {
							onClose(fileName)
						}
					}

					go truncateLogSource(slock, lck, src, cmd.limit)
				}
			}
		}
	}

	for {
		opCnt++

		select {
		case ent, ok := <-mg.writeChan:
			if ok {
				onWrite(ent)
			}

		case logFileToClose, ok := <-closeLogFileChan:
			if ok {
				onClose(logFileToClose)
			}

		case cmd, ok := <-mg.readChan:
			if ok {
				onRead(cmd)
			}

		case sz, ok := <-mg.sizeChan:
			if ok {
				onSize(sz)
			}

		case cmd, ok := <-mg.truncateChan:
			if ok {
				onTruncate(cmd)
			}

		case ack := <-mg.closeChan:
			logManagerTrace.Debug("Closing")

			for ent := range mg.writeChan {
				onWrite(ent)
			}

			for cmd := range mg.readChan {
				onRead(cmd)
			}

			for sz := range mg.sizeChan {
				onSize(sz)
			}

			for _, logFile := range openLogFiles {
				logFile.Close()
			}

			ack <- true
			close(ack)
			return
		}
	}
}
