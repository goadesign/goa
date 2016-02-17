package goa

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
)

var (
	// Log is the logger used by goa to log informational and error messages.
	// The default logger logs to Stderr.
	Log Logger
)

// Initialize default logger
func init() {
	Log = &DefaultLogger{Logger: log.New(os.Stderr, "", log.LstdFlags)}
}

type (
	// Logger is the logger interface used by goa to log informational and error messages.
	// Adapters to different logging backends are provided in the logging package.
	Logger interface {
		// Info logs a message with optional contextual data.
		// The contextual data consists of key/value pairs (so the size of data is always an even number).
		Info(ctx context.Context, msg string, data ...KV)
		// Info logs a message with optional contextual data.
		// The contextual data consists of key/value pairs (so the size of data is always an even number).
		Error(ctx context.Context, msg string, data ...KV)
	}

	// KV is a key/value pair to be logged
	KV struct {
		// Key of pair
		Key string
		// Value of pair
		Value interface{}
	}

	// DefaultLogger is the default goa logger implementation
	DefaultLogger struct {
		*log.Logger
	}
)

// Info logs the given informational message and accompanying data.
func Info(ctx context.Context, msg string, data ...KV) {
	if Log != nil {
		Log.Info(ctx, msg, data...)
	}
}

// Error logs the given error message and accompanying data.
func Error(ctx context.Context, msg string, data ...KV) {
	if Log != nil {
		Log.Error(ctx, msg, data...)
	}
}

// Info logs informational messages such as service startup
func (l *DefaultLogger) Info(ctx context.Context, msg string, data ...KV) {
	data = append(LogContext(ctx), data...)
	format, v := data2fmt(msg, data...)
	l.Printf("[INFO] "+format, v...)
}

// Error logs unhandled errors
func (l *DefaultLogger) Error(ctx context.Context, msg string, data ...KV) {
	data = append(LogContext(ctx), data...)
	format, v := data2fmt(msg, data...)
	l.Printf("[ERROR] "+format, v...)
}

func data2fmt(msg string, data ...KV) (format string, v []interface{}) {
	format = msg
	v = make([]interface{}, len(data))
	for i := 0; i < len(data); i++ {
		format += fmt.Sprintf("\t%v: %%v", data[i].Key)
		v[i] = data[i].Value
	}
	return
}
