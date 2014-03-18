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
