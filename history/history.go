// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package history

import (
	"time"
)

type History struct {
	unitOfMeasure time.Duration
	tail          *segment
	head          *segment
}

type segment struct {
	start time.Time
	end   time.Time
	next  *segment
}

func New(unitOfMeasure time.Duration) *History {
	return &History{unitOfMeasure, nil, nil}
}

func (h *History) Append(point time.Time) {
	rpoint := point.Truncate(h.unitOfMeasure)

	if h.head == nil {
		h.head = &segment{rpoint, rpoint.Add(h.unitOfMeasure), nil}
		h.tail = h.head

	} else {
		diff := rpoint.Sub(h.head.end)

		if diff > 0 {
			newHead := &segment{rpoint, rpoint.Add(h.unitOfMeasure), nil}
			h.head.next = newHead
			h.head = newHead

		} else if diff == 0 {
			h.head.end = rpoint.Add(h.unitOfMeasure)
		}
	}
}

func (h *History) Start() time.Time {
	if h.tail == nil {
		panic("History is empty.")
	}

	return h.tail.start
}

func (h *History) End() time.Time {
	if h.head == nil {
		panic("History is empty.")
	}

	return h.head.end
}

func (h *History) IsEmpty() bool {
	return h.head == nil
}

func (h *History) Delete(ts time.Time) {
	if h.IsEmpty() {
		return
	}

	gapStart := ts.Truncate(h.unitOfMeasure)
	cur := h.tail
	var prev *segment = nil

	for cur != nil {
		if gapStart.Equal(cur.start) || (gapStart.After(cur.start) && gapStart.Before(cur.end)) {
			break
		}

		prev = cur
		cur = cur.next
	}

	if cur != nil {
		if gapStart.Equal(cur.start) {
			//Deleting a point at the start of a segment.

			newStart := gapStart.Add(h.unitOfMeasure)

			if newStart.Before(cur.end) {
				//If the segment is longer than a single point,
				//updating its start.

				cur.start = newStart
			} else {
				//Otherwise, removing whole segment.

				if prev != nil {
					prev.next = cur.next
				}

				if h.tail == cur {
					h.tail = cur.next
				}

				if h.head == cur {
					h.head = prev
				}
			}
		} else {
			gapEnd := gapStart.Add(h.unitOfMeasure)

			if gapEnd.Equal(cur.end) {
				//Deleting the point at the end of a segment.

				cur.end = gapStart
			} else {
				//If the point is in the middle of a segment,
				//the segment needs to be split.

				newSeg := &segment{gapEnd, cur.end, cur.next}
				cur.end = gapStart
				cur.next = newSeg

				if h.head == cur {
					h.head = newSeg
				}
			}
		}
	}
}

func (h *History) Truncate(limit time.Time) {
	rlimit := limit.Truncate(h.unitOfMeasure).Add(h.unitOfMeasure)
	cur := h.tail

	for cur != nil {
		if cur.start.After(rlimit) || cur.start.Equal(rlimit) {
			break
		}

		if rlimit.After(cur.start) && rlimit.Before(cur.end) {
			cur.start = rlimit
			break
		}

		cur = cur.next
	}

	h.tail = cur

	if cur == nil {
		h.head = nil
	}
}
