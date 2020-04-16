// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debugapi

import (
	"net/http"

	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/multiformats/go-multiaddr"
)

// swagger:response addressesResponse
//
// Node addresses.
type addressesResponse struct {
	// Hex-encoded node overlay address.
	// in: body
	Overlay swarm.Address `json:"overlay"`
	// A list of node's libp2p underlay multiaddresses.
	// in: body
	Underlay []multiaddr.Multiaddr `json:"underlay"`
}

// swagger:route GET /addresses p2p addresses
//
// Node addresses
//
// Returns overlay and underlay addresses of a running node.
//
// Produces:
// - application/json
//
// Responses:
//   200: addressesResponse
//   default: statusResponse
func (s *server) addressesHandler(w http.ResponseWriter, r *http.Request) {
	underlay, err := s.P2P.Addresses()
	if err != nil {
		s.Logger.Debugf("debug api: p2p addresses: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	jsonhttp.OK(w, addressesResponse{
		Overlay:  s.Overlay,
		Underlay: underlay,
	})
}
