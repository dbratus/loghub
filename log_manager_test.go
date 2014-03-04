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
	"os"
	"testing"
	"time"
)

func getTestLogHome() string {
	return os.TempDir() + "loghub.test.home"
}

func makeTestLogHome() {
	homePath := getTestLogHome()

	//println(homePath)

	if stat, err := os.Stat(homePath); err != nil {
		if os.IsNotExist(err) {
			if e := os.Mkdir(homePath, 0777); e != nil {
				panic(e.Error())
			}
		} else {
			panic(err.Error())
		}
	} else if !stat.IsDir() {
		panic(homePath + " already exists and its not a directory.")
	} else {
		os.RemoveAll(homePath)

		if e := os.Mkdir(homePath, 0777); e != nil {
			panic(e.Error())
		}
	}
}

func deleteTestLogHome() {
	homePath := getTestLogHome()
	os.RemoveAll(homePath)
}

func TestWriteReadLog(t *testing.T) {
	makeTestLogHome()
	defer deleteTestLogHome()

	sources := [...]string{ "src1", "src2", "src3" }

	entriesPerSource := 10
	logManager := NewDefaultLogManager(getTestLogHome())
	beforeWrite := time.Now()

	for i := 0; i < entriesPerSource; i++ {
		for _, src := range sources {
			ent := &LogEntry{0, 1, src, EncodingPlain, []byte("Message")}

			logManager.WriteLog(ent)
		}
	}
	
	logManager.Close()
	afterWrite := time.Now()

	logManager = NewDefaultLogManager(getTestLogHome())

	qResult := make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{beforeWrite.UnixNano(), afterWrite.UnixNano(), 1, 1, "src.", qResult})
	entCnt := 0

	for _ = range qResult {
		entCnt++
	}

	if entCnt < entriesPerSource*3 {
		t.Errorf("Failed to read entries from log manager: expected %d, got %d.", entriesPerSource, entCnt)
		t.FailNow()
	}

	for i := 0; i < entriesPerSource; i++ {
		for _, src := range sources {
			ent := &LogEntry{0, 1, src, EncodingPlain, []byte("Message")}

			logManager.WriteLog(ent)
		}
	}

	logManager.Close()

	afterWrite = time.Now()

	logManager = NewDefaultLogManager(getTestLogHome())
	qResult = make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{beforeWrite.UnixNano(), afterWrite.UnixNano(), 1, 1, "src.", qResult})
	entCnt = 0

	for _ = range qResult {
		entCnt++
	}

	if entCnt < entriesPerSource*3*2 {
		t.Errorf("Failed to read entries from log manager: expected %d, got %d.", entriesPerSource*3*2, entCnt)
		t.FailNow()
	}

	logManager.Close()
}