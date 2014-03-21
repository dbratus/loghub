// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
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
