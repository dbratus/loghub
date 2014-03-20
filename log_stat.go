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

func startLogStatSender(addr string, log LogManager, port int, resistanceLevel int64, sendInterval time.Duration) (func(), error) {
	closeChan := make(chan chan bool)

	var addrUdp *net.UDPAddr

	if a, err := net.ResolveUDPAddr("udp4", addr); err != nil {
		return nil, err
	} else {
		addrUdp = a
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
				if n, err := conn.WriteToUDP(msgBuf.Bytes(), addrUdp); err != nil {
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
