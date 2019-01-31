package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type (
	// RequestIDOption uses a constructor pattern to customize middleware.
	RequestIDOption func(*RequestIDOptions) *RequestIDOptions

	// RequestIDOptions is the struct storing all the options.
	RequestIDOptions struct {
		// useRequestID if true indicates the middleware to use the RequestIDKey
		// context key to infer the request ID instead of always generating
		// unique IDs. Defaults to always-generate.
		useRequestID bool
		// requestIDLimit if positive truncates the request ID at the specified
		// length. Defaults to no limit.
		requestIDLimit int
	}
)

// NewRequestIDOptions initializes the options for the request ID middleware.
func NewRequestIDOptions(options ...RequestIDOption) *RequestIDOptions {
	o := new(RequestIDOptions)
	for _, option := range options {
		o = option(o)
	}
	return o
}

// GenerateRequestID initializes the given context with a unique value under
// the RequestIDKey key. If UseRequestIDOption is set to true, it uses the
// RequestIDKey key in the context (if present) instead of generating a new ID.
func GenerateRequestID(ctx context.Context, o *RequestIDOptions) context.Context {
	var id string
	{
		if o.useRequestID {
			if i := ctx.Value(RequestIDKey); i != nil {
				id = i.(string)
				if o.requestIDLimit > 0 && len(id) > o.requestIDLimit {
					id = id[:o.requestIDLimit]
				}
			}
		}
		if id == "" {
			id = shortID()
		}
	}
	return context.WithValue(ctx, RequestIDKey, id)
}

// UseRequestIDOption enables/disables using RequestID context key to store
// the unique request ID.
func UseRequestIDOption(f bool) RequestIDOption {
	return func(o *RequestIDOptions) *RequestIDOptions {
		o.useRequestID = f
		return o
	}
}

// RequestIDLimitOption sets the option for truncating the request ID stored
// in the context at the specified length.
func RequestIDLimitOption(limit int) RequestIDOption {
	return func(o *RequestIDOptions) *RequestIDOptions {
		o.requestIDLimit = limit
		return o
	}
}

// IsUseRequestID returns the request ID option.
func (o *RequestIDOptions) IsUseRequestID() bool {
	return o.useRequestID
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
