package expr

import (
	"goa.design/goa/v3/eval"
)

type (
	// HTTPErrorExpr defines a HTTP error response including its name,
	// status, headers and result type.
	HTTPErrorExpr struct {
		// ErrorExpr is the underlying goa design error expression.
		*ErrorExpr
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
	case *HTTPEndpointExpr:
		if p.MethodExpr.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the method", e.Name)
		}
	case *HTTPServiceExpr:
		if p.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the service", e.Name)
		}
	case *RootExpr:
		if Root.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the API", e.Name)
		}
	}
	return verr
}

// Finalize looks up the corresponding method error expression.
func (e *HTTPErrorExpr) Finalize(a *HTTPEndpointExpr) {
	var ee *ErrorExpr
	switch p := e.Response.Parent.(type) {
	case *HTTPEndpointExpr:
		ee = p.MethodExpr.Error(e.Name)
	case *HTTPServiceExpr:
		ee = p.Error(e.Name)
	case *RootExpr:
		ee = Root.Error(e.Name)
	}
	e.ErrorExpr = ee
	e.Response.Finalize(a, e.AttributeExpr)
	if e.Response.Body == nil {
		e.Response.Body = httpErrorResponseBody(a, e)
	}

	// Initialize response content type if result is media type.
	if e.Response.Body.Type == Empty {
		return
	}
	if e.Response.ContentType != "" {
		return
	}
	mt, ok := e.Response.Body.Type.(*ResultTypeExpr)
	if !ok {
		return
	}
	e.Response.ContentType = mt.Identifier
}

// Dup creates a copy of the error expression.
func (e *HTTPErrorExpr) Dup() *HTTPErrorExpr {
	return &HTTPErrorExpr{
		ErrorExpr: e.ErrorExpr,
		Name:      e.Name,
		Response:  e.Response.Dup(),
	}
}
