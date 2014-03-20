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
	"bytes"
	"compress/flate"
)

const (
	EncodingPlain   = 0
	EncodingDeflate = 1
)

const PlainMessageMaxLength = 512

type LogEntry struct {
	Timestamp int64
	Severity  int
	Source    string
	Encoding  int
	Message   []byte
}

type LogQuery struct {
	From        int64
	To          int64
	MinSeverity int
	MaxSeverity int
	Source      string
}

type LogWriter interface {
	WriteLog(*LogEntry)
}

type LogReader interface {
	ReadLog(*LogQuery, chan *LogEntry)
}

type Logger interface {
	LogReader
	LogWriter
}

type LogManager interface {
	Logger

	Close()
	Size() int64
}

func EncodeMessage(msg string) ([]byte, int) {
	if len(msg) > PlainMessageMaxLength {
		buf := new(bytes.Buffer)

		if compressor, err := flate.NewWriter(buf, flate.DefaultCompression); err == nil {
			compressor.Write([]byte(msg))
			compressor.Flush()

			return buf.Bytes(), EncodingDeflate
		} else {
			return []byte(msg), EncodingPlain
		}
	} else {
		return []byte(msg), EncodingPlain
	}
}

func DecodeMessage(msg []byte, encoding int) string {
	if encoding == EncodingDeflate {
		reader := flate.NewReader(bytes.NewBuffer(msg))
		buf := make([]byte, PlainMessageMaxLength)
		result := new(bytes.Buffer)

		for {
			n, _ := reader.Read(buf)

			result.Write(buf[:n])

			if n < len(buf) {
				return string(result.Bytes())
			}
		}
	} else {
		return string(msg)
	}
}
