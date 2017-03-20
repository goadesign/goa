package rest

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
		// Name of error, we need a separate copy of the name to match it
		// up with the appropriate ErrorExpr.
		Name string
		// Response is the corresponding HTTP response.
		Response *HTTPResponseExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (e *HTTPErrorExpr) EvalName() string {
	return "HTTP error " + e.Name
}

// Validate makes sure there is a error expression that matches the HTTP error
// expression.
func (e *HTTPErrorExpr) Validate() *eval.ValidationErrors {
	var ee *design.ErrorExpr
	switch p := e.Response.Parent.(type) {
	case *ActionExpr:
		ee = p.EndpointExpr.Error(e.Name)
	case *ResourceExpr:
		ee = p.ServiceExpr.Error(e.Name)
	case *RootExpr:
		ee = design.Root.Error(e.Name)
	}
	if ee == nil {
		verr := new(eval.ValidationErrors)
		verr.Add(e, "Error %#v does not match an error defined in the endpoint, service or API", e.Name)
		return verr
	}
	e.ErrorExpr = ee
	return nil
}
