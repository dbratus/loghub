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
	"sync/atomic"
	"time"
)

var (
	logStatSenderTrace   = trace.New("LogStatSender")
	logStatReceiverTrace = trace.New("LogStatReceiver")
)

func startLogStatSender(hubAddr string, log LogManager, addr string, lim int64, lastTransferId *int64, sendInterval time.Duration) (func(), error) {
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
				timeToTimestamp(time.Now()),
				log.Size(),
				lim,
				addr,
				atomic.LoadInt64(lastTransferId),
			}
			encoder := gob.NewEncoder(msgBuf)

			if err := encoder.Encode(stat); err != nil {
				logStatSenderTrace.Errorf("Failed to encode LogStat: %s.", err.Error())
			} else {
				if _, err := conn.WriteToUDP(msgBuf.Bytes(), hubAddrUdp); err != nil {
					logStatSenderTrace.Errorf("Failed to write LogStat: %s.", err.Error())
				} else {
					logStatSenderTrace.Debugf("Stat sent to %s: sz=%d, lim=%d, trid=%d.", hubAddrUdp.String(), stat.Size, stat.Limit, stat.LastTransferId)
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
			if n, _, err := conn.ReadFromUDP(buf); err == nil {
				msgBuf := bytes.NewBuffer(buf[:n])
				decoder := gob.NewDecoder(msgBuf)

				var stat LogStat

				if err := decoder.Decode(&stat); err != nil {
					logStatReceiverTrace.Errorf("Failed to decode LogStat: %s.", err.Error())
				} else {
					logStatReceiverTrace.Debugf("Stat received from %s: sz=%d, lim=%d, trid=%d.", stat.Addr, stat.Size, stat.Limit, stat.LastTransferId)
					hub.SetLogStat(&stat)
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
