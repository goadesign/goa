package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// GRPCServiceExpr describes a gRPC service.
	GRPCServiceExpr struct {
		eval.DSLFunc
		// ServiceExpr is the service expression that backs this service.
		ServiceExpr *ServiceExpr
		// Name of parent service if any
		ParentName string
		// GRPCEndpoints is the list of service endpoints.
		GRPCEndpoints []*GRPCEndpointExpr
		// GRPCErrors lists gRPC errors that apply to all endpoints.
		GRPCErrors []*GRPCErrorExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator.
		Meta MetaExpr
	}
)

// Name of service (service)
func (svc *GRPCServiceExpr) Name() string {
	return svc.ServiceExpr.Name
}

// Description of service (service)
func (svc *GRPCServiceExpr) Description() string {
	return svc.ServiceExpr.Description
}

// Endpoint returns the service endpoint with the given name or nil if there
// isn't one.
func (svc *GRPCServiceExpr) Endpoint(name string) *GRPCEndpointExpr {
	for _, a := range svc.GRPCEndpoints {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// EndpointFor builds the endpoint for the given method.
func (svc *GRPCServiceExpr) EndpointFor(name string, m *MethodExpr) *GRPCEndpointExpr {
	if a := svc.Endpoint(name); a != nil {
		return a
	}
	a := &GRPCEndpointExpr{
		MethodExpr: m,
		Service:    svc,
	}
	svc.GRPCEndpoints = append(svc.GRPCEndpoints, a)
	return a
}

// Error returns the error with the given name.
func (svc *GRPCServiceExpr) Error(name string) *ErrorExpr {
	for _, erro := range svc.ServiceExpr.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return Root.Error(name)
}

// GRPCError returns the service gRPC error with given name if any.
func (svc *GRPCServiceExpr) GRPCError(name string) *GRPCErrorExpr {
	for _, erro := range svc.GRPCErrors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// EvalName returns the generic definition name used in error messages.
func (svc *GRPCServiceExpr) EvalName() string {
	if svc.Name() == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", svc.Name())
}

// Prepare initializes the error responses.
func (svc *GRPCServiceExpr) Prepare() {
	for _, er := range svc.GRPCErrors {
		er.Response.Prepare()
	}
}

// Validate makes sure the service is valid.
func (svc *GRPCServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	// Validate errors
	for _, er := range svc.GRPCErrors {
		verr.Merge(er.Validate())
	}
	for _, er := range Root.API.GRPC.Errors {
		// This may result in the same error being validated multiple
		// times however service is the top level expression being
		// walked and errors cannot be walked until all expressions have
		// run. Another solution could be to append a new dynamically
		// generated root that the eval engine would process after. Keep
		// things simple for now.
		verr.Merge(er.Validate())
	}
	return verr
}
