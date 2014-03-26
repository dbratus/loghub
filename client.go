// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"container/list"
	"github.com/dbratus/loghub/trace"
	"net"
	"runtime"
	"sync/atomic"
)

var clientTrace = trace.New("LogHubClient")

type logHubClient struct {
	connPoolSize   int
	address        string
	getConnChan    chan chan net.Conn
	putConnChan    chan net.Conn
	brokenConnChan chan net.Conn
	closeChan      chan chan bool
	isClosed       *int32
	activeOps      *int32
}

func NewLogHubClient(address string, connPoolSize int) MessageHandler {
	cli := &logHubClient{
		connPoolSize,
		address,
		make(chan chan net.Conn),
		make(chan net.Conn),
		make(chan net.Conn),
		make(chan chan bool),
		new(int32),
		new(int32),
	}

	go cli.connPool()

	return cli
}

func (cli *logHubClient) connPool() {
	connections := list.New()
	waiters := list.New()
	totalConnections := 0

	if conn, err := net.Dial("tcp", cli.address); err == nil {
		connections.PushBack(conn)
		totalConnections++
	} else {
		clientTrace.Errorf("Failed to connect to %s: %s", cli.address, err.Error())
	}

	for {
		select {
		case connChan, ok := <-cli.getConnChan:
			if ok {
				if connections.Len() > 0 {
					connChan <- connections.Remove(connections.Front()).(net.Conn)
				} else {
					if totalConnections < cli.connPoolSize {
						if conn, err := net.Dial("tcp", cli.address); err == nil {
							connChan <- conn
							totalConnections++
						} else {
							clientTrace.Errorf("Failed to connect to %s: %s.", cli.address, err.Error())
							close(connChan)
						}
					} else {
						waiters.PushBack(connChan)
					}
				}
			}
		case conn, ok := <-cli.putConnChan:
			if ok {
				if waiters.Len() == 0 {
					connections.PushBack(conn)
				} else {
					connChan := waiters.Remove(waiters.Front()).(chan net.Conn)
					connChan <- conn
				}
			}
		case conn, ok := <-cli.brokenConnChan:
			if ok {
				conn.Close()
				totalConnections--
			}
		case ack := <-cli.closeChan:
			for connChan := range cli.getConnChan {
				close(connChan)
			}

			for conn := range cli.putConnChan {
				conn.Close()
			}

			for conn := range cli.brokenConnChan {
				conn.Close()
			}

			for conn := connections.Front(); conn != nil; conn = conn.Next() {
				conn.Value.(net.Conn).Close()
			}

			ack <- true
			return
		}
	}
}

func (cli *logHubClient) getConn() (net.Conn, bool) {
	if atomic.LoadInt32(cli.isClosed) > 0 {
		return nil, false
	}

	connChan := make(chan net.Conn)
	cli.getConnChan <- connChan
	conn, ok := <-connChan

	return conn, ok
}

func (cli *logHubClient) Write(entries chan *IncomingLogEntryJSON) {
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeIncomingLogEntryJSON(entries)
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionWrite}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeIncomingLogEntryJSON(entries)
		return
	}

	go func() {
		ok := true

		for ent := range entries {
			if ok {
				if err := writer.WriteJSON(ent); err != nil {
					clientTrace.Errorf("Failed to write message: %s.", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				clientTrace.Errorf("Failed to write message: %s.", err.Error())
				cli.brokenConnChan <- conn
			} else {
				cli.putConnChan <- conn
			}
		} else {
			cli.brokenConnChan <- conn
		}

		atomic.AddInt32(cli.activeOps, -1)
	}()
}

func (cli *logHubClient) Read(queries chan *LogQueryJSON, entries chan *OutgoingLogEntryJSON) {
	//TODO: Get rid of duplication.
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionRead}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	}

	go func() {
		ok := true

		for q := range queries {
			if ok {
				if err := writer.WriteJSON(q); err != nil {
					clientTrace.Errorf("Failed to write message: %s.", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				clientTrace.Errorf("Failed to write message: %s.", err.Error())
				cli.brokenConnChan <- conn
			} else {
				reader := NewJSONStreamReader(conn)

				for {
					ent := new(OutgoingLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != ErrStreamDelimiter {
							clientTrace.Errorf("Failed to read message: %s.", err.Error())
							ok = false
						}
						break
					}

					entries <- ent
				}

				if ok {
					cli.putConnChan <- conn
				} else {
					cli.brokenConnChan <- conn
				}
			}
		} else {
			cli.brokenConnChan <- conn
		}

		close(entries)
		atomic.AddInt32(cli.activeOps, -1)
	}()
}

func (cli *logHubClient) InternalRead(queries chan *LogQueryJSON, entries chan *InternalLogEntryJSON) {
	//TODO: Get rid of duplication.
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionInternalRead}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	}

	go func() {
		ok := true

		for q := range queries {
			if ok {
				if err := writer.WriteJSON(q); err != nil {
					clientTrace.Errorf("Failed to write message: %s.", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				clientTrace.Errorf("Failed to write message: %s.", err.Error())
				cli.brokenConnChan <- conn
			} else {
				reader := NewJSONStreamReader(conn)

				for {
					ent := new(InternalLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != ErrStreamDelimiter {
							clientTrace.Errorf("Failed to read message: %s.", err.Error())
							ok = false
						}
						break
					}

					entries <- ent
				}

				if ok {
					cli.putConnChan <- conn
				} else {
					cli.brokenConnChan <- conn
				}
			}
		} else {
			cli.brokenConnChan <- conn
		}

		close(entries)
		atomic.AddInt32(cli.activeOps, -1)
	}()
}

func (cli *logHubClient) Truncate(cmd *TruncateJSON) {
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)
	defer atomic.AddInt32(cli.activeOps, -1)

	if c, ok := cli.getConn(); !ok {
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionTruncate}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	if err := writer.WriteJSON(cmd); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	cli.putConnChan <- conn
}

func (cli *logHubClient) Transfer(cmd *TransferJSON) {
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)
	defer atomic.AddInt32(cli.activeOps, -1)

	if c, ok := cli.getConn(); !ok {
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionTransfer}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	if err := writer.WriteJSON(cmd); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	cli.putConnChan <- conn
}

func (cli *logHubClient) Accept(cmd *AcceptJSON, entries chan *InternalLogEntryJSON, resultChan chan *AcceptResultJSON) {
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	respond := func(result bool) {
		resultChan <- &AcceptResultJSON{result}
		close(resultChan)
	}

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeInternalLogEntryJSON(entries)
		go respond(false)
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)
	reader := NewJSONStreamReader(conn)

	header := MessageHeaderJSON{ActionAccept}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeInternalLogEntryJSON(entries)
		go respond(false)
	}

	go func() {
		defer atomic.AddInt32(cli.activeOps, -1)

		if err := writer.WriteJSON(cmd); err != nil {
			clientTrace.Errorf("Failed to write message: %s.", err.Error())
			respond(false)
			cli.brokenConnChan <- conn
			return
		}

		ok := true

		for ent := range entries {
			if ok {
				if err := writer.WriteJSON(ent); err != nil {
					clientTrace.Errorf("Failed to write message: %s.", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				clientTrace.Errorf("Failed to write message: %s.", err.Error())
				respond(false)
				cli.brokenConnChan <- conn
				return
			}

			var acceptResult AcceptResultJSON

			if err := reader.ReadJSON(&acceptResult); err != nil {
				clientTrace.Errorf("Failed to read message: %s.", err.Error())
				respond(false)
				cli.brokenConnChan <- conn
				return
			}

			respond(acceptResult.Result)
			cli.putConnChan <- conn
		} else {
			respond(false)
			cli.brokenConnChan <- conn
		}
	}()
}

func (cli *logHubClient) Close() {
	atomic.AddInt32(cli.isClosed, 1)

	close(cli.getConnChan)

	for atomic.LoadInt32(cli.activeOps) > 0 {
		runtime.Gosched()
	}

	close(cli.putConnChan)
	close(cli.brokenConnChan)

	ack := make(chan bool)
	cli.closeChan <- ack
	<-ack

	close(cli.closeChan)
}
