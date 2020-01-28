// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protobuf

import (
	"fmt"
	"io"

	"github.com/ethersphere/bee/pkg/p2p"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
)

const delimitedReaderMaxSize = 128 * 1024 // max message size

type Message = proto.Message

func NewWriterAndReader(s p2p.Stream) (w ggio.Writer, r ggio.Reader) {
	r = ggio.NewDelimitedReader(s, delimitedReaderMaxSize)
	w = ggio.NewDelimitedWriter(s)
	return w, r
}

func NewReader(r io.Reader) ggio.Reader {
	return ggio.NewDelimitedReader(r, delimitedReaderMaxSize)
}

func NewWriter(w io.Writer) ggio.Writer {
	return ggio.NewDelimitedWriter(w)
}

func ReadMessages(r io.Reader, newMessage func() Message) (m []Message, err error) {
	pr := NewReader(r)
	for {
		msg := newMessage()
		if err := pr.ReadMsg(msg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		m = append(m, msg)
	}
	return m, nil
}

func Request(w ggio.Writer, r ggio.Reader, req, resp Message) error {
	if err := w.WriteMsg(req); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	if err := r.ReadMsg(resp); err != nil {
		return fmt.Errorf("read message: %w", err)
	}
	return nil
}

func Respond(w ggio.Writer, r ggio.Reader, req Message, f func() (resp Message, err error)) error {
	if err := r.ReadMsg(req); err != nil {
		return fmt.Errorf("read message: %w", err)
	}
	resp, err := f()
	if err != nil {
		return err
	}
	if err := w.WriteMsg(resp); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}
