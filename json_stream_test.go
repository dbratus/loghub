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
	"bytes"
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