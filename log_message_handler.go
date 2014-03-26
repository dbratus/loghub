// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

type logMessageHandler struct {
	logManager LogManager
}

func NewLogMessageHandler(logManager LogManager) MessageHandler {
	return &logMessageHandler{logManager}
}

func (mh *logMessageHandler) Write(entries chan *IncomingLogEntryJSON) {
	for ent := range entries {
		mh.logManager.WriteLog(IncomingLogEntryJSONToLogEntry(ent))
	}
}

func (mh *logMessageHandler) query(queries chan *LogQueryJSON) chan *LogEntry {
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

func (mh *logMessageHandler) Read(queries chan *LogQueryJSON, resultJSON chan *OutgoingLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToOutgoingLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logMessageHandler) InternalRead(queries chan *LogQueryJSON, resultJSON chan *InternalLogEntryJSON) {
	result := mh.query(queries)

	if result != nil {
		for ent := range result {
			resultJSON <- LogEntryToInternalLogEntryJSON(ent)
		}
	}

	close(resultJSON)
}

func (mh *logMessageHandler) Truncate(cmd *TruncateJSON) {
	mh.logManager.Truncate(cmd.Src, cmd.Lim)
}

func (mh *logMessageHandler) Transfer(cmd *TransferJSON) {
	//TODO: Implement.
}

func (mh *logMessageHandler) Accept(cmd *AcceptJSON, entries chan *InternalLogEntryJSON, result chan *AcceptResultJSON) {
	//TODO: Implement.
}

func (mh *logMessageHandler) Close() {
	mh.logManager.Close()
}
