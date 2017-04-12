/*
Package goakit contains an adapter that makes it possible to configure goa so it uses the go-kit
log package as logger backend.
Usage:

    // Initialize logger using github.com/go-kit/kit/log package
    logger := log.NewLogfmtLogger(w)
    // Initialize goa service logger using adapter
    service.WithLogger(goakit.New(logger))
    // ... Proceed with configuring and starting the goa service

    // In middlewares:
    goakit.Logger(ctx).Log("foo", "bar")
*/
package goakit

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
)

// adapter is the go-kit log goa logger adapter.
type adapter struct {
	log.Logger
}

// New wraps a go-kit logger into a goa logger.
func New(logger log.Logger) goa.LogAdapter {
	return &adapter{logger}
}

// Logger returns the go-kit logger stored in the given context if any, nil otherwise.
func Logger(ctx context.Context) log.Logger {
	logger := goa.ContextLogger(ctx)
	if a, ok := logger.(*adapter); ok {
		return a.Logger
	}
	return nil
}

// Info logs informational messages using go-kit.
func (a *adapter) Info(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "info", "msg", msg}
	ctx = append(ctx, data...)
	a.Logger.Log(ctx...)
}

// Error logs error messages using go-kit.
func (a *adapter) Error(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "error", "msg", msg}
	ctx = append(ctx, data...)
	a.Logger.Log(ctx...)
}

// New instantiates a new logger from the given context.
func (a *adapter) New(data ...interface{}) goa.LogAdapter {
	return &adapter{Logger: log.With(a.Logger, data...)}
}
