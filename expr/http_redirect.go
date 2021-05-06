package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// HTTPRedirectExpr defines an endpoint that replies to the request with a redirect.
	HTTPRedirectExpr struct {
		// URL is the URL that is being redirected to.
		URL string
		// StatusCode is the HTTP status code.
		StatusCode int
		// Parent expression, one of HTTPEndpointExpr or HTTPFileServerExpr.
		Parent eval.Expression
	}
)

// EvalName returns the generic definition name used in error messages.
func (r *HTTPRedirectExpr) EvalName() string {
	suffix := fmt.Sprintf("redirect to %s with status code %d", r.URL, r.StatusCode)
	var prefix string
	if r.Parent != nil {
		prefix = r.Parent.EvalName() + " "
	}
	return prefix + suffix
}
