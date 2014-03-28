// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func runHubTest(t *testing.T, logsCount int, test func(Hub, []*testProtocolHandler)) {
	hub := NewDefaultHub()
	defer hub.Close()

	closeFuncs := make([]func(), 0, logsCount)
	protocolHandlers := make([]*testProtocolHandler, 0, logsCount)

	for i := 0; i < logsCount; i++ {
		port := 10000 + i
		serverAddress := ":" + strconv.Itoa(port)
		protocolHandler := newTestProtocolHandler()

		protocolHandlers = append(protocolHandlers, protocolHandler)

		if c, err := startServer(serverAddress, protocolHandler); err != nil {
			t.Errorf("Failed to start LogHub server", err.Error())
			t.FailNow()
		} else {
			closeFuncs = append(closeFuncs, c)
		}

		hub.SetLogStat(net.IPv4(127, 0, 0, 1), &LogStat{timeToTimestamp(time.Now()), 0, 0, port, 1})
	}

	defer func() {
		for _, c := range closeFuncs {
			c()
		}
	}()

	test(hub, protocolHandlers)
}

func TestHubSetStatRead(t *testing.T) {
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
