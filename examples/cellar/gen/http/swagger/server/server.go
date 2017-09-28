// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// swagger HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design

package server

import (
	"context"
	"net/http"

	swagger "goa.design/goa/examples/cellar/gen/swagger"
	goahttp "goa.design/goa/http"
)

// Server lists the swagger service endpoint HTTP handlers.
type Server struct {
}

// New instantiates HTTP handlers for all the swagger service endpoints.
func New(
	e *swagger.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) *Server {
	return &Server{}
}

// Mount configures the mux to serve the swagger endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountGenHTTPOpenapiJSON(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../gen/http/openapi.json")
	}))
}

// MountGenHTTPOpenapiJSON configures the mux to serve GET request made to
// "/swagger/swagger.json".
func MountGenHTTPOpenapiJSON(mux goahttp.Muxer, h http.Handler) {
	mux.Handle("GET", "/swagger/swagger.json", h.ServeHTTP)
}
