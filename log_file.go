// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/binary"
	"github.com/dbratus/loghub/trace"
	"os"
	"regexp"
	"runtime"
	"sync/atomic"
)

const defaultEntryBufferSize = 1024
const entryPayloadBase = 16
const timestampLength = 8
const entryHeaderLength = entryPayloadBase + timestampLength
const hopLength = 64

var logFileTrace = trace.New("LogFile")

type LogFile struct {
	writeChan chan *LogEntry
	readChan  chan readLogCmd
	sizeChan  chan chan int64
	closeChan chan chan bool
}

type entryHeader struct {
	payloadLength     int32
	prevPayloadLength int32
	hop               int32
	backHop           int32
	timestamp         int64
}

func (h *entryHeader) NextOffset() int64 {
	return int64(h.payloadLength) + int64(entryPayloadBase)
}

func (h *entryHeader) PrevOffset() int64 {
	return int64(h.prevPayloadLength) + int64(entryPayloadBase)
}

func OpenLogFile(path string, create bool) (*LogFile, error) {
	flags := os.O_RDWR

	if create {
		flags |= os.O_CREATE
	}

	if f, err := os.OpenFile(path, flags, 0660); err == nil {
		log := &LogFile{
			make(chan *LogEntry),
			make(chan readLogCmd),
			make(chan chan int64),
			make(chan chan bool),
		}

		go log.run(f)

		return log, nil
	} else {
		return nil, err
	}
}

func writeUtf8(buf []byte, str string) []byte {
	var int32Buf [4]byte

	binary.BigEndian.PutUint32(int32Buf[:], uint32(len(str)))
	buf = append(buf, int32Buf[:]...)
	buf = append(buf, []byte(str)...)

	return buf
}

func writeEntry(buf []byte, ent *LogEntry, prevPayloadLen int32, backHop int32) ([]byte, int32) {
	for i := 0; i < entryPayloadBase; i++ {
		buf = append(buf, byte(0))
	}

	binary.BigEndian.PutUint32(buf[4:], uint32(prevPayloadLen))
	binary.BigEndian.PutUint32(buf[12:], uint32(backHop))

	var int64Buf [8]byte
	binary.BigEndian.PutUint64(int64Buf[:], uint64(ent.Timestamp))

	buf = append(buf, int64Buf[:]...)
	buf = append(buf, byte(ent.Severity))
	buf = writeUtf8(buf, ent.Source)
	buf = append(buf, byte(ent.Encoding))

	binary.BigEndian.PutUint32(int64Buf[:4], uint32(len(ent.Message)))
	buf = append(buf, int64Buf[:4]...)
	buf = append(buf, ent.Message...)

	payloadLen := int32(len(buf) - entryPayloadBase)
	binary.BigEndian.PutUint32(buf, uint32(payloadLen))

	return buf, payloadLen
}

func writeHop(file *os.File, currentOffset int64, hopStartOffset int64) bool {
	var int32Buf [4]byte

	binary.BigEndian.PutUint32(int32Buf[:], uint32(currentOffset-hopStartOffset))

	if _, err := file.WriteAt(int32Buf[:], hopStartOffset+int64(8)); err != nil {
		logFileTrace.Errorf("Failed to write a hop: %s.", err.Error())
		return false
	}

	return true
}

func readEntryHeader(file *os.File, offset int64) (header entryHeader, ok bool) {
	var headerBuf [entryHeaderLength]byte

	if _, err := file.ReadAt(headerBuf[:], offset); err != nil {
		logFileTrace.Errorf("Failed to read entry header at offset %d: %s.", offset, err.Error())
		ok = false
		return
	}

	header.payloadLength = int32(binary.BigEndian.Uint32(headerBuf[:]))
	header.prevPayloadLength = int32(binary.BigEndian.Uint32(headerBuf[4:]))
	header.hop = int32(binary.BigEndian.Uint32(headerBuf[8:]))
	header.backHop = int32(binary.BigEndian.Uint32(headerBuf[12:]))
	header.timestamp = int64(binary.BigEndian.Uint64(headerBuf[16:]))

	ok = true
	return
}

func readBytes(file *os.File, offset int64) ([]byte, error) {
	var int32Buf [4]byte

	if _, err := file.ReadAt(int32Buf[:], offset); err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(int32Buf[:])
	result := make([]byte, l)

	if _, err := file.ReadAt(result, offset+int64(4)); err != nil {
		return nil, err
	}

	return result, nil
}

func readEntry(file *os.File, offset int64) (ent *LogEntry, ok bool) {
	var int64Buf [8]byte

	ent = new(LogEntry)
	ok = true

	offset += int64(entryPayloadBase)

	if _, err := file.ReadAt(int64Buf[:], offset); err != nil {
		logFileTrace.Errorf("Failed to read log entry timestamp at offset %d: %s.", offset, err.Error())
		ok = false
		return
	}

	ent.Timestamp = int64(binary.BigEndian.Uint64(int64Buf[:]))
	offset += 8

	if _, err := file.ReadAt(int64Buf[:1], offset); err != nil {
		logFileTrace.Errorf("Failed to read log entry severity at offset %d: %s.", offset, err.Error())
		ok = false
		return
	}

	ent.Severity = int(int64Buf[0])
	offset += 1

	if s, err := readBytes(file, offset); err != nil {
		logFileTrace.Errorf("Failed to read log entry source at offset %d: %s.", offset, err.Error())
		ok = false
		return
	} else {
		ent.Source = string(s)
		offset += int64(len(s) + 4)
	}

	if _, err := file.ReadAt(int64Buf[:1], offset); err != nil {
		logFileTrace.Errorf("Failed to read log entry encoding at offset %d: %s.", offset, err.Error())
		ok = false
		return
	}

	ent.Encoding = int(int64Buf[0])
	offset += 1

	if m, err := readBytes(file, offset); err != nil {
		logFileTrace.Errorf("Failed to read log entry message at offset %d: %s.", offset, err.Error())
		ok = false
		return
	} else {
		ent.Message = m
	}

	return
}

func readEntries(file *os.File, cmd readLogCmd, offset int64, lastEntryOffset int64) {
	for {
		header, ok := readEntryHeader(file, offset)

		if !ok {
			close(cmd.entries)
			return
		}

		var srcRegexp *regexp.Regexp = nil
		filterBySource := false

		if cmd.query.Source != "" {
			filterBySource = true

			if re, err := regexp.Compile(cmd.query.Source); err == nil {
				srcRegexp = re
			}
		}

		if header.timestamp >= cmd.query.From && header.timestamp <= cmd.query.To {
			if ent, ok := readEntry(file, offset); ok {
				if (!filterBySource || (srcRegexp != nil && srcRegexp.MatchString(ent.Source))) &&
					ent.Severity >= cmd.query.MinSeverity &&
					ent.Severity <= cmd.query.MaxSeverity {
					cmd.entries <- ent
				}
			} else {
				close(cmd.entries)
				return
			}
		} else {
			close(cmd.entries)
			return
		}

		if offset < lastEntryOffset {
			offset += header.NextOffset()
		} else {
			close(cmd.entries)
			return
		}
	}
}

func findLeftBoundFromLeft(file *os.File, cmd readLogCmd, offset int64, lastEntryOffset int64) {

	//Go right without hops.
	for {
		header, ok := readEntryHeader(file, offset)

		if !ok {
			close(cmd.entries)
			return
		}

		if header.timestamp >= cmd.query.From && header.timestamp <= cmd.query.To {
			readEntries(file, cmd, offset, lastEntryOffset)
			return
		}

		if offset == lastEntryOffset {
			close(cmd.entries)
			return
		}

		if header.NextOffset() == 0 {
			panic("Failed to go forth in the log. Offset of the next entry is unknown.")
		}

		offset += header.NextOffset()
	}
}

func findAndReadEntries(file *os.File, lastEntryOffset int64, cmd readLogCmd, readsCounter *int32) {
	defer atomic.AddInt32(readsCounter, -1)

	offset := lastEntryOffset

	//Go left doing hops.
	for {
		header, ok := readEntryHeader(file, offset)

		if !ok {
			close(cmd.entries)
			return
		}

		if header.timestamp < cmd.query.From {
			findLeftBoundFromLeft(file, cmd, offset, lastEntryOffset)
			return
		}

		if offset == 0 {
			if header.timestamp >= cmd.query.From && header.timestamp <= cmd.query.To {
				readEntries(file, cmd, offset, lastEntryOffset)
				return
			} else {
				close(cmd.entries)
				return
			}
		}

		if header.backHop != 0 {
			offset -= int64(header.backHop)
		} else {
			offset -= header.PrevOffset()
		}
	}
}

func initLogFile(file *os.File) (lastTimestampWritten int64, prevPayloadLen int32, nextHopStartOffset int64, hopCounter int, initialized bool) {
	initialized = true
	lastTimestampWritten = 0
	prevPayloadLen = 0
	nextHopStartOffset = 0
	hopCounter = hopLength

	offset := int64(0)
	var fileSize int64

	if stat, err := file.Stat(); err != nil {
		logFileTrace.Errorf("Failed to get stat of a log file: %s.", err.Error())
		initialized = false
		return
	} else {
		fileSize = stat.Size()
	}

	if fileSize == 0 {
		return
	}

	for {
		header, ok := readEntryHeader(file, offset)
		nextOffset := offset + header.NextOffset()

		//Checking if the last entry is written completely.
		if !ok || nextOffset > fileSize {
			//The file is broken and needs to be fixed.
			//By truncating the file, we remove the last record
			//that was partially written.
			if err := file.Truncate(offset); err != nil {
				logFileTrace.Errorf("Failed to truncate a log file: %s.", err.Error())
				initialized = false
				return
			}

			//Setting the cursor to the end of the file.
			if _, err := file.Seek(0, 2); err != nil {
				logFileTrace.Errorf("Failed to seek to the end of a log file: %s.", err.Error())
				initialized = false
				return
			}

			return
		}

		lastTimestampWritten = header.timestamp
		prevPayloadLen = header.payloadLength

		//Checking the end of the file.
		if nextOffset == fileSize {
			//Setting the cursor to the end of the file.
			if _, err := file.Seek(0, 2); err != nil {
				logFileTrace.Errorf("Failed to seek to the end of a log file: %s.", err.Error())
				initialized = false
			}

			return
		}

		if nextOffset <= offset {
			logFileTrace.Errorf("Next offset must be greater than the current: nextOffset=%d, offset=%d.", nextOffset, offset)
			panic("Next offset must be greater than the current.")
		}

		//Moving next.
		if header.hop > 0 {
			hopOffset := offset + int64(header.hop)
			hopCounter = hopLength

			if hopOffset <= fileSize {
				offset = hopOffset
				nextHopStartOffset = offset
			} else {
				nextHopStartOffset = offset
				offset = nextOffset
			}

		} else {
			offset = nextOffset
			hopCounter--
		}
	}

	return
}

func (log *LogFile) run(file *os.File) {
	buf := make([]byte, 0, defaultEntryBufferSize)

	lastTimestampWritten, prevPayloadLen, nextHopStartOffset, hopCounter, initialized := initLogFile(file)
	currentOffset, _ := file.Seek(0, 1)
	readsCounter := new(int32)

	onWrite := func(ent *LogEntry) {
		if initialized && ent.Timestamp >= lastTimestampWritten {
			backHop := int32(0)

			if hopCounter--; hopCounter == 0 {
				backHop = int32(currentOffset - nextHopStartOffset)
			}

			buf, prevPayloadLen = writeEntry(buf, ent, prevPayloadLen, backHop)

			if _, err := file.Write(buf); err != nil {
				logFileTrace.Errorf("Failed to wrtie log entry: %s.", err.Error())
				currentOffset, _ = file.Seek(0, 1)

			} else {
				if hopCounter == 0 {
					if writeHop(file, currentOffset, nextHopStartOffset) {
						nextHopStartOffset = currentOffset
						hopCounter = hopLength
					}
				}

				currentOffset += int64(len(buf))
			}

			lastTimestampWritten = ent.Timestamp
			buf = buf[:0]
		}
	}

	onRead := func(cmd readLogCmd) {
		if cmd.query.From >= cmd.query.To || cmd.query.MinSeverity > cmd.query.MaxSeverity {
			close(cmd.entries)
		}

		lastEntryOffset := currentOffset - int64(prevPayloadLen+entryPayloadBase)

		atomic.AddInt32(readsCounter, 1)
		go findAndReadEntries(file, lastEntryOffset, cmd, readsCounter)
	}

	onSize := func(sz chan int64) {
		sz <- currentOffset
		close(sz)
	}

	for {
		select {
		case ent, ok := <-log.writeChan:
			if ok {
				onWrite(ent)
			}

		case cmd, ok := <-log.readChan:
			if ok {
				onRead(cmd)
			}
		case sz, ok := <-log.sizeChan:
			if ok {
				onSize(sz)
			}

		case ack := <-log.closeChan:
			for ent := range log.writeChan {
				onWrite(ent)
			}

			for cmd := range log.readChan {
				onRead(cmd)
			}

			for sz := range log.sizeChan {
				onSize(sz)
			}

			for cnt := atomic.LoadInt32(readsCounter); cnt > 0; cnt = atomic.LoadInt32(readsCounter) {
				runtime.Gosched()
			}

			file.Close()
			ack <- true
			close(ack)
			return
		}
	}
}

func (log *LogFile) WriteLog(entry *LogEntry) {
	log.writeChan <- entry
}

func (log *LogFile) ReadLog(q *LogQuery, entries chan *LogEntry) {
	log.readChan <- readLogCmd{q, entries}
}

func (log *LogFile) Size() int64 {
	result := make(chan int64)
	log.sizeChan <- result
	return <-result
}

func (log *LogFile) Close() {
	close(log.writeChan)
	close(log.readChan)
	close(log.sizeChan)

	ack := make(chan bool)
	log.closeChan <- ack
	<-ack

	close(log.closeChan)
}
