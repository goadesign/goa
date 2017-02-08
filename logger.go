package goa

import (
	"bytes"
	"golang.org/x/net/context"
	"fmt"
	"log"
)

// errMissingLogValue is the value used to log keys with missing values
const errMissingLogValue = "MISSING"

type (
	// Logger is the logging interface used by goa to produce log entries.
	Logger interface {
		// Info logs informational messages.
		Info(ctx context.Context, keyvals ...interface{})
		// Error logs error messages.
		Error(ctx context.Context, keyvals ...interface{})
	}

	// adapter is a thin wrapper around the stdlib logger that adapts it to
	// the Logger interface.
	adapter struct {
		*log.Logger
	}
)

// AdaptStdLogger creates a logger backed by a stdlib logger.
func AdaptStdLogger(l *log.Logger) Logger {
	return &adapter{l}
}

// Info logs an informational message.
func (a *adapter) Info(ctx context.Context, keyvals ...interface{}) {
	a.log("INFO", keyvals...)
}

// Error logs an error message.
func (a *adapter) Error(ctx context.Context, keyvals ...interface{}) {
	a.log("EROR", keyvals...)
}

// Log renders the log entries using the std lib logger.
func (a *adapter) log(lvl string, keyvals ...interface{}) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, errMissingLogValue)
	}
	var fm bytes.Buffer
	fm.WriteString("[%s]")
	vals := make([]interface{}, n+1)
	vals[0] = lvl
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2+1] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
}
