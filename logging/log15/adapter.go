/*
Package goalog15 contains an adapter that makes it possible to configure goa so it uses log15
as logger backend.
Usage:

    logger := log15.New()
    // ... Initialize logger handler using log15 package
    service.WithLogger(goalog15.New(logger))
    // ... Proceed with configuring and starting the goa service

    // In handlers:
    goalog15.Logger(ctx).Info("foo")
*/
package goalog15

import (
	"context"

	"github.com/goadesign/goa"
	"github.com/inconshreveable/log15"
)

// adapter is the log15 goa adapter logger.
type adapter struct {
	log15.Logger
}

// New wraps a log15 logger into a goa logger adapter.
func New(logger log15.Logger) goa.LogAdapter {
	return &adapter{Logger: logger}
}

// Logger returns the log15 logger stored in the given context if any, nil otherwise.
func Logger(ctx context.Context) log15.Logger {
	logger := goa.ContextLogger(ctx)
	if a, ok := logger.(*adapter); ok {
		return a.Logger
	}
	return nil
}

// Info logs informational messages using log15.
func (a *adapter) Info(msg string, data ...interface{}) {
	a.Logger.Info(msg, data...)
}

// Error logs error messages using log15.
func (a *adapter) Error(msg string, data ...interface{}) {
	a.Logger.Error(msg, data...)
}

// New creates a new logger given a context.
func (a *adapter) New(data ...interface{}) goa.LogAdapter {
	return &adapter{Logger: a.Logger.New(data...)}
}
