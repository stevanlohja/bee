// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debugapi

import (
	"net/http"

	"github.com/ethersphere/bee/pkg/jsonhttp"
)

// swagger:response nodeStatusResponse
//
// Node status response.
type statusResponse struct {
	// Status message.
	// Example: ok
	// in: body
	Status string `json:"status"`
}

// swagger:route GET /health status health
//
// Service health status
//
// Returns service status to be used by health check tools.
//
// Produces:
// - application/json
//
// Responses:
//   200: nodeStatusResponse
//   default: statusResponse
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, statusResponse{
		Status: "ok",
	})
}

// swagger:route GET /readiness status readiness
//
// Service readiness status
//
// Returns service status to be used by readiness check tools.
//
// Produces:
// - application/json
//
// Responses:
//   200: nodeStatusResponse
//   default: statusResponse
func (s *server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, statusResponse{
		Status: "ok",
	})
}
