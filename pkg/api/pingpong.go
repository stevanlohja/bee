// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/ethersphere/bee/pkg/httputil"
	"github.com/ethersphere/bee/pkg/jsonhttp"
	"github.com/ethersphere/bee/pkg/p2p"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/gorilla/mux"
)

// swagger:response pingpongResponse
//
// Pingpong result.
type pingpongResponse struct {
	// Round trip time in time duration format.
	// Example: 1.45s
	RTT string `json:"rtt"`
}

// swagger:parameters pingpong
type pingpongPath struct {
	// Hex-encoded peer overlay address
	// in: path
	Addr swarm.Address `json:"addr"`
}

// swagger:route POST /pingpong/{addr} pingpong pingpong
//
// Pingpong exchange
//
// Measure round trip time to peer by exchanging P2P messages.
//
// Produces:
// - application/json
//
// Responses:
//   200: pingpongResponse
//   default: statusResponse
func (s *server) pingpongHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span, logger, ctx := s.Tracer.StartSpanFromContext(ctx, "pingpong-api", s.Logger)
	defer span.Finish()

	var path pingpongPath
	if err := httputil.UnmarshalMuxVars(mux.Vars(r), &path); err != nil {
		logger.Debugf("pingpong: path params: %v", err)
		jsonhttp.NotFound(w, nil)
		return
	}

	rtt, err := s.Pingpong.Ping(ctx, path.Addr, "hey", "there", ",", "how are", "you", "?")
	if err != nil {
		logger.Debugf("pingpong: ping %s: %v", path.Addr, err)
		if errors.Is(err, p2p.ErrPeerNotFound) {
			jsonhttp.NotFound(w, "peer not found")
			return
		}

		logger.Errorf("pingpong failed to peer %s", path.Addr)
		jsonhttp.InternalServerError(w, nil)
		return
	}
	s.metrics.PingRequestCount.Inc()

	logger.Infof("pingpong succeeded to peer %s", path.Addr)
	jsonhttp.OK(w, pingpongResponse{
		RTT: rtt.String(),
	})
}
