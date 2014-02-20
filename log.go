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

const (
	EncodingPlain   = 0
	EncodingDeflate = 1
)

type LogEntry struct {
	Timestamp int64
	Severity  int
	Source    string
	Encoding  int
	Message   []byte
}

type Logger interface {
	WriteLog(*LogEntry)
	ReadLog(from int64, to int64, minSeverity int, maxSeverity int, source string) chan *LogEntry
}
