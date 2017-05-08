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
	verr := new(eval.ValidationErrors)
	switch p := e.Response.Parent.(type) {
	case *ActionExpr:
		if p.EndpointExpr.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the endpoint", e.Name)
		}
	case *ResourceExpr:
		if p.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the service", e.Name)
		}
	case *RootExpr:
		if design.Root.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the API", e.Name)
		}
	}
	return verr
}

// Finalize looks up the corresponding endpoint error expression.
func (e *HTTPErrorExpr) Finalize() {
	var ee *design.ErrorExpr
	switch p := e.Response.Parent.(type) {
	case *ActionExpr:
		ee = p.EndpointExpr.Error(e.Name)
	case *ResourceExpr:
		ee = p.Error(e.Name)
	case *RootExpr:
		ee = design.Root.Error(e.Name)
	}
	e.ErrorExpr = ee
}

// Dup creates a copy of the error expression.
func (e *HTTPErrorExpr) Dup() *HTTPErrorExpr {
	return &HTTPErrorExpr{
		ErrorExpr: e.ErrorExpr,
		Name:      e.Name,
		Response:  e.Response.Dup(),
	}
}
