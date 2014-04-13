// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/balancer"
	"github.com/dbratus/loghub/lhproto"
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
	rebalancingInterval     = time.Second
)

type Hub interface {
	ReadLog([]*LogQuery, chan *LogEntry)
	SetLogStat(net.IP, *LogStat)
	Truncate(string, int64)
	GetStats() map[string]*LogStat
	ForEachLog(func(lhproto.ProtocolHandler))
	SetCredentials(lhproto.Credentials)

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
	readChan           chan readLogMultiSrcCmd
	statChan           chan setLogStatCmd
	truncateChan       chan truncateLogCmd
	getStatsChan       chan chan map[string]*LogStat
	credChan           chan lhproto.Credentials
	iterChan           chan func(lhproto.ProtocolHandler)
	closeChan          chan chan bool
	useTls             bool
	skipCertValidation bool
}

func NewDefaultHub(useTls bool, skipCertValidation bool) Hub {
	h := &defaultHub{
		make(chan readLogMultiSrcCmd),
		make(chan setLogStatCmd),
		make(chan truncateLogCmd),
		make(chan chan map[string]*LogStat),
		make(chan lhproto.Credentials),
		make(chan func(lhproto.ProtocolHandler)),
		make(chan chan bool),
		useTls,
		skipCertValidation,
	}

	go h.run()

	return h
}

func (h *defaultHub) run() {
	type logInfo struct {
		stat           *LogStat
		client         lhproto.ProtocolHandler
		timeout        time.Time
		usersCount     *int32
		lastTransferId int64
	}

	logs := make(map[string]*logInfo)
	logBalancer := balancer.New()
	cred := lhproto.Credentials{"", ""}

	readLog := func(queries []*LogQuery, client lhproto.ProtocolHandler, entries chan *LogEntry, usersCount *int32) {
		defer atomic.AddInt32(usersCount, -1)

		queriesJSON := make(chan *lhproto.LogQueryJSON)
		results := make(chan *lhproto.InternalLogEntryJSON)

		client.InternalRead(&cred, queriesJSON, results)

		for _, q := range queries {
			queriesJSON <- LogQueryToLogQueryJSON(q)
		}

		close(queriesJSON)

		for ent := range results {
			entries <- InternalLogEntryJSONToLogEntry(ent)
		}

		close(entries)
	}

	onRead := func(cmd readLogMultiSrcCmd) {
		var results chan *LogEntry = nil

		for _, log := range logs {
			entries := make(chan *LogEntry)

			atomic.AddInt32(log.usersCount, 1)
			go readLog(cmd.queries, log.client, entries, log.usersCount)

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

	onSetLogStat := func(cmd setLogStatCmd) {
		addr := cmd.addr.String() + ":" + strconv.Itoa(cmd.stat.Port)

		hubTrace.Debugf("Got stat from %s: sz=%d, lim=%d, trid=%d.", addr, cmd.stat.Size, cmd.stat.Limit, cmd.stat.LastTransferId)

		if log, found := logs[addr]; found {
			if cmd.stat.Timestamp > log.stat.Timestamp {
				hubTrace.Debugf("Updating stat of %s.", addr)

				log.stat = cmd.stat
				log.timeout = time.Now().Add(logCloseTimeout)

				logBalancer.UpdateHost(addr, cmd.stat.Size, cmd.stat.Limit)

				if log.lastTransferId != cmd.stat.LastTransferId {
					hubTrace.Debugf("Transfer %d at %s complete.", cmd.stat.LastTransferId, addr)

					log.lastTransferId = cmd.stat.LastTransferId
					logBalancer.TransferComplete(cmd.stat.LastTransferId)
				}
			}
		} else {
			hubTrace.Debugf("Creating stat of %s.", addr)

			logs[addr] = &logInfo{
				cmd.stat,
				lhproto.NewClient(addr, maxConnectionsPerClient, h.useTls, h.skipCertValidation),
				time.Now().Add(logCloseTimeout),
				new(int32),
				cmd.stat.LastTransferId,
			}

			logBalancer.UpdateHost(addr, cmd.stat.Size, cmd.stat.Limit)
		}
	}

	onCheckTimeouts := func() {
		now := time.Now()

		for ip, log := range logs {
			if now.After(log.timeout) {
				delete(logs, ip)

				go func(log *logInfo) {
					for atomic.LoadInt32(log.usersCount) > 0 {
						runtime.Gosched()
					}
					log.client.Close()
				}(log)
			}
		}
	}

	onTruncate := func(cmd truncateLogCmd) {
		for _, log := range logs {
			atomic.AddInt32(log.usersCount, 1)

			go func(log *logInfo) {
				defer atomic.AddInt32(log.usersCount, -1)

				log.client.Truncate(&cred, &lhproto.TruncateJSON{cmd.source, cmd.limit})
			}(log)
		}
	}

	onRebalance := func() {
		if cred.User != "" {
			for _, transfer := range logBalancer.MakeTransfers() {
				hubTrace.Debugf("Transfering %d from %s to %s, id %d.", transfer.Amount, transfer.From, transfer.To, transfer.Id)

				if log, found := logs[transfer.From]; found {
					atomic.AddInt32(log.usersCount, 1)

					go func(log *logInfo, transfer *balancer.Transfer) {
						defer atomic.AddInt32(log.usersCount, -1)

						log.client.Transfer(&cred, &lhproto.TransferJSON{transfer.Id, transfer.To, transfer.Amount})
					}(log, transfer)
				}
			}
		}
	}

	onGetStats := func(statsChan chan map[string]*LogStat) {
		stats := make(map[string]*LogStat)

		for addr, log := range logs {
			stats[addr] = log.stat
		}

		statsChan <- stats
		close(statsChan)
	}

	onIter := func(iter func(lhproto.ProtocolHandler)) {
		for _, log := range logs {
			atomic.AddInt32(log.usersCount, 1)

			go func(log *logInfo) {
				defer atomic.AddInt32(log.usersCount, -1)

				iter(log.client)
			}(log)
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

		case cmd, ok := <-h.truncateChan:
			if ok {
				onTruncate(cmd)
			}

		case statsChan, ok := <-h.getStatsChan:
			if ok {
				onGetStats(statsChan)
			}

		case newCred, ok := <-h.credChan:
			if ok {
				cred = newCred

				hubTrace.Info("Credentials changed.")
			}

		case iter, ok := <-h.iterChan:
			if ok {
				onIter(iter)
			}

		case <-time.After(logTimeoutsCheckInteval):
			onCheckTimeouts()

		case <-time.After(rebalancingInterval):
			onRebalance()

		case ack := <-h.closeChan:
			for cmd := range h.readChan {
				onRead(cmd)
			}

			for statsChan := range h.getStatsChan {
				onGetStats(statsChan)
			}

			for iter := range h.iterChan {
				onIter(iter)
			}

			ack <- true
			close(ack)
			return
		}
	}
}

func (h *defaultHub) SetLogStat(addr net.IP, stat *LogStat) {
	h.statChan <- setLogStatCmd{addr, stat}
}

func (h *defaultHub) ReadLog(queries []*LogQuery, entries chan *LogEntry) {
	h.readChan <- readLogMultiSrcCmd{queries, entries}
}

func (h *defaultHub) Truncate(source string, limit int64) {
	h.truncateChan <- truncateLogCmd{source, limit}
}

func (h *defaultHub) GetStats() map[string]*LogStat {
	statsChan := make(chan map[string]*LogStat)
	h.getStatsChan <- statsChan
	return <-statsChan
}

func (h *defaultHub) ForEachLog(iter func(lhproto.ProtocolHandler)) {
	h.iterChan <- iter
}

func (h *defaultHub) SetCredentials(cred lhproto.Credentials) {
	h.credChan <- cred
}

func (h *defaultHub) Close() {
	close(h.readChan)
	close(h.statChan)
	close(h.getStatsChan)
	close(h.credChan)
	close(h.iterChan)

	ack := make(chan bool)
	h.closeChan <- ack
	<-ack

	close(h.closeChan)
}
