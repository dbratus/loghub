// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package lhproto

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

	//The body is \0 terminated TruncateJSON.
	//The action has no response.
	ActionTruncate = "truncate"

	//The body is \0 terminated TransferJSON.
	ActionTransfer = "transfer"

	//The body is \0 terminated AcceptJSON followed by
	//a sequnce of InternalLogEntryJSON.
	//<AcceptJSON>\0<InternalLogEntryJSON>\0<InternalLogEntryJSON>\0...\0<InternalLogEntryJSON>\0\0
	//
	//The response is \0 terminated AcceptResultJSON.
	ActionAccept = "accept"

	//The body is a sequence of StatJSON terminated by \0.
	//<StatJSON>\0<StatJSON>\0...\0<StatJSON>\0\0
	ActionStat = "stat"
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

//The truncation command.
type TruncateJSON struct {
	Src string
	Lim int64
}

//The transfer command.
type TransferJSON struct {
	Id   int64
	Addr string
	Lim  int64
}

//The accept command.
type AcceptJSON struct {
	Chunk      string
	TransferId int64
}

//The result of 'accept' action.
type AcceptResultJSON struct {
	Result bool
}

//The result of 'stat' action.
type StatJSON struct {
	Addr string
	Sz   int64
	Lim  int64
}

//The interface which a protocol handler must implement.
type ProtocolHandler interface {
	Write(chan *IncomingLogEntryJSON)
	Read(chan *LogQueryJSON, chan *OutgoingLogEntryJSON)
	InternalRead(chan *LogQueryJSON, chan *InternalLogEntryJSON)
	Truncate(*TruncateJSON)
	Transfer(*TransferJSON)
	Accept(*AcceptJSON, chan *InternalLogEntryJSON, chan *AcceptResultJSON)
	Stat(chan *StatJSON)
	Close()
}
