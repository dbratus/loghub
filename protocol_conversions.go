// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"github.com/dbratus/loghub/lhproto"
)

func IncomingLogEntryJSONToLogEntry(in *lhproto.IncomingLogEntryJSON) *LogEntry {
	ent := new(LogEntry)

	ent.Severity = in.Sev
	ent.Source = in.Src

	ent.Message, ent.Encoding = EncodeMessage(in.Msg)

	return ent
}

func LogEntryToOutgoingLogEntryJSON(in *LogEntry) *lhproto.OutgoingLogEntryJSON {
	return &lhproto.OutgoingLogEntryJSON{
		lhproto.IncomingLogEntryJSON{
			in.Severity,
			in.Source,
			DecodeMessage(in.Message, in.Encoding),
		},
		in.Timestamp,
	}
}

func LogEntryToInternalLogEntryJSON(in *LogEntry) *lhproto.InternalLogEntryJSON {
	var msg string

	if in.Encoding == EncodingDeflate {
		msg = base64.StdEncoding.EncodeToString(in.Message)
	} else {
		msg = string(in.Message)
	}

	return &lhproto.InternalLogEntryJSON{
		in.Severity,
		in.Source,
		in.Encoding,
		msg,
		in.Timestamp,
	}
}

func InternalLogEntryJSONToLogEntry(in *lhproto.InternalLogEntryJSON) *LogEntry {
	var msg []byte

	if in.Enc == EncodingDeflate {
		if m, err := base64.StdEncoding.DecodeString(in.Msg); err == nil {
			msg = m
		} else {
			msg = []byte{}
		}
	} else {
		msg = []byte(in.Msg)
	}

	return &LogEntry{in.Ts, in.Sev, in.Src, in.Enc, msg}
}

func LogQueryJSONToLogQuery(in *lhproto.LogQueryJSON) *LogQuery {
	return &LogQuery{in.From, in.To, in.MinSev, in.MaxSev, in.Src}
}

func LogQueryToLogQueryJSON(in *LogQuery) *lhproto.LogQueryJSON {
	return &lhproto.LogQueryJSON{in.From, in.To, in.MinSeverity, in.MaxSeverity, in.Source}
}
