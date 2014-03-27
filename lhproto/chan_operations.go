// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package lhproto

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
