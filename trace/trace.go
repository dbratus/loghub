// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package trace

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

const debugChanBacklogLen = 100

var (
	traceLevel *int32
	criticals  chan *traceEntry
	debug      chan *traceEntry
)

var severityNames = map[int]string{
	LevelError: "E",
	LevelWarn:  "W",
	LevelInfo:  "I",
	LevelDebug: "D",
}

type Tracer interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}

type traceEntry struct {
	timestamp time.Time
	severity  int
	source    string
	message   string
}

type trace struct {
	source string
}

func init() {
	traceLevel = new(int32)
	*traceLevel = LevelInfo

	criticals = make(chan *traceEntry)
	debug = make(chan *traceEntry, debugChanBacklogLen)

	go run()
}

func run() {
	for {
		select {
		case ent := <-criticals:
			write(ent)
		case ent := <-debug:
			write(ent)
		}
	}
}

func write(ent *traceEntry) {
	ts := ent.timestamp.Format("2006-01-02 15:04:05.999")

	fmt.Fprintf(os.Stderr, "%s [%s] %s: %s\n", ts, ent.source, severityNames[ent.severity], ent.message)
}

func SetTraceLevel(level int) {
	atomic.StoreInt32(traceLevel, int32(level))
}

func GetTraceLevel() int {
	return int(atomic.LoadInt32(traceLevel))
}

func New(source string) Tracer {
	return &trace{source}
}

func (t *trace) Debug(a ...interface{}) {
	if GetTraceLevel() >= LevelDebug {
		debug <- &traceEntry{time.Now(), LevelDebug, t.source, fmt.Sprint(a...)}
	}
}

func (t *trace) Debugf(format string, a ...interface{}) {
	if GetTraceLevel() >= LevelDebug {
		debug <- &traceEntry{time.Now(), LevelDebug, t.source, fmt.Sprintf(format, a...)}
	}
}

func (t *trace) Info(a ...interface{}) {
	if GetTraceLevel() >= LevelInfo {
		criticals <- &traceEntry{time.Now(), LevelInfo, t.source, fmt.Sprint(a...)}
	}
}

func (t *trace) Infof(format string, a ...interface{}) {
	if GetTraceLevel() >= LevelInfo {
		criticals <- &traceEntry{time.Now(), LevelInfo, t.source, fmt.Sprintf(format, a...)}
	}
}

func (t *trace) Warn(a ...interface{}) {
	if GetTraceLevel() >= LevelWarn {
		criticals <- &traceEntry{time.Now(), LevelWarn, t.source, fmt.Sprint(a...)}
	}
}

func (t *trace) Warnf(format string, a ...interface{}) {
	if GetTraceLevel() >= LevelWarn {
		criticals <- &traceEntry{time.Now(), LevelWarn, t.source, fmt.Sprintf(format, a...)}
	}
}

func (t *trace) Error(a ...interface{}) {
	if GetTraceLevel() >= LevelError {
		criticals <- &traceEntry{time.Now(), LevelError, t.source, fmt.Sprint(a...)}
	}
}

func (t *trace) Errorf(format string, a ...interface{}) {
	if GetTraceLevel() >= LevelError {
		criticals <- &traceEntry{time.Now(), LevelError, t.source, fmt.Sprintf(format, a...)}
	}
}
