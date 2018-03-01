// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// openapi HTTP server
//
// Command:
// $ goa gen goa.design/goa/examples/calc/design

package server

import (
	"context"
	"net/http"

	openapi "goa.design/goa/examples/calc/gen/openapi"
	goahttp "goa.design/goa/http"
)

// Server lists the openapi service endpoint HTTP handlers.
type Server struct {
	Mounts []*MountPoint
}

// MountPoint holds information about the mounted endpoints.
type MountPoint struct {
	// Method is the name of the service method served by the mounted HTTP handler.
	Method string
	// Verb is the HTTP method used to match requests to the mounted handler.
	Verb string
	// Pattern is the HTTP request path pattern used to match requests to the
	// mounted handler.
	Pattern string
}

// New instantiates HTTP handlers for all the openapi service endpoints.
func New(
	e *openapi.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"../../gen/http/openapi.json", "GET", "/swagger.json"},
		},
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "openapi" }

// Mount configures the mux to serve the openapi endpoints.
func Mount(mux goahttp.Muxer) {
	MountGenHTTPOpenapiJSON(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../gen/http/openapi.json")
	}))
}

// MountGenHTTPOpenapiJSON configures the mux to serve GET request made to
// "/swagger.json".
func MountGenHTTPOpenapiJSON(mux goahttp.Muxer, h http.Handler) {
	mux.Handle("GET", "/swagger.json", h.ServeHTTP)
}
