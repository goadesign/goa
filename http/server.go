package http

import (
	"net/http"
	"net/url"
	"strings"
)

type (
	// Server is the HTTP server interface used to wrap the server handlers
	// with the given middleware.
	Server interface {
		Use(func(http.Handler) http.Handler)
	}

	// Mounter is the interface for servers that allow mounting their endpoints
	// into a muxer.
	Mounter interface {
		Mount(Muxer)
	}

	// Servers is a list of servers.
	Servers []Server
)

// Use wraps the servers with the given middleware.
func (s Servers) Use(m func(http.Handler) http.Handler) {
	for _, v := range s {
		v.Use(m)
	}
}

// Mount will go through all the servers and mount them into the Muxer. It will
// panic unless all servers satisfy the Mounter interface.
func (s Servers) Mount(mux Muxer) {
	for _, v := range s {
		m := v.(Mounter)
		m.Mount(mux)
	}
}

// Replace returns a handler that serves HTTP requests by replacing the
// request URL's Path (and RawPath if set) and invoking the handler h.
// The logic is the same as the standard http package StripPrefix function.
func Replace(old, nw string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p, rp string
		if old != "" {
			p = strings.Replace(r.URL.Path, old, nw, 1)
			rp = strings.Replace(r.URL.RawPath, old, nw, 1)
		} else {
			p = nw
			if r.URL.RawPath != "" {
				rp = nw
			}
		}
		if p != r.URL.Path && (r.URL.RawPath == "" || rp != r.URL.RawPath) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}
