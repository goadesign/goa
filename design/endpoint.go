package design

import (
	"fmt"

	"github.com/goadesign/goa/eval"
)

type (
	// EndpointExpr defines a single endpoint.
	EndpointExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of endpoint.
		Name string
		// Description of endpoint for consumption by humans.
		Description string
		// Request payload type.
		Request *AttributeExpr
		// Response payload type.
		Response *AttributeExpr
		// Service that owns endpoint.
		Service *ServiceExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
		Metadata map[string]string
	}
)

// EvalName returns the generic expression name used in error messages.
func (e *EndpointExpr) EvalName() string {
	var prefix, suffix string
	if e.Name != "" {
		suffix = fmt.Sprintf("endpoint %#v", e.Name)
	} else {
		suffix = "unnamed endpoint"
	}
	if e.Service != nil {
		prefix = e.Service.EvalName() + " "
	}
	return prefix + suffix
}
