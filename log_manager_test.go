// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/trace"
	"os"
	"testing"
	"time"
)

func getTestLogHome() string {
	return os.TempDir() + "loghub.test.home"
}

func getAltTestLogHome() string {
	return os.TempDir() + "loghub.test.home.alt"
}

func makeTestLogHome(homePath string) {
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

func deleteTestLogHome(homePath string) {
	os.RemoveAll(homePath)
}

func TestWriteReadLog(t *testing.T) {
	makeTestLogHome(getTestLogHome())
	defer deleteTestLogHome(getTestLogHome())

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
	logManager.ReadLog(&LogQuery{timeToTimestamp(beforeWrite), timeToTimestamp(afterWrite), 1, 1, "src."}, qResult)
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
	logManager.ReadLog(&LogQuery{timeToTimestamp(beforeWrite), timeToTimestamp(afterWrite), 1, 1, "src."}, qResult)
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

	makeTestLogHome(getTestLogHome())
	defer deleteTestLogHome(getTestLogHome())

	sources := [...]string{"src1", "src2", "src3"}

	entriesPerSource := 128
	logManager := NewDefaultLogManager(getTestLogHome())
	defer logManager.Close()

	baseTs := time.Now()

	for _, src := range sources {
		for i := 0; i < entriesPerSource; i++ {
			ent := &LogEntry{
				timeToTimestamp(baseTs.Add(time.Hour * time.Duration(i))),
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
		timeToTimestamp(baseTs),
		timeToTimestamp(baseTs.Add(time.Hour * time.Duration(entriesPerSource))),
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

	initialSize := logManager.Size()

	logManager.Truncate("src.", timeToTimestamp(baseTs.Add(time.Hour*time.Duration(entriesPerSource/2))))

	qResult = make(chan *LogEntry)
	logManager.ReadLog(&LogQuery{
		timeToTimestamp(baseTs),
		timeToTimestamp(baseTs.Add(time.Hour * time.Duration(entriesPerSource))),
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

	if sz := logManager.Size(); sz >= initialSize {
		t.Errorf("Log must shrink after truncation. Was %d, became %d.", initialSize, sz)
		t.FailNow()
	}
}

func TestTransfer(t *testing.T) {
	trace.SetTraceLevel(trace.LevelDebug)
	defer trace.SetTraceLevel(trace.LevelInfo)

	makeTestLogHome(getTestLogHome())
	defer deleteTestLogHome(getTestLogHome())
	makeTestLogHome(getAltTestLogHome())
	defer deleteTestLogHome(getAltTestLogHome())

	sources := [...]string{"src1", "src2", "src3"}

	entriesPerSource := 10
	logManager := NewDefaultLogManager(getTestLogHome())
	logManagerAlt := NewDefaultLogManager(getAltTestLogHome())

	defer logManager.Close()
	defer logManagerAlt.Close()

	nHours := 3
	thisHour := time.Now().Truncate(time.Hour)

	populateLogs := func(logManager LogManager, shift time.Duration) {
		for i := 0; i < nHours; i++ {
			baseTs := thisHour.Add(time.Hour * time.Duration(i-nHours))

			for _, src := range sources {
				for j := 0; j < entriesPerSource; j++ {
					ent := &LogEntry{
						timeToTimestamp(baseTs.Add(time.Minute*time.Duration(j) + shift)),
						1,
						src,
						EncodingPlain,
						[]byte("Message"),
					}

					logManager.WriteLog(ent)
				}
			}
		}
	}

	populateLogs(logManager, time.Millisecond)
	populateLogs(logManagerAlt, time.Duration(0))

	entries := make(chan *LogEntry)
	gb := int64(1024 * 1024 * 1024)

	if chunkId, found := logManager.GetTransferChunk(gb, entries); found {
		cnt := 0
		initialLogSize := logManager.Size()
		initialAltLogSize := logManagerAlt.Size()

		inpEntries := make(chan *LogEntry)
		ack := logManagerAlt.AcceptTransferChunk(chunkId, inpEntries)

		for ent := range entries {
			cnt++
			inpEntries <- ent
		}

		close(inpEntries)

		if !<-ack {
			t.Errorf("Failed to accept transfer chunk.")
			t.FailNow()
		}

		if cnt < entriesPerSource {
			t.Errorf("Transfer chunk has not enough entries. Expected %d, got %d.", entriesPerSource, cnt)
			t.FailNow()
		}

		logManager.DeleteTransferChunk(chunkId)

		if sz := logManager.Size(); sz >= initialLogSize {
			t.Errorf("Source log must shrink after transfer chunk deletion.. Was %d, became %d.", initialLogSize, sz)
			t.FailNow()
		}

		if sz := logManagerAlt.Size(); sz <= initialAltLogSize {
			t.Errorf("Destination log must grow after transfer chunk deletion. Was %d, became %d.", initialAltLogSize, sz)
			t.FailNow()
		}
	} else {
		t.Error("Transfer chunk not found.")
		t.FailNow()
	}
}
