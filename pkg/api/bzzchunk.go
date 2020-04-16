// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ethersphere/bee/pkg/httputil"
	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/gorilla/mux"
)

// swagger:parameters chunkUpload chunkGet
type chunkPath struct {
	// Hex-encoded peer overlay address
	// in: path
	Addr swarm.Address `json:"addr"`
}

// swagger:route POST /bzz-chunk/{addr} bzz-chunk chunkUpload
//
// Upload chunk
//
// Upload chunk data with a given address by sending raw data.
//
// Consumes:
// - application/octet-stream
//
// Produces:
// - application/json
//
// Responses:
//   200: statusResponse
//   default: statusResponse
func (s *server) chunkUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var path chunkPath
	if err := httputil.UnmarshalMuxVars(mux.Vars(r), &path); err != nil {
		s.Logger.Debugf("bzz-chunk: path params: %v", err)
		jsonhttp.NotFound(w, nil)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Logger.Debugf("bzz-chunk: read chunk data error: %v, addr %s", err, path.Addr)
		s.Logger.Error("bzz-chunk: read chunk data error")
		jsonhttp.InternalServerError(w, "cannot read chunk data")
		return

	}

	_, err = s.Storer.Put(ctx, storage.ModePutUpload, swarm.NewChunk(path.Addr, data))
	if err != nil {
		s.Logger.Debugf("bzz-chunk: chunk write error: %v, addr %s", err, path.Addr)
		s.Logger.Error("bzz-chunk: chunk write error")
		jsonhttp.BadRequest(w, "chunk write error")
		return
	}

	jsonhttp.OK(w, nil)
}

// swagger:route GET /bzz-chunk/{addr} bzz-chunk chunkGet
//
// Download chunk
//
// Download chunk data with given address.
//
// Produces:
// - application/octet-stream
//
// Responses:
//   200:
//   default: statusResponse
func (s *server) chunkGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var path chunkPath
	if err := httputil.UnmarshalMuxVars(mux.Vars(r), &path); err != nil {
		s.Logger.Debugf("bzz-chunk: path params: %v", err)
		jsonhttp.NotFound(w, nil)
		return
	}

	chunk, err := s.Storer.Get(ctx, storage.ModeGetRequest, path.Addr)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			s.Logger.Trace("bzz-chunk: chunk not found. addr %s", path.Addr)
			jsonhttp.NotFound(w, "chunk not found")
			return

		}
		s.Logger.Debugf("bzz-chunk: chunk read error: %v ,addr %s", err, path.Addr)
		s.Logger.Error("bzz-chunk: chunk read error")
		jsonhttp.InternalServerError(w, "chunk read error")
		return
	}
	w.Header().Set("Content-Type", "binary/octet-stream")
	_, _ = io.Copy(w, bytes.NewReader(chunk.Data()))
}
