// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/gob"
	"github.com/dbratus/loghub/trace"
	"net"
	"time"
)

var (
	logStatSenderTrace   = trace.New("LogStatSender")
	logStatReceiverTrace = trace.New("LogStatReceiver")
)

type LogStat struct {
	Timestamp       int64
	Size            int64
	ResistanceLevel int64
	Port            int
}

func startLogStatSender(hubAddr string, log LogManager, port int, resistanceLevel int64, sendInterval time.Duration) (func(), error) {
	closeChan := make(chan chan bool)

	var hubAddrUdp *net.UDPAddr

	if a, err := net.ResolveUDPAddr("udp4", hubAddr); err != nil {
		return nil, err
	} else {
		hubAddrUdp = a
	}

	var conn *net.UDPConn

	if c, err := net.ListenUDP("udp4", nil); err != nil {
		return nil, err
	} else {
		conn = c
	}

	go func() {
		msgBuf := new(bytes.Buffer)

		sendStat := func() {
			stat := &LogStat{
				time.Now().UnixNano(),
				log.Size(),
				resistanceLevel,
				port,
			}
			encoder := gob.NewEncoder(msgBuf)

			if err := encoder.Encode(stat); err != nil {
				logStatSenderTrace.Errorf("Failed to encode LogStat: %s.", err.Error())
			} else {
				if n, err := conn.WriteToUDP(msgBuf.Bytes(), hubAddrUdp); err != nil {
					logStatSenderTrace.Errorf("Failed to write LogStat: %s.", err.Error())
				} else {
					logStatSenderTrace.Debugf("%d of %d bytes sent.", n, msgBuf.Len())
				}
			}

			msgBuf.Reset()
		}

		sendStat()

		for {
			select {
			case <-time.After(sendInterval):
				sendStat()

			case ack := <-closeChan:
				conn.Close()
				ack <- true
				return
			}
		}
	}()

	return func() {
		ack := make(chan bool)
		closeChan <- ack
		<-ack
	}, nil
}

func startLogStatReceiver(addr string, hub Hub) (func(), error) {

	var addrUdp *net.UDPAddr

	if a, err := net.ResolveUDPAddr("udp4", addr); err != nil {
		return nil, err
	} else {
		addrUdp = a
	}

	var conn *net.UDPConn

	if c, err := net.ListenUDP("udp4", addrUdp); err != nil {
		return nil, err
	} else {
		conn = c
	}

	buf := make([]byte, 1024)

	go func() {
		for {
			if n, senderAddr, err := conn.ReadFromUDP(buf); err == nil {
				logStatReceiverTrace.Debugf("Received %d bytes from %s.", n, senderAddr.IP.String())

				msgBuf := bytes.NewBuffer(buf[:n])
				decoder := gob.NewDecoder(msgBuf)

				var stat LogStat

				if err := decoder.Decode(&stat); err != nil {
					logStatSenderTrace.Errorf("Failed to decode LogStat: %s.", err.Error())
				} else {
					hub.SetLogStat(senderAddr.IP, &stat)
				}

			} else {
				return
			}
		}
	}()

	return func() {
		conn.Close()
	}, nil
}