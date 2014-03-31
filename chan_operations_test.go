// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestMergeLogs(t *testing.T) {
	genLog := func(length int, offset int64, mult int64, log chan *LogEntry) {
		for i := 0; i < length; i++ {
			ent := new(LogEntry)
			ent.Timestamp = int64(i)*mult + offset

			log <- ent
		}

		close(log)
	}

	checkMergedSorted := func(merged []*LogEntry) {
		for i := 1; i < len(merged); i++ {
			for j := 0; j < i; j++ {
				if merged[i].Timestamp < merged[j].Timestamp {
					t.Errorf("MergeLogs failed:")
					for idx, ent := range merged {
						t.Errorf("%d:%d", idx, ent.Timestamp)
					}
					t.FailNow()
				}
			}
		}
	}

	in1 := make(chan *LogEntry, 5)
	in2 := make(chan *LogEntry, 5)

	go genLog(5, 0, 2, in1)
	go genLog(5, 1, 2, in2)

	out := make(chan *LogEntry, 10)
	merged := make([]*LogEntry, 0, 10)

	go MergeLogs(in1, in2, out)

	for ent := range out {
		merged = append(merged, ent)
	}

	checkMergedSorted(merged)

	in1 = make(chan *LogEntry, 5)
	in2 = make(chan *LogEntry, 5)

	go genLog(5, 0, 2, in1)
	close(in2)

	out = make(chan *LogEntry, 10)
	merged = make([]*LogEntry, 0, 10)

	go MergeLogs(in1, in2, out)

	checkMergedSorted(merged)

	in1 = make(chan *LogEntry, 5)
	in2 = make(chan *LogEntry, 5)

	close(in1)
	go genLog(5, 1, 2, in2)

	out = make(chan *LogEntry, 10)
	merged = make([]*LogEntry, 0, 10)

	go MergeLogs(in1, in2, out)

	checkMergedSorted(merged)
}

func TestMergeDuplicates(t *testing.T) {
	genLog := func(length int, src string, log chan *LogEntry) {
		for i := 0; i < length; i++ {
			ent := new(LogEntry)
			ent.Timestamp = int64(i)
			ent.Source = src

			log <- ent
		}

		close(log)
	}

	entriesCount := 10
	chan1 := make(chan *LogEntry)
	chan2 := make(chan *LogEntry)
	chan3 := make(chan *LogEntry)

	go genLog(entriesCount, "src1", chan1)
	go genLog(entriesCount, "src1", chan2)
	go genLog(entriesCount, "src2", chan3)

	merged := make(chan *LogEntry)

	go MergeLogs(chan1, chan2, merged)

	result := make(chan *LogEntry)

	go MergeLogs(merged, chan3, result)

	cnt := 0
	for _ = range result {
		cnt++
	}

	if cnt != entriesCount*2 {
		t.Errorf("Duplicates after merge. Expected %d, got %d.", entriesCount*2, cnt)
		t.FailNow()
	}
}
