/*
Package goalogrus contains an adapter that makes it possible to configure goa so it uses logrus
as logger backend.
Usage:

    logger := logrus.New()
    // Initialize logger handler using logrus package
    service.UseLogger(goalogrus.New(logger))
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
	*logrus.Logger
}

// New wraps a logrus logger into a goa logger.
func New(logger *logrus.Logger) goa.Logger {
	return &Logger{Logger: logger}
}

// Info logs messages using logrus.
func (l *Logger) Info(msg string, data ...interface{}) {
	l.Logger.WithFields(data2rus(data)).Info(msg)
}

// Error logs errors using logrus.
func (l *Logger) Error(msg string, data ...interface{}) {
	l.Logger.WithFields(data2rus(data)).Error(msg)
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
