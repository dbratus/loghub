// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/lhproto"
	"sync/atomic"
)

type logProtocolHandler struct {
	logManager     LogManager
	lastTransferId *int64
	limit          int64
}

func NewLogProtocolHandler(logManager LogManager, lastTransferId *int64, limit int64) lhproto.ProtocolHandler {
	return &logProtocolHandler{logManager, lastTransferId, limit}
}

func (mh *logProtocolHandler) Write(entries chan *lhproto.IncomingLogEntryJSON) {
	for ent := range entries {
		mh.logManager.WriteLog(IncomingLogEntryJSONToLogEntry(ent))
	}
}

func (mh *logProtocolHandler) query(queries chan *lhproto.LogQueryJSON) chan *LogEntry {
	var results chan *LogEntry = nil

	for qJSON := range queries {
		q := LogQueryJSONToLogQuery(qJSON)
		var res chan *LogEntry

		if results == nil {
			results = make(chan *LogEntry)
			res = results
		} else {
			res = make(chan *LogEntry)
			mergedResults := make(chan *LogEntry)
			go MergeLogs(results, res, mergedResults)
			results = mergedResults
		}

		mh.logManager.ReadLog(q, res)
	}

	return results
}

func (mh *logProtocolHandler) Read(queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logProtocolHandler) InternalRead(queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logProtocolHandler) Truncate(cmd *lhproto.TruncateJSON) {
	mh.logManager.Truncate(cmd.Src, cmd.Lim)
}

func (mh *logProtocolHandler) Transfer(cmd *lhproto.TransferJSON) {
	lim := cmd.Lim
	cli := lhproto.NewClient(cmd.Addr, 1)
	defer cli.Close()

	for {
		entries := make(chan *LogEntry)

		if chunkId, chunkSize, found := mh.logManager.GetTransferChunk(lim, entries); found {
			acceptResult := make(chan *lhproto.AcceptResultJSON)
			acceptEntries := make(chan *lhproto.InternalLogEntryJSON)

			cli.Accept(&lhproto.AcceptJSON{chunkId, cmd.Id}, acceptEntries, acceptResult)

			for ent := range entries {
				acceptEntries <- LogEntryToInternalLogEntryJSON(ent)
			}

			if res := <-acceptResult; res.Result {
				mh.logManager.DeleteTransferChunk(chunkId)

				lim -= chunkSize

				if lim <= 0 {
					break
				}
			} else {
				break
			}
		}
	}

	atomic.StoreInt64(mh.lastTransferId, cmd.Id)
}

func (mh *logProtocolHandler) Accept(cmd *lhproto.AcceptJSON, entries chan *lhproto.InternalLogEntryJSON, result chan *lhproto.AcceptResultJSON) {
	entriesToAccept := make(chan *LogEntry)

	ack := mh.logManager.AcceptTransferChunk(cmd.Chunk, entriesToAccept)

	for ent := range entries {
		entriesToAccept <- InternalLogEntryJSONToLogEntry(ent)
	}

	result <- &lhproto.AcceptResultJSON{<-ack}
	atomic.StoreInt64(mh.lastTransferId, cmd.TransferId)
}

func (mh *logProtocolHandler) Stat(stats chan *lhproto.StatJSON) {
	stats <- &lhproto.StatJSON{
		"",
		mh.logManager.Size(),
		mh.limit,
	}
	close(stats)
}

func (mh *logProtocolHandler) Close() {
	mh.logManager.Close()
}
