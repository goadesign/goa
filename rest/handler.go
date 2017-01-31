package rest

import "net/http"

// Handler is a HTTP handler that makes it possible to mount middleware that is
// executed after the request context is initialized. Such middleware may take
// advantage of the ContextXXX functions to retrieve information from the
// context.
//
// Note: the process of initializing the context involves reading the request
// body so that it cannot be read again (however the data is available via the
// ContextRequest function).
type Handler struct {
	middlewares []func(http.Handler) http.Handler
}

// Use mounts a HTTP middleware for execution after the request context
// initialization has occurred.
func (h *Handler) Use(m func(http.Handler) http.Handler) {
	h.middlewares = append(h.middlewares, m)
}

// ServeHTTP applies any middleware mounted via Use the first time it is called
// then invokes the embedded handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
