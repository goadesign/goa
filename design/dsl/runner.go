package dsl

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	. "github.com/raphael/goa/design"
)

var (
	// DSLErrors contains the DSL execution errors if any.
	DSLErrors multiError

	// Global DSL evaluation stack
	ctxStack contextStack
)

type (
	// DSL evaluation contexts stack
	contextStack []DSLDefinition

	// multiError collects all DSL errors. It implements error.
	multiError []*dslError

	// A DSL error with name of the file and line number of where error occurred.
	dslError struct {
		GoError error
		File    string
		Line    int
	}
)

// Reset resets the runner internal evaluation stack with the given stack.
// Mainly useful for tests.
func Reset(stack []DSLDefinition) {
	ctxStack = stack
	DSLErrors = nil
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) current() DSLDefinition {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Error returns the error message.
func (m multiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (de *dslError) Error() (res string) {
	if err := de.GoError; err != nil {
		res = fmt.Sprintf("[%s:%d] %s", de.File, de.Line, err.Error())
	}
	return
}

// executeDSL runs DSL in given evaluation context and returns true if successful.
// It appends to DSLErrors in case of failure (and returns false).
func executeDSL(dsl func(), ctx DSLDefinition) bool {
	if dsl == nil {
		return true
	}
	initCount := len(DSLErrors)
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(DSLErrors) <= initCount
}

// incompatibleDsl should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDsl(dslFunc string) {
	elems := strings.Split(dslFunc, ".")
	var suffix string
	if ctxStack.current() != nil {
		suffix = fmt.Sprintf(" in %s", ctxStack.current().Context())
	}
	appendError(fmt.Errorf("Invalid use of %s%s", elems[len(elems)-1], suffix))
}

// invalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func invalidArgError(expected string, actual interface{}) {
	appendError(fmt.Errorf("cannot use %#v (type %s) as type %s in argument to Attribute",
		actual, reflect.TypeOf(actual), expected))
}

// appendError records a DSL error for reporting post DSL execution.
// TBD: REMOVE ANY CONTEXT FROM ERROR MESSAGES AND ADD CONTEXT GENERICALLY
// EITHER HERE OR BY ADDING SOME NEW CONSTRUCTS.
func appendError(err error) {
	file, line := computeErrorLocation()
	DSLErrors = append(DSLErrors, &dslError{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// computeErrorLocation implements a heuristic to find the location in the user code where the
// error occurred. It walks back the callstack until the file doesn't match "/goa/design/*.go".
// When successful it returns the file name and line number.
func computeErrorLocation() (file string, line int) {
	depth := 2
	_, file, line, _ = runtime.Caller(depth)
	ok := strings.HasSuffix(file, "_test.go") // Be nice with tests
	if !ok {
		nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
		ok = !nok
	}
	for !ok {
		depth++
		_, file, line, _ = runtime.Caller(depth)
		ok = strings.HasSuffix(file, "_test.go")
		if !ok {
			nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
			ok = !nok
		}
	}
	gopath := os.Getenv("GOPATH") + "/src/"
	if strings.HasPrefix(file, gopath) {
		file = file[len(gopath):]
	}
	return
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
	m, ok := ctxStack.current().(*MediaTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDsl(caller())
	}
	return m, ok
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
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "<unknown>"
	}
	return runtime.FuncForPC(pc).Name()
}
