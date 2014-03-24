// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	//"github.com/dbratus/loghub/trace"
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

	sources := [...]string{"src1", "src2", "src3"}

	entriesPerSource := 10
	logManager := NewDefaultLogManager(getTestLogHome())
	initialLogSize := logManager.Size()

	beforeWrite := time.Now()

	for i := 0; i < entriesPerSource; i++ {
		for _, src := range sources {
			ent := &LogEntry{0, 1, src, EncodingPlain, []byte("Message")}

			logManager.WriteLog(ent)
		}
	}

	newLogManagerSize := logManager.Size()

	if newLogManagerSize == initialLogSize {
		t.Error("Log size after write must change.")
		t.FailNow()
	}

	logManager.Close()

	afterWrite := time.Now()

	logManager = NewDefaultLogManager(getTestLogHome())

	qResult := make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{beforeWrite.UnixNano(), afterWrite.UnixNano(), 1, 1, "src."}, qResult)
	entCnt := 0

	for _ = range qResult {
		entCnt++
	}

	if entCnt < entriesPerSource*3 {
		t.Errorf("Failed to read entries from log manager: expected %d, got %d.", entriesPerSource, entCnt)
		t.FailNow()
	}

	initialLogSize = logManager.Size()

	if initialLogSize != newLogManagerSize {
		t.Error("Log size after open must not change.")
		t.FailNow()
	}

	for i := 0; i < entriesPerSource; i++ {
		for _, src := range sources {
			ent := &LogEntry{0, 1, src, EncodingPlain, []byte("Message")}

			logManager.WriteLog(ent)
		}
	}

	if logManager.Size() == initialLogSize {
		t.Error("Log size after write must change.")
		t.FailNow()
	}

	logManager.Close()

	afterWrite = time.Now()

	logManager = NewDefaultLogManager(getTestLogHome())
	qResult = make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{beforeWrite.UnixNano(), afterWrite.UnixNano(), 1, 1, "src."}, qResult)
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

func TestTruncate(t *testing.T) {
	//trace.SetTraceLevel(trace.LevelDebug)
	//defer trace.SetTraceLevel(trace.LevelInfo)

	makeTestLogHome()
	defer deleteTestLogHome()

	sources := [...]string{"src1", "src2", "src3"}

	entriesPerSource := 128
	logManager := NewDefaultLogManager(getTestLogHome())

	baseTs := time.Now()

	for _, src := range sources {
		for i := 0; i < entriesPerSource; i++ {
			ent := &LogEntry{
				baseTs.Add(time.Hour * time.Duration(i)).UnixNano(),
				1,
				src,
				EncodingPlain,
				[]byte("Message"),
			}

			logManager.WriteLog(ent)
		}
	}

	qResult := make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{
		baseTs.UnixNano(),
		baseTs.Add(time.Hour * time.Duration(entriesPerSource)).UnixNano(),
		1,
		1,
		"src.",
	}, qResult)
	entCnt := 0

	for _ = range qResult {
		entCnt++
	}

	if entCnt < entriesPerSource*len(sources) {
		t.Errorf("Failed to read entries from log manager: expected %d, got %d.", entriesPerSource*len(sources), entCnt)
		t.FailNow()
	}

	logManager.Truncate(baseTs.Add(time.Hour*time.Duration(entriesPerSource/2)).UnixNano(), "src.")

	qResult = make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{
		baseTs.UnixNano(),
		baseTs.Add(time.Hour * time.Duration(entriesPerSource)).UnixNano(),
		1,
		1,
		"src.",
	}, qResult)
	entCnt = 0

	for _ = range qResult {
		entCnt++
	}

	if entCnt > (entriesPerSource/2)*len(sources) {
		t.Errorf("Failed to read entries from log manager: expected %d, got %d.", (entriesPerSource/2)*len(sources), entCnt)
		t.FailNow()
	}

	logManager.Close()
}
