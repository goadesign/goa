package dsl

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"

	. "github.com/raphael/goa/design"
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
	if dsl == nil {
		return true
	}
	initCount := len(dslErrors)
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(dslErrors) <= initCount
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
		depth++
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
		fmt.Printf("%d: %d: %s\n", err.Line, err.Line, err.Error.Error())
	}
	os.Exit(1)
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition(failIfNotAction bool) (*ActionDefinition, bool) {
	a, ok := ctxStack.current().(*ActionDefinition)
	if !ok && failIfNotAction {
		incompatibleDsl(caller())
	}
	return a, ok
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition(failIfNotAPI bool) (*APIDefinition, bool) {
	a, ok := ctxStack.current().(*APIDefinition)
	if !ok && failIfNotAPI {
		incompatibleDsl(caller())
	}
	return a, ok
}

// mediaTypeDefinition returns true and current context if it is a MediaTypeDefinition,
// nil and false otherwise.
func mediaTypeDefinition(failIfNotMT bool) (*MediaTypeDefinition, bool) {
	a, ok := ctxStack.current().(*MediaTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDsl(caller())
	}
	return a, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition(failIfNotAttribute bool) (*AttributeDefinition, bool) {
	a, ok := ctxStack.current().(*AttributeDefinition)
	if !ok && failIfNotAttribute {
		incompatibleDsl(caller())
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition(failIfNotResource bool) (*ResourceDefinition, bool) {
	r, ok := ctxStack.current().(*ResourceDefinition)
	if !ok && failIfNotResource {
		incompatibleDsl(caller())
	}
	return r, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition(failIfNotResponse bool) (*ResponseDefinition, bool) {
	r, ok := ctxStack.current().(*ResponseDefinition)
	if !ok && failIfNotResponse {
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
