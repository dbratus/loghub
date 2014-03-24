// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

type hubMessageHandler struct {
	hub Hub
}

func NewHubMessageHandler(hub Hub) MessageHandler {
	return &hubMessageHandler{hub}
}

func (mh *hubMessageHandler) Write(entries chan *IncomingLogEntryJSON) {
	PurgeIncomingLogEntryJSON(entries)
}

func (mh *hubMessageHandler) query(queriesJSON chan *LogQueryJSON) chan *LogEntry {
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

func (mh *hubMessageHandler) Read(queries chan *LogQueryJSON, resultJSON chan *OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubMessageHandler) InternalRead(queries chan *LogQueryJSON, resultJSON chan *InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *hubMessageHandler) Truncate(cmd *TruncateJSON) {
	mh.hub.Truncate(cmd.Src, cmd.Lim)
}

func (mh *hubMessageHandler) Close() {
	mh.hub.Close()
}
