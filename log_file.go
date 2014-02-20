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
	"os"
	"encoding/binary"
)

const defaultEntryBufferSize = 1024
const entryPayloadBase = 16
const hopLength = 64

type logQuery struct {
	from        int64
	to     		int64
	minSeverity int
	maxSeverity int
	source      string
	result      chan *LogEntry
}

type logClose struct {
	ack chan bool
}

type LogFile struct {
	writeChan chan *LogEntry
	readChan  chan *logQuery
	closeChan chan *logClose
}

func OpenCreateLogFile(path string) (*LogFile, error) {
	if f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE, 0660); err == nil {
		log := &LogFile{ make(chan *LogEntry), make(chan *logQuery), make(chan *logClose) }

		go log.run(f)

		return log, nil
	} else {
		return nil, err
	}
}

func writeUtf8(buf []byte, str string) []byte {
	var varintBuf [binary.MaxVarintLen64]byte

	l := binary.PutUvarint(varintBuf[:], uint64(len(str)))
	buf = append(buf, varintBuf[:l]...)
	buf = append(buf, []byte(str)...)

	return buf
}

func writeEntry(buf []byte, ent *LogEntry, prevPayloadLen int, backHop int64) ([]byte, int) {
	for i := 0; i < entryPayloadBase; i++ {
		buf = append(buf, byte(0))
	}

	binary.BigEndian.PutUint32(buf[4:], uint32(prevPayloadLen))

	if backHop != 0 {
		binary.BigEndian.PutUint32(buf[12:], uint32(backHop))
	}

	var int64Buf [8]byte
	binary.BigEndian.PutUint64(int64Buf[:], uint64(ent.Timestamp))

	buf = append(buf, int64Buf[:]...)
	buf = append(buf, byte(ent.Severity))
	buf = writeUtf8(buf, ent.Source)
	buf = append(buf, byte(ent.Encoding))

	binary.BigEndian.PutUint64(int64Buf[:], uint64(len(ent.Message)))
	buf = append(buf, int64Buf[:]...)

	buf = append(buf, ent.Message...)

	payloadLen := len(buf) - entryPayloadBase
	binary.BigEndian.PutUint32(buf, uint32(payloadLen))

	return buf, payloadLen
}

func writeHop(file *os.File, hopStartOffset int64) bool {
	var int64Buf [8]byte
	binary.BigEndian.PutUint64(int64Buf[:], uint64(hopStartOffset))

	if _, err := file.WriteAt(int64Buf[:], hopStartOffset + int64(8)); err != nil {
		println("Failed to write a hop:", err)
		return false
	}

	return true
}

func readEntryHeader(file *os.File, offset int64) (ts int64, payloadLen int, nextOffset int64, ok bool) {
	var int64Buf [8]byte

	if n, _ := file.ReadAt(int64Buf[:], offset); n < 8 {
		ts = int64(0)
		nextOffset = offset + int64(n)
		payloadLen = 0
		ok = false
		return
	}

	payloadLen = int(binary.BigEndian.Uint32(int64Buf[:]))
	nextOffset = offset + int64(entryPayloadBase) + int64(payloadLen)

	if n, _ := file.ReadAt(int64Buf[:], offset + int64(entryPayloadBase)); n < 8 {
		ts = int64(0)
		ok = false
		return
	}

	ts = int64(binary.BigEndian.Uint64(int64Buf[:]))
	ok = true
	return
}

func initLogger(file *os.File) (lastTimestampWritten int64, prevPayloadLen int, initialized bool) {
	initialized = true
	lastTimestampWritten = 0
	prevPayloadLen = 0

	offset := int64(0)
	var fileSize int64

	if stat, err := file.Stat(); err != nil {
		println("Failed to get stat of a log file:", err)
		initialized = false
		return
	} else {
		fileSize = stat.Size()
	}
	
	for {
		ts, payloadLen, nextOffset, ok := readEntryHeader(file, offset)

		if !ok || nextOffset > fileSize {
			//The file is broken and needs to be fixed.
			//By truncating the file, we remove the last record
			//that was partially written.
			if err := file.Truncate(offset); err != nil {
				println("Failed to truncate a log file:", err)
				initialized = false
				return
			}

			if _, err := file.Seek(0, 2); err != nil {
				println("Failed to seek to the end of a log file:", err)
				initialized = false
				return
			}

			return
		} else if nextOffset == fileSize {
			return
		}

		if nextOffset <= offset {
			println("Next offset must be greater than the current: nextOffset =", nextOffset, "offset =", offset)
			panic("Next offset must be greater than the current.")
		}

		offset = nextOffset
		lastTimestampWritten = ts
		prevPayloadLen = payloadLen
	}

	return
}

func (log *LogFile) run(file *os.File) {
	buf := make([]byte, 0, defaultEntryBufferSize)

	stop := false
	lastTimestampWritten, prevPayloadLen, initialized := initLogger(file)
	hopCounter := hopLength
	nextHopStartOffset := int64(0) //TODO: Get from init.

	for !stop {
		select {
			case ent := <- log.writeChan:
				if initialized {
					isAppending := ent.Timestamp >= lastTimestampWritten
					backHop := int64(0)
					
					if hopCounter--; hopCounter == 0 && isAppending {
						backHop = nextHopStartOffset
					}

					buf, prevPayloadLen = writeEntry(buf, ent, prevPayloadLen, backHop)

					if isAppending {
						currentOffset, _ := file.Seek(0, 1)

						if _, err := file.Write(buf); err != nil {
							println("Failed to wrtie log entry:", err)
						
						} else {
							if hopCounter == 0 {
								if writeHop(file, nextHopStartOffset) {
									nextHopStartOffset = currentOffset
									hopCounter = hopLength
								}
							}
						}

						lastTimestampWritten = ent.Timestamp

					} else {
						// TODO: Shift the tail, insert the record.
					}

					buf = buf[:0]
				}

			//case q := <- log.readChan:

			case cmd := <- log.closeChan:
				file.Close()
				cmd.ack <- true
				stop = true
		}
	}
}

func (log *LogFile) WriteLog(entry *LogEntry) {
	log.writeChan <- entry
}

func (log *LogFile) ReadLog(from int64, to int64, minSeverity int, maxSeverity int, source string) chan *LogEntry {
	result := make(chan *LogEntry)
	q := &logQuery { from, to, minSeverity, maxSeverity, source, result}

	log.readChan <- q

	return result
}

func (log *LogFile) Close() {
	ack := make(chan bool)
	closeCmd := &logClose { ack }

	log.closeChan <- closeCmd
	<- ack

	close(log.writeChan)
	close(log.readChan)
	close(log.closeChan)
}