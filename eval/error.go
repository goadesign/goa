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

// normalizeFileForPackageMatch strips @version segments from module cache paths
// so that package matching works regardless of where the module is cached.
// For example: ".../goa/v3@v3.23.2/dsl/..." becomes ".../goa/v3/dsl/...".
func normalizeFileForPackageMatch(file string) string {
	file = filepath.ToSlash(file)
	parts := strings.Split(file, "/")
	for i, p := range parts {
		if at := strings.IndexByte(p, '@'); at >= 0 {
			parts[i] = p[:at]
		}
	}
	return strings.Join(parts, "/")
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go" or one of the DSL package paths.
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
func computeErrorLocation() (file string, line int) {
	skipFunc := func(pc uintptr, file string) bool {
		if strings.HasSuffix(file, "_test.go") { // Be nice with tests
			return false
		}
		file = filepath.ToSlash(file)
		fn := runtime.FuncForPC(pc)
		var name string
		if fn != nil {
			name = fn.Name()
		}
		normalized := normalizeFileForPackageMatch(file)
		for _, pkg := range Context.dslPackages {
			if strings.Contains(file, pkg) || strings.Contains(normalized, pkg) || strings.Contains(name, pkg) {
				return true
			}
		}
		return false
	}
	depth := 3
	pc, file, line, _ := runtime.Caller(depth)
	for skipFunc(pc, file) {
		depth++
		pc, file, line, _ = runtime.Caller(depth)
	}
	wd, err := os.Getwd()
	if err != nil {
		return file, line
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return file, line
	}
	f, err := filepath.Rel(wd, file)
	if err != nil {
		return file, line
	}
	file = f
	return file, line
}
