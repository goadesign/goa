package goa

import "net/http"

type GetterFunc func(string) interface{}
type SetterFunc func(string, interface{})
type MiddlewareFunc func(http.Handler) http.Handler

func Middleware(getter GetterFunc, setter SetterFunc) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// wrap response writer so we can get headers / status
			// use sync.Pool for generating wrappers
			next.ServeHTTP(w, r)
			// Check headers status, binder sets action in response writer
			// (checks if response writer is a goa response writer and if so
			// uses that response writer SetAction method)
		})
	}
}
