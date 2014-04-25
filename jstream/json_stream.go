// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package jstream

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

var ErrStreamDelimiter = errors.New("Stream delimiter encountered")

var streamDelimiter = []byte{byte(0)}

type Reader interface {
	ReadJSON(interface{}) error
}

type Writer interface {
	WriteJSON(interface{}) error
	WriteDelimiter() error
}

type streamReader struct {
	reader *bufio.Reader
}

type streamWriter struct {
	writer io.Writer
}

func NewReader(reader io.Reader) Reader {
	return &streamReader{bufio.NewReader(reader)}
}

func NewWriter(writer io.Writer) Writer {
	return &streamWriter{writer}
}

func (r *streamReader) ReadJSON(target interface{}) error {
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

func (w *streamWriter) WriteJSON(source interface{}) error {
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

func (w *streamWriter) WriteDelimiter() error {
	if _, err := w.writer.Write(streamDelimiter); err != nil {
		return err
	}

	return nil
}
