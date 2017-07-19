package design

import (
	"fmt"

	"goa.design/goa.v2/eval"
)

type (
	// MethodExpr defines a single method.
	MethodExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of method.
		Name string
		// Description of method for consumption by humans.
		Description string
		// Docs points to the method external documentation if any.
		Docs *DocsExpr
		// Payload attribute
		Payload *AttributeExpr
		// Result attribute
		Result *AttributeExpr
		// Errors lists the error responses.
		Errors []*ErrorExpr
		// Service that owns method.
		Service *ServiceExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
		Metadata MetadataExpr
	}
)

// Error returns the error with the given name. It looks up recursively in the
// enpoint then the service and finally the root expression.
func (e *MethodExpr) Error(name string) *ErrorExpr {
	for _, err := range e.Errors {
		if err.Name == name {
			return err
		}
	}
	return e.Service.Error(name)
}

// EvalName returns the generic expression name used in error messages.
func (e *MethodExpr) EvalName() string {
	var prefix, suffix string
	if e.Name != "" {
		suffix = fmt.Sprintf("method %#v", e.Name)
	} else {
		suffix = "unnamed method"
	}
	if e.Service != nil {
		prefix = e.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Finalize makes sure the method payload and result types are set.
func (e *MethodExpr) Finalize() {
	if e.Payload == nil {
		e.Payload = &AttributeExpr{Type: Empty}
	}
	if e.Result == nil {
		e.Result = &AttributeExpr{Type: Empty}
	}
}
