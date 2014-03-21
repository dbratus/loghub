// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func TestEncodeDecodeMessage(t *testing.T) {
	shortMessage := "Short message"

	encodedMessage, encoding := EncodeMessage(shortMessage)

	if encoding != EncodingPlain {
		t.Error("Failed to encode short message. Invalid encoding.")
		t.FailNow()
	}

	decodedMessage := DecodeMessage(encodedMessage, encoding)

	if decodedMessage != shortMessage {
		t.Error("Failed to decode short message.")
		t.FailNow()
	}

	longMessage := strings.Repeat("Long message", PlainMessageMaxLength)

	encodedMessage, encoding = EncodeMessage(longMessage)

	if encoding != EncodingDeflate {
		t.Error("Failed to encode long message. Invalid encoding.")
		t.FailNow()
	}

	decodedMessage = DecodeMessage(encodedMessage, encoding)

	if decodedMessage != longMessage {
		t.Error("Failed to decode long message.")
		t.FailNow()
	}
}

func TestLogEntryMapping(t *testing.T) {
	sev := 1
	src := "Source"
	msg := strings.Repeat("Message", PlainMessageMaxLength)
	ts := int64(10000)

	incomingLogEntry := &IncomingLogEntryJSON{sev, src, msg}
	logEntry := IncomingLogEntryJSONToLogEntry(incomingLogEntry)
	logEntry.Timestamp = ts
	outgoingLogEntry := LogEntryToOutgoingLogEntryJSON(logEntry)

	enc := logEntry.Encoding

	if outgoingLogEntry.Sev != sev {
		t.Error("Severity is not mapped.")
		t.FailNow()
	}

	if outgoingLogEntry.Src != src {
		t.Error("Source is not mapped.")
		t.FailNow()
	}

	if outgoingLogEntry.Ts != ts {
		t.Error("Timestamp is not mapped.")
		t.FailNow()
	}

	if outgoingLogEntry.Msg != msg {
		t.Error("Message is not mapped.")
		t.FailNow()
	}

	internalLogEntry := LogEntryToInternalLogEntryJSON(logEntry)

	if internalLogEntry.Sev != sev {
		t.Error("Severity is not mapped.")
		t.FailNow()
	}

	if internalLogEntry.Src != src {
		t.Error("Source is not mapped.")
		t.FailNow()
	}

	if internalLogEntry.Ts != ts {
		t.Error("Timestamp is not mapped.")
		t.FailNow()
	}

	if internalLogEntry.Enc != enc {
		t.Error("Encoding is not mapped.")
		t.FailNow()
	}

	logEntry = InternalLogEntryJSONToLogEntry(internalLogEntry)

	if logEntry.Severity != sev {
		t.Error("Severity is not mapped.")
		t.FailNow()
	}

	if logEntry.Source != src {
		t.Error("Source is not mapped.")
		t.FailNow()
	}

	if logEntry.Timestamp != ts {
		t.Error("Timestamp is not mapped.")
		t.FailNow()
	}

	if DecodeMessage(logEntry.Message, enc) != msg {
		t.Error("Message is not mapped.")
		t.FailNow()
	}
}

func TestLogQueryMapping(t *testing.T) {
	logQueryJSON := LogQueryJSON{1, 2, 3, 4, "Source"}
	logQuery := LogQueryJSONToLogQuery(&logQueryJSON)

	if logQuery.From != logQueryJSON.From {
		t.Error("From is not mapped.")
		t.FailNow()
	}

	if logQuery.To != logQueryJSON.To {
		t.Error("To is not mapped.")
		t.FailNow()
	}

	if logQuery.MinSeverity != logQueryJSON.MinSev {
		t.Error("MinSeverity is not mapped.")
		t.FailNow()
	}

	if logQuery.MaxSeverity != logQueryJSON.MaxSev {
		t.Error("MaxSeverity is not mapped.")
		t.FailNow()
	}

	if logQuery.Source != logQueryJSON.Src {
		t.Error("Source is not mapped.")
		t.FailNow()
	}
}
