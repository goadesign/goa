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

// Log a message at the given level with context key/value pairs
func Debug(msg string, ctx ...interface{}) { Log.Debug(msg, ctx...) }
func Info(msg string, ctx ...interface{})  { Log.Info(msg, ctx...) }
func Warn(msg string, ctx ...interface{})  { Log.Warn(msg, ctx...) }
func Error(msg string, ctx ...interface{}) { Log.Error(msg, ctx...) }
func Crit(msg string, ctx ...interface{})  { Log.Crit(msg, ctx...) }
