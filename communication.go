/*
	This file is part of LogHub.

	LogHub is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	LogHub is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with LogHub.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

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