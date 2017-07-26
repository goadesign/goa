// +build !appengine

package middleware

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
)

// AppEngineContext is a middleware that ensures that the goa contexts are valid App Engine contexts
// It is safe to use in a non-appengine environment
func AppEngineContext(h goa.Handler) goa.Handler {
	return h
}
