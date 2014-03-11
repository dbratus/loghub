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
	"container/list"
	"net"
	"runtime"
	"sync/atomic"
)

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
		println("Failed to connect to", cli.address, ":", err.Error())
	}

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
						println("Failed to connect to", cli.address, ":", err.Error())
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
		for _ = range entries {
		}
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionWrite}
	if err := writer.WriteJSON(&header); err != nil {
		println("Failed to write message", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		return
	}

	go func() {
		ok := true

		for ent := range entries {
			if ok {
				if err := writer.WriteJSON(ent); err != nil {
					println("Failed to write message", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				println("Failed to write message", err.Error())
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
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		for _ = range queries {
		}
		close(entries)
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionRead}
	if err := writer.WriteJSON(&header); err != nil {
		println("Failed to write message", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		return
	}

	go func() {
		ok := true

		for q := range queries {
			if ok {
				if err := writer.WriteJSON(q); err != nil {
					println("Failed to write message", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				println("Failed to write message", err.Error())
				cli.brokenConnChan <- conn
			} else {
				reader := NewJSONStreamReader(conn)

				for {
					ent := new(OutgoingLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != ErrStreamDelimiter {
							println("Failed to read message", err.Error())
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
	var conn net.Conn
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		for _ = range queries {
		}
		close(entries)
		return
	} else {
		conn = c
	}

	writer := NewJSONStreamWriter(conn)

	header := MessageHeaderJSON{ActionInternalRead}
	if err := writer.WriteJSON(&header); err != nil {
		println("Failed to write message", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		return
	}

	go func() {
		ok := true

		for q := range queries {
			if ok {
				if err := writer.WriteJSON(q); err != nil {
					println("Failed to write message", err.Error())
					ok = false
				}
			}
		}

		if ok {
			if err := writer.WriteDelimiter(); err != nil {
				println("Failed to write message", err.Error())
				cli.brokenConnChan <- conn
			} else {
				reader := NewJSONStreamReader(conn)

				for {
					ent := new(InternalLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != ErrStreamDelimiter {
							println("Failed to read message", err.Error())
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
