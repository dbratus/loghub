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
	"fmt"
	"os"
	"testing"
)

func TestWriteReadLogFile(t *testing.T) {
	testLogFilePath := os.TempDir() + "/loghub.test.log"

	//println(testLogFilePath)

	var log *LogFile

	if l, err := OpenCreateLogFile(testLogFilePath); err != nil {
		t.Error("OpenCreateLogFile failed.", err)
		t.FailNow()
	} else {
		if l == nil {
			t.Error("Nil returned by OpenCreateLogFile successful call.")
			t.FailNow()
		}

		log = l
	}

	cleanup := func() {
		log.Close()
		os.Remove(testLogFilePath)
	}

	defer cleanup()

	entriesCount := hopLength * 3

	for i := 0; i < entriesCount; i++ {
		msg := fmt.Sprint("Message", i)
		ent := &LogEntry{int64(i + 1), 1, "Test", EncodingPlain, []byte(msg)}

		log.WriteLog(ent)
	}

	var logEntries chan *LogEntry
	var cnt int
	var minTs, maxTs int64
	entriesToRead := entriesCount / 2

	minTs = -1
	maxTs = int64(entriesToRead)
	logEntries = log.ReadLog(minTs, maxTs, 1, 1, "Test")
	cnt = 0

	for {
		ent, ok := <-logEntries

		if !ok {
			break
		}

		if ent.Timestamp < minTs || ent.Timestamp > maxTs {
			t.Errorf("Timestamp %d exceeds the range %d-%d.", ent.Timestamp, minTs, maxTs)
			t.FailNow()
		}

		cnt++
	}

	if cnt != entriesToRead {
		t.Errorf("%d entries read of %d expected.", cnt, entriesToRead)
		t.FailNow()
	}

	minTs = int64(entriesToRead)
	maxTs = int64(entriesCount * 2)
	logEntries = log.ReadLog(minTs, maxTs, 1, 1, "Test")
	cnt = 0

	for {
		ent, ok := <-logEntries

		if !ok {
			break
		}

		if ent.Timestamp < minTs || ent.Timestamp > maxTs {
			t.Errorf("Timestamp %d exceeds the range %d-%d.", ent.Timestamp, minTs, maxTs)
			t.FailNow()
		}

		cnt++
	}

	if cnt != entriesToRead+1 {
		t.Errorf("%d entries read of %d expected.", cnt, entriesToRead)
		t.FailNow()
	}

	entriesToRead = hopLength * 2
	minTs = int64(hopLength / 2)
	maxTs = int64(hopLength*2 + hopLength/2)
	logEntries = log.ReadLog(minTs, maxTs, 1, 1, "")
	cnt = 0

	for {
		ent, ok := <-logEntries

		if !ok {
			break
		}

		if ent.Timestamp < minTs || ent.Timestamp > maxTs {
			t.Errorf("Timestamp %d exceeds the range %d-%d.", ent.Timestamp, minTs, maxTs)
			t.FailNow()
		}

		cnt++
	}

	if cnt != entriesToRead+1 {
		t.Errorf("%d entries read of %d expected.", cnt, entriesToRead)
		t.FailNow()
	}

	logEntries = log.ReadLog(minTs, maxTs, 1, 1, "X")
	cnt = 0

	for {
		_, ok := <-logEntries

		if !ok {
			break
		}

		cnt++
	}

	if cnt != 0 {
		t.Errorf("%d entries read of %d expected.", cnt, 0)
		t.FailNow()
	}

	logEntries = log.ReadLog(minTs, maxTs, 0, 0, "Test")
	cnt = 0

	for {
		_, ok := <-logEntries

		if !ok {
			break
		}

		cnt++
	}

	if cnt != 0 {
		t.Errorf("%d entries read of %d expected.", cnt, 0)
		t.FailNow()
	}
}
