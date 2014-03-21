// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package rnglock

type RangeLock struct {
	lockRequests   chan lockRequest
	unlockRequests chan LockId
	closeChan      chan chan bool
}

type rng struct {
	start int64
	end   int64
}

type LockId int64

const InvalidLockId = LockId(-1)

const blocksCapacity = 10

type lockRequest struct {
	rng    rng
	isRead bool
	result chan LockId
}

func (a *rng) overlaps(b rng) bool {
	return !(a.end < b.start || b.end < a.start)
}

func New() *RangeLock {
	rl := &RangeLock{make(chan lockRequest), make(chan LockId), make(chan chan bool)}

	go rl.processRequests()

	return rl
}

func (rl *RangeLock) processRequests() {
	type lockData struct {
		id  LockId
		rng rng

		//The channel from which the lock owner
		//gets the lock id. As soon as the id is
		//sent, the lock owner is allowed to pass.
		resultChan chan LockId

		//How many locks, blocks this lock.
		blockersCnt int
		//The locks blocked by this lock.
		blocks []*lockData

		//Whether this is a read lock.
		isRead bool

		prev *lockData
		next *lockData
	}

	var locks *lockData = nil
	locksById := make(map[LockId]*lockData)
	nextLockId := LockId(1)

	lock := func(req lockRequest) {
		//Creating new lock.
		newLock := &lockData{
			nextLockId,
			req.rng,
			req.result,
			0,
			make([]*lockData, 0, blocksCapacity),
			req.isRead,
			nil,
			nil,
		}
		locksById[nextLockId] = newLock

		//Incrementing ids conter.
		nextLockId++

		//Looking for conflicting locks.
		var cur *lockData
		for cur = locks; cur != nil; cur = cur.next {
			if cur.rng.overlaps(req.rng) && !(cur.isRead && req.isRead) {
				newLock.blockersCnt++
				cur.blocks = append(cur.blocks, newLock)
			}
		}

		//Appending the new lock to the locks list.
		newLock.prev = cur

		if cur != nil {
			cur.next = newLock
		} else {
			locks = newLock
		}

		//If the new lock is not blocked,
		//allowing the lock owner to pass.
		if newLock.blockersCnt == 0 {
			req.result <- newLock.id
			close(req.result)
		}
	}

	unlock := func(lck LockId) {
		if lock, found := locksById[lck]; found {
			//Deleting the lock.
			delete(locksById, lck)

			if lock.prev != nil {
				lock.prev.next = lock.next
			} else {
				locks = lock.next
			}

			if lock.next != nil {
				lock.next.prev = lock.prev
			}

			//Releasing the blocked locks.
			for _, blk := range lock.blocks {
				blk.blockersCnt--

				if blk.blockersCnt == 0 {
					blk.resultChan <- blk.id
					close(blk.resultChan)
				}
			}
		}
	}

	for {
		select {
		case lck, ok := <-rl.unlockRequests:
			if ok {
				unlock(lck)
			}
		case req, ok := <-rl.lockRequests:
			if ok {
				lock(req)
			}
		case ack := <-rl.closeChan:
			for req := range rl.lockRequests {
				req.result <- InvalidLockId
				close(req.result)
			}

			ack <- true
			return
		}
	}
}

func (rl *RangeLock) Lock(start, end int64, isRead bool) LockId {
	if start >= end {
		return InvalidLockId
	}

	result := make(chan LockId)
	rl.lockRequests <- lockRequest{rng{start, end}, isRead, result}
	return <-result
}

func (rl *RangeLock) Unlock(lock LockId) {
	if lock == InvalidLockId {
		return
	}

	rl.unlockRequests <- lock
}

func (rl *RangeLock) Close() {
	close(rl.lockRequests)
	close(rl.unlockRequests)

	ack := make(chan bool)
	rl.closeChan <- ack
	<-ack

	close(rl.closeChan)
}
