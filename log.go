// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"compress/flate"
	"time"
)

const (
	EncodingPlain   = 0
	EncodingDeflate = 1
)

const PlainMessageMaxLength = 512

const maxTimestamp = int64(^uint64(0) ^ 1<<63)
const minTimestamp = -int64(^uint64(0)^1<<63) - 1

const minSeverity = 0
const maxSeverity = 255

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

type LogStat struct {
	Timestamp      int64
	Size           int64
	Limit          int64
	Addr           string
	LastTransferId int64
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

type LogStorage interface {
	Logger

	Close()
	Size() int64
}

type LogManager interface {
	LogStorage

	Truncate(source string, limit int64)

	GetTransferChunk(maxSize int64, entries chan *LogEntry) (id string, size int64, found bool)
	AcceptTransferChunk(id string, entries chan *LogEntry) chan bool
	DeleteTransferChunk(id string)
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

func timeToTimestamp(t time.Time) int64 {
	return t.UnixNano()
}

func timestampToLocalTime(t int64) time.Time {
	return time.Unix(0, t)
}

func timestampToTime(t int64) time.Time {
	return time.Unix(0, t).UTC()
}
