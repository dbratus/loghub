// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/jstream"
	"github.com/dbratus/loghub/lhproto"
	"net"
)

func startServer(address string, handler lhproto.ProtocolHandler) (func(), error) {
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

func handleConnection(conn net.Conn, handler lhproto.ProtocolHandler) {
	reader := jstream.NewReader(conn)
	writer := jstream.NewWriter(conn)

	for {
		var header lhproto.MessageHeaderJSON

		if err := reader.ReadJSON(&header); err != nil {
			conn.Close()
			break
		}

		switch header.Action {
		case lhproto.ActionWrite:
			entChan := make(chan *lhproto.IncomingLogEntryJSON)

			go handler.Write(entChan)

			for {
				ent := new(lhproto.IncomingLogEntryJSON)

				if err := reader.ReadJSON(ent); err != nil {
					close(entChan)

					if err != jstream.ErrStreamDelimiter {
						conn.Close()
						return
					}

					break
				}

				entChan <- ent
			}
		case lhproto.ActionRead:
			qChan := make(chan *lhproto.LogQueryJSON)
			entChan := make(chan *lhproto.OutgoingLogEntryJSON)

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

		case lhproto.ActionInternalRead:
			qChan := make(chan *lhproto.LogQueryJSON)
			entChan := make(chan *lhproto.InternalLogEntryJSON)

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

		case lhproto.ActionTruncate:
			var cmd lhproto.TruncateJSON

			if err := reader.ReadJSON(&cmd); err != nil {
				conn.Close()
				return
			}

			go handler.Truncate(&cmd)

		case lhproto.ActionTransfer:
			var cmd lhproto.TransferJSON

			if err := reader.ReadJSON(&cmd); err != nil {
				conn.Close()
				return
			}

			go handler.Transfer(&cmd)

		case lhproto.ActionAccept:
			var cmd lhproto.AcceptJSON

			if err := reader.ReadJSON(&cmd); err != nil {
				conn.Close()
				return
			}

			entChan := make(chan *lhproto.InternalLogEntryJSON)
			resultChan := make(chan *lhproto.AcceptResultJSON)

			go handler.Accept(&cmd, entChan, resultChan)

			for {
				ent := new(lhproto.InternalLogEntryJSON)

				if err := reader.ReadJSON(ent); err != nil {
					close(entChan)

					if err != jstream.ErrStreamDelimiter {
						conn.Close()
						return
					}

					break
				}

				entChan <- ent
			}

			result := <-resultChan

			if err := writer.WriteJSON(result); err != nil {
				conn.Close()
				return
			}
		}
	}
}

func readLogQueryJSONChannel(reader jstream.Reader, handler lhproto.ProtocolHandler, qChan chan *lhproto.LogQueryJSON) bool {
	for {
		q := new(lhproto.LogQueryJSON)

		if err := reader.ReadJSON(q); err != nil {
			close(qChan)

			if err != jstream.ErrStreamDelimiter {
				return false
			}

			break
		}

		qChan <- q
	}

	return true
}
