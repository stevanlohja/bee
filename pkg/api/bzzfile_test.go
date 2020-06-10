// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/jsonhttp/jsonhttptest"
	"github.com/ethersphere/bee/pkg/storage/mock"
	"github.com/ethersphere/bee/pkg/storage/mock/validator"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/tags"
)

func TestBzz(t *testing.T) {
	var (
		resource             = func(addr swarm.Address) string { return "/bzz-chunk/" + addr.String() }
		validHash            = swarm.MustParseHexAddress("aabbcc")
		invalidHash          = swarm.MustParseHexAddress("bbccdd")
		validContent         = []byte("bbaatt")
		invalidContent       = []byte("bbaattss")
		mockValidator        = validator.NewMockValidator(validHash, append(newSpan(uint64(len(validContent))), validContent...))
		tag                  = tags.NewTags()
		mockValidatingStorer = mock.NewValidatingStorer(mockValidator, tag)
		client               = newTestServer(t, testServerOptions{
			Storer: mockValidatingStorer,
			Tags:   tag,
		})
	)

	t.Run("ok", func(t *testing.T) {
		jsonhttptest.ResponseDirect(t, client, http.MethodPost, resource(validHash), bytes.NewReader(validContent), http.StatusOK, jsonhttp.StatusResponse{
			Message: http.StatusText(http.StatusOK),
			Code:    http.StatusOK,
		})

		// try to fetch the same chunk
		resp := request(t, client, http.MethodGet, resource(validHash), nil, http.StatusOK)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(validContent, data) {
			t.Fatal("data retrieved doesnt match uploaded content")
		}
	})
}
