package goa

import (
	"bytes"
	"fmt"
	"log"
)

// ErrMissingLogValue is the value used to log keys with missing values
const ErrMissingLogValue = "MISSING"

type (
	// Logger is the interface used internally by goa for all log operations.
	Logger interface {
		// Log creates a new log entry with the given key value pairs.
		Log(keyvals ...interface{})
	}

	// logger is a thin wrapper around the stdlib logger that adapts it to
	// the Logger interface.
	logger struct {
		*log.Logger
	}
)

// AdaptLogger creates a logger that write log entries to the given stdlib logger.
func AdaptLogger(l *log.Logger) Logger {
	return &logger{l}
}

func (l *logger) Log(keyvals ...interface{}) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, ErrMissingLogValue)
	}
	var fm bytes.Buffer
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	l.Logger.Printf(fm.String(), vals...)
}
