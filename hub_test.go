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

func runHubTest(t *testing.T, logsCount int, test func(Hub, []*testMessageHandler)) {
	hub := NewDefaultHub()
	defer hub.Close()

	closeFuncs := make([]func(), 0, logsCount)
	messageHandlers := make([]*testMessageHandler, 0, logsCount)

	for i := 0; i < logsCount; i++ {
		port := 10000 + i
		serverAddress := ":" + strconv.Itoa(port)
		messageHandler := newTestMessageHandler()

		messageHandlers = append(messageHandlers, messageHandler)

		if c, err := startServer(serverAddress, messageHandler); err != nil {
			t.Errorf("Failed to start LogHub server", err.Error())
			t.FailNow()
		} else {
			closeFuncs = append(closeFuncs, c)
		}

		hub.SetLogStat(net.IPv4(127, 0, 0, 1), &LogStat{time.Now().UnixNano(), 0, 0, port})
	}

	defer func() {
		for _, c := range closeFuncs {
			c()
		}
	}()

	test(hub, messageHandlers)
}

func TestHubSetStatRead(t *testing.T) {
	logsCount := 10

	runHubTest(t, logsCount, func(hub Hub, messageHandlers []*testMessageHandler) {
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

	runHubTest(t, logsCount, func(hub Hub, messageHandlers []*testMessageHandler) {
		hub.Truncate("hub", 10000)

		cntChan := make(chan bool)
		cnt := 0
		for i := 0; i < len(messageHandlers); i++ {
			go func(i int) {
				<-messageHandlers[i].truncations
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
