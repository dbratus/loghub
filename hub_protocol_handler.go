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

func (mh *hubProtocolHandler) Write(cred *lhproto.Credentials, entries chan *lhproto.IncomingLogEntryJSON) {
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

func (mh *hubProtocolHandler) Read(cred *lhproto.Credentials, queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubProtocolHandler) InternalRead(cred *lhproto.Credentials, queries chan *lhproto.LogQueryJSON, resultJSON chan *lhproto.InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubProtocolHandler) Truncate(cred *lhproto.Credentials, cmd *lhproto.TruncateJSON) {
	mh.hub.Truncate(cmd.Src, cmd.Lim)
}

func (mh *hubProtocolHandler) Transfer(cred *lhproto.Credentials, cmd *lhproto.TransferJSON) {
}

func (mh *hubProtocolHandler) Accept(cred *lhproto.Credentials, cmd *lhproto.AcceptJSON, entries chan *lhproto.InternalLogEntryJSON, result chan *lhproto.AcceptResultJSON) {
	lhproto.PurgeInternalLogEntryJSON(entries)
	result <- &lhproto.AcceptResultJSON{false}
}

func (mh *hubProtocolHandler) Stat(cred *lhproto.Credentials, stats chan *lhproto.StatJSON) {
	for addr, stat := range mh.hub.GetStats() {
		stats <- &lhproto.StatJSON{
			addr,
			stat.Size,
			stat.Limit,
		}
	}

	close(stats)
}

func (mh *hubProtocolHandler) User(cred *lhproto.Credentials, usr *lhproto.UserInfoJSON) {
}

func (mh *hubProtocolHandler) Password(cred *lhproto.Credentials, pass *lhproto.PasswordJSON) {
}

func (mh *hubProtocolHandler) Close() {
	mh.hub.Close()
}
