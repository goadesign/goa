/*
Package goalogrus contains an adapter that makes it possible to configure goa so it uses logrus
as logger backend.
Usage:

    logger := logrus.New()
    // Initialize logger handler using logrus package
    service.WithLogger(goalogrus.New(logger))
    // ... Proceed with configuring and starting the goa service
*/
package goalogrus

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/goadesign/goa"
)

// Logger is the logrus goa adapter logger.
type Logger struct {
	*logrus.Entry
}

// New wraps a logrus logger into a goa logger.
func New(logger *logrus.Logger) goa.LogAdapter {
	return FromEntry(logrus.NewEntry(logger))
}

// FromEntry wraps a logrus log entry into a goa logger.
func FromEntry(entry *logrus.Entry) goa.LogAdapter {
	return &Logger{Entry: entry}
}

// Info logs messages using logrus.
func (l *Logger) Info(msg string, data ...interface{}) {
	l.Entry.WithFields(data2rus(data)).Info(msg)
}

// Error logs errors using logrus.
func (l *Logger) Error(msg string, data ...interface{}) {
	l.Entry.WithFields(data2rus(data)).Error(msg)
}

// New creates a new logger given a context.
func (l *Logger) New(data ...interface{}) goa.LogAdapter {
	return &Logger{Entry: l.Entry.WithFields(data2rus(data))}
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
