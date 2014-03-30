// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package balancer

import (
	"github.com/yasushi-saito/rbtree"
)

type Balancer struct {
	hosts            *rbtree.Tree
	hostsByAddr      map[string]*hostInfo
	currentTransfers map[int64]*Transfer
	nextTransferId   int64
}

type Transfer struct {
	Id     int64
	From   string
	To     string
	Amount int64
}

type hostInfo struct {
	addr  string
	size  int64
	limit int64
}

func (from *hostInfo) getAmountToTransfer(to *hostInfo) int64 {
	canGive := from.size - from.limit/2
	canGet := to.limit - to.size

	if canGet > canGive {
		return canGive
	} else {
		return canGet
	}
}

func (h *hostInfo) load() float64 {
	return float64(h.size) / float64(h.limit)
}

func compareHostInfo(a, b rbtree.Item) int {
	aload := a.(*hostInfo).load()
	bload := b.(*hostInfo).load()

	if aload > bload {
		return 1
	} else if aload < bload {
		return -1
	}

	return 0
}

func New() *Balancer {
	return &Balancer{
		rbtree.NewTree(compareHostInfo),
		make(map[string]*hostInfo),
		make(map[int64]*Transfer),
		1,
	}
}

func (bl *Balancer) MakeTransfers() (transfers []*Transfer) {
	transfers = make([]*Transfer, 0, 100)

	for bl.hosts.Len() > 0 {
		ilmost := bl.hosts.Min()
		irmost := bl.hosts.Max()
		lmost := ilmost.Item().(*hostInfo)
		rmost := irmost.Item().(*hostInfo)

		if rmost != lmost && rmost.load() > 1.0 && lmost.load() < 1.0 {
			bl.hosts.DeleteWithIterator(ilmost)
			bl.hosts.DeleteWithIterator(irmost)

			trf := &Transfer{
				bl.nextTransferId,
				rmost.addr,
				lmost.addr,
				rmost.getAmountToTransfer(lmost),
			}
			bl.currentTransfers[trf.Id] = trf
			bl.nextTransferId++

			transfers = append(transfers, trf)
		} else {
			break
		}
	}

	return
}

func (bl *Balancer) TransferComplete(transferId int64) {
	if trf, found := bl.currentTransfers[transferId]; found {
		if host, found := bl.hostsByAddr[trf.From]; found {
			bl.hosts.Insert(host)
		}

		if host, found := bl.hostsByAddr[trf.To]; found {
			bl.hosts.Insert(host)
		}

		delete(bl.currentTransfers, transferId)
	}
}

func (bl *Balancer) deleteHostFromTree(host *hostInfo) bool {
	iter := bl.hosts.FindLE(host)

	for !iter.Limit() && iter.Item() != host {
		iter = iter.Next()
	}

	if !iter.Limit() {
		bl.hosts.DeleteWithIterator(iter)
		return true
	}

	return false
}

func (bl *Balancer) UpdateHost(hostAddr string, size int64, lim int64) {
	if host, found := bl.hostsByAddr[hostAddr]; found {
		isHostInTree := bl.deleteHostFromTree(host)

		host.size = size
		host.limit = lim

		if isHostInTree {
			bl.hosts.Insert(host)
		}

	} else {
		host = &hostInfo{
			hostAddr,
			size,
			lim,
		}

		bl.hostsByAddr[hostAddr] = host
		bl.hosts.Insert(host)
	}
}

func (bl *Balancer) RemoveHost(hostAddr string) {
	if host, found := bl.hostsByAddr[hostAddr]; found {
		bl.deleteHostFromTree(host)
		delete(bl.hostsByAddr, hostAddr)
	}
}
