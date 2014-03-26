// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

func MergeLogs(leftIn chan *LogEntry, rightIn chan *LogEntry, out chan *LogEntry) {
	var left, right *LogEntry = nil, nil

	for {
		if left == nil {
			if l, ok := <-leftIn; !ok {
				if right != nil {
					out <- right
				}

				for ent := range rightIn {
					out <- ent
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
					out <- left
				}

				for ent := range leftIn {
					out <- ent
				}

				close(out)
				break
			} else {
				right = r
			}
		}

		if left.Timestamp < right.Timestamp {
			out <- left
			left = nil
		} else {
			out <- right
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

func PurgeIncomingLogEntryJSON(entries chan *IncomingLogEntryJSON) {
	for _ = range entries {
	}
}

func PurgeLogQueryJSON(queries chan *LogQueryJSON) {
	for _ = range queries {
	}
}

func PurgeOutgoingLogEntryJSON(entries chan *OutgoingLogEntryJSON) {
	for _ = range entries {
	}
}

func PurgeInternalLogEntryJSON(entries chan *InternalLogEntryJSON) {
	for _ = range entries {
	}
}
