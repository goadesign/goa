package design

import (
	goa "github.com/goadesign/goa/design"
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
		Type DataType
		// Response body media type if any
		MediaType string
		// Response view name if MediaType is MediaTypeExpr
		ViewName string
		// Response header expressions
		Headers *AttributeExpr
		// Parent action or resource
		Parent eval.Expression
		// Metadata is a list of key/value pairs
		Metadata goa.MetadataExpr
		// Standard is true if the response expression comes from the goa default responses
		Standard bool
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
)
