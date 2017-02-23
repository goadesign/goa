package main

import (
	"context"

	"github.com/go-kit/kit/log"
)

// Logger wraps a go-kit logger and makes it compatible with goa's Logger
// interface.
type Logger struct {
	log.Logger
}

// Info logs informational messages.
func (l *Logger) Info(_ context.Context, keyvals ...interface{}) {
	kv := append([]interface{}{"lvl", "info"}, keyvals...)
	l.Logger.Log(kv...)
}

// Info logs errors.
func (l *Logger) Error(_ context.Context, keyvals ...interface{}) {
	kv := append([]interface{}{"lvl", "error"}, keyvals...)
	l.Logger.Log(kv...)
}
