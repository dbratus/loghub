// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/lhproto"
	"github.com/dbratus/loghub/trace"
	"sync/atomic"
)

var logProtocolHanderTrace = trace.New("LogProtocolHander")

type logProtocolHandler struct {
	logManager     LogManager
	lastTransferId *int64
	limit          int64
}

func NewLogProtocolHandler(logManager LogManager, lastTransferId *int64, limit int64) lhproto.ProtocolHandler {
	return &logProtocolHandler{logManager, lastTransferId, limit}
}

func (mh *logProtocolHandler) Write(cred *lhproto.Credentials, entries chan *lhproto.IncomingLogEntryJSON) {
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

func (mh *logProtocolHandler) Read(cred *lhproto.Credentials, queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logProtocolHandler) InternalRead(cred *lhproto.Credentials, queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logProtocolHandler) Truncate(cred *lhproto.Credentials, cmd *lhproto.TruncateJSON) {
	mh.logManager.Truncate(cmd.Src, cmd.Lim)
}

func (mh *logProtocolHandler) Transfer(cred *lhproto.Credentials, cmd *lhproto.TransferJSON) {
	logProtocolHanderTrace.Debugf("Got transfer of %d to %s, id %d.", cmd.Lim, cmd.Addr, cmd.Id)

	lim := cmd.Lim
	cli := lhproto.NewClient(cmd.Addr, 1, false, false)
	defer cli.Close()

	for {
		entries := make(chan *LogEntry)

		if chunkId, chunkSize, found := mh.logManager.GetTransferChunk(lim, entries); found {
			logProtocolHanderTrace.Debugf("Transfering %s, sz=%d.", chunkId, chunkSize)

			acceptResult := make(chan *lhproto.AcceptResultJSON)
			acceptEntries := make(chan *lhproto.InternalLogEntryJSON)

			cli.Accept(cred, &lhproto.AcceptJSON{chunkId, cmd.Id}, acceptEntries, acceptResult)

			for ent := range entries {
				acceptEntries <- LogEntryToInternalLogEntryJSON(ent)
			}

			close(acceptEntries)

			if res := <-acceptResult; res.Result {
				if res.Result {
					logProtocolHanderTrace.Debugf("Transfer of %s completed successfully.", chunkId)
				} else {
					logProtocolHanderTrace.Debugf("Transfer of %s failed.", chunkId)
				}

				mh.logManager.DeleteTransferChunk(chunkId)

				lim -= chunkSize

				if lim <= 0 {
					logProtocolHanderTrace.Debugf("Transfer %d complete. No more data to transfer.", cmd.Id)
					break
				}

				logProtocolHanderTrace.Debugf("Continuing transfer %d, %d bytes to transfer left.", cmd.Id, lim)
			} else {
				logProtocolHanderTrace.Debugf("Transfer %d complete. Accept failed.", cmd.Id)
				break
			}
		} else {
			logProtocolHanderTrace.Debugf("Transfer %d complete. No no chunk found.", cmd.Id)
			break
		}
	}

	atomic.StoreInt64(mh.lastTransferId, cmd.Id)
}

func (mh *logProtocolHandler) Accept(cred *lhproto.Credentials, cmd *lhproto.AcceptJSON, entries chan *lhproto.InternalLogEntryJSON, result chan *lhproto.AcceptResultJSON) {
	logProtocolHanderTrace.Debugf("Accepting %s of transfer %d.", cmd.Chunk, cmd.TransferId)

	entriesToAccept := make(chan *LogEntry)

	ack := mh.logManager.AcceptTransferChunk(cmd.Chunk, entriesToAccept)

	for ent := range entries {
		entriesToAccept <- InternalLogEntryJSONToLogEntry(ent)
	}

	close(entriesToAccept)

	acceptResult := <-ack

	if acceptResult {
		logProtocolHanderTrace.Debugf("Accepted %s of transfer %d successfully.", cmd.Chunk, cmd.TransferId)
	} else {
		logProtocolHanderTrace.Debugf("Accept %s of transfer %d failed.", cmd.Chunk, cmd.TransferId)
	}

	result <- &lhproto.AcceptResultJSON{acceptResult}
}

func (mh *logProtocolHandler) Stat(cred *lhproto.Credentials, stats chan *lhproto.StatJSON) {
	stats <- &lhproto.StatJSON{
		"",
		mh.logManager.Size(),
		mh.limit,
	}
	close(stats)
}

func (mh *logProtocolHandler) User(cred *lhproto.Credentials, usr *lhproto.UserInfoJSON) {
}

func (mh *logProtocolHandler) Password(cred *lhproto.Credentials, pass *lhproto.PasswordJSON) {
}

func (mh *logProtocolHandler) Close() {
	mh.logManager.Close()
}
