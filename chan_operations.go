// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
)

func MergeLogs(leftIn chan *LogEntry, rightIn chan *LogEntry, out chan *LogEntry) {
	var left, right, lastYield *LogEntry = nil, nil, nil

	yield := func(ent *LogEntry) {
		if lastYield == nil ||
			lastYield.Timestamp != ent.Timestamp ||
			lastYield.Source != ent.Source ||
			bytes.Compare(lastYield.Message, ent.Message) != 0 {

			out <- ent
			lastYield = ent
		}
	}

	for {
		if left == nil {
			if l, ok := <-leftIn; !ok {
				if right != nil {
					yield(right)
				}

				for ent := range rightIn {
					yield(ent)
				}

				close(out)
				break
			} else {
				left = l
			}
		}

		if right == nil {
			if r, ok := <-rightIn; !ok {
				if left != nil {
					yield(left)
				}

				for ent := range leftIn {
					yield(ent)
				}

				close(out)
				break
			} else {
				right = r
			}
		}

		if left.Timestamp < right.Timestamp {
			yield(left)
			left = nil
		} else {
			yield(right)
			right = nil
		}
	}
}

func ForwardLog(from chan *LogEntry, to chan *LogEntry) {
	for ent := range from {
		to <- ent
	}

	close(to)
}

func PurgeLog(log chan *LogEntry) {
	for _ = range log {
	}
}
