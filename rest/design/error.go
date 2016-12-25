package design

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

type (
	// HTTPErrorExpr defines a HTTP error response including its name,
	// status, headers and media type.
	HTTPErrorExpr struct {
		// ErrorExpr is the underlying goa design error expression.
		*design.ErrorExpr
		// HTTP status
		Status int
		// Headers maps the ErrorExpr type attribues to HTTP headers.
		// Each entry is of the form "attribute name:header name". If the
		// : is omitted then the string defines both the attribute and
		// header name.
		Headers []string
		// Fields maps the ErrorExpr type attributes to HTTP body fields.
		// The mapping syntax is the same as the one used by Headers.
		Fields []string
		// Parent resource or action
		Parent eval.Expression
	}
)

// EvalName returns the generic definition name used in error messages.
func (r *HTTPErrorExpr) EvalName() string {
	return "HTTP error " + r.Name
}
