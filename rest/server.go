package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"goa.design/goa.v2"
)

type (
	// HTTPServer implements a server which accepts HTTP requests.
	HTTPServer struct {
		// Context is the base context.
		Context context.Context
		// Service which runs the server.
		Service *goa.Service
		// Mux is the server HTTP mux (a.k.a. router).
		Mux ServeMux
		// Middleware is the server specific middleware.
		Middleware []Middleware
		// Decoder identifies the unmarshaler needed for the request and
		// invokes it.
		Decoder HTTPDecoder
		// Encoder is the response encoder.
		Encoder HTTPEncoder
		// MaxRequestBodyLength is the maximum number of bytes a request
		// may contain.
		MaxRequestBodyLength int64
	}

	// Handler is the HTTP request handler function.
	Handler func(rw http.ResponseWriter, req *http.Request) error

	// Unmarshaler defines the request payload unmarshaler signatures.
	Unmarshaler func(*goa.Service, *http.Request) error
)

// Send encodes the response using the server encoder and writes the HTTP
// response using the given HTTP status code.
func (svr *HTTPServer) Send(ctx context.Context, code int, body interface{}) error {
	r := ContextResponse(ctx)
	if r == nil {
		return fmt.Errorf("no response data in context")
	}
	r.WriteHeader(code)
	accept := ContextRequest(ctx).Header.Get("Accept")
	return svr.Encoder.Encode(body, ContextResponse(ctx), accept)
}

// ServeFiles replies to the request with the contents of the named file or
// directory. See FileHandler for details.
func (svr *HTTPServer) ServeFiles(path, filename string) error {
	if strings.Contains(path, ":") {
		return fmt.Errorf("path may only include wildcards that match the entire end of the URL (e.g. *filepath)")
	}
	svr.Service.LogAdapter.Info("mount file", "name", filename, "route", fmt.Sprintf("GET %s", path))
	handler := func(rw http.ResponseWriter, req *http.Request) error {
		return FileHandler(path, filename)(rw, req)
	}
	svr.Mux.Handle("GET", path, svr.MuxHandler("serve", handler, nil))
	return nil
}

// MuxHandler wraps a request handler into a MuxHandler. The MuxHandler
// initializes the request context by loading the request state, invokes the
// handler and in case of error invokes the controller (if there is one) or
// Service error handler. This function is intended for the generated code. User
// code should not need to call it directly.
func (svr *HTTPServer) MuxHandler(name string, hdlr Handler, unm Unmarshaler) MuxHandler {
	// Use closure to enable late computation of handlers to ensure all
	// middleware has been registered.
	var handler Handler

	return func(rw http.ResponseWriter, req *http.Request, params url.Values) {
		// Build handler middleware chains on first invocation
		if handler == nil {
			handler = func(rw http.ResponseWriter, req *http.Request) error {
				if !ContextResponse(req.Context()).Written() {
					return hdlr(rw, req)
				}
				return nil
			}
			ml := len(svr.Middleware)
			for i := range svr.Middleware {
				handler = svr.Middleware[ml-i-1](handler)
			}
		}

		// Build context
		ctx := NewContext(WithAction(svr.Context, name), rw, req, params)

		// Protect against request bodies with unreasonable length
		if svr.MaxRequestBodyLength > 0 {
			req.Body = http.MaxBytesReader(rw, req.Body, svr.MaxRequestBodyLength)
		}

		// Load body if any
		if req.ContentLength > 0 && unm != nil {
			if err := unm(svr.Service, req); err != nil {
				if err.Error() == "http: request body too large" {
					msg := fmt.Sprintf("request body length exceeds %d bytes", svr.MaxRequestBodyLength)
					err = ErrRequestBodyTooLarge(msg)
				} else {
					err = ErrBadRequest(err)
				}
				// TBD
				// ctx = WithError(ctx, err)
			}
		}

		// Invoke handler
		req = req.WithContext(ctx)
		if err := handler(ContextResponse(ctx), req); err != nil {
			goa.LogError(ctx, "uncaught error", "err", err)
			respBody := fmt.Sprintf("Internal error: %s", err) // Sprintf catches panics
			svr.Send(ctx, 500, respBody)
		}
	}
}
