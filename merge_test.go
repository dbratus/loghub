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
