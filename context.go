package goa

import (
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
	logKey
	logContextKey
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

// NewContext builds a new goa request context.
// If ctx is nil then context.Background() is used.
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

// ContextRequest extracts the request data from the given context.
func ContextRequest(ctx context.Context) *RequestData {
	if r := ctx.Value(reqKey); r != nil {
		return r.(*RequestData)
	}
	return nil
}

// ContextResponse extracts the response data from the given context.
func ContextResponse(ctx context.Context) *ResponseData {
	if r := ctx.Value(respKey); r != nil {
		return r.(*ResponseData)
	}
	return nil
}

// ContextLogger extracts the logger from the given context.
func ContextLogger(ctx context.Context) Logger {
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

// Written returns true if the response was written, false otherwise.
func (r *ResponseData) Written() bool {
	return r.Status != 0
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
