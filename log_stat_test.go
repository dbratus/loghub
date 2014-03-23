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

func (mg *logManagerForStatTest) Truncate(int64, string) {
}

func (mg *logManagerForStatTest) Size() int64 {
	return mg.size
}

type hubForStatTest struct {
	stat *LogStat
}

func (h *hubForStatTest) ReadLog([]*LogQuery, chan *LogEntry) {
}

func (h *hubForStatTest) SetLogStat(addr net.IP, stat *LogStat) {
	h.stat = stat
}

func (h *hubForStatTest) Close() {
}

func TestLogStatSenderReceiver(t *testing.T) {
	logManager := &logManagerForStatTest{1000}

	senderPort := 9999
	resistanceLevel := int64(20000)

	if cl, err := startLogStatSender(":10000", logManager, senderPort, resistanceLevel, time.Second); err != nil {
		t.Errorf("Failed to start LogStat sender: %s.", err.Error())
		t.FailNow()
	} else {
		defer cl()
	}

	hub := new(hubForStatTest)

	if cl, err := startLogStatReceiver(":10000", hub); err != nil {
		t.Errorf("Failed to start LogStat receiver: %s.", err.Error())
		t.FailNow()
	} else {
		defer cl()
	}

	time.Sleep(time.Millisecond * 100)

	if hub.stat == nil {
		t.Error("LogStat has not arrived.")
		t.FailNow()
	}

	if hub.stat.Port != senderPort {
		t.Errorf("Wrong LogStat port. Expected %d, got %d.", senderPort, hub.stat.Port)
	}

	if hub.stat.ResistanceLevel != resistanceLevel {
		t.Errorf("Wrong LogStat resistence level. Expected %d, got %d.", resistanceLevel, hub.stat.ResistanceLevel)
	}

	if hub.stat.Size != logManager.size {
		t.Errorf("Wrong LogStat size. Expected %d, got %d.", logManager.size, hub.stat.Size)
	}
}
