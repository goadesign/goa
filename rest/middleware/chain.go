package middleware

import "net/http"

// Chain coalesces the given middlewares into a single middleware where the
// first middleware calls the second which calls the next etc.
func Chain(ms ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	if len(ms) == 0 {
		return nil
	}
	return func(h http.Handler) http.Handler {
		ln := len(ms)
		for i := 0; i < ln; i++ {
			h = ms[ln-i-1](h)
		}
		return h
	}
}
