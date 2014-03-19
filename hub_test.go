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
	"net"
	"strconv"
	"testing"
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

		hub.SetLogStat(net.IPv4(127, 0, 0, 1), &LogStat{0, 0, port})
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
