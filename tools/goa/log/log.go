package log

import "gopkg.in/inconshreveable/log15.v2"

var (
	// Log is the logger used by the generator.
	// The default handler logs to STDOUT.
	Log log15.Logger
)

// Log to STDOUT by default.
func init() {
	Log = log15.New()
	Log.SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StdoutHandler))
}

// Debug logs a message at the debug level with context key/value pairs.
func Debug(msg string, ctx ...interface{}) { Log.Debug(msg, ctx...) }

// Info logs a message at the info level with context key/value pairs.
func Info(msg string, ctx ...interface{}) { Log.Info(msg, ctx...) }

// Warn logs a message at the warning level with context key/value pairs.
func Warn(msg string, ctx ...interface{}) { Log.Warn(msg, ctx...) }

// Error logs a message at the error level with context key/value pairs.
func Error(msg string, ctx ...interface{}) { Log.Error(msg, ctx...) }

// Crit logs a message at the critical level with context key/value pairs.
func Crit(msg string, ctx ...interface{}) { Log.Crit(msg, ctx...) }
