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

func (mh *logMessageHandler) Read(chan *LogQueryJSON, chan *OutgoingLogEntryJSON) {
	//TODO: Implement.
}

func (mh *logMessageHandler) InternalRead(chan *LogQueryJSON, chan *InternalLogEntryJSON) {
	//TODO: Implement.
}

func (mh *logMessageHandler) Close() {
	mh.logManager.Close()
}
