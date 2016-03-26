/*
Package goakit contains an adapter that makes it possible to configure goa so it uses the go-kit
log package as logger backend.
Usage:

    // Initialize logger using github.com/go-kit/kit/log package
    logger := log.NewLogfmtLogger(w)
    // Initialize goa service logger using adapter
    service.UseLogger(goakit.New(logger))
    // ... Proceed with configuring and starting the goa service
*/
package goakit

import (
	"github.com/go-kit/kit/log"
	"github.com/goadesign/goa"
)

// Logger is the go-kit log goa adapter logger.
type Logger struct {
	log.Logger
}

// New wraps a go-kit logger into a goa logger.
func New(logger log.Logger) goa.Logger {
	return &Logger{Logger: logger}
}

// Info logs informational messages using log15.
func (l *Logger) Info(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "info", "msg", msg}
	ctx = append(ctx, data...)
	l.Logger.Log(ctx...)
}

// Error logs error messages using log15.
func (l *Logger) Error(msg string, data ...interface{}) {
	ctx := []interface{}{"lvl", "error", "msg", msg}
	ctx = append(ctx, data...)
	l.Logger.Log(ctx...)
}
