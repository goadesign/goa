package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type (
	// Error represents an error that occurred while evaluating the DSL.
	// It contains the name of the file and line number of where the error
	// occurred as well as the original Go error.
	Error struct {
		// GoError is the original error returned by the DSL function.
		GoError error
		// File is the path to the file containing the user code that
		// caused the error.
		File string
		// Line is the line number  that caused the error.
		Line int
	}

	// MultiError collects multiple DSL errors. It implements error.
	MultiError []*Error
)

// Error returns the error message.
func (m MultiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (e *Error) Error() string {
	if err := e.GoError; err != nil {
		if e.File == "" {
			return err.Error()
		}
		return fmt.Sprintf("[%s:%d] %s", e.File, e.Line, err.Error())
	}
	return ""
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go" or one of the DSL package paths.
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
func computeErrorLocation() (file string, line int) {
	skipFunc := func(file string) bool {
		if strings.HasSuffix(file, "_test.go") { // Be nice with tests
			return false
		}
		file = filepath.ToSlash(file)
		for _, pkg := range Context.dslPackages {
			if strings.Contains(file, pkg) {
				return true
			}
		}
		return false
	}
	depth := 3
	_, file, line, _ = runtime.Caller(depth)
	for skipFunc(file) {
		depth++
		_, file, line, _ = runtime.Caller(depth)
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return
	}
	f, err := filepath.Rel(wd, file)
	if err != nil {
		return
	}
	file = f
	return
}
