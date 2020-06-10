// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/ethersphere/bee/pkg/collection/entry"
	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/file/joiner"
	"github.com/ethersphere/bee/pkg/file/splitter"
	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
)

const (
	MultiPartFormData = "multipart/form-data"
)

type FileUploadResponse struct {
	Address swarm.Address `json:"address"`
}

func (s *server) bzzFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	contentType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != MultiPartFormData {
		s.Logger.Debugf("bzz-file: no mutlipart: %v", err)
		s.Logger.Error("bzz-file: no mutlipart ")
		jsonhttp.BadRequest(w, "not a mutlipart/form-data")
		return
	}

	mr := multipart.NewReader(r.Body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			s.Logger.Debugf("bzz-file: read mutlipart: %v", err)
			s.Logger.Error("bzz-file: read mutlipart ")
			jsonhttp.BadRequest(w, "read a mutlipart/form-data")
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			fileName = part.FormName()
		}

		// set up splitter to process the data
		splitr := splitter.NewSimpleSplitter(s.Storer)
		ctx := context.Background()

		// first store the file and get the reference
		var reader io.ReadCloser
		reader = part
		fr, err := s.storeFileData(ctx, reader, r.ContentLength)
		if err != nil {
			s.Logger.Debugf("bzz-file: file store: %v, file name %s", err, fileName)
			s.Logger.Error("bzz-file: file store store ")
			jsonhttp.InternalServerError(w, "could not store metadata")
			return
		}

		// then create and store metadata.
		m := entry.NewMetadata(fileName)
		contentType := part.Header.Get("Content-Type")
		if contentType == "" {
			mt, err := mimetype.DetectReader(r.Body)
			if err != nil {
				s.Logger.Debugf("bzz-file: mimetype detection: %v, file name %s", err, fileName)
				s.Logger.Error("bzz-file: mimetype detection ")
				jsonhttp.InternalServerError(w, "could not detect minetype")
				return
			}
			contentType = mt.String()
		}
		m.SetMimeType(contentType)
		mr, err := s.storeMetaData(ctx, splitr, m)
		if err != nil {
			s.Logger.Debugf("bzz-file: metadata store: %v, file name %s", err, fileName)
			s.Logger.Error("bzz-file: metadata store ")
			jsonhttp.InternalServerError(w, "could not store metadata")
			return
		}

		// now join both references to create entry and store it.
		entrie := entry.New(fr, mr)
		addr, err := s.storeEntry(ctx, splitr, entrie)
		if err != nil {
			s.Logger.Debugf("bzz-file: entry store: %v, file name %s", err, fileName)
			s.Logger.Error("bzz-file: entry store ")
			jsonhttp.InternalServerError(w, "could not store entry")
			return
		}

		jsonhttp.OK(w, &FileUploadResponse{Address: addr})
	}
}

func (s *server) bzzFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	address, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("bzz-file: parse file address %s: %v", addr, err)
		s.Logger.Error("bzz-file: parse file address")
		jsonhttp.BadRequest(w, "invalid file address")
		return
	}

	// read entry.
	j := joiner.NewSimpleJoiner(s.Storer)
	buf := bytes.NewBuffer(nil)
	err = file.JoinReadAll(j, address, buf)
	if err != nil {
		s.Logger.Debugf("bzz-file: read entry %s: %v", addr, err)
		s.Logger.Error("bzz-file: read entry")
		jsonhttp.InternalServerError(w, "error reading entry")
		return
	}
	e := &entry.Entry{}
	err = e.UnmarshalBinary(buf.Bytes())
	if err != nil {
		s.Logger.Debugf("bzz-file: unmarshall entry %s: %v", addr, err)
		s.Logger.Error("bzz-file: unmarshall entry")
		jsonhttp.InternalServerError(w, "error unmarshalling entry")
		return
	}

	// If none match header is set always send the reply as not modified
	// TODO: when SOC comes, we need to revisit this concept
	noneMatchEtag := r.Header.Get("If-None-Match")
	if noneMatchEtag != "" {
		if e.Reference().Equal(address) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Read metadata.
	buf = bytes.NewBuffer(nil)
	err = file.JoinReadAll(j, e.Metadata(), buf)
	if err != nil {
		s.Logger.Debugf("bzz-file: read metadata %s: %v", addr, err)
		s.Logger.Error("bzz-file: read netadata")
		jsonhttp.InternalServerError(w, "error reading metadata")
		return
	}
	metaData := &entry.Metadata{}
	err = json.Unmarshal(buf.Bytes(), metaData)
	if err != nil {
		s.Logger.Debugf("bzz-file: unmarshall metadata %s: %v", addr, err)
		s.Logger.Error("bzz-file: unmarshall metadata")
		jsonhttp.InternalServerError(w, "error unmarshalling metadata")
		return
	}

	// send the file data back in the response
	outBuffer := io.ReadWriter(nil)
	err = file.JoinReadAll(j, e.Reference(), outBuffer)
	if err != nil {
		s.Logger.Debugf("bzz-file: data read %s: %v", addr, err)
		s.Logger.Error("bzz-file: data read")
		jsonhttp.InternalServerError(w, "error reading data")
		return
	}
	w.Header().Set("ETag", fmt.Sprintf("%q", e.Reference()))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", metaData.Filename))
	w.Header().Set("Content-Type", metaData.MimeType)
	_, _ = io.Copy(w, outBuffer)
}

func (s *server) storeFileData(ctx context.Context, r io.ReadCloser, len int64) (swarm.Address, error) {
	//resp, err := s.splitUpload(ctx, r, len)
	return swarm.NewAddress([]byte{0}), nil
}

func (s *server) storeMetaData(ctx context.Context, splitr file.Splitter, m *entry.Metadata) (swarm.Address, error) {
	// serialize metadata and send it to splitter
	metadataBytes, err := json.Marshal(m)
	if err != nil {
		return swarm.NewAddress([]byte{0}), err
	}

	// first add metadata
	metadataBuf := bytes.NewBuffer(metadataBytes)
	metadataReader := io.LimitReader(metadataBuf, int64(len(metadataBytes)))
	metadataReadCloser := ioutil.NopCloser(metadataReader)
	metadataAddr, err := splitr.Split(ctx, metadataReadCloser, int64(len(metadataBytes)))
	if err != nil {
		return swarm.NewAddress([]byte{0}), err
	}
	return metadataAddr, nil
}

func (s *server) storeEntry(ctx context.Context, splitr file.Splitter, ent *entry.Entry) (swarm.Address, error) {
	fileEntryBytes, err := ent.MarshalBinary()
	if err != nil {
		return swarm.NewAddress([]byte{0}), err
	}
	fileEntryBuf := bytes.NewBuffer(fileEntryBytes)
	fileEntryReader := io.LimitReader(fileEntryBuf, int64(len(fileEntryBytes)))
	fileEntryReadCloser := ioutil.NopCloser(fileEntryReader)
	fileEntryAddr, err := splitr.Split(ctx, fileEntryReadCloser, int64(len(fileEntryBytes)))
	if err != nil {
		return swarm.NewAddress([]byte{0}), err
	}
	return fileEntryAddr, nil
}
