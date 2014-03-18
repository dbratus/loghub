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
	"net"
)

const maxConnectionsPerClient = 10

type Hub interface {
	ReadLog(chan *LogQuery, chan *LogEntry)
	SetLogStat(net.IP, *LogStat)
	Close()
}

type setLogStatCmd struct {
	addr net.IP
	stat *LogStat
}

type readLogMultiSrcCmd struct {
	queries chan *LogQuery
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

	onRead := func(cmd *readLogMultiSrcCmd) {
		close(cmd.entries)
	}

	onSetLogStat := func(cmd *setLogStatCmd) {
	}

	for {
		select {
		case cmd := <-h.readChan:
			onRead(cmd)

		case cmd := <-h.statChan:
			onSetLogStat(cmd)

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

func (h *defaultHub) ReadLog(queries chan *LogQuery, entries chan *LogEntry) {
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
