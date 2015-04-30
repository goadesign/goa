package design

import (
	"fmt"
	"runtime"
)

var (
	ctxStack contextStack // Global DSL evaluation stack
	dslError error        // Last evaluation error if any
)

// DSL evaluation contexts stack
type contextStack []interface{}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) Current() interface{} {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// executeDSL runs DSL in given evaluation context and returns true if successful.
// It initializes dslError in case of failure (and returns false).
func executeDSL(dsl func(), ctx interface{}) bool {
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return dslError == nil
}

// incompatibleDsl should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDsl() {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		dslError = "invalid definition"
		return
	}
	dslFunc := runtime.FuncForPC(pc).Name()
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		dslError = fmt.Errorf("invalid use of %s", dslFunc)
		return
	}
	dslError = fmt.Errorf("Invalid use of %s in %s:%d", dslFunc, file, line)
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition() (*ActionDefinition, bool) {
	a, ok := ctxStack.Current().(*ActionDefinition)
	if !ok {
		incompatibleDsl()
	}
	return a, ok
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition() (*APIDefinition, bool) {
	a, ok := ctxStack.Current().(*APIDefinition)
	if !ok {
		incompatibleDsl()
	}
	return a, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attribute() (*Attribute, bool) {
	a, ok := ctxStack.Current().(*Attribute)
	if !ok {
		incompatibleDsl()
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition() (*ResourceDefinition, bool) {
	r, ok := ctxStack.Current().(*ResourceDefinition)
	if !ok {
		incompatibleDsl()
	}
	return r, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition() (*ResponseDefinition, bool) {
	r, ok := ctxStack.Current().(*ResponseDefinition)
	if !ok {
		incompatibleDsl()
	}
	return r, ok
}
