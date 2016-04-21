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
	"github.com/goadesign/goa"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
)

// Adapter is the log15 goa adapter logger.
type Adapter struct {
	log15.Logger
}

// New wraps a log15 logger into a goa logger adapter.
func New(logger log15.Logger) goa.LogAdapter {
	return &Adapter{Logger: logger}
}

// Logger returns the log15 logger stored in the given context if any, nil otherwise.
func Logger(ctx context.Context) log15.Logger {
	logger := goa.ContextLogger(ctx)
	if a, ok := logger.(*Adapter); ok {
		return a.Logger
	}
	return nil
}

// Info logs informational messages using log15.
func (l *Adapter) Info(msg string, data ...interface{}) {
	l.Logger.Info(msg, data...)
}

// Error logs error messages using log15.
func (l *Adapter) Error(msg string, data ...interface{}) {
	l.Logger.Error(msg, data...)
}

// New creates a new logger given a context.
func (l *Adapter) New(data ...interface{}) goa.LogAdapter {
	return &Adapter{Logger: l.Logger.New(data...)}
}
