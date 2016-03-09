package goa

import (
	"bytes"
	"fmt"
	"log"

	"golang.org/x/net/context"
)

type (
	// Logger is the logger interface used by goa to log informational and error messages.
	// Adapters to different logging backends are provided in the logging package.
	Logger interface {
		// Info logs an informational message.
		Info(msg string, keyvals ...interface{})
		// Error logs an error.
		Error(msg string, keyvals ...interface{})
	}

	// stdLogger uses the stdlib logger.
	stdLogger struct {
		*log.Logger
	}
)

// ErrMissingLogValue is the value used to log keys with missing values
const ErrMissingLogValue = "MISSING"

// Info extracts the logger from the given context and calls Info on it.
// In general this shouldn't be needed (the client code should already have a handle on the logger)
// This is mainly useful for "out-of-band" code like middleware.
func Info(ctx context.Context, msg string, keyvals ...interface{}) {
	logit(ctx, msg, keyvals, false)
}

// Error extracts the logger from the given context and calls Error on it.
// In general this shouldn't be needed (the client code should already have a handle on the logger)
// This is mainly useful for "out-of-band" code like middleware.
func Error(ctx context.Context, msg string, keyvals ...interface{}) {
	logit(ctx, msg, keyvals, true)
}

func logit(ctx context.Context, msg string, keyvals []interface{}, aserror bool) {
	if l := ctx.Value(logKey); l != nil {
		if logger, ok := l.(Logger); ok {
			var logctx []interface{}
			if lctx := ctx.Value(logContextKey); lctx != nil {
				logctx = lctx.([]interface{})
			}
			data := append(logctx, keyvals...)
			if aserror {
				logger.Error(msg, data...)
			} else {
				logger.Info(msg, data...)
			}
		}
	}
}

// LogWith stores logging context to be used by all Log invocations using the returned context.
func LogWith(ctx context.Context, keyvals ...interface{}) context.Context {
	return context.WithValue(ctx, logContextKey, keyvals)
}

// NewStdLogger returns an implementation of Logger backed by a stdlib Logger.
func NewStdLogger(logger *log.Logger) Logger {
	return &stdLogger{Logger: logger}
}

func (l *stdLogger) Info(msg string, keyvals ...interface{}) {
	l.logit(msg, keyvals, false)
}

func (l *stdLogger) Error(msg string, keyvals ...interface{}) {
	l.logit(msg, keyvals, true)
}

func (l *stdLogger) logit(msg string, keyvals []interface{}, iserror bool) {
	n := (len(keyvals) + 1) / 2
	var fm bytes.Buffer
	lvl := "INFO"
	if iserror {
		lvl = "ERROR"
	}
	fm.WriteString(fmt.Sprintf("[%s] %s", lvl, msg))
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = ErrMissingLogValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		vals[i/2] = v
		fm.WriteString(" ")
		fm.WriteString(fmt.Sprintf("%s=%%v", k))
	}
	l.Logger.Printf(fm.String(), vals...)
}
