// +build appengine

package middleware

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
	"google.golang.org/appengine"
)

// AppEngineContext is a middleware that ensures that the goa contexts are valid App Engine contexts
// It is safe to use in a non-appengine environment
func AppEngineContext(h goa.Handler) goa.Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		return h(appengine.WithContext(ctx, req), rw, req)
	}
}
