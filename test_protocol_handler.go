// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/lhproto"
	"sync/atomic"
	"time"
)

const (
	testLogEntriesCount = 10
	testLogQueriesCount = 10
	testMaxTries        = 100
)

type testProtocolHandler struct {
	entriesWritten        chan *lhproto.IncomingLogEntryJSON
	outgoingEntriesToRead []*lhproto.OutgoingLogEntryJSON
	internalEntriesToRead []*lhproto.InternalLogEntryJSON
	truncations           chan *lhproto.TruncateJSON
	transfers             chan *lhproto.TransferJSON
	accepts               chan *lhproto.AcceptJSON
	entriesAccepted       chan *lhproto.InternalLogEntryJSON
	isClosed              *int32
}

func newTestProtocolHandler() *testProtocolHandler {
	outgoingEntriesToRead := make([]*lhproto.OutgoingLogEntryJSON, 0, testLogEntriesCount)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &lhproto.OutgoingLogEntryJSON{lhproto.IncomingLogEntryJSON{1, "Source", "Message"}, timeToTimestamp(time.Now())}

		outgoingEntriesToRead = append(outgoingEntriesToRead, m)
	}

	internalEntriesToRead := make([]*lhproto.InternalLogEntryJSON, 0, testLogEntriesCount)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &lhproto.InternalLogEntryJSON{1, "Source", EncodingPlain, "Message", timeToTimestamp(time.Now())}

		internalEntriesToRead = append(internalEntriesToRead, m)
	}

	return &testProtocolHandler{
		make(chan *lhproto.IncomingLogEntryJSON),
		outgoingEntriesToRead,
		internalEntriesToRead,
		make(chan *lhproto.TruncateJSON),
		make(chan *lhproto.TransferJSON),
		make(chan *lhproto.AcceptJSON),
		make(chan *lhproto.InternalLogEntryJSON),
		new(int32),
	}
}

func (mh *testProtocolHandler) Write(entries chan *lhproto.IncomingLogEntryJSON) {
	if !mh.IsClosed() {
		for ent := range entries {
			mh.entriesWritten <- ent
		}
	}
}

func (mh *testProtocolHandler) Read(queries chan *lhproto.LogQueryJSON, result chan *lhproto.OutgoingLogEntryJSON) {
	for _ = range queries {
	}

	for _, ent := range mh.outgoingEntriesToRead {
		result <- ent
	}

	close(result)
}

func (mh *testProtocolHandler) InternalRead(queries chan *lhproto.LogQueryJSON, result chan *lhproto.InternalLogEntryJSON) {
	for _ = range queries {
	}

	for _, ent := range mh.internalEntriesToRead {
		result <- ent
	}

	close(result)
}

func (mh *testProtocolHandler) Truncate(cmd *lhproto.TruncateJSON) {
	if !mh.IsClosed() {
		mh.truncations <- cmd
	}
}

func (mh *testProtocolHandler) Transfer(cmd *lhproto.TransferJSON) {
	if !mh.IsClosed() {
		mh.transfers <- cmd
	}
}

func (mh *testProtocolHandler) Accept(cmd *lhproto.AcceptJSON, entries chan *lhproto.InternalLogEntryJSON, result chan *lhproto.AcceptResultJSON) {
	if !mh.IsClosed() {
		mh.accepts <- cmd

		for ent := range entries {
			mh.entriesAccepted <- ent
		}

		result <- &lhproto.AcceptResultJSON{true}
	}
}

func (mh *testProtocolHandler) Close() {
	atomic.StoreInt32(mh.isClosed, 1)
	close(mh.entriesWritten)
	close(mh.truncations)
}

func (mh *testProtocolHandler) IsClosed() bool {
	return atomic.LoadInt32(mh.isClosed) > 0
}