/*
Package goakit contains an adapter that makes it possible to configure goa so it uses the go-kit
log package as logger backend.
Usage:

    // Initialize logger using github.com/go-kit/kit/log package
    logger := log.NewLogfmtLogger(w)
    // Initialize goa service logger using adapter
    service.WithLogger(goakit.New(logger))
    // ... Proceed with configuring and starting the goa service

    // In handlers:
    goakit.Context(ctx).Log("foo", "bar")
*/
package goakit

import (
	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
)

// adapter is the go-kit log goa logger adapter.
type adapter struct {
	*log.Context
}

// New wraps a go-kit logger into a goa logger.
func New(logger log.Logger) goa.LogAdapter {
	return FromContext(log.NewContext(logger))
}

// FromContext wraps a go-kit log context into a goa logger.
func FromContext(ctx *log.Context) goa.LogAdapter {
	return &adapter{Context: ctx}
}

// Context returns the go-kit log context stored in the given context if any, nil otherwise.
func Context(ctx context.Context) *log.Context {
	logger := goa.ContextLogger(ctx)
	if a, ok := logger.(*adapter); ok {
		return a.Context
	}
	return nil
}

// Info logs informational messages using go-kit.
func (a *adapter) Info(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "info", "msg", msg}
	ctx = append(ctx, data...)
	a.Context.Log(ctx...)
}

// Error logs error messages using go-kit.
func (a *adapter) Error(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "error", "msg", msg}
	ctx = append(ctx, data...)
	a.Context.Log(ctx...)
}

// New instantiates a new logger from the given context.
func (a *adapter) New(data ...interface{}) goa.LogAdapter {
	return &adapter{Context: a.Context.With(data...)}
}
