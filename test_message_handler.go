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
