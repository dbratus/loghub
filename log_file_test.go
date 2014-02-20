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
	"os"
	"time"
	"fmt"
)

func TestWriteReadLogFile(t *testing.T) {
	testLogFilePath := os.TempDir() + "/loghub.test.log"

	var log *LogFile

	if l, err := OpenCreateLogFile(testLogFilePath); err != nil {
		t.Error("OpenCreateLogFile failed.", err)
		t.FailNow()
	} else {
		if l == nil {
			t.Error("Nil returned by OpenCreateLogFile successful call.")
			t.FailNow()
		}

		log = l
	}

	cleanup := func() {
		log.Close()
		os.Remove(testLogFilePath)
	}

	defer cleanup()

	entriesCount := 128

	for i := 0; i < entriesCount; i++ {
		msg := fmt.Sprint("Message", i)
		ent := &LogEntry{ time.Now().Unix(), 1, "Test", EncodingPlain, []byte(msg) }

		log.WriteLog(ent)
	}
}