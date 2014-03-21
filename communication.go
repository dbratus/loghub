// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
)

/*
This file describes the data structures of the LogHub communication protocol.
These structures are directly marshalled to JSON.

The LogHub protocol is a sequence of JSON objects separated by \0.

The sequence can be devided onto messages of the form:

<MessageHeaderJSON>\0<message body>

The action in the message header determines the format of
the message body and the reponse.

The client sends a message to the server and reads the response
then sends another message etc. To terminate the communication
the client sends \0.
*/

//Actions of the messages.
const (
	//The body is a sequnce of IncomingLogEntryJSON terminated by \0.
	//<IncomingLogEntryJSON>\0<IncomingLogEntryJSON>\0...\0<IncomingLogEntryJSON>\0\0
	//
	//The message has no response.
	ActionWrite = "write"

	//The body is a sequence of LogQueryJSON terminated by \0.
	//<LogQueryJSON>\0<LogQueryJSON>\0...\0<LogQueryJSON>\0\0
	//
	//The response is a sequnce of OutgoingLogEntryJSON.
	//<OutgoingLogEntryJSON>\0<OutgoingLogEntryJSON>\0...\0<OutgoingLogEntryJSON>\0\0
	ActionRead = "read"

	//The body is a sequence of LogQueryJSON terminated by \0.
	//<LogQueryJSON>\0<LogQueryJSON>\0...\0<LogQueryJSON>\0\0
	//
	//The response is a sequnce of InternalLogEntryJSON.
	//<InternalLogEntryJSON>\0<InternalLogEntryJSON>\0...\0<InternalLogEntryJSON>\0\0
	ActionInternalRead = "iread"
)

//The message header that each message starts with.
type MessageHeaderJSON struct {
	//The action to perform (one of Action... contants).
	//The action determines the structures of the body.
	Action string
}

//The log entries written by applications.
type IncomingLogEntryJSON struct {
	Sev int
	Src string

	//The message is always in plain text.
	//Compressing and decompressing of the messages is
	//the responsibility of the hub.
	Msg string
}

//Validates the incoming message.
func (m *IncomingLogEntryJSON) IsValid() bool {
	return len(m.Src) > 0 && len(m.Msg) > 0
}

//The log entries returned by the hub to its clients.
type OutgoingLogEntryJSON struct {
	IncomingLogEntryJSON

	Ts int64
}

//The log entries for the internal communication.
type InternalLogEntryJSON struct {
	Sev int
	Src string
	Enc int

	//The value is base64 string if the Enc is EncodingDeflate;
	//if the End is EncodingPlain, then Msg is plain text.
	Msg string
	Ts  int64
}

//The log query.
type LogQueryJSON struct {
	From   int64
	To     int64
	MinSev int
	MaxSev int
	Src    string
}

//The interface which a protocol message handler must implement.
type MessageHandler interface {
	Write(chan *IncomingLogEntryJSON)
	Read(chan *LogQueryJSON, chan *OutgoingLogEntryJSON)
	InternalRead(chan *LogQueryJSON, chan *InternalLogEntryJSON)
	Close()
}

func IncomingLogEntryJSONToLogEntry(in *IncomingLogEntryJSON) *LogEntry {
	ent := new(LogEntry)

	ent.Severity = in.Sev
	ent.Source = in.Src

	ent.Message, ent.Encoding = EncodeMessage(in.Msg)

	return ent
}

func LogEntryToOutgoingLogEntryJSON(in *LogEntry) *OutgoingLogEntryJSON {
	return &OutgoingLogEntryJSON{
		IncomingLogEntryJSON{
			in.Severity,
			in.Source,
			DecodeMessage(in.Message, in.Encoding),
		},
		in.Timestamp,
	}
}

func LogEntryToInternalLogEntryJSON(in *LogEntry) *InternalLogEntryJSON {
	var msg string

	if in.Encoding == EncodingDeflate {
		msg = base64.StdEncoding.EncodeToString(in.Message)
	} else {
		msg = string(in.Message)
	}

	return &InternalLogEntryJSON{
		in.Severity,
		in.Source,
		in.Encoding,
		msg,
		in.Timestamp,
	}
}

func InternalLogEntryJSONToLogEntry(in *InternalLogEntryJSON) *LogEntry {
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

func LogQueryJSONToLogQuery(in *LogQueryJSON) *LogQuery {
	return &LogQuery{in.From, in.To, in.MinSev, in.MaxSev, in.Src}
}

func LogQueryToLogQueryJSON(in *LogQuery) *LogQueryJSON {
	return &LogQueryJSON{in.From, in.To, in.MinSeverity, in.MaxSeverity, in.Source}
}
