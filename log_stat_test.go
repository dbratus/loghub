// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"net"
	"testing"
	"time"
)

type logManagerForStatTest struct {
	size int64
}

func (mg *logManagerForStatTest) WriteLog(*LogEntry) {
}

func (mg *logManagerForStatTest) ReadLog(*LogQuery, chan *LogEntry) {
}

func (mg *logManagerForStatTest) Close() {
}

func (mg *logManagerForStatTest) Truncate(string, int64) {
}

func (mg *logManagerForStatTest) Size() int64 {
	return mg.size
}

func (mg *logManagerForStatTest) GetTransferChunk(maxSize int64, entries chan *LogEntry) (id string, size int64, found bool) {
	return "", 0, false
}

func (mg *logManagerForStatTest) AcceptTransferChunk(id string, entries chan *LogEntry) chan bool {
	return nil
}

func (mg *logManagerForStatTest) DeleteTransferChunk(id string) {
}

type hubForStatTest struct {
	stat chan *LogStat
}

func (h *hubForStatTest) ReadLog([]*LogQuery, chan *LogEntry) {
}

func (h *hubForStatTest) SetLogStat(addr net.IP, stat *LogStat) {
	h.stat <- stat
}

func (h *hubForStatTest) Truncate(source string, limit int64) {
}

func (h *hubForStatTest) Close() {
}

func TestLogStatSenderReceiver(t *testing.T) {
	logManager := &logManagerForStatTest{1000}

	senderPort := 9999
	lim := int64(20000)
	lastTransferId := new(int64)

	if cl, err := startLogStatSender(":10000", logManager, senderPort, lim, lastTransferId, time.Second); err != nil {
		t.Errorf("Failed to start LogStat sender: %s.", err.Error())
		t.FailNow()
	} else {
		defer cl()
	}

	hub := &hubForStatTest{make(chan *LogStat)}

	if cl, err := startLogStatReceiver(":10000", hub); err != nil {
		t.Errorf("Failed to start LogStat receiver: %s.", err.Error())
		t.FailNow()
	} else {
		defer cl()
	}

	select {
	case <-time.After(time.Second * 10):
		t.Error("LogStat has not arrived.")
		t.FailNow()
	case stat := <-hub.stat:
		if stat.Port != senderPort {
			t.Errorf("Wrong LogStat port. Expected %d, got %d.", senderPort, stat.Port)
		}

		if stat.Limit != lim {
			t.Errorf("Wrong LogStat limit. Expected %d, got %d.", lim, stat.Limit)
		}

		if stat.Size != logManager.size {
			t.Errorf("Wrong LogStat size. Expected %d, got %d.", logManager.size, stat.Size)
		}
	}
}
