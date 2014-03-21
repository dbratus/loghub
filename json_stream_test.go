// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"testing"
)

func TestWriteReadJSONStream(t *testing.T) {
	var data bytes.Buffer
	writer := NewJSONStreamWriter(&data)

	header := MessageHeaderJSON{ActionWrite}
	entry := IncomingLogEntryJSON{1, "source", "Message"}

	if err := writer.WriteJSON(&header); err != nil {
		t.Errorf("Failed to write JSON %s.", err.Error())
		t.FailNow()
	}

	for i := 0; i < 3; i++ {
		if err := writer.WriteJSON(&entry); err != nil {
			t.Errorf("Failed to write JSON %s.", err.Error())
			t.FailNow()
		}
	}

	if err := writer.WriteDelimiter(); err != nil {
		t.Errorf("Failed to write JSON %s.", err.Error())
		t.FailNow()
	}

	reader := NewJSONStreamReader(&data)

	if err := reader.ReadJSON(&header); err != nil {
		t.Errorf("Failed to read JSON %s.", err.Error())
		t.FailNow()
	}

	for i := 0; i < 3; i++ {
		if err := reader.ReadJSON(&entry); err != nil {
			t.Errorf("Failed to read JSON %s.", err.Error())
			t.FailNow()
		}
	}

	if err := reader.ReadJSON(&entry); err != ErrStreamDelimiter {
		t.Errorf("Failed to read JSON %s.", err.Error())
		t.FailNow()
	}
}
