package goa

import (
	"bytes"
	"fmt"
	"log"

	"context"
)

// ErrMissingLogValue is the value used to log keys with missing values
const ErrMissingLogValue = "MISSING"

type (
	// LogAdapter is the logger interface used by goa to log informational and error messages.
	// Adapters to different logging backends are provided in the logging sub-packages.
	// goa takes care of initializing the logging context with the service, controller and
	// action names.
	LogAdapter interface {
		// Info logs an informational message.
		Info(msg string, keyvals ...interface{})
		// Error logs an error.
		Error(msg string, keyvals ...interface{})
		// New appends to the logger context and returns the updated logger logger.
		New(keyvals ...interface{}) LogAdapter
	}

	// adapter is the stdlib logger adapter.
	adapter struct {
		*log.Logger
		keyvals []interface{}
	}
)

// NewLogger returns a goa log adpater backed by a log logger.
func NewLogger(logger *log.Logger) LogAdapter {
	return &adapter{Logger: logger}
}

// Logger returns the logger stored in the context if any, nil otherwise.
func Logger(ctx context.Context) *log.Logger {
	logger := ContextLogger(ctx)
	if a, ok := logger.(*adapter); ok {
		return a.Logger
	}
	return nil
}

func (a *adapter) Info(msg string, keyvals ...interface{}) {
	a.logit(msg, keyvals, false)
}

func (a *adapter) Error(msg string, keyvals ...interface{}) {
	a.logit(msg, keyvals, true)
}

func (a *adapter) New(keyvals ...interface{}) LogAdapter {
	if len(keyvals) == 0 {
		return a
	}
	kvs := append(a.keyvals, keyvals...)
	if len(kvs)%2 != 0 {
		kvs = append(kvs, ErrMissingLogValue)
	}
	return &adapter{
		Logger: a.Logger,
		// Limiting the capacity of the stored keyvals ensures that a new
		// backing array is created if the slice must grow.
		keyvals: kvs[:len(kvs):len(kvs)],
	}
}

func (a *adapter) logit(msg string, keyvals []interface{}, iserror bool) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, ErrMissingLogValue)
	}
	m := (len(a.keyvals) + 1) / 2
	n += m
	var fm bytes.Buffer
	lvl := "INFO"
	if iserror {
		lvl = "EROR" // Not a typo. It ensures all level strings are 4-chars long.
	}
	fm.WriteString(fmt.Sprintf("[%s] %s", lvl, msg))
	vals := make([]interface{}, n)
	offset := len(a.keyvals)
	for i := 0; i < offset; i += 2 {
		k := a.keyvals[i]
		v := a.keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2+offset/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
}

// LogInfo extracts the logger from the given context and calls Info on it.
// This is intended for code that needs portable logging such as the internal code of goa and
// middleware. User code should use the log adapters instead.
func LogInfo(ctx context.Context, msg string, keyvals ...interface{}) {
	if l := ctx.Value(logKey); l != nil {
		if logger, ok := l.(LogAdapter); ok {
			logger.Info(msg, keyvals...)
		}
	}
}

// LogError extracts the logger from the given context and calls Error on it.
// This is intended for code that needs portable logging such as the internal code of goa and
// middleware. User code should use the log adapters instead.
func LogError(ctx context.Context, msg string, keyvals ...interface{}) {
	if l := ctx.Value(logKey); l != nil {
		if logger, ok := l.(LogAdapter); ok {
			logger.Error(msg, keyvals...)
		}
	}
}
