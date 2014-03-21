// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

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

			go handler.Write(entChan)

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

			go handler.Read(qChan, entChan)

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

			go handler.InternalRead(qChan, entChan)

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
