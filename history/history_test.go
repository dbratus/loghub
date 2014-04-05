// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package history

import (
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Truncate(time.Hour)

	hist.Insert(start)
	hist.Insert(start)
	hist.Insert(start.Add(time.Hour * 2))
	hist.Insert(start.Add(time.Hour * 1))
	hist.Insert(start.Add(time.Hour * 4))
	hist.Insert(start.Add(-time.Hour))
	hist.Insert(start.Add(-time.Hour * 3))

	if !hist.Start().Equal(start.Add(-time.Hour * 3)) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start.Add(-time.Hour * 3))

	if !hist.Start().Equal(start.Add(-time.Hour)) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start.Add(-time.Hour))

	if !hist.Start().Equal(start) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start)

	if !hist.Start().Equal(start.Add(time.Hour * 1)) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start.Add(time.Hour * 1))

	if !hist.Start().Equal(start.Add(time.Hour * 2)) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start.Add(time.Hour * 2))

	if !hist.Start().Equal(start.Add(time.Hour * 4)) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	hist.Truncate(start.Add(time.Hour * 4))

	if !hist.IsEmpty() {
		t.Error("History must be empty.")
		t.FailNow()
	}
}

func TestDelete(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Truncate(time.Hour)

	hist.Insert(start)
	hist.Insert(start.Add(time.Hour * 2))

	hist.Delete(start)

	if !hist.Start().Equal(start.Add(time.Hour * 2)) {
		t.Error("Invalid start after delete.")
		t.FailNow()
	}

	hist.Insert(start.Add(time.Hour * 3))
	hist.Insert(start.Add(time.Hour * 4))

	hist.Delete(start.Add(time.Hour * 3))
	hist.Truncate(start.Add(time.Hour * 2))

	if !hist.Start().Equal(start.Add(time.Hour * 4)) {
		t.Error("Invalid start after delete in the middle.")
		t.FailNow()
	}
}

func TestTruncate(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Truncate(time.Hour)
	pointCnt := 10

	for i := 0; i < pointCnt; i++ {
		hist.Insert(start.Add(time.Hour * time.Duration(i)))
	}

	limit := start.Add(time.Hour * time.Duration(pointCnt/2))
	cnt := 0

	for !hist.IsEmpty() && (hist.Start().Before(limit) || hist.Start().Equal(limit)) {
		hist.Truncate(hist.Start())
		cnt++
	}

	if cnt < pointCnt/2 {
		t.Errorf("Expected %d truncations, got %d.", pointCnt/2, cnt)
		t.FailNow()
	}
}

func TestRounding(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Truncate(time.Hour)
	point := start

	for i := 0; i < 60; i++ {
		hist.Insert(point)
		point = point.Add(time.Minute)
	}

	if !hist.Start().Add(time.Hour).Equal(hist.End()) {
		t.Error("End must be greater than start by 1 hour.")
		t.FailNow()
	}

	if !hist.Start().Equal(start) {
		t.Error("Invalid start of the history.")
		t.FailNow()
	}
}

func printHistRelTo(hist *History, start time.Time) {
	for cur := hist.tail; cur != nil; cur = cur.next {
		rstart := int(cur.start.Sub(start).Hours())
		rend := int(cur.end.Sub(start).Hours())

		println(rstart, "-", rend)
	}
}
