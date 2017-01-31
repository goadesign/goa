package rest

import (
	"net/http"

	"golang.org/x/net/context"
)

// Keys used to store data in context.
const (
	reqKey key = iota + 1
	respKey
	endpointKey
	serviceKey
)

type (
	// RequestData provides access to the underlying HTTP request.
	RequestData struct {
		// Request is the raw underlying http request. This field is
		// initialized by the final handler and is thus not available to
		// middlewares until the next handler has executed. Middlewares
		// must use the request passed to them as argument if needed.
		*http.Request
		// Payload returns the decoded request body.
		Payload interface{}
		// Params contains the captured path parameters indexed by name.
		Params map[string]string
	}

	// ResponseData provides access to the underlying HTTP response.
	ResponseData struct {
		http.ResponseWriter

		// ErrorCode is the code of the error
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
func NewContext(
	ctx context.Context,
	resp *ResponseData,
	req *RequestData,
	service, endpoint string,
) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, respKey, resp)
	ctx = context.WithValue(ctx, reqKey, req)
	ctx = context.WithValue(ctx, serviceKey, service)
	ctx = context.WithValue(ctx, endpointKey, endpoint)

	return ctx
}

// ContextService extracts the controller name from the given context.
func ContextService(ctx context.Context) string {
	if c := ctx.Value(serviceKey); c != nil {
		return c.(string)
	}
	return "<unknown>"
}

// ContextEndpoint extracts the action name from the given context.
func ContextEndpoint(ctx context.Context) string {
	if a := ctx.Value(endpointKey); a != nil {
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
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write records the amount of data written and calls the underlying writer.
func (r *ResponseData) Write(b []byte) (int, error) {
	r.Length += len(b)
	return r.ResponseWriter.Write(b)
}
