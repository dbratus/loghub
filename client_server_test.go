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
	"testing"
	"time"
)

const (
	testLogEntriesCount = 10
	testLogQueriesCount = 10
	testMaxTries        = 100
)

type testMessageHandler struct {
	entriesWritten        []*IncomingLogEntryJSON
	outgoingEntriesToRead []*OutgoingLogEntryJSON
	internalEntriesToRead []*InternalLogEntryJSON
	isClosed              bool
}

func newTestMessageHandler() *testMessageHandler {
	outgoingEntriesToRead := make([]*OutgoingLogEntryJSON, 0, testLogEntriesCount)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &OutgoingLogEntryJSON{IncomingLogEntryJSON{1, "Source", "Message"}, time.Now().UnixNano()}

		outgoingEntriesToRead = append(outgoingEntriesToRead, m)
	}

	internalEntriesToRead := make([]*InternalLogEntryJSON, 0, testLogEntriesCount)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &InternalLogEntryJSON{1, "Source", EncodingPlain, "Message", time.Now().UnixNano()}

		internalEntriesToRead = append(internalEntriesToRead, m)
	}

	return &testMessageHandler{
		make([]*IncomingLogEntryJSON, 0, testLogEntriesCount),
		outgoingEntriesToRead,
		internalEntriesToRead,
		false,
	}
}

func (mh *testMessageHandler) Write(entries chan *IncomingLogEntryJSON) {
	for ent := range entries {
		mh.entriesWritten = append(mh.entriesWritten, ent)
	}
}

func (mh *testMessageHandler) Read(queries chan *LogQueryJSON, result chan *OutgoingLogEntryJSON) {
	for _ = range queries {
	}

	for _, ent := range mh.outgoingEntriesToRead {
		result <- ent
	}

	close(result)
}

func (mh *testMessageHandler) InternalRead(queries chan *LogQueryJSON, result chan *InternalLogEntryJSON) {
	for _ = range queries {
	}

	for _, ent := range mh.internalEntriesToRead {
		result <- ent
	}

	close(result)
}

func (mh *testMessageHandler) Close() {
	mh.isClosed = true
}

func TestClientServer(t *testing.T) {
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

	for triesCount := 0; len(messageHandler.entriesWritten) < testLogEntriesCount; triesCount++ {
		if triesCount == testMaxTries {
			t.Errorf(
				"Failed to write log entries through the JSON client. %d written, expected %d.",
				len(messageHandler.entriesWritten),
				testLogEntriesCount,
			)
			t.FailNow()
		}

		<-time.After(time.Millisecond * 10)
	}

	client.Close()
	closeServer()

	if !messageHandler.isClosed {
		t.Error("Failed to close JSON message handler.")
		t.FailNow()
	}
}

func TestClientWithoutServer(t *testing.T) {
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
