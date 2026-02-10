package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	shouldSkip := func(file, name string) bool {
		if strings.HasSuffix(file, "_test.go") { // Be nice with tests
			return false
		}
		if isGoaSourceFile(file) {
			return true
		}
		file = filepath.ToSlash(file)
		normalized := normalizeFileForPackageMatch(file)
		for _, pkg := range Context.dslPackages {
			if strings.Contains(file, pkg) || strings.Contains(normalized, pkg) || strings.Contains(name, pkg) {
				return true
			}
		}
		return false
	}

	// Start scanning just above computeErrorLocation itself. This is robust to
	// inlining and avoids hardcoding assumptions about the exact call depth.
	const skip = 2
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if frame.File == "" || frame.Line == 0 {
			if !more {
				break
			}
			continue
		}
		if !shouldSkip(frame.File, frame.Function) {
			return relativeToWorkdir(frame.File), frame.Line
		}
		if !more {
			break
		}
	}
	return "", 0
}

// isGoaSourceFile reports whether file points into the Goa module sources.
//
// This is used to robustly skip internal Goa frames even when the runtime
// reports a call location (file:line) inside an inlined Goa function but the
// corresponding frame Function name does not include an import path.
func isGoaSourceFile(file string) bool {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return false
	}
	root := filepath.Dir(filepath.Dir(thisFile))
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// validationErrorLocation returns the location of the DSL that declared the
// given expression when available.
//
// The location is derived from the expression DSL function pointer and is used
// to annotate validation errors (i.e. errors returned by Validate()).
func validationErrorLocation(expr Expression) (file string, line int, ok bool) {
	source, ok := expr.(Source)
	if !ok {
		return "", 0, false
	}
	fn := source.DSL()
	if fn == nil {
		return "", 0, false
	}
	return dslFuncLocation(fn)
}

// dslFuncLocation returns the file and line where the given DSL function is
// declared.
//
// The returned file is relative to the current working directory when possible.
func dslFuncLocation(fn func()) (file string, line int, ok bool) {
	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "", 0, false
	}
	file, line = f.FileLine(pc)
	if file == "" || line == 0 {
		return "", 0, false
	}
	return relativeToWorkdir(file), line, true
}

// relativeToWorkdir returns file relative to the current working directory when
// possible, otherwise it returns file unchanged.
func relativeToWorkdir(file string) string {
	wd, err := os.Getwd()
	if err != nil {
		return file
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return file
	}
	rel, err := filepath.Rel(wd, file)
	if err != nil {
		return file
	}
	return rel
}
