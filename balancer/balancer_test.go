// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package balancer

import (
	"testing"
)

func TestBalancer(t *testing.T) {
	blc := New()

	var hosts = map[string]int64{
		"h1": 1,
		"h2": 2,
		"h3": 3,
		"h4": 11,
		"h5": 12,
		"h6": 13,
	}

	for h, sz := range hosts {
		blc.UpdateHost(h, sz, 10)
	}

	transfers := blc.MakeTransfers()
	if len(transfers) < 3 {
		t.Errorf("Not enough transfers. Expected %d, got %d.", 3, len(transfers))
		t.FailNow()
	}

	for h, sz := range hosts {
		blc.UpdateHost(h, sz, 10)
	}

	transfersTwice := blc.MakeTransfers()
	if len(transfersTwice) > 0 {
		t.Errorf("Too many transfers. Expected %d, got %d.", 0, len(transfersTwice))
		t.FailNow()
	}

	for _, tr := range transfers {
		hosts[tr.To] += tr.Amount
		hosts[tr.From] -= tr.Amount
	}

	for h, sz := range hosts {
		blc.UpdateHost(h, sz, 10)
	}

	transfersTwice = blc.MakeTransfers()

	if len(transfersTwice) > 0 {
		t.Errorf("Too many transfers. Expected %d, got %d.", 0, len(transfers))
		t.FailNow()
	}

	for _, t := range transfers {
		blc.TransferComplete(t.Id)
	}

	transfers = blc.MakeTransfers()

	if len(transfers) > 0 {
		t.Errorf("Too many transfers. Expected %d, got %d.", 0, len(transfers))
		t.FailNow()
	}
}
