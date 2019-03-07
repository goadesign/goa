// Package xray contains the AWS X-Ray segment document type populated by the
// transport-specific X-Ray middleware.
package xray

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	// SegKey is the request context key used to store the segments if any.
	SegKey contextKey = iota + 1
)

type (
	// private type used to define context keys.
	contextKey int
)

// Connect creates a goroutine to periodically re-dial a connection, so the
// hostname can be re-resolved if the IP changes. Returns a func that provides
// the latest Conn value.
func Connect(ctx context.Context, renewPeriod time.Duration, dial func() (net.Conn, error)) (func() net.Conn, error) {
	var (
		err error

		// guard access to c
		mu sync.RWMutex
		c  net.Conn
	)

	// get an initial connection
	if c, err = dial(); err != nil {
		return nil, err
	}

	// periodically re-dial
	go func() {
		ticker := time.NewTicker(renewPeriod)
		for {
			select {
			case <-ticker.C:
				newConn, err := dial()
				if err != nil {
					continue // we don't have anything better to replace `c` with
				}
				mu.Lock()
				c = newConn
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() net.Conn {
		mu.RLock()
		defer mu.RUnlock()
		return c
	}, nil
}

// NewID is a span ID creation algorithm which produces values that are
// compatible with AWS X-Ray.
func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// NewTraceID is a trace ID creation algorithm which produces values that are
// compatible with AWS X-Ray.
func NewTraceID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%d-%x-%s", 1, time.Now().Unix(), fmt.Sprintf("%x", b))
}
