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

func TestHubSetStatRead(t *testing.T) {
	hub := NewDefaultHub()
	defer hub.Close()

	logsCount := 10
	closeFuncs := make([]func(), 0, logsCount)

	for i := 0; i < logsCount; i++ {
		port := 10000 + i
		serverAddress := ":" + strconv.Itoa(port)
		messageHandler := newTestMessageHandler()

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

	entries := make(chan *LogEntry)
	var queries = []*LogQuery{&LogQuery{}}

	hub.ReadLog(queries, entries)

	cnt := 0
	for _ = range entries {
		cnt++
	}

	if cnt < testLogEntriesCount*logsCount {
		t.Errorf("Failed to get entries from the hub. Expected %d, got %d.", testLogEntriesCount*logsCount, cnt)
		t.FailNow()
	}
}
