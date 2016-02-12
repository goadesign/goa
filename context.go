package goa

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
)

// Keys used to store data in context.
const (
	reqKey key = iota + 1
	respKey
)

type (
	// RequestData provides access to the underlying HTTP request.
	RequestData struct {
		*http.Request

		// Action is the name of the resource action targeted by the request
		Action string
		// Controller is the name of the resource controller targeted by the request
		Controller string
		// Service is the service targeted by the request
		Service Service
		// Payload returns the decoded request body.
		Payload interface{}
		// Params is the path and querystring request parameters.
		Params url.Values
	}

	// ResponseData provides access to the underlying HTTP response.
	ResponseData struct {
		http.ResponseWriter
		ctx context.Context // for access to the encoder

		// Status is the response HTTP status code
		Status int
		// Len is the response body length
		Len int
	}

	// key is the type used to store internal values in the context.
	// Context provides typed accessor methods to these values.
	key int
)

// NewContext builds a goa context from the given context.Context and request state.
// If gctx is nil then RootContext is used.
func NewContext(gctx context.Context, ctrl *ApplicationController, action string,
	req *http.Request, rw http.ResponseWriter, params url.Values) context.Context {

	if gctx == nil {
		gctx = RootContext
	}
	request := &RequestData{
		Request:    req,
		Params:     params,
		Action:     action,
		Controller: ctrl.Name,
		Service:    ctrl.app,
	}
	response := &ResponseData{ResponseWriter: rw}
	gctx = context.WithValue(gctx, reqKey, request)
	gctx = context.WithValue(gctx, respKey, response)
	response.ctx = gctx

	return gctx
}

// Request gives access to the underlying HTTP request.
func Request(ctx context.Context) *RequestData {
	r := ctx.Value(reqKey)
	if r != nil {
		return r.(*RequestData)
	}
	return nil
}

// Response gives access to the underlying HTTP response.
func Response(ctx context.Context) *ResponseData {
	r := ctx.Value(respKey)
	if r != nil {
		return r.(*ResponseData)
	}
	return nil
}

// SetPayload initializes the unmarshaled request body value.
func (r *RequestData) SetPayload(payload interface{}) {
	r.Payload = payload
}

// SwitchResponseWriter overrides the underlying response writer. It returns the response
// writer that was previously set.
func (r *ResponseData) SwitchResponseWriter(rw http.ResponseWriter) http.ResponseWriter {
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
func (r *ResponseData) Send(code int, body interface{}) error {
	r.WriteHeader(code)
	return Request(r.ctx).Service.EncodeResponse(r.ctx, body)
}

// BadRequest sends a HTTP response with status code 400 and the given error as body.
func (r *ResponseData) BadRequest(err *BadRequestError) error {
	return r.Send(400, err.Error())
}

// Bug sends a HTTP response with status code 500 and the given body.
// The body can be set using a format and substituted values a la fmt.Printf.
func (r *ResponseData) Bug(format string, a ...interface{}) error {
	body := fmt.Sprintf(format, a...)
	return r.Send(500, body)
}

// WriteHeader records the response status code and calls the underlying writer.
func (r *ResponseData) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write records the amount of data written and calls the underlying writer.
func (r *ResponseData) Write(b []byte) (int, error) {
	r.Len += len(b)
	return r.ResponseWriter.Write(b)
}
