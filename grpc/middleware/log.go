package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

type (
	// Logger is the logging interface used by the middleware to produce
	// log entries.
	Logger interface {
		// Log creates a log entry using a sequence of alternating keys
		// and values.
		Log(keyvals ...interface{})
	}

	// adapter is a thin wrapper around the stdlib logger that adapts it to
	// the Logger interface.
	adapter struct {
		*log.Logger
	}
)

// Log returns a middleware that logs incoming gRPC requests and outgoing
// responses. The middleware uses the request ID set by the RequestID middleware
// or creates a short unique request ID if missing for each incoming request and
// logs it with the request and corresponding response details.
//
// The middleware logs the incoming requests gRPC method. It also logs the
// response gRPC status code, message length (in bytes), and timing information.
func Log(l Logger) grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		reqID := ctx.Value(RequestIDKey)
		if reqID == nil {
			reqID = shortID()
		}
		started := time.Now()

		// before executing rpc
		l.Log("id", reqID,
			"method", info.FullMethod)

		// invoke rpc
		h, err := handler(ctx, req)

		// after executing rpc
		s, _ := status.FromError(err)
		l.Log("id", reqID,
			"status", s.Code(),
			"time", time.Since(started).String())
		return h, err
	})
}

// NewLogger creates a Logger backed by a stdlib logger.
func NewLogger(l *log.Logger) Logger {
	return &adapter{l}
}

func (a *adapter) Log(keyvals ...interface{}) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
	}
	var fm bytes.Buffer
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
}

// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
