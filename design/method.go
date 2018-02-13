package design

import (
	"fmt"

	"goa.design/goa/eval"
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
// endpoint then the service and finally the root expression.
func (m *MethodExpr) Error(name string) *ErrorExpr {
	for _, err := range m.Errors {
		if err.Name == name {
			return err
		}
	}
	return m.Service.Error(name)
}

// EvalName returns the generic expression name used in error messages.
func (m *MethodExpr) EvalName() string {
	var prefix, suffix string
	if m.Name != "" {
		suffix = fmt.Sprintf("method %#v", m.Name)
	} else {
		suffix = "unnamed method"
	}
	if m.Service != nil {
		prefix = m.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Validate validates the method payloads, results, and errors (if any).
func (m *MethodExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if m.Payload != nil {
		verr.Merge(m.Payload.Validate("payload", m))
	}
	if m.Result != nil {
		verr.Merge(m.Result.Validate("result", m))
	}
	for _, e := range m.Errors {
		if err := e.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
	}
	return verr
}

// Finalize makes sure the method payload and result types are set.
func (m *MethodExpr) Finalize() {
	if m.Payload == nil {
		m.Payload = &AttributeExpr{Type: Empty}
	}
	if m.Result == nil {
		m.Result = &AttributeExpr{Type: Empty}
	}
	for _, e := range m.Errors {
		e.Finalize()
	}
}
