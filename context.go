package goa

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/net/context"
)

// Keys used to store data in context.
const (
	reqKey key = iota + 1
	respKey
	paramsKey
	serviceKey
	logKey
	logContextKey
	securitySchemeKey
	securityScopesKey
)

type (
	// RequestData provides access to the underlying HTTP request.
	RequestData struct {
		*http.Request

		// Payload returns the decoded request body.
		Payload interface{}
		// Params is the path and querystring request parameters.
		Params url.Values
	}

	// ResponseData provides access to the underlying HTTP response.
	ResponseData struct {
		http.ResponseWriter

		// Status is the response HTTP status code
		Status int
		// Length is the response body length
		Length int
	}

	// key is the type used to store internal values in the context.
	// Context provides typed accessor methods to these values.
	key int
)

// NewContext builds a new goa request context. The parent context may include
// log context data.
// If parent is nil then RootContext is used.
func NewContext(ctx context.Context, rw http.ResponseWriter, req *http.Request, params url.Values) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	request := &RequestData{Request: req, Params: params}
	response := &ResponseData{ResponseWriter: rw}
	ctx = context.WithValue(ctx, respKey, response)
	ctx = context.WithValue(ctx, reqKey, request)

	return ctx
}

// UseLogger sets the request context logger and returns the resulting new context.
func UseLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

// Request gives access to the underlying HTTP request.
func Request(ctx context.Context) *RequestData {
	if r := ctx.Value(reqKey); r != nil {
		return r.(*RequestData)
	}
	return nil
}

// Response gives access to the underlying HTTP response.
func Response(ctx context.Context) *ResponseData {
	if r := ctx.Value(respKey); r != nil {
		return r.(*ResponseData)
	}
	return nil
}

// RequestService returns the service tageted by the request with the given context.
func RequestService(ctx context.Context) *Service {
	if r := ctx.Value(serviceKey); r != nil {
		return r.(*Service)
	}
	return nil
}

// RequestLogger returns the logger used by the request context.
func RequestLogger(ctx context.Context) Logger {
	if v := ctx.Value(logKey); v != nil {
		return v.(Logger)
	}
	return nil
}

// SwitchWriter overrides the underlying response writer. It returns the response
// writer that was previously set.
func (r *ResponseData) SwitchWriter(rw http.ResponseWriter) http.ResponseWriter {
	rwo := r.ResponseWriter
	r.ResponseWriter = rw
	return rwo
}

// Written returns true if the response was written.
func (r *ResponseData) Written() bool {
	return r.Status != 0
}

// Send serializes the given body matching the request Accept header against the service
// encoders. It uses the default service encoder if no match is found.
func (r *ResponseData) Send(ctx context.Context, code int, body interface{}) error {
	r.WriteHeader(code)
	return RequestService(ctx).EncodeResponse(ctx, body)
}

// BadRequest sends a HTTP response with status code 400 and the given error as body.
func (r *ResponseData) BadRequest(ctx context.Context, err *HTTPError) error {
	return r.Send(ctx, 400, err)
}

// Bug sends a HTTP response with status code 500 and the given body.
// The body can be set using a format and substituted values a la fmt.Printf.
func (r *ResponseData) Bug(ctx context.Context, format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	return r.Send(ctx, 500, &HTTPError{ErrInternal, msg})
}

// WriteHeader records the response status code and calls the underlying writer.
func (r *ResponseData) WriteHeader(status int) {
	go IncrCounter([]string{"goa", "response", strconv.Itoa(status)}, 1.0)
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write records the amount of data written and calls the underlying writer.
func (r *ResponseData) Write(b []byte) (int, error) {
	r.Length += len(b)
	return r.ResponseWriter.Write(b)
}
