/*
Package goalogrus contains an adapter that makes it possible to configure goa so it uses logrus
as logger backend.
Usage:

    logger := logrus.New()
    // Initialize logger handler using logrus package
    service.WithLogger(goalogrus.New(logger))
    // ... Proceed with configuring and starting the goa service

    // In handlers:
    goalogrus.Entry(ctx).Info("foo", "bar")
*/
package goalogrus

import (
	"fmt"

	"context"

	"github.com/Sirupsen/logrus"
	"github.com/goadesign/goa"
)

// adapter is the logrus goa logger adapter.
type adapter struct {
	*logrus.Entry
}

// New wraps a logrus logger into a goa logger.
func New(logger *logrus.Logger) goa.LogAdapter {
	return FromEntry(logrus.NewEntry(logger))
}

// FromEntry wraps a logrus log entry into a goa logger.
func FromEntry(entry *logrus.Entry) goa.LogAdapter {
	return &adapter{Entry: entry}
}

// Entry returns the logrus log entry stored in the given context if any, nil otherwise.
func Entry(ctx context.Context) *logrus.Entry {
	logger := goa.ContextLogger(ctx)
	if a, ok := logger.(*adapter); ok {
		return a.Entry
	}
	return nil
}

// Info logs messages using logrus.
func (a *adapter) Info(msg string, data ...interface{}) {
	a.Entry.WithFields(data2rus(data)).Info(msg)
}

// Error logs errors using logrus.
func (a *adapter) Error(msg string, data ...interface{}) {
	a.Entry.WithFields(data2rus(data)).Error(msg)
}

// New creates a new logger given a context.
func (a *adapter) New(data ...interface{}) goa.LogAdapter {
	return &adapter{Entry: a.Entry.WithFields(data2rus(data))}
}

func data2rus(keyvals []interface{}) logrus.Fields {
	n := (len(keyvals) + 1) / 2
	res := make(logrus.Fields, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = goa.ErrMissingLogValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}
