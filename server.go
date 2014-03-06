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

import (
	"net"
)

func startServer(address string, handler MessageHandler) (func(), error) {
	if listener, err := net.Listen("tcp", address); err == nil {
		go func() {
			for {
				if conn, err := listener.Accept(); err == nil {
					go handleConnection(conn, handler)
				} else {
					break
				}
			}
		}()

		return func() {
			listener.Close()
			handler.Close()
		}, nil

	} else {
		return nil, err
	}
}

func handleConnection(conn net.Conn, handler MessageHandler) {
	reader := NewJSONStreamReader(conn)
	writer := NewJSONStreamWriter(conn)

	for {
		var header MessageHeaderJSON

		if err := reader.ReadJSON(&header); err != nil {
			conn.Close()
			break
		}

		switch header.Action {
			case ActionWrite:
				entChan := make(chan *IncomingLogEntryJSON)

				go handler.HandleWrite(entChan)

				for {
					ent := new(IncomingLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						close(entChan)

						if err != ErrStreamDelimiter {
							conn.Close()
							return
						}

						break
					}

					entChan <- ent
				}
			case ActionRead:
				qChan := make(chan *LogQueryJSON)
				entChan := make(chan *OutgoingLogEntryJSON)

				go handler.HandleRead(qChan, entChan)

				if !readLogQueryJSONChannel(reader, handler, qChan) {
					conn.Close()
					return
				}

				continueWriting := true

				for ent := range entChan {
					if continueWriting {
						if err := writer.WriteJSON(ent); err != nil {
							conn.Close()
							continueWriting = false
						}
					}
				}

				if continueWriting {
					writer.WriteDelimiter()
				}

			case ActionInternalRead:
				qChan := make(chan *LogQueryJSON)
				entChan := make(chan *InternalLogEntryJSON)

				go handler.HandleInternalRead(qChan, entChan)

				if !readLogQueryJSONChannel(reader, handler, qChan) {
					conn.Close()
					return
				}

				continueWriting := true

				for ent := range entChan {
					if continueWriting {
						if err := writer.WriteJSON(ent); err != nil {
							conn.Close()
							continueWriting = false
						}
					}
				}

				if continueWriting {
					writer.WriteDelimiter()
				}
		}
	}
}

func readLogQueryJSONChannel(reader JSONStreamReader, handler MessageHandler, qChan chan *LogQueryJSON) bool {
	for {
		q := new(LogQueryJSON)

		if err := reader.ReadJSON(q); err != nil {
			close(qChan)

			if err != ErrStreamDelimiter {
				return false
			}
			
			break
		}

		qChan <- q
	}

	return true
}