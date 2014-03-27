// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/lhproto"
)

type hubProtocolHandler struct {
	hub Hub
}

func NewHubProtocolHandler(hub Hub) lhproto.ProtocolHandler {
	return &hubProtocolHandler{hub}
}

func (mh *hubProtocolHandler) Write(entries chan *lhproto.IncomingLogEntryJSON) {
	lhproto.PurgeIncomingLogEntryJSON(entries)
}

func (mh *hubProtocolHandler) query(queriesJSON chan *lhproto.LogQueryJSON) chan *LogEntry {
	queries := make([]*LogQuery, 0, 10)

	for qJSON := range queriesJSON {
		queries = append(queries, LogQueryJSONToLogQuery(qJSON))
	}

	if len(queries) > 0 {
		entries := make(chan *LogEntry)

		mh.hub.ReadLog(queries, entries)

		return entries
	} else {
		return nil
	}
}

func (mh *hubProtocolHandler) Read(queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubProtocolHandler) InternalRead(queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubProtocolHandler) Truncate(cmd *lhproto.TruncateJSON) {
	mh.hub.Truncate(cmd.Src, cmd.Lim)
}

func (mh *hubProtocolHandler) Transfer(cmd *lhproto.TransferJSON) {
}

func (mh *hubProtocolHandler) Accept(cmd *lhproto.AcceptJSON, entries chan *lhproto.InternalLogEntryJSON, result chan *lhproto.AcceptResultJSON) {
	lhproto.PurgeInternalLogEntryJSON(entries)
	result <- &lhproto.AcceptResultJSON{false}
}

func (mh *hubProtocolHandler) Close() {
	mh.hub.Close()
}
