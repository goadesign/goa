package http

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

type (
	// Server is the HTTP server interface used to wrap the server handlers
	// with the given middleware.
	Server interface {
		Use(func(http.Handler) http.Handler)
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

// ReplacePrefix returns a handler that serves HTTP requests by replacing the
// prefix from the request URL's Path (and RawPath if set) and invoking the
// handler h. The logic is the same as the standard http package StripPrefix
// function.
func ReplacePrefix(old, nw string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, old)
		rp := strings.TrimPrefix(r.URL.RawPath, old)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = path.Join(nw, p)
			r2.URL.RawPath = path.Join(nw, rp)
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}
