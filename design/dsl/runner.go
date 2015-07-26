package dsl

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
)

var (
	ctxStack  contextStack // Global DSL evaluation stack
	dslErrors []*dslError  // DSL evaluation errors
)

// DSL evaluation contexts stack
type contextStack []interface{}

// A DSL error with name of the file and line number of where error occurred.
type dslError struct {
	Error error
	File  string
	Line  int
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) current() interface{} {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// executeDSL runs DSL in given evaluation context and returns true if successful.
// It appends to dslErrors in case of failure (and returns false).
func executeDSL(dsl func(), ctx interface{}) bool {
	errorCount := len(dslErrors)
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(dslErrors) > errorCount
}

// incompatibleDsl should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDsl(dslFunc string) {
	appendError(fmt.Errorf("Invalid use of %s", dslFunc))
}

// invalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func invalidArgError(expected string, actual interface{}) {
	appendError(fmt.Errorf("cannot use %v (type %s) as type %s in argument to Attribute",
		actual, reflect.TypeOf(actual), expected))
}

// appendError records a DSL error for reporting post DSL execution.
func appendError(err error) {
	file, line := computeErrorLocation()
	dslErrors = append(dslErrors, &dslError{
		Error: err,
		File:  file,
		Line:  line,
	})
}

// computeErrorLocation implements a heuristic to find the location in the user code where the
// error occurred. It walks back the callstack until the file doesn't match "/goa/design/*.go".
// When successful it returns the file name and line number.
func computeErrorLocation() (string, int) {
	depth := 2
	_, file, line, ok := runtime.Caller(depth)
	if ok {
		ok, _ = regexp.MatchString(`/goa/design/.+\.go$`, file)
	}
	for ok {
		depth += 1
		_, file, line, ok = runtime.Caller(depth)
		ok, _ = regexp.MatchString(`/goa/design/.+\.go$`, file)
	}
	if !ok {
		return "<unknown>", 0
	}
	return file, line
}

// reportErrors prints the DSL errors and exits the process.
func reportErrors() {
	for _, err := range dslErrors {
		fmt.Printf("%s: %d: %s\n", err.Line, err.Line, err.Error.Error())
	}
	os.Exit(1)
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition() (*design.ActionDefinition, bool) {
	a, ok := ctxStack.current().(*design.ActionDefinition)
	if !ok {
		incompatibleDsl(caller())
	}
	return a, ok
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition() (*design.APIDefinition, bool) {
	a, ok := ctxStack.current().(*design.APIDefinition)
	if !ok {
		incompatibleDsl(caller())
	}
	return a, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition() (*design.AttributeDefinition, bool) {
	a, ok := ctxStack.current().(*design.AttributeDefinition)
	if !ok {
		incompatibleDsl(caller())
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition() (*design.ResourceDefinition, bool) {
	r, ok := ctxStack.current().(*design.ResourceDefinition)
	if !ok {
		incompatibleDsl(caller())
	}
	return r, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition() (*design.ResponseDefinition, bool) {
	r, ok := ctxStack.current().(*design.ResponseDefinition)
	if !ok {
		incompatibleDsl(caller())
	}
	return r, ok
}

// Name of calling function.
func caller() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "<unknown>"
	}
	return runtime.FuncForPC(pc).Name()
}
