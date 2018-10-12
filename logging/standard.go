package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type stdlogger struct {
	prefix string
	logger *log.Logger
}

// NewStdLogger returns the adapted standard logger
func NewStdLogger(out io.Writer, prefix string, flag int) Logger {

	return &stdlogger{logger: log.New(out, prefix, flag), prefix: prefix}
}

func (l *stdlogger) Debug(args ...interface{}) {
	l.write("DEBUG", args...)
}

func (l *stdlogger) Info(args ...interface{}) {
	l.write("INFO ", args...)
}

func (l *stdlogger) Warn(args ...interface{}) {
	l.write("WARN ", args...)
}

func (l *stdlogger) Error(args ...interface{}) {
	l.write("ERROR", args...)
}

func (l *stdlogger) Fatal(args ...interface{}) {
	l.write("FATAL", args...)
}

func (l *stdlogger) Debugf(template string, args ...interface{}) {
	l.writef("DEBUG", template, args...)
}

func (l *stdlogger) Infof(template string, args ...interface{}) {
	l.writef("INFO ", template, args...)
}

func (l *stdlogger) Warnf(template string, args ...interface{}) {
	l.writef("WARN ", template, args...)
}

func (l *stdlogger) Errorf(template string, args ...interface{}) {
	l.writef("ERROR", template, args...)
}

func (l *stdlogger) Fatalf(template string, args ...interface{}) {
	l.writef("FATAL", template, args...)
}

func (l *stdlogger) write(level string, args ...interface{}) {
	var fm bytes.Buffer
	fm.WriteString(fmt.Sprintf("%s %s ", l.prefix, level))
	l.logger.SetPrefix(fm.String())
	l.logger.Println(args...)
}

func (l *stdlogger) writef(level string, template string, args ...interface{}) {
	var fm bytes.Buffer
	fm.WriteString(fmt.Sprintf("%s %s ", l.prefix, level))
	l.logger.SetPrefix(fm.String())
	l.logger.Printf(template, args...)
}
