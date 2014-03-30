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
	rounded := point.Truncate(h.unitOfMeasure)

	if h.head == nil {
		h.head = &segment{rounded, rounded, nil}
		h.tail = h.head
	} else {
		diff := rounded.Sub(h.head.end)

		if diff > h.unitOfMeasure {
			newHead := &segment{rounded, rounded, nil}
			h.head.next = newHead
			h.head = newHead
		} else if diff > 0 {
			h.head.end = rounded
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

	rounded := ts.Truncate(h.unitOfMeasure)
	cur := h.tail
	var prev *segment = nil

	for cur != nil {
		if rounded.Equal(cur.start) {
			break
		}

		prev = cur
		cur = cur.next
	}

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

func (h *History) Truncate(limit time.Time) {
	rounded := limit.Truncate(h.unitOfMeasure).Add(h.unitOfMeasure)
	cur := h.tail

	for cur != nil {
		if rounded.Equal(cur.start) {
			break
		}

		if rounded.After(cur.start) && (rounded.Before(cur.end) || rounded.Equal(cur.end)) {
			cur.start = rounded
			break
		}

		cur = cur.next
	}

	h.tail = cur

	if cur == nil {
		h.head = nil
	}
}
