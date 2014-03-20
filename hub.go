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
	"github.com/dbratus/loghub/trace"
	"net"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var hubTrace = trace.New("Hub")

const (
	maxConnectionsPerClient = 10
	logCloseTimeout         = time.Minute * 30
	logTimeoutsCheckInteval = time.Second * 10
)

type Hub interface {
	ReadLog([]*LogQuery, chan *LogEntry)
	SetLogStat(net.IP, *LogStat)
	Close()
}

type setLogStatCmd struct {
	addr net.IP
	stat *LogStat
}

type readLogMultiSrcCmd struct {
	queries []*LogQuery
	entries chan *LogEntry
}

type defaultHub struct {
	readChan  chan *readLogMultiSrcCmd
	statChan  chan *setLogStatCmd
	closeChan chan chan bool
}

func NewDefaultHub() Hub {
	h := &defaultHub{
		make(chan *readLogMultiSrcCmd),
		make(chan *setLogStatCmd),
		make(chan chan bool),
	}

	go h.run()

	return h
}

func (h *defaultHub) run() {
	type logInfo struct {
		stat         *LogStat
		client       MessageHandler
		timeout      time.Time
		readersCount *int32
	}

	logs := make(map[string]*logInfo)

	readLog := func(queries []*LogQuery, client MessageHandler, entries chan *LogEntry, readersCount *int32) {
		atomic.AddInt32(readersCount, 1)
		defer atomic.AddInt32(readersCount, -1)

		queriesJSON := make(chan *LogQueryJSON)
		results := make(chan *InternalLogEntryJSON)

		client.InternalRead(queriesJSON, results)

		for _, q := range queries {
			queriesJSON <- LogQueryToLogQueryJSON(q)
		}

		close(queriesJSON)

		for ent := range results {
			entries <- InternalLogEntryJSONToLogEntry(ent)
		}

		close(entries)
	}

	onRead := func(cmd *readLogMultiSrcCmd) {
		var results chan *LogEntry = nil

		for _, stat := range logs {
			entries := make(chan *LogEntry)

			go readLog(cmd.queries, stat.client, entries, stat.readersCount)

			if results == nil {
				results = entries
			} else {
				merged := make(chan *LogEntry)
				go MergeLogs(entries, results, merged)
				results = merged
			}
		}

		go ForwardLog(results, cmd.entries)
	}

	onSetLogStat := func(cmd *setLogStatCmd) {
		addr := cmd.addr.String() + ":" + strconv.Itoa(cmd.stat.Port)

		hubTrace.Debugf("Got stat from %s: SZ=%d, RL=%d.", addr, cmd.stat.Size, cmd.stat.ResistanceLevel)

		if log, found := logs[addr]; found {
			if cmd.stat.Timestamp > log.stat.Timestamp {
				log.stat = cmd.stat
				log.timeout = time.Now().Add(logCloseTimeout)
			}
		} else {
			logs[addr] = &logInfo{
				cmd.stat,
				NewLogHubClient(addr, maxConnectionsPerClient),
				time.Now().Add(logCloseTimeout),
				new(int32),
			}
		}
	}

	onCheckTimeouts := func() {
		now := time.Now()

		for ip, stat := range logs {
			if now.After(stat.timeout) {
				delete(logs, ip)

				go func() {
					for atomic.LoadInt32(stat.readersCount) > 0 {
						runtime.Gosched()
					}
					stat.client.Close()
				}()
			}
		}
	}

	for {
		select {
		case cmd, ok := <-h.readChan:
			if ok {
				onRead(cmd)
			}

		case cmd, ok := <-h.statChan:
			if ok {
				onSetLogStat(cmd)
			}

		case <-time.After(logTimeoutsCheckInteval):
			onCheckTimeouts()

		case ack := <-h.closeChan:
			for cmd := range h.readChan {
				onRead(cmd)
			}

			ack <- true
			close(ack)
			return
		}
	}
}

func (h *defaultHub) SetLogStat(addr net.IP, stat *LogStat) {
	h.statChan <- &setLogStatCmd{addr, stat}
}

func (h *defaultHub) ReadLog(queries []*LogQuery, entries chan *LogEntry) {
	h.readChan <- &readLogMultiSrcCmd{queries, entries}
}

func (h *defaultHub) Close() {
	close(h.readChan)
	close(h.statChan)

	ack := make(chan bool)
	h.closeChan <- ack
	<-ack

	close(h.closeChan)
}
