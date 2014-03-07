/*
	This file is part of LogHub.

	LogHub is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	LogHub is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with LogHub.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

var logFileCloseDelay = time.Second * 10

type LogManagerStat struct {
	Capacity int64
	Size     int64
}

type LogManager interface {
	Logger

	Close()
	Stat() *LogManagerStat
}

type defaultLogManager struct {
	home      string
	writeChan chan *LogEntry
	readChan  chan *LogQuery
	closeChan chan *logClose
}

func NewDefaultLogManager(home string) LogManager {
	var mg = &defaultLogManager{home, make(chan *LogEntry), make(chan *LogQuery), make(chan *logClose)}

	go mg.run()

	return mg
}

func (mg *defaultLogManager) WriteLog(ent *LogEntry) {
	mg.writeChan <- ent
}

func (mg *defaultLogManager) ReadLog(q *LogQuery) {
	mg.readChan <- q
}

func (mg *defaultLogManager) Stat() *LogManagerStat {
	//TODO: Implement.
	return nil
}

func (mg *defaultLogManager) Close() {
	close(mg.writeChan)
	close(mg.readChan)

	ack := make(chan bool)
	closeCmd := &logClose{ack}

	mg.closeChan <- closeCmd
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
	return fmt.Sprintf("%d.%d.%d.%d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func getLogFileNameForEntry(entry *LogEntry) string {
	return getSourceDirName(entry.Source) + "/" + getFileNameForTimestamp(entry.Timestamp)
}

func initLogManager(home string) (logSources map[string]bool, initialized bool) {
	initialized = false
	logSources = make(map[string]bool, 0)

	if homeDir, err := os.Open(home); err == nil {
		if dirnames, err := homeDir.Readdirnames(0); err == nil {
			for _, dir := range dirnames {
				if src, err := getSourceNameByDir(dir); err == nil {
					logSources[src] = true
				}
			}

			initialized = true
		} else {
			println("Failed to initialize log manager:", err.Error())
		}
	} else {
		println("Failed to initialize log manager:", err.Error())
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

func (mg *defaultLogManager) run() {
	logSources, initialized := initLogManager(mg.home)
	openLogFiles := make(map[string]*LogFile)
	closeLogFileChan := make(chan string)
	closeLogFileChanClosed := new(int32)

	waitAndCloseFile := func(fileName string) {
		<-time.After(logFileCloseDelay)

		if atomic.LoadInt32(closeLogFileChanClosed) == 0 {
			closeLogFileChan <- fileName
		}
	}

	getLogFile := func(fileName string, create bool) (*LogFile, error) {
		if logFile, found := openLogFiles[fileName]; !found {
			srcDir := mg.home + "/" + fileName[0:strings.LastIndex(fileName, "/")]

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

			if logFile, err := OpenLogFile(mg.home+"/"+fileName, create); err == nil {
				openLogFiles[fileName] = logFile

				go waitAndCloseFile(fileName)

				return logFile, nil
			} else {
				return nil, err
			}
		} else {
			//TODO: Prolongate log file timeout on access.
			return logFile, nil
		}
	}

	queryLogSources := func(sources []string, q *LogQuery) {
		fileNames := make(chan string)

		go getLogFileNamesForRange(sources, q.From, q.To, fileNames)

		var results chan *LogEntry = nil

		for fileName := range fileNames {
			if logFile, err := getLogFile(fileName, false); err == nil {
				subQuery := *q
				subQuery.Result = make(chan *LogEntry)

				logFile.ReadLog(&subQuery)

				if results == nil {
					results = subQuery.Result
				} else {
					merged := make(chan *LogEntry)
					go MergeLogs(results, subQuery.Result, merged)
					results = merged
				}
			} else {
				if !os.IsNotExist(err) {
					println("Failed to get log file:", err.Error())
				}
			}
		}

		go func() {
			if results != nil {
				for ent := range results {
					q.Result <- ent
				}
			}

			close(q.Result)
		}()
	}

	onWrite := func(ent *LogEntry) {
		if initialized {
			ent.Timestamp = time.Now().UnixNano()
			logFileToWrite := getLogFileNameForEntry(ent)

			if logFile, err := getLogFile(logFileToWrite, true); err == nil {
				logFile.WriteLog(ent)
				logSources[ent.Source] = true
			} else {
				println("Failed to obtain log file :", err.Error())
			}
		}
	}

	onRead := func(q *LogQuery) {
		if initialized {
			sources := make([]string, 0, 100)

			if q.Source != "" {
				if re, err := regexp.Compile(q.Source); err == nil {
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

			if len(sources) > 0 {
				queryLogSources(sources, q)
			} else {
				close(q.Result)
			}
		} else {
			close(q.Result)
		}
	}

	for run := true; run; {
		select {
		case ent, ok := <-mg.writeChan:
			if ok {
				onWrite(ent)
			}

		case logFileToClose, ok := <-closeLogFileChan:
			if ok {
				//TODO: Prolongate log file timeout on access.
				if logFile, found := openLogFiles[logFileToClose]; found {
					logFile.Close()
				}
			}

		case q, ok := <-mg.readChan:
			if ok {
				onRead(q)
			}

		case cmd := <-mg.closeChan:
			for ent := range mg.writeChan {
				onWrite(ent)
			}

			for q := range mg.readChan {
				onRead(q)
			}

			atomic.AddInt32(closeLogFileChanClosed, 1)
			close(closeLogFileChan)

			for _, logFile := range openLogFiles {
				logFile.Close()
			}

			cmd.ack <- true
			run = false
		}
	}
}
