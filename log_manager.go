// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dbratus/loghub/history"
	"github.com/dbratus/loghub/rlimit"
	"github.com/dbratus/loghub/rnglock"
	"github.com/dbratus/loghub/trace"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

const mergedFileSuffix = ".merged"

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

type getTransferChunkCmd struct {
	maxSize   int64
	entries   chan *LogEntry
	chunkChan chan logChunkInfo
}

type logChunkInfo struct {
	id   string
	size int64
}

type acceptTransferChunkCmd struct {
	id      string
	entries chan *LogEntry
	ack     chan bool
}

type delTransferChunkCmd struct {
	id  string
	ack chan bool
}

type defaultLogManager struct {
	home                    string
	writeChan               chan *LogEntry
	readChan                chan readLogCmd
	sizeChan                chan chan int64
	truncateChan            chan truncateLogCmd
	getTransferChunkChan    chan getTransferChunkCmd
	acceptTransferChunkChan chan acceptTransferChunkCmd
	delTransferChunkChan    chan delTransferChunkCmd
	closeChan               chan chan bool
}

type logSourceInfo struct {
	lock    *rnglock.RangeLock
	history *history.History
}

func newLogSourceInfo() *logSourceInfo {
	return &logSourceInfo{rnglock.New(), history.New(time.Hour)}
}

func NewDefaultLogManager(home string) LogManager {
	var mg = &defaultLogManager{
		home,
		make(chan *LogEntry),
		make(chan readLogCmd),
		make(chan chan int64),
		make(chan truncateLogCmd),
		make(chan getTransferChunkCmd),
		make(chan acceptTransferChunkCmd),
		make(chan delTransferChunkCmd),
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

func (mg *defaultLogManager) GetTransferChunk(maxSize int64, entries chan *LogEntry) (id string, size int64, found bool) {
	chunkChan := make(chan logChunkInfo)
	mg.getTransferChunkChan <- getTransferChunkCmd{maxSize, entries, chunkChan}
	chunk, found := <-chunkChan

	id = chunk.id
	size = chunk.size

	return
}

func (mg *defaultLogManager) AcceptTransferChunk(id string, entries chan *LogEntry) chan bool {
	ack := make(chan bool)
	mg.acceptTransferChunkChan <- acceptTransferChunkCmd{id, entries, ack}
	return ack
}

func (mg *defaultLogManager) DeleteTransferChunk(id string) {
	ack := make(chan bool)
	mg.delTransferChunkChan <- delTransferChunkCmd{id, ack}
	<-ack
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
	close(mg.getTransferChunkChan)
	close(mg.acceptTransferChunkChan)
	close(mg.delTransferChunkChan)

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
	t := timestampToTime(timestamp)
	return fmt.Sprintf("%.4d.%.2d.%.2d.%.2d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func getRangeByFileName(fileName string) (start int64, end int64) {
	if st, err := time.Parse("2006.01.02.15", fileName); err != nil {
		start = minTimestamp
		end = maxTimestamp

	} else {
		start = timeToTimestamp(st)
		end = timeToTimestamp(st.Add(time.Hour - time.Nanosecond))
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

func getLogFileNamesForRange(sources []string, minTimestamp int64, maxTimestamp int64, fileNames chan string) {
	minHour := timeToTimestamp(timestampToTime(minTimestamp).Truncate(time.Hour))
	maxHour := timeToTimestamp(timestampToTime(maxTimestamp).Truncate(time.Hour))

	for _, src := range sources {
		for curTs := minHour; curTs <= maxHour; curTs += int64(time.Hour) {
			fileNames <- getSourceDirName(src) + "/" + getFileNameForTimestamp(curTs)
		}
	}

	close(fileNames)
}

func initLogManager(home string) (logSources map[string]*logSourceInfo, size *int64, initialized bool) {
	initialized = true
	logSources = make(map[string]*logSourceInfo)
	size = new(int64)

	if homeDir, err := os.Open(home); err == nil {
		defer homeDir.Close()

		if dirnames, err := homeDir.Readdirnames(0); err == nil {
			for _, dir := range dirnames {
				if src, err := getSourceNameByDir(dir); err == nil {
					srcInfo := newLogSourceInfo()

					logSources[src] = srcInfo

					srcDirName := home + "/" + dir

					if srcDir, err := os.Open(srcDirName); err == nil {
						defer srcDir.Close()

						if logFiles, err := srcDir.Readdir(0); err == nil {
							for _, dirFile := range logFiles {
								if strings.HasSuffix(dirFile.Name(), mergedFileSuffix) {
									//The file is being merged.

									logFileName := strings.TrimSuffix(dirFile.Name(), mergedFileSuffix)
									logFilePath := srcDirName + "/" + logFileName

									//If the source of the merge exists,
									//the file may not have been merged completely,
									//so it must be removed; otherwise, it must be
									//remaned and treated as a normal log file.
									if _, err = os.Stat(logFileName); os.IsNotExist(err) {
										os.Rename(srcDirName+"/"+dirFile.Name(), logFilePath)

										*size += logFileSize(dirFile)
										start, _ := getRangeByFileName(logFileName)

										srcInfo.history.Insert(timestampToTime(start))
									} else {
										os.Remove(srcDirName + "/" + dirFile.Name())
									}

								} else {
									//The file is a normal log file.

									*size += logFileSize(dirFile)
									start, _ := getRangeByFileName(dirFile.Name())

									srcInfo.history.Insert(timestampToTime(start))
								}
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

func (mg *defaultLogManager) run() {
	logSources, closedSize, initialized := initLogManager(mg.home)
	openLogFiles := make(map[string]LogStorage)
	logFileTimeouts := make(map[string]*int64)
	closeLogFileChan := make(chan string, 1000)

	maxOpenFiles := rlimit.GetMaxOpenFiles()

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
		for ts := atomic.LoadInt64(timeout); ; {
			now := time.Now()
			tm := timestampToTime(ts)

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
			if atomic.AddInt64(closedSize, logFile.Size()) < 0 {
				logManagerTrace.Warnf("Closed size became negative on closing of %s.", logFileToClose)
			}

			logFile.Close()

			delete(openLogFiles, logFileToClose)
			delete(logFileTimeouts, logFileToClose)
		}
	}

	getLogFile := func(fileName string, create bool, register bool) (LogStorage, error) {
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
				if srcInfo, found := logSources[src]; found {
					start, end := getRangeByFileName(fileName)

					//logManagerTrace.Debugf("Locking range %s %d-%d for write.", src, start, end)
					lck := srcInfo.lock.Lock(opCnt, start, end, false)

					onClose(logFileToClose)

					//logManagerTrace.Debugf("Unlocking range %s %d-%d.", src, start, end)
					srcInfo.lock.Unlock(lck)
				}
			}

			//Opening the file.
			if logFile, err := OpenLogFile(mg.home+"/"+fileName, create); err == nil {
				if register {
					openLogFiles[fileName] = logFile

					//Tracking the closed files' size.
					if atomic.AddInt64(closedSize, -logFile.Size()) < 0 {
						logManagerTrace.Warnf("Closed size became negative on opening of %s.", fileName)
					}

					timeout := new(int64)
					*timeout = timeToTimestamp(time.Now().Add(logFileTimeout))
					logFileTimeouts[fileName] = timeout

					go waitAndCloseFile(fileName, timeout)
				}

				return logFile, nil
			} else {
				return nil, err
			}
		} else {
			if timeout, found := logFileTimeouts[fileName]; found {
				atomic.StoreInt64(timeout, timeToTimestamp(time.Now().Add(logFileTimeout)))
			}

			return logFile, nil
		}
	}

	queryLogSources := func(sources []string, cmd readLogCmd) {
		fileNames := make(chan string)
		locksHold := make(map[string]rnglock.LockId)

		for _, src := range sources {
			if srcInfo, found := logSources[src]; found {
				//logManagerTrace.Debugf("Locking range %s %d-%d for read.", src, cmd.query.From, cmd.query.To)
				locksHold[src] = srcInfo.lock.Lock(opCnt, cmd.query.From, cmd.query.To, true)
			}
		}

		unlockAll := func() {
			for src, lck := range locksHold {
				if srcInfo, found := logSources[src]; found {
					//logManagerTrace.Debugf("Unlocking range %s %d-%d.", src, cmd.query.From, cmd.query.To)
					srcInfo.lock.Unlock(lck)
				}
			}
		}

		go getLogFileNamesForRange(sources, cmd.query.From, cmd.query.To, fileNames)

		var results chan *LogEntry = nil

		for fileName := range fileNames {
			if logFile, err := getLogFile(fileName, false, true); err == nil {
				src, _ := parseLogFileName(fileName)

				res := make(chan *LogEntry)
				resWithSrc := make(chan *LogEntry)

				//Sources already selected, so the source filter
				//is not required. Moreover, as the entries have no
				//sources stored due to optimisation, non-empty source
				//filter will always return empty result.
				cmd.query.Source = ""

				logFile.ReadLog(cmd.query, res)

				go func() {
					for ent := range res {
						//Restoring the source.
						ent.Source = src
						resWithSrc <- ent
					}

					close(resWithSrc)
				}()

				if results == nil {
					results = resWithSrc
				} else {
					merged := make(chan *LogEntry)
					go MergeLogs(results, resWithSrc, merged)
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
				ent.Timestamp = timeToTimestamp(time.Now())
			}

			logFileToWrite := getLogFileNameForEntry(ent)
			var srcInfo *logSourceInfo

			if inf, found := logSources[ent.Source]; !found {
				srcInfo = newLogSourceInfo()
				logSources[ent.Source] = srcInfo
			} else {
				srcInfo = inf
			}

			srcInfo.history.Insert(timestampToTime(ent.Timestamp))

			if logFile, err := getLogFile(logFileToWrite, true, true); err == nil {
				//Removing source to make the entry smaller.
				//The source can be restored from the log file directory name.
				ent.Source = ""

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

	onSize := func(szChan chan int64) {
		size := atomic.LoadInt64(closedSize)

		for _, logFile := range openLogFiles {
			size += logFile.Size()
		}

		szChan <- size
		close(szChan)
	}

	deleteLogSource := func(src string) {
		if srcInfo, found := logSources[src]; found {
			srcInfo.lock.Close()
			delete(logSources, src)
		}
	}

	getFileNamesForTruncation := func(hist *history.History, src string, limit int64) (fileNames []string, removeDir bool) {
		fileNames = make([]string, 0, 100)
		removeDir = false

		if hist.IsEmpty() {
			removeDir = true
			return
		}

		srcDirName := getSourceDirName(src)
		minTs := timeToTimestamp(hist.Start())

		for minTs <= limit {
			fileNames = append(fileNames, srcDirName+"/"+getFileNameForTimestamp(minTs))

			hist.Truncate(timestampToTime(minTs))

			if hist.IsEmpty() {
				removeDir = true
				deleteLogSource(src)
				break
			} else {
				minTs = timeToTimestamp(hist.Start())
			}
		}

		return
	}

	truncateLogSource := func(lock *rnglock.RangeLock, lck rnglock.LockId, src string, fileNames []string, removeDir bool) {
		defer lock.Unlock(lck)

		srcDirName := getSourceDirName(src)

		logManagerTrace.Debugf("Truncating log source %s.", src)

		for _, fileName := range fileNames {
			fileName = mg.home + "/" + fileName

			logManagerTrace.Debugf("Deleting log file %s.", fileName)

			size := int64(0)

			if stat, err := os.Stat(fileName); err == nil {
				size = logFileSize(stat)
			} else {
				logManagerTrace.Errorf("Failed to get stat of log file: %s.", err.Error())
			}

			if err := os.Remove(fileName); err == nil {
				if atomic.AddInt64(closedSize, -size) < 0 {
					logManagerTrace.Warnf("Closed size became negative on truncating of %s.", fileName)
				}
			} else {
				logManagerTrace.Errorf("Failed to remove log file: %s.", err.Error())
			}
		}

		if removeDir {
			logManagerTrace.Debugf("Deleting source directory %s.", srcDirName)

			if err := os.RemoveAll(mg.home + "/" + srcDirName); err != nil {
				logManagerTrace.Errorf("Failed to remove source directory on truncation: %s.", err.Error())
			}
		}
	}

	onTruncate := func(cmd truncateLogCmd) {
		sources := filterLogSources(cmd.source)

		if len(sources) > 0 {
			for _, src := range sources {
				if srcInfo, found := logSources[src]; found {
					lck := srcInfo.lock.Lock(opCnt, minTimestamp, cmd.limit, false)

					fileNames, removeDir := getFileNamesForTruncation(srcInfo.history, src, cmd.limit)

					for _, fileName := range fileNames {
						if _, found := openLogFiles[fileName]; found {
							onClose(fileName)
						}
					}

					go truncateLogSource(srcInfo.lock, lck, src, fileNames, removeDir)
				}
			}
		}
	}

	onGetTransferChunk := func(cmd getTransferChunkCmd) {
		for src, srcInfo := range logSources {
			if !srcInfo.history.IsEmpty() && srcInfo.history.Start().Before(time.Now().Truncate(time.Hour)) {
				chunkId := getSourceDirName(src) + "/" + getFileNameForTimestamp(timeToTimestamp(srcInfo.history.Start()))
				fileName := mg.home + "/" + chunkId

				if stat, err := os.Stat(fileName); err == nil && logFileSize(stat) < cmd.maxSize {
					rangeStart := timeToTimestamp(srcInfo.history.Start())
					rangeEnd := timeToTimestamp(srcInfo.history.Start().Add(time.Hour - time.Nanosecond))

					lck := srcInfo.lock.Lock(opCnt, rangeStart, rangeEnd, true)

					if logFile, err := getLogFile(chunkId, false, true); err == nil {
						cmd.chunkChan <- logChunkInfo{chunkId, logFile.Size()}
						close(cmd.chunkChan)

						results := make(chan *LogEntry)
						logFile.ReadLog(&LogQuery{rangeStart, rangeEnd, minSeverity, maxSeverity, ""}, results)

						go func() {
							ForwardLog(results, cmd.entries)
							srcInfo.lock.Unlock(lck)
						}()

						return
					} else {
						srcInfo.lock.Unlock(lck)
					}
				}
			}
		}

		close(cmd.entries)
		close(cmd.chunkChan)
	}

	onAcceptTransferChunk := func(cmd acceptTransferChunkCmd) {
		src, ts := parseLogFileName(cmd.id)

		var srcInfo *logSourceInfo

		if inf, found := logSources[src]; !found {
			srcInfo = newLogSourceInfo()
		} else {
			srcInfo = inf
		}

		fileName := mg.home + "/" + cmd.id
		rangeStart, rangeEnd := getRangeByFileName(ts)

		lck := srcInfo.lock.Lock(opCnt, rangeStart, rangeEnd, false)
		srcInfo.history.Insert(timestampToTime(rangeStart))

		onError := func() {
			PurgeLog(cmd.entries)
			srcInfo.lock.Unlock(lck)
			cmd.ack <- false
		}

		onClose(cmd.id)

		if stat, err := os.Stat(fileName); os.IsNotExist(err) {
			if logFile, err := getLogFile(cmd.id, true, false); err == nil {
				go func() {
					for ent := range cmd.entries {
						//Removing source to make the entry smaller.
						//The source can be restored from the log file directory name.
						ent.Source = ""

						logFile.WriteLog(ent)
					}

					logFile.Close()

					if stat, err = os.Stat(fileName); err == nil {
						if atomic.AddInt64(closedSize, logFileSize(stat)) < 0 {
							logManagerTrace.Warnf("Closed size became negative on accepting/creating of %s.", fileName)
						}
					}

					srcInfo.lock.Unlock(lck)
					cmd.ack <- true
				}()
			} else {
				logManagerTrace.Errorf("Failed to get log file for transfer chunk: %s.", err.Error())
				go onError()
				return
			}
		} else if err == nil {
			if atomic.AddInt64(closedSize, -logFileSize(stat)) < 0 {
				logManagerTrace.Warnf("Closed size became negative on accepting/merging of %s.", fileName)
			}

			if logFile, err := getLogFile(cmd.id, false, false); err == nil {
				entriesRead := make(chan *LogEntry)
				mergedEntries := make(chan *LogEntry)

				logFile.ReadLog(&LogQuery{rangeStart, rangeEnd, minSeverity, maxSeverity, ""}, entriesRead)
				go MergeLogs(cmd.entries, entriesRead, mergedEntries)

				mergedFileName := fileName + mergedFileSuffix

				if mergedLogFile, err := OpenLogFile(mergedFileName, true); err == nil {
					go func() {
						for ent := range mergedEntries {
							//Removing source to make the entry smaller.
							//The source can be restored from the log file directory name.
							ent.Source = ""

							mergedLogFile.WriteLog(ent)
						}

						mergedLogFile.Close()
						logFile.Close()

						if err := os.Remove(fileName); err != nil {
							logManagerTrace.Errorf("Failed to remove log file: %s.", err.Error())
							srcInfo.lock.Unlock(lck)
							cmd.ack <- false
							return
						}

						if err := os.Rename(mergedFileName, fileName); err != nil {
							logManagerTrace.Errorf("Failed to rename log file: %s.", err.Error())
							srcInfo.lock.Unlock(lck)
							cmd.ack <- false
							return
						}

						if stat, err = os.Stat(fileName); err == nil {
							if atomic.AddInt64(closedSize, logFileSize(stat)) < 0 {
								logManagerTrace.Warnf("Closed size became negative on merging of %s.", fileName)
							}
						} else {
							logManagerTrace.Errorf("Failed to get stat of renamed file: %s.", err.Error())
						}

						srcInfo.lock.Unlock(lck)
						cmd.ack <- true
					}()
				} else {
					logManagerTrace.Errorf("Failed to open log file for merge: %s.", err.Error())

					go func() {
						PurgeLog(mergedEntries)
						srcInfo.lock.Unlock(lck)
						cmd.ack <- false
					}()
				}
			} else {
				logManagerTrace.Errorf("Failed to get log file for merge: %s.", err.Error())
				go onError()
				return
			}
		} else {
			logManagerTrace.Errorf("Failed to get stat: %s.", err.Error())
			go onError()
			return
		}
	}

	onDeleteTransferChunk := func(cmd delTransferChunkCmd) {
		src, ts := parseLogFileName(cmd.id)

		if srcInfo, found := logSources[src]; found {
			fileName := mg.home + "/" + cmd.id
			dirName := mg.home + "/" + getSourceDirName(src)

			onClose(cmd.id)

			if stat, err := os.Stat(fileName); err == nil {
				rangeStart, rangeEnd := getRangeByFileName(ts)

				lck := srcInfo.lock.Lock(opCnt, rangeStart, rangeEnd, false)

				srcInfo.history.Delete(timestampToTime(rangeStart))
				removeDir := false

				if srcInfo.history.IsEmpty() {
					removeDir = true
					deleteLogSource(src)
				}

				go func(removeDir bool) {
					logManagerTrace.Debugf("Removing log file %s on transfer chunk deletion.", fileName)

					if err := os.Remove(fileName); err == nil {
						if atomic.AddInt64(closedSize, -logFileSize(stat)) < 0 {
							logManagerTrace.Warnf("Closed size became negative on deleting of %s.", fileName)
						}
					} else {
						logManagerTrace.Errorf("Failed to remove log file on transfer chunk deletion: %s.", err.Error())
					}

					if removeDir {
						logManagerTrace.Debugf("Removing source directory %s on transfer chunk deletion.", dirName)

						if err := os.RemoveAll(dirName); err != nil {
							logManagerTrace.Errorf("Failed to remove source directory on transfer chunk deletion: %s.", err.Error())
						}
					}

					srcInfo.lock.Unlock(lck)
					cmd.ack <- true
					close(cmd.ack)
				}(removeDir)
			} else {
				logManagerTrace.Errorf("Failed to get log file stat on transfer chunk deletion: %s.", err.Error())

				cmd.ack <- false
				close(cmd.ack)
			}
		} else {
			cmd.ack <- false
			close(cmd.ack)
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

		case cmd, ok := <-mg.getTransferChunkChan:
			if ok {
				onGetTransferChunk(cmd)
			}

		case cmd, ok := <-mg.acceptTransferChunkChan:
			if ok {
				onAcceptTransferChunk(cmd)
			}

		case cmd, ok := <-mg.delTransferChunkChan:
			if ok {
				onDeleteTransferChunk(cmd)
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

			for cmd := range mg.getTransferChunkChan {
				close(cmd.entries)
				close(cmd.chunkChan)
			}

			for cmd := range mg.acceptTransferChunkChan {
				PurgeLog(cmd.entries)
				close(cmd.ack)
			}

			for cmd := range mg.delTransferChunkChan {
				onDeleteTransferChunk(cmd)
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
