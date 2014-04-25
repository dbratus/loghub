// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"strconv"
	"testing"
	"time"
	//"github.com/dbratus/loghub/trace"
)

func runHubTest(t *testing.T, logsCount int, test func(Hub, []*testProtocolHandler)) {
	hub := NewDefaultHub(false, false, "key")
	defer hub.Close()

	closeFuncs := make([]func(), 0, logsCount)
	protocolHandlers := make([]*testProtocolHandler, 0, logsCount)

	for i := 0; i < logsCount; i++ {
		port := 10000 + i
		serverAddress := "127.0.0.1:" + strconv.Itoa(port)
		protocolHandler := newTestProtocolHandler()

		protocolHandlers = append(protocolHandlers, protocolHandler)

		if c, err := startServer(serverAddress, protocolHandler, nil, ""); err != nil {
			t.Errorf("Failed to start LogHub server %s.", err.Error())
			t.FailNow()
		} else {
			closeFuncs = append(closeFuncs, c)
		}

		hub.SetLogStat(&LogStat{timeToTimestamp(time.Now()), 0, 0, serverAddress, 1})
	}

	defer func(closers []func()) {
		for _, c := range closers {
			c()
		}

		<-time.After(time.Millisecond * 100)
	}(closeFuncs)

	test(hub, protocolHandlers)
}

func TestHubSetStatRead(t *testing.T) {
	//trace.SetTraceLevel(trace.LevelDebug)
	//defer trace.SetTraceLevel(trace.LevelInfo)
	logsCount := 10

	runHubTest(t, logsCount, func(hub Hub, protocolHandlers []*testProtocolHandler) {
		entries := make(chan *LogEntry)
		var queries = []*LogQuery{&LogQuery{}}

		hub.ReadLog(queries, entries)

		cnt := 0
		for _ = range entries {
			cnt++
		}

		if cnt < testLogEntriesCount*logsCount {
			t.Errorf("Failed to get entries from the hub. Expected %d, got %d.", testLogEntriesCount*logsCount, cnt)
			t.Fail()
		}
	})
}

func TestHubTruncate(t *testing.T) {
	logsCount := 10

	runHubTest(t, logsCount, func(hub Hub, protocolHandlers []*testProtocolHandler) {
		hub.Truncate("hub", 10000)

		cntChan := make(chan bool)
		cnt := 0
		for i := 0; i < len(protocolHandlers); i++ {
			go func(i int) {
				<-protocolHandlers[i].truncations
				cntChan <- true
			}(i)
		}

		for cnt < logsCount {
			select {
			case <-cntChan:
				cnt++

			case <-time.After(time.Second * 30):
				t.Errorf("Failed to truncate hub. Expected %d truncations, got %d.", logsCount, cnt)
				t.FailNow()
			}
		}
	})
}
