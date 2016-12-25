package design

import (
	"fmt"

	"goa.design/goa.v2/eval"
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
		// Docs points to the endpoint external documentation if any.
		Docs *DocsExpr
		// Request type.
		Request UserType
		// Response type.
		Response UserType
		// Errors lists the error responses.
		Errors []*ErrorExpr
		// Service that owns endpoint.
		Service *ServiceExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
		Metadata MetadataExpr
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

// Finalize makes sure the endpoint request and response types are set.
func (e *EndpointExpr) Finalize() {
	if e.Request == nil {
		e.Request = Empty
	}
	if e.Response == nil {
		e.Response = Empty
	}
}
