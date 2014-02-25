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
	"fmt"
	"time"
)

type LogManagerStat struct {
	Capacity int64
	Size     int64
}

type LogManager interface {
	Logger

	Stat() *LogManagerStat
}

type defaultLogManager struct {
	home      string
	writeChan chan *LogEntry
	readChan  chan *logQuery
	closeChan chan *logClose
}

func NewDefaultLogManager(home string) LogManager {
	var mg = &defaultLogManager{home, make(chan *LogEntry), make(chan *logQuery), make(chan *logClose)}

	go mg.run()

	return mg
}

func (mg *defaultLogManager) WriteLog(*LogEntry) {

}

func (mg *defaultLogManager) ReadLog(from int64, to int64, minSeverity int, maxSeverity int, source string) chan *LogEntry {
	return nil
}

func (mg *defaultLogManager) Stat() *LogManagerStat {
	return nil
}

func getSourceDirName(source string) string {
	return base64.URLEncoding.EncodeToString([]byte(source))
}

func getFileNameForTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return fmt.Sprintf("%d.%d.%d.%d", t.Year(), int(t.Month()), t.Day(), t.Hour())
}

func getLogFileNameForEntry(entry *LogEntry) string {
	return fmt.Sprintf("%s/%s", getSourceDirName(entry.Source), getFileNameForTimestamp(entry.Timestamp))
}

func initLogManager() (initialized bool) {
	initialized = false
	return
}

func (mg *defaultLogManager) run() {
	stop := false
	initialized := initLogManager()
	openLogFiles := make(map[string]*LogFile)
	closeLogFileChan := make(chan string)

	waitAndCloseFile := func(fileName string) {
		<-time.After(time.Second * 10)
		closeLogFileChan <- fileName
	}

	getLogFile := func(fileName string) (*LogFile, error) {
		if logFile, found := openLogFiles[fileName]; !found {
			if logFile, err := OpenCreateLogFile(fmt.Sprintf("%s/%s", mg.home, fileName)); err == nil {
				openLogFiles[fileName] = logFile

				go waitAndCloseFile(fileName)

				return logFile, nil
			} else {
				return nil, err
			}
		} else {
			return logFile, nil
		}
	}

	for !stop {
		select {
		case ent := <-mg.writeChan:
			if initialized {
				if logFile, err := getLogFile(getLogFileNameForEntry(ent)); err == nil {
					logFile.WriteLog(ent)
				} else {
					println("Failed to obtain log file :", err)
				}
			}

		case logFileToClose := <-closeLogFileChan:
			if logFile, found := openLogFiles[logFileToClose]; found {
				logFile.Close()
			}

		//case q := <-log.readChan:
		//TODO: Read the log.

		case cmd := <-mg.closeChan:
			//TODO: Cleanup.

			cmd.ack <- true
			stop = true
		}
	}
}
