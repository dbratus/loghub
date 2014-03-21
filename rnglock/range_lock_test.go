// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package rnglock

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestOverlaps(t *testing.T) {
	if !(&rng{1, 4}).overlaps(rng{2, 3}) {
		t.Error("Overlap 1,4 - 2,3.")
		t.FailNow()
	}

	if !(&rng{2, 3}).overlaps(rng{1, 4}) {
		t.Error("Overlap 2,3 - 1,4.")
		t.FailNow()
	}

	if !(&rng{1, 3}).overlaps(rng{2, 4}) {
		t.Error("Overlap 1,3 - 2,4.")
		t.FailNow()
	}

	if !(&rng{2, 4}).overlaps(rng{1, 3}) {
		t.Error("Overlap 1,3 - 2,4.")
		t.FailNow()
	}

	if (&rng{1, 2}).overlaps(rng{3, 4}) {
		t.Error("Overlap 1,2 - 3,4.")
		t.FailNow()
	}

	if (&rng{3, 4}).overlaps(rng{1, 2}) {
		t.Error("Overlap 3,4 - 1,2.")
		t.FailNow()
	}
}

func rangeLockTest(t *testing.T, l1Read, l2Read, mustCollide bool, iterCnt int) {
	var l1ReadStr, l2ReadStr string

	if l1Read {
		l1ReadStr = "read"
	} else {
		l1ReadStr = "write"
	}

	if l2Read {
		l2ReadStr = "read"
	} else {
		l2ReadStr = "write"
	}

	cnt := new(int32)
	lock := New()
	defer lock.Close()

	for i := 0; i < iterCnt; i++ {
		lockId := lock.Lock(1, 4, l1Read)

		go func() {
			lockId := lock.Lock(2, 3, l2Read)
			defer lock.Unlock(lockId)
			atomic.AddInt32(cnt, 1)
		}()

		<-time.After(time.Millisecond * 10)

		if mustCollide {
			if atomic.LoadInt32(cnt) > int32(i) {
				t.Errorf("Lock %s x %s failed.", l1ReadStr, l2ReadStr)
				t.FailNow()
			}
		} else {
			if atomic.LoadInt32(cnt) == int32(i) {
				t.Errorf("Lock %s x %s failed.", l1ReadStr, l2ReadStr)
				t.FailNow()
			}
		}

		lock.Unlock(lockId)

		<-time.After(time.Millisecond * 10)

		if atomic.LoadInt32(cnt) == int32(i) {
			t.Errorf("Unlock %s x %s failed.", l1ReadStr, l2ReadStr)
			t.FailNow()
		}
	}
}

func TestRangeLocks(t *testing.T) {
	rangeLockTest(t, true, true, false, 10)
	rangeLockTest(t, true, false, true, 10)
	rangeLockTest(t, false, true, true, 10)
	rangeLockTest(t, false, false, true, 10)
}
