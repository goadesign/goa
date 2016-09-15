package design

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

type (
	// ResponseExpr defines a HTTP response status and optional validation rules.
	ResponseExpr struct {
		// Response name
		Name string
		// HTTP status
		Status int
		// Response description
		Description string
		// Response body type if any
		Type design.DataType
		// Response body media type if any
		MediaType string
		// Response view name if MediaType is MediaTypeExpr
		ViewName string
		// Response header expressions
		Headers *design.AttributeExpr
		// Parent action or resource
		Parent eval.Expression
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}

	// ResponseTemplateExpr defines a response template.
	// A response template is a function that takes an arbitrary number
	// of strings and returns a response expression.
	ResponseTemplateExpr struct {
		// Response template name
		Name string
		// Response template function
		Template func(params ...string) *ResponseExpr
	}

	// ResponseIterator is the type of functions given to IterateResponses.
	ResponseIterator func(r *ResponseDefinition) error
)
