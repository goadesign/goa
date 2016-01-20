package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/raphael/goa/design"
)

var (
	// Errors contains the DSL execution errors if any.
	Errors MultiError

	// Global DSL evaluation stack
	ctxStack contextStack
)

type (
	// MultiError collects all DSL errors. It implements error.
	MultiError []*Error

	// Error represents an error that occurred while running the API DSL.
	// It contains the name of the file and line number of where the error
	// occurred as well as the original Go error.
	Error struct {
		GoError error
		File    string
		Line    int
	}

	// DSL evaluation contexts stack
	contextStack []design.Definition
)

// RunDSL runs the given root definitions. It iterates over the definition sets multiple times to
// first execute the DSL, the validate the resulting definitions and finally finalize them.
// The executed DSL may append new roots to the Roots Design package variable to have them be
// executed (last) in the same run.
func RunDSL() error {
	if len(design.Roots) == 0 {
		return nil
	}
	Errors = nil

	executed := 0
	recursed := 0
	for executed < len(design.Roots) {
		recursed++
		start := executed
		executed = len(design.Roots)
		for _, root := range design.Roots[start:] {
			root.IterateSets(runSet)
		}
		if recursed > 100 {
			// Let's cross that bridge once we get there
			return fmt.Errorf("too many generated roots, infinite loop?")
		}
	}
	if Errors != nil {
		return Errors
	}
	for _, root := range design.Roots {
		root.IterateSets(validateSet)
	}
	if Errors != nil {
		return Errors
	}
	for _, root := range design.Roots {
		root.IterateSets(finalizeSet)
	}

	return nil
}

// ExecuteDSL runs the given DSL to initialize the given definition. It returns true on success.
// It returns false and appends to Errors on failure.
// Note that `RunDSL` takes care of calling `ExecuteDSL` on all definitions that implement Source.
// This function is intended for use by definitions that run the DSL at declaration time rather than
// store the DSL for execution by the engine (usually simple independent definitions).
// The DSL should use ReportError to record DSL execution errors.
func ExecuteDSL(dsl func(), def design.Definition) bool {
	if dsl == nil {
		return true
	}
	initCount := len(Errors)
	ctxStack = append(ctxStack, def)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(Errors) <= initCount
}

// CurrentDefinition returns the definition whose initialization DSL is currently being executed.
func CurrentDefinition() design.Definition {
	return ctxStack.Current()
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) Current() design.Definition {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// ReportError records a DSL error for reporting post DSL execution.
func ReportError(fm string, vals ...interface{}) {
	var suffix string
	if cur := ctxStack.Current(); cur != nil {
		if ctx := cur.Context(); ctx != "" {
			suffix = fmt.Sprintf(" in %s", ctx)
		}
	} else {
		suffix = " (top level)"
	}
	err := fmt.Errorf(fm+suffix, vals...)
	file, line := computeErrorLocation()
	Errors = append(Errors, &Error{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// Error returns the error message.
func (m MultiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (de *Error) Error() (res string) {
	if err := de.GoError; err != nil {
		res = fmt.Sprintf("[%s:%d] %s", de.File, de.Line, err.Error())
	}
	return
}

// incompatibleDSL should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDSL(dslFunc string) {
	elems := strings.Split(dslFunc, ".")
	ReportError("invalid use of %s", elems[len(elems)-1])
}

// invalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func invalidArgError(expected string, actual interface{}) {
	ReportError("cannot use %#v (type %s) as type %s",
		actual, reflect.TypeOf(actual), expected)
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go".
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
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

// runSet executes the DSL for all definitions in the given set. The definition DSLs may append to
// the set as they execute.
func runSet(set design.DefinitionSet) error {
	executed := 0
	recursed := 0
	for executed < len(set) {
		recursed++
		for _, def := range set[executed:] {
			executed++
			if source, ok := def.(design.Source); ok {
				ExecuteDSL(source.DSL(), source)
			}
		}
		if recursed > 100 {
			return fmt.Errorf("too many generated definitions, infinite loop?")
		}
	}
	return nil
}

// validateSet runs the validation on all the set definitions that define one.
func validateSet(set design.DefinitionSet) error {
	for _, def := range set {
		if validate, ok := def.(design.Validate); ok {
			validate.Validate()
		}
	}
	return nil
}

// finalizeSet runs the validation on all the set definitions that define one.
func finalizeSet(set design.DefinitionSet) error {
	for _, def := range set {
		if finalize, ok := def.(design.Finalize); ok {
			finalize.Finalize()
		}
	}
	return nil
}

// topLevelDefinition returns true if the currently evaluated DSL is a root
// DSL (i.e. is not being run in the context of another definition).
func topLevelDefinition(failItNotTopLevel bool) bool {
	top := ctxStack.Current() == nil
	if failItNotTopLevel && !top {
		incompatibleDSL(caller())
	}
	return top
}

// Name of calling function.
func caller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "<unknown>"
	}
	return runtime.FuncForPC(pc).Name()
}
