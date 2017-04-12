package goa

import (
	"net/http"
	"net/url"
	"strconv"

	"context"
)

// Keys used to store data in context.
const (
	reqKey key = iota + 1
	respKey
	ctrlKey
	actionKey
	paramsKey
	logKey
	logContextKey
	errKey
	securityScopesKey
)

type (
	// RequestData provides access to the underlying HTTP request.
	RequestData struct {
		*http.Request

		// Payload returns the decoded request body.
		Payload interface{}
		// Params contains the raw values for the parameters defined in the design including
		// path parameters, query string parameters and header parameters.
		Params url.Values
	}

	// ResponseData provides access to the underlying HTTP response.
	ResponseData struct {
		http.ResponseWriter

		// The service used to encode the response.
		Service *Service
		// ErrorCode is the code of the error returned by the action if any.
		ErrorCode string
		// Status is the response HTTP status code.
		Status int
		// Length is the response body length.
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

// WithAction creates a context with the given action name.
func WithAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, actionKey, action)
}

// WithLogger sets the request context logger and returns the resulting new context.
func WithLogger(ctx context.Context, logger LogAdapter) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

// WithLogContext instantiates a new logger by appending the given key/value pairs to the context
// logger and setting the resulting logger in the context.
func WithLogContext(ctx context.Context, keyvals ...interface{}) context.Context {
	logger := ContextLogger(ctx)
	if logger == nil {
		return ctx
	}
	nl := logger.New(keyvals...)
	return WithLogger(ctx, nl)
}

// WithError creates a context with the given error.
func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errKey, err)
}

// ContextController extracts the controller name from the given context.
func ContextController(ctx context.Context) string {
	if c := ctx.Value(ctrlKey); c != nil {
		return c.(string)
	}
	return "<unknown>"
}

// ContextAction extracts the action name from the given context.
func ContextAction(ctx context.Context) string {
	if a := ctx.Value(actionKey); a != nil {
		return a.(string)
	}
	return "<unknown>"
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
func ContextLogger(ctx context.Context) LogAdapter {
	if v := ctx.Value(logKey); v != nil {
		return v.(LogAdapter)
	}
	return nil
}

// ContextError extracts the error from the given context.
func ContextError(ctx context.Context) error {
	if err := ctx.Value(errKey); err != nil {
		return err.(error)
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
