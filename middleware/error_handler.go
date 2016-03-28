package middleware

import (
	"net/http"

	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// ErrorHandler turns a Go error into an HTTP response. It should be placed in the middleware chain
// below the logger middleware so the logger properly logs the HTTP response. ErrorHandler
// understands instances of goa.Error and returns the status and response body embodied in them,
// it turns other Go error types into a 500 internal error response.
// If suppressInternal is true the details of internal errors is not included in HTTP responses.
func ErrorHandler(suppressInternal bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			e := h(ctx, rw, req)
			if e == nil {
				return nil
			}

			status := http.StatusInternalServerError
			var respBody interface{}
			if err, ok := e.(*goa.Error); ok {
				status = err.Status
				respBody = err
				rw.Header().Set("Content-Type", goa.ErrorMediaIdentifier)
			} else {
				respBody = e.Error()
				rw.Header().Set("Content-Type", "text/plain")
			}
			if status >= 500 && status < 600 {
				reqID := ctx.Value(reqIDKey)
				if reqID == nil {
					reqID = shortID()
					ctx = context.WithValue(ctx, reqIDKey, reqID)
				}
				goa.LogError(ctx, "uncaught error", "id", reqID, "msg", respBody)
				if suppressInternal {
					rw.Header().Set("Content-Type", goa.ErrorMediaIdentifier)
					respBody = goa.ErrInternal("internal error [%s]", reqID)
				}
			}
			return goa.ContextResponse(ctx).Send(ctx, status, respBody)
		}
	}
}
