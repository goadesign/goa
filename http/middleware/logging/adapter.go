package logging

import (
	"bytes"
	"context"
	"fmt"
	"log"
)

type (
	// Adapter is the logging interface used by the middleware to produce
	// log entries.
	Adapter interface {
		// Info logs informational messages.
		Info(ctx context.Context, keyvals ...interface{})
		// Error logs error messages.
		Error(ctx context.Context, keyvals ...interface{})
	}

	// adapter is a thin wrapper around the stdlib logger that adapts it to
	// the Adapter interface.
	adapter struct {
		*log.Logger
	}
)

// Adapt creates a Adapter backed by a stdlib logger.
func Adapt(l *log.Logger) Adapter {
	return &adapter{l}
}

func (a *adapter) Info(_ context.Context, keyvals ...interface{})  { a.log("INFO", keyvals...) }
func (a *adapter) Error(_ context.Context, keyvals ...interface{}) { a.log("ERROR", keyvals...) }

func (a *adapter) log(lvl string, keyvals ...interface{}) {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
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
