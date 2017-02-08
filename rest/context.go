package rest

import (
	"context"
	"net/http"
)

type ContextKey int

const (
	// keyRequest is the key used to store the raw http.Request object in
	// the request context.
	keyRequest ContextKey = iota + 1

	// keyResponse is the key used to store the raw http.ResponseWriter
	// object in the request context.
	keyResponse
)

type (
	// key is the private type used to store values in the context.
	key int
)

// NewContext builds a HTTP request context that stores the underlying HTTP
// request and response writer. Use ContextRequest and ContextResponse to extract
// the corresponding objects from the returned context.
func NewContext(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, keyResponse, w)
	return context.WithValue(ctx, keyRequest, r)
}

// ContextRequest extracts the underlying HTTP request from the context.
func ContextRequest(ctx context.Context) *http.Request {
	if r := ctx.Value(keyRequest); r != nil {
		return r.(*http.Request)
	}
	return nil
}

// ContextResponse extracts the underlying HTTP response writer from the context.
func ContextResponse(ctx context.Context) http.ResponseWriter {
	if r := ctx.Value(keyResponse); r != nil {
		return r.(http.ResponseWriter)
	}
	return nil
}
