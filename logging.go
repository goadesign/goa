package goa

import (
	"bytes"
	"fmt"
	"log"

	"golang.org/x/net/context"
)

type (
	// LogAdapter is the logger interface used by goa to log informational and error messages.
	// Adapters to different logging backends are provided in the logging package.
	// goa takes care of initializing the logging context with the service, controller and
	// action name. Additional logging context values may be set via WithValue.
	LogAdapter interface {
		// Info logs an informational message.
		Info(msg string, keyvals ...interface{})
		// Error logs an error.
		Error(msg string, keyvals ...interface{})
		// New appends to the logger context and returns the updated logger adapter.
		New(keyvals ...interface{}) LogAdapter
	}

	// stdLogger uses the stdlib logger.
	stdLogger struct {
		*log.Logger
		keyvals []interface{}
	}
)

// ErrMissingLogValue is the value used to log keys with missing values
const ErrMissingLogValue = "MISSING"

// LogInfo extracts the logger from the given context and calls Info on it.
// In general this shouldn't be needed (the client code should already have a handle on the logger)
// This is mainly useful for "out-of-band" code like middleware.
func LogInfo(ctx context.Context, msg string, keyvals ...interface{}) {
	logit(ctx, msg, keyvals, false)
}

// LogError extracts the logger from the given context and calls Error on it.
// In general this shouldn't be needed (the client code should already have a handle on the logger)
// This is mainly useful for "out-of-band" code like middleware.
func LogError(ctx context.Context, msg string, keyvals ...interface{}) {
	logit(ctx, msg, keyvals, true)
}

func logit(ctx context.Context, msg string, keyvals []interface{}, aserror bool) {
	if l := ctx.Value(logKey); l != nil {
		if logger, ok := l.(LogAdapter); ok {
			if aserror {
				logger.Error(msg, keyvals...)
			} else {
				logger.Info(msg, keyvals...)
			}
		}
	}
}

// NewStdLogger returns an implementation of Logger backed by a stdlib Logger.
func NewStdLogger(logger *log.Logger) LogAdapter {
	return &stdLogger{Logger: logger}
}

func (l *stdLogger) Info(msg string, keyvals ...interface{}) {
	l.logit(msg, keyvals, false)
}

func (l *stdLogger) Error(msg string, keyvals ...interface{}) {
	l.logit(msg, keyvals, true)
}

func (l *stdLogger) New(keyvals ...interface{}) LogAdapter {
	if len(keyvals) == 0 {
		return l
	}
	kvs := append(l.keyvals, keyvals...)
	if len(kvs)%2 != 0 {
		kvs = append(kvs, ErrMissingLogValue)
	}
	return &stdLogger{
		Logger: l.Logger,
		// Limiting the capacity of the stored keyvals ensures that a new
		// backing array is created if the slice must grow.
		keyvals: kvs[:len(kvs):len(kvs)],
	}
}

func (l *stdLogger) logit(msg string, keyvals []interface{}, iserror bool) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, ErrMissingLogValue)
	}
	var fm bytes.Buffer
	lvl := "INFO"
	if iserror {
		lvl = "ERROR"
	}
	fm.WriteString(fmt.Sprintf("[%s] %s", lvl, msg))
	vals := make([]interface{}, n)
	offset := len(l.keyvals)
	for i := 0; i < offset; i += 2 {
		k := l.keyvals[i]
		v := l.keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2+offset/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	l.Logger.Printf(fm.String(), vals...)
}
