// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/trace"
	"testing"
	"time"
)

func TestClientServer(t *testing.T) {
	//trace.SetTraceLevel(trace.LevelDebug)
	//defer trace.SetTraceLevel(trace.LevelError)

	serverAddress := ":9999"
	messageHandler := newTestMessageHandler()
	var closeServer func()

	if c, err := startServer(serverAddress, messageHandler); err != nil {
		t.Errorf("Failed to start LogHub server", err.Error())
		t.FailNow()
	} else {
		closeServer = c
	}

	client := NewLogHubClient(serverAddress, 1)

	entriesToWrite := make(chan *IncomingLogEntryJSON)
	client.Write(entriesToWrite)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &IncomingLogEntryJSON{1, "Source", "Message"}

		entriesToWrite <- m
	}

	close(entriesToWrite)

	incomingMsgCnt := 0
	for incomingMsgCnt < testLogEntriesCount {
		select {
		case <-messageHandler.entriesWritten:
			incomingMsgCnt++

		case <-time.After(time.Second * 10):
			t.Errorf("Incomming entries have not arrived. Expected %d, got %d.", testLogEntriesCount, incomingMsgCnt)
			t.FailNow()
		}
	}

	truncateCmd := TruncateJSON{"src", 10000}
	client.Truncate(&truncateCmd)

	select {
	case cmd := <-messageHandler.truncations:
		if cmd.Lim != truncateCmd.Lim {
			t.Error("Lim doesn't match.")
			t.FailNow()
		}

		if cmd.Src != truncateCmd.Src {
			t.Error("Src doesn't match.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("Truncation has not arrived.")
		t.FailNow()
	}

	queries := make(chan *LogQueryJSON)
	result := make(chan *OutgoingLogEntryJSON)

	client.Read(queries, result)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(LogQueryJSON)
	}

	close(queries)

	resultLen := 0
	for _ = range result {
		resultLen++
	}

	if resultLen < testLogEntriesCount {
		t.Errorf(
			"Failed to read outgoing log entries through the JSON client. %d read, expected %d.",
			resultLen,
			testLogEntriesCount,
		)
		t.FailNow()
	}

	queries = make(chan *LogQueryJSON)
	resultInternal := make(chan *InternalLogEntryJSON)

	client.InternalRead(queries, resultInternal)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(LogQueryJSON)
	}

	close(queries)

	resultLen = 0
	for _ = range resultInternal {
		resultLen++
	}

	if resultLen < testLogEntriesCount {
		t.Errorf(
			"Failed to read internal log entries through the JSON client. %d read, expected %d.",
			resultLen,
			testLogEntriesCount,
		)
		t.FailNow()
	}

	client.Close()
	closeServer()

	if !messageHandler.isClosed {
		t.Error("Failed to close JSON message handler.")
		t.FailNow()
	}
}

func TestClientWithoutServer(t *testing.T) {
	trace.SetTraceLevel(-1)
	defer trace.SetTraceLevel(trace.LevelError)

	client := NewLogHubClient(":9999", 1)

	entriesToWrite := make(chan *IncomingLogEntryJSON)
	client.Write(entriesToWrite)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &IncomingLogEntryJSON{1, "Source", "Message"}

		entriesToWrite <- m
	}

	close(entriesToWrite)

	queries := make(chan *LogQueryJSON)
	result := make(chan *OutgoingLogEntryJSON)

	client.Read(queries, result)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(LogQueryJSON)
	}

	close(queries)

	PurgeOutgoingLogEntryJSON(result)

	queries = make(chan *LogQueryJSON)
	resultInternal := make(chan *InternalLogEntryJSON)

	client.InternalRead(queries, resultInternal)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(LogQueryJSON)
	}

	close(queries)

	PurgeInternalLogEntryJSON(resultInternal)

	client.Close()
}
