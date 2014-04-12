// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package lhproto

import (
	"container/list"
	"crypto/tls"
	"github.com/dbratus/loghub/jstream"
	"github.com/dbratus/loghub/trace"
	"io"
	"net"
	"runtime"
	"sync/atomic"
)

var clientTrace = trace.New("LogHubClient")

type logHubClient struct {
	connPoolSize   int
	address        string
	getConnChan    chan chan io.ReadWriteCloser
	putConnChan    chan io.ReadWriteCloser
	brokenConnChan chan io.ReadWriteCloser
	closeChan      chan chan bool
	isClosed       *int32
	activeOps      *int32
	tlsConfig      *tls.Config
}

func NewClient(address string, connPoolSize int, useTls bool, skipCertValidation bool) ProtocolHandler {
	var tlsConfig *tls.Config

	if useTls {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: skipCertValidation,
		}
	}

	cli := &logHubClient{
		connPoolSize,
		address,
		make(chan chan io.ReadWriteCloser),
		make(chan io.ReadWriteCloser),
		make(chan io.ReadWriteCloser),
		make(chan chan bool),
		new(int32),
		new(int32),
		tlsConfig,
	}

	go cli.connPool()

	return cli
}

func (cli *logHubClient) newConnection() (io.ReadWriteCloser, error) {
	if cli.tlsConfig == nil {
		return net.Dial("tcp", cli.address)
	} else {
		return tls.Dial("tcp", cli.address, cli.tlsConfig)
	}
}

func (cli *logHubClient) connPool() {
	connections := list.New()
	waiters := list.New()
	totalConnections := 0

	if conn, err := cli.newConnection(); err == nil {
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
					connChan <- connections.Remove(connections.Front()).(io.ReadWriteCloser)
				} else {
					if totalConnections < cli.connPoolSize {
						if conn, err := cli.newConnection(); err == nil {
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
					connChan := waiters.Remove(waiters.Front()).(chan io.ReadWriteCloser)
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
				conn.Value.(io.ReadWriteCloser).Close()
			}

			ack <- true
			return
		}
	}
}

func (cli *logHubClient) getConn() (io.ReadWriteCloser, bool) {
	if atomic.LoadInt32(cli.isClosed) > 0 {
		return nil, false
	}

	connChan := make(chan io.ReadWriteCloser)
	cli.getConnChan <- connChan
	conn, ok := <-connChan

	return conn, ok
}

func (cli *logHubClient) Write(cred *Credentials, entries chan *IncomingLogEntryJSON) {
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeIncomingLogEntryJSON(entries)
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionWrite, cred.User, cred.Password}
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

func (cli *logHubClient) Read(cred *Credentials, queries chan *LogQueryJSON, entries chan *OutgoingLogEntryJSON) {
	//TODO: Get rid of duplication.
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionRead, cred.User, cred.Password}
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
				reader := jstream.NewReader(conn)

				for {
					ent := new(OutgoingLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != jstream.ErrStreamDelimiter {
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

func (cli *logHubClient) InternalRead(cred *Credentials, queries chan *LogQueryJSON, entries chan *InternalLogEntryJSON) {
	//TODO: Get rid of duplication.
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		go PurgeLogQueryJSON(queries)
		close(entries)
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionInternalRead, cred.User, cred.Password}
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
				reader := jstream.NewReader(conn)

				for {
					ent := new(InternalLogEntryJSON)

					if err := reader.ReadJSON(ent); err != nil {
						if err != jstream.ErrStreamDelimiter {
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

func (cli *logHubClient) Truncate(cred *Credentials, cmd *TruncateJSON) {
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)
	defer atomic.AddInt32(cli.activeOps, -1)

	if c, ok := cli.getConn(); !ok {
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionTruncate, cred.User, cred.Password}
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

func (cli *logHubClient) Transfer(cred *Credentials, cmd *TransferJSON) {
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)
	defer atomic.AddInt32(cli.activeOps, -1)

	if c, ok := cli.getConn(); !ok {
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionTransfer, cred.User, cred.Password}
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

func (cli *logHubClient) Accept(cred *Credentials, cmd *AcceptJSON, entries chan *InternalLogEntryJSON, resultChan chan *AcceptResultJSON) {
	var conn io.ReadWriteCloser
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

	writer := jstream.NewWriter(conn)
	reader := jstream.NewReader(conn)

	header := MessageHeaderJSON{ActionAccept, cred.User, cred.Password}
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

func (cli *logHubClient) Stat(cred *Credentials, stats chan *StatJSON) {
	var conn io.ReadWriteCloser
	atomic.AddInt32(cli.activeOps, 1)

	if c, ok := cli.getConn(); !ok {
		atomic.AddInt32(cli.activeOps, -1)
		close(stats)
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)
	reader := jstream.NewReader(conn)

	header := MessageHeaderJSON{ActionStat, cred.User, cred.Password}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		atomic.AddInt32(cli.activeOps, -1)
		close(stats)
		return
	}

	go func() {
		ok := true

		for {
			stat := new(StatJSON)

			if err := reader.ReadJSON(stat); err != nil {
				if err != jstream.ErrStreamDelimiter {
					clientTrace.Errorf("Failed to read message: %s.", err.Error())
					ok = false
				}
				break
			}

			stats <- stat
		}

		if ok {
			cli.putConnChan <- conn
		} else {
			cli.brokenConnChan <- conn
		}

		close(stats)
		atomic.AddInt32(cli.activeOps, -1)
	}()
}

func (cli *logHubClient) User(cred *Credentials, usr *UserInfoJSON) {
	var conn io.ReadWriteCloser

	if c, ok := cli.getConn(); !ok {
		return
	} else {
		conn = c
	}

	writer := jstream.NewWriter(conn)

	header := MessageHeaderJSON{ActionUser, cred.User, cred.Password}
	if err := writer.WriteJSON(&header); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	if err := writer.WriteJSON(&usr); err != nil {
		clientTrace.Errorf("Failed to write message: %s.", err.Error())
		cli.brokenConnChan <- conn
		return
	}

	cli.putConnChan <- conn
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
