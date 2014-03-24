// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"time"
)

const (
	testLogEntriesCount = 10
	testLogQueriesCount = 10
	testMaxTries        = 100
)

type testMessageHandler struct {
	entriesWritten        chan *IncomingLogEntryJSON
	outgoingEntriesToRead []*OutgoingLogEntryJSON
	internalEntriesToRead []*InternalLogEntryJSON
	truncations           chan *TruncateJSON
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
		make(chan *IncomingLogEntryJSON),
		outgoingEntriesToRead,
		internalEntriesToRead,
		make(chan *TruncateJSON),
		false,
	}
}

func (mh *testMessageHandler) Write(entries chan *IncomingLogEntryJSON) {
	for ent := range entries {
		mh.entriesWritten <- ent
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

func (mh *testMessageHandler) Truncate(cmd *TruncateJSON) {
	mh.truncations <- cmd
}

func (mh *testMessageHandler) Close() {
	close(mh.entriesWritten)
	close(mh.truncations)
	mh.isClosed = true
}
