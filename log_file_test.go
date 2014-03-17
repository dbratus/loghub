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

func TestWriteCloseWriteReadLogFile(t *testing.T) {
	testLogFilePath := os.TempDir() + "loghub.write-close-write-read-test.log"

	//println(testLogFilePath)

	var log *LogFile

	if l, err := OpenLogFile(testLogFilePath, true); err != nil {
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
		os.Remove(testLogFilePath)
	}

	defer cleanup()

	entriesCount := 3

	for i := 0; i < entriesCount; i++ {
		var msg string

		if i == 2 {
			msg = "Message Сообщение"
		} else {
			msg = "Message"
		}

		ent := &LogEntry{int64(i + 1), 1, "Test", EncodingPlain, []byte(msg)}

		log.WriteLog(ent)
	}

	log.Close()

	if l, err := OpenLogFile(testLogFilePath, true); err != nil {
		t.Error("OpenCreateLogFile failed.", err)
		t.FailNow()
	} else {
		if l == nil {
			t.Error("Nil returned by OpenCreateLogFile successful call.")
			t.FailNow()
		}

		log = l
	}

	for i := 0; i < entriesCount; i++ {
		var msg string

		if i == 2 {
			msg = "Message Сообщение"
		} else {
			msg = "Message"
		}

		ent := &LogEntry{int64(entriesCount + i + 1), 1, "Test", EncodingPlain, []byte(msg)}

		log.WriteLog(ent)
	}

	minTs := int64(-1)
	maxTs := int64(entriesCount*2 + 1)
	logEntries := make(chan *LogEntry)
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test", logEntries})
	cnt := 0

	for _ = range logEntries {
		cnt++
	}

	if cnt != entriesCount*2 {
		t.Errorf("%d entries read of %d expected.", cnt, entriesCount*2)
		t.FailNow()
	}
}

func TestWriteReadLogFile(t *testing.T) {
	testLogFilePath := os.TempDir() + "loghub.test.log"

	//println(testLogFilePath)

	var log *LogFile

	if l, err := OpenLogFile(testLogFilePath, true); err != nil {
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
		os.Remove(testLogFilePath)
	}

	defer cleanup()

	fileSize := log.Size()

	entriesCount := hopLength * 3

	for i := 0; i < entriesCount; i++ {
		msg := fmt.Sprint("Message", i)
		ent := &LogEntry{int64(i + 1), 1, "Test", EncodingPlain, []byte(msg)}

		log.WriteLog(ent)
	}

	if log.Size() == fileSize {
		t.Error("Size of the log file must change after write.")
		t.FailNow()
	}

	log.Close()

	if l, err := OpenLogFile(testLogFilePath, true); err != nil {
		t.Error("OpenCreateLogFile failed.", err)
		t.FailNow()
	} else {
		if l == nil {
			t.Error("Nil returned by OpenCreateLogFile successful call.")
			t.FailNow()
		}

		log = l
	}

	var logEntries chan *LogEntry
	var cnt int
	var minTs, maxTs int64
	entriesToRead := entriesCount / 2

	minTs = -1
	maxTs = int64(entriesToRead)
	logEntries = make(chan *LogEntry)
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test", logEntries})
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

	logEntries = make(chan *LogEntry)
	minTs = int64(entriesToRead)
	maxTs = int64(entriesCount * 2)
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test", logEntries})
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

	logEntries = make(chan *LogEntry)
	entriesToRead = hopLength * 2
	minTs = int64(hopLength / 2)
	maxTs = int64(hopLength*2 + hopLength/2)
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "", logEntries})
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

	logEntries = make(chan *LogEntry)
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "X", logEntries})
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

	logEntries = make(chan *LogEntry)
	log.ReadLog(&LogQuery{minTs, maxTs, 0, 0, "Test", logEntries})
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

	log.Close()
}
