/*
Package goalog15 contains an adapter that makes it possible to configure goa so it uses log15
as logger backend.
Usage:

    logger := log15.New()
    // ... Initialize logger handler using log15 package
    service.UseLogger(goalog15.New(logger))
    // ... Proceed with configuring and starting the goa service
*/
package goalog15

import (
	"github.com/goadesign/goa"
	"gopkg.in/inconshreveable/log15.v2"
)

// Logger is the log15 goa adapter logger.
type Logger struct {
	log15.Logger
}

// New wraps a log15 logger into a goa logger.
func New(logger log15.Logger) goa.Logger {
	return &Logger{Logger: logger}
}

// Info logs informational messages using log15.
func (l *Logger) Info(msg string, data ...interface{}) {
	l.Logger.Info(msg, data...)
}

// Error logs error messages using log15.
func (l *Logger) Error(msg string, data ...interface{}) {
	l.Logger.Error(msg, data...)
}
