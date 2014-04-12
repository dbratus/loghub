// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/dbratus/loghub/tmpdir"
	"testing"
)

func TestWriteCloseWriteReadLogFile(t *testing.T) {
	var log *LogFile

	home := tmpdir.GetPath("loghub.test.home")

	tmpdir.Make(home)
	defer func() {
		if log != nil {
			log.Close()
		}

		tmpdir.Rm(home)
	}()

	testLogFilePath := home + "/write-close-write-read-test.log"

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
	log = nil

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
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test"}, logEntries)
	cnt := 0

	for _ = range logEntries {
		cnt++
	}

	if cnt != entriesCount*2 {
		t.Errorf("%d entries read of %d expected.", cnt, entriesCount*2)
		t.FailNow()
	}

	log.Close()
	log = nil
}

func TestWriteReadLogFile(t *testing.T) {
	var log *LogFile

	home := tmpdir.GetPath("loghub.test.home")
	tmpdir.Make(home)
	defer func() {
		if log != nil {
			log.Close()
		}

		tmpdir.Rm(home)
	}()

	testLogFilePath := home + "/loghub.test.log"

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
	log = nil

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
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test"}, logEntries)
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
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "Test"}, logEntries)
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
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, ""}, logEntries)
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
	log.ReadLog(&LogQuery{minTs, maxTs, 1, 1, "X"}, logEntries)
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
	log.ReadLog(&LogQuery{minTs, maxTs, 0, 0, "Test"}, logEntries)
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
	log = nil
}
