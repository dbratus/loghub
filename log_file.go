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
const timestampLength = 8
const entryHeaderLength = entryPayloadBase + timestampLength
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

	binary.BigEndian.PutUint64(int64Buf[:], uint64(len(ent.Message)))
	buf = append(buf, int64Buf[:]...)

	buf = append(buf, ent.Message...)

	payloadLen := int32(len(buf) - entryPayloadBase)
	binary.BigEndian.PutUint32(buf, uint32(payloadLen))

	return buf, payloadLen
}

func writeHop(file *os.File, currentOffset int64, hopStartOffset int64) bool {
	var int32Buf [4]byte

	binary.BigEndian.PutUint32(int32Buf[:], uint32(currentOffset - hopStartOffset))

	if _, err := file.WriteAt(int32Buf[:], hopStartOffset + int64(8)); err != nil {
		println("Failed to write a hop:", err)
		return false
	}

	return true
}

func readEntryHeader(file *os.File, offset int64) (header entryHeader, ok bool) {
	var headerBuf [entryHeaderLength]byte

	if n, _ := file.ReadAt(headerBuf[:], offset); n < entryHeaderLength {
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

func initLogger(file *os.File) (lastTimestampWritten int64, prevPayloadLen int32, nextHopStartOffset int64, hopCounter int, initialized bool) {
	initialized = true
	lastTimestampWritten = 0
	prevPayloadLen = 0
	nextHopStartOffset = 0
	hopCounter = hopLength

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
		header, ok := readEntryHeader(file, offset)
		nextOffset := offset + header.NextOffset()

		if !ok || nextOffset > fileSize {
			//The file is broken and needs to be fixed.
			//By truncating the file, we remove the last record
			//that was partially written.
			if err := file.Truncate(offset); err != nil {
				println("Failed to truncate a log file:", err)
				initialized = false
				return
			}

			//Setting the cursor to the end of the file.
			if _, err := file.Seek(0, 2); err != nil {
				println("Failed to seek to the end of a log file:", err)
				initialized = false
				return
			}

			return
		} else if nextOffset == fileSize {
			//Setting the cursor to the end of the file.
			if _, err := file.Seek(0, 2); err != nil {
				println("Failed to seek to the end of a log file:", err)
				initialized = false
			}

			return
		}

		if nextOffset <= offset {
			println("Next offset must be greater than the current: nextOffset =", nextOffset, "offset =", offset)
			panic("Next offset must be greater than the current.")
		}

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
		
		lastTimestampWritten = header.timestamp
		prevPayloadLen = header.payloadLength
	}

	return
}

func (log *LogFile) run(file *os.File) {
	buf := make([]byte, 0, defaultEntryBufferSize)

	stop := false
	lastTimestampWritten, prevPayloadLen, nextHopStartOffset, hopCounter, initialized := initLogger(file)
	currentOffset, _ := file.Seek(0, 1)

	for !stop {
		select {
			case ent := <- log.writeChan:
				if initialized && ent.Timestamp >= lastTimestampWritten {
					backHop := int32(0)

					if hopCounter--; hopCounter == 0 {
						backHop = int32(currentOffset - nextHopStartOffset)
					}

					buf, prevPayloadLen = writeEntry(buf, ent, prevPayloadLen, backHop)

					if _, err := file.Write(buf); err != nil {
						println("Failed to wrtie log entry:", err)
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