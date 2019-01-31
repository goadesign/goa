package middleware

import (
	"bytes"
	"fmt"
	"log"
)

type (
	// Logger is the logging interface used by the middleware to produce
	// log entries.
	Logger interface {
		// Log creates a log entry using a sequence of alternating keys
		// and values.
		Log(keyvals ...interface{}) error
	}

	// adapter is a thin wrapper around the stdlib logger that adapts it to
	// the Logger interface.
	adapter struct {
		*log.Logger
	}
)

// NewLogger creates a Logger backed by a stdlib logger.
func NewLogger(l *log.Logger) Logger {
	return &adapter{l}
}

func (a *adapter) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
	}
	var fm bytes.Buffer
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
	return nil
}
