// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package history

import (
	"testing"
	"time"
)

func TestHistory(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Round(time.Hour)
	point := start

	for i := 0; i < 3; i++ {
		hist.Append(point)
		point = point.Add(time.Hour)
	}

	point = point.Add(time.Hour)

	for i := 0; i < 3; i++ {
		hist.Append(point)
		point = point.Add(time.Hour)
	}

	if hist.IsEmpty() {
		t.Error("The history must not be empty.")
		t.FailNow()
	}

	if !hist.Start().Equal(start) {
		t.Error("Invalid start.")
		t.FailNow()
	}

	if !hist.End().Equal(start.Add(time.Hour * 6)) {
		t.Error("Invalid end.")
		t.FailNow()
	}

	limit := start.Add(time.Hour * 3)
	expectedNewStart := start.Add(time.Hour * 4)

	hist.Truncate(limit)

	if !hist.Start().Equal(expectedNewStart) {
		t.Error("Invalid start after truncation.")
		t.FailNow()
	}

	hist.Truncate(start.Add(time.Hour * 6))

	if !hist.IsEmpty() {
		t.Error("The history must be empty.")
		t.FailNow()
	}
}

func TestRounding(t *testing.T) {
	hist := New(time.Hour)

	start := time.Now().Round(time.Hour)
	point := start

	for i := 0; i < 60; i++ {
		hist.Append(point)
		point = point.Add(time.Minute)
	}

	if !hist.Start().Equal(hist.End()) {
		t.Error("Start and end must be equal.")
		t.FailNow()
	}

	if !hist.Start().Equal(start) {
		t.Error("Invalid start of the history.")
		t.FailNow()
	}
}
