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
	"io"
	"errors"
	"bufio"
	"encoding/json"
)

var ErrStreamDelimiter = errors.New("Stream delimiter encountered")

var jsonStreamDelimiter = [...]byte{byte(0)}

type JSONStreamReader interface {
	ReadJSON(interface{}) error
}

type JSONStreamWriter interface {
	WriteJSON(interface{}) error
	WriteDelimiter() error
}

type jsonStreamReader struct {
	reader *bufio.Reader
}

type jsonStreamWriter struct {
	writer io.Writer
}

func NewJSONStreamReader(reader io.Reader) JSONStreamReader {
	return &jsonStreamReader{bufio.NewReader(reader)}
}

func NewJSONStreamWriter(writer io.Writer) JSONStreamWriter {
	return &jsonStreamWriter{writer}
}

func (r *jsonStreamReader) ReadJSON(target interface{}) error {
	if bytes, err := r.reader.ReadBytes(byte(0)); err == nil {
		if len(bytes) == 1 { //The buffer contains only the delimiter.
			return ErrStreamDelimiter
		}

		if err = json.Unmarshal(bytes[:len(bytes)-1], target); err != nil {
			return err
		}

		return nil
	} else {
		return err
	}
}

func (w *jsonStreamWriter) WriteJSON(source interface{}) error {
	if bytes, err := json.Marshal(source); err == nil {
		if _, err = w.writer.Write(bytes); err != nil {
			return err
		}

		if err = w.WriteDelimiter(); err != nil {
			return err
		}

		return nil
	} else {
		return err
	}
}

func (w *jsonStreamWriter) WriteDelimiter() error {
	if _, err := w.writer.Write(jsonStreamDelimiter[:]); err != nil {
		return err
	}

	return nil
}