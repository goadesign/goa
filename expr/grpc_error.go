// This file binds reusable gRPC error-response policy to the concrete error
// returned by each endpoint method.
package expr

import (
	"goa.design/goa/v3/eval"
)

type (
	// GRPCErrorExpr defines a gRPC error response including its name,
	// status, and result type.
	GRPCErrorExpr struct {
		// ErrorExpr is the underlying goa design error expression.
		*ErrorExpr
		// Name of error to match it up with the appropriate ErrorExpr.
		Name string
		// Response is the corresponding gRPC response.
		Response *GRPCResponseExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (e *GRPCErrorExpr) EvalName() string {
	return "gRPC error " + e.Name
}

// Validate makes sure there is a error expression that matches the gRPC error
// expression.
func (e *GRPCErrorExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	switch p := e.Response.Parent.(type) {
	case *GRPCEndpointExpr:
		if p.MethodExpr.Error(e.Name) == nil {
			verr.Add(e, "Error %#v does not match an error defined in the method", e.Name)
		}
	case *GRPCServiceExpr:
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
func (e *GRPCErrorExpr) Finalize(a *GRPCEndpointExpr) {
	e.ErrorExpr = a.MethodExpr.Error(e.Name)
	e.Response.Finalize(a, e.AttributeExpr)
}

// mappedError returns the error declaration that owns this reusable gRPC
// response policy before the policy is applied to an endpoint method.
func (e *GRPCErrorExpr) mappedError() (*ErrorExpr, string) {
	switch parent := e.Response.Parent.(type) {
	case *GRPCEndpointExpr:
		return parent.MethodExpr.Error(e.Name), "method"
	case *GRPCServiceExpr:
		return parent.Error(e.Name), "service"
	case *GRPCExpr, *RootExpr:
		return Root.Error(e.Name), "API"
	}
	return nil, ""
}

// Dup creates a copy of the error expression.
func (e *GRPCErrorExpr) Dup() *GRPCErrorExpr {
	return &GRPCErrorExpr{
		ErrorExpr: e.ErrorExpr,
		Name:      e.Name,
		Response:  e.Response.Dup(),
	}
}
