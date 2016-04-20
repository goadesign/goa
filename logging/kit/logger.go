/*
Package goakit contains an adapter that makes it possible to configure goa so it uses the go-kit
log package as logger backend.
Usage:

    // Initialize logger using github.com/go-kit/kit/log package
    logger := log.NewLogfmtLogger(w)
    // Initialize goa service logger using adapter
    service.WithLogger(goakit.New(logger))
    // ... Proceed with configuring and starting the goa service
*/
package goakit

import (
	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
)

// Logger is the go-kit log goa adapter logger.
type Logger struct {
	*log.Context
}

// New wraps a go-kit logger into a goa logger.
func New(logger log.Logger) goa.LogAdapter {
	return FromContext(log.NewContext(logger))
}

// FromContext wraps a go-kit log context into a goa logger.
func FromContext(ctx *log.Context) goa.LogAdapter {
	return &Logger{Context: ctx}
}

// Info logs informational messages using go-kit.
func (l *Logger) Info(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "info", "msg", msg}
	ctx = append(ctx, data...)
	l.Context.Log(ctx...)
}

// Error logs error messages using go-kit.
func (l *Logger) Error(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "error", "msg", msg}
	ctx = append(ctx, data...)
	l.Context.Log(ctx...)
}

// New instantiates a new logger from the given context.
func (l *Logger) New(data ...interface{}) goa.LogAdapter {
	return &Logger{Context: l.Context.With(data...)}
}
