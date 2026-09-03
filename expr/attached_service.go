// This file checks and finishes services that generators add after the design
// DSL has run.
package expr

import (
	"fmt"
	"slices"

	"goa.design/goa/v3/eval"
)

// EvaluateAttachedServices prepares, checks, and finishes services that a
// generator added to r. The services and types must already belong to r. No
// service is finished unless every added expression is valid.
func (r *RootExpr) EvaluateAttachedServices(services []*ServiceExpr, types ...UserType) error {
	sets, err := r.attachedServiceExpressions(services, types)
	if err != nil {
		return err
	}
	prepareExpressions(sets)
	if err := validateExpressions(r, sets); err != nil {
		return err
	}
	finalizeExpressions(sets)
	return nil
}

// attachedServiceExpressions verifies that each service and type belongs to r
// and that every endpoint points to a method on its service. It returns the
// expressions in the order that Prepare, Validate, and Finalize must run.
func (r *RootExpr) attachedServiceExpressions(
	services []*ServiceExpr,
	types []UserType,
) ([]eval.ExpressionSet, error) {
	selected := make(map[*ServiceExpr]struct{}, len(services))
	methods := make(eval.ExpressionSet, 0)
	for _, service := range services {
		if !slices.Contains(r.Services, service) {
			return nil, fmt.Errorf("service %q is not part of this design", service.Name)
		}
		if r.Service(service.Name) != service {
			return nil, fmt.Errorf("service name %q is already used in this design", service.Name)
		}
		if _, ok := selected[service]; ok {
			return nil, fmt.Errorf("service %q was provided more than once", service.Name)
		}
		selected[service] = struct{}{}
		for _, method := range service.Methods {
			if method.Service != service {
				return nil, fmt.Errorf("method %q belongs to a different service", method.Name)
			}
			methods = append(methods, method)
		}
	}

	typeExpressions := make(eval.ExpressionSet, len(types))
	for index, userType := range types {
		if !slices.Contains(r.Types, userType) {
			return nil, fmt.Errorf("type %q is not part of this design", userType.Name())
		}
		typeExpressions[index] = userType.Attribute()
	}

	var (
		httpServices, httpEndpoints, httpFileServers          eval.ExpressionSet
		jsonrpcServices, jsonrpcEndpoints, jsonrpcFileServers eval.ExpressionSet
		grpcServices, grpcEndpoints                           eval.ExpressionSet
		err                                                   error
	)
	if r.API.HTTP != nil {
		httpServices, httpEndpoints, httpFileServers, err = collectHTTPExpressions(
			r.API.HTTP.Services,
			r.API.HTTP,
			selected,
		)
		if err != nil {
			return nil, err
		}
	}
	if r.API.JSONRPC != nil {
		jsonrpcServices, jsonrpcEndpoints, jsonrpcFileServers, err = collectHTTPExpressions(
			r.API.JSONRPC.Services,
			&r.API.JSONRPC.HTTPExpr,
			selected,
		)
		if err != nil {
			return nil, err
		}
	}
	if r.API.GRPC != nil {
		grpcServices, grpcEndpoints, err = collectGRPCExpressions(r.API.GRPC.Services, selected)
		if err != nil {
			return nil, err
		}
	}

	for service := range selected {
		service.design = r
	}
	return []eval.ExpressionSet{
		typeExpressions,
		eval.ToExpressionSet(services),
		methods,
		httpServices,
		httpEndpoints,
		httpFileServers,
		jsonrpcServices,
		jsonrpcEndpoints,
		jsonrpcFileServers,
		grpcServices,
		grpcEndpoints,
	}, nil
}

// collectHTTPExpressions returns the selected HTTP services, endpoints, and
// file servers. It rejects a child that points to another service.
func collectHTTPExpressions(
	transports []*HTTPServiceExpr,
	httpRoot *HTTPExpr,
	selected map[*ServiceExpr]struct{},
) (eval.ExpressionSet, eval.ExpressionSet, eval.ExpressionSet, error) {
	services := make(eval.ExpressionSet, 0)
	endpoints := make(eval.ExpressionSet, 0)
	fileServers := make(eval.ExpressionSet, 0)
	for _, transport := range transports {
		if _, ok := selected[transport.ServiceExpr]; !ok {
			continue
		}
		if transport.Root != httpRoot {
			return nil, nil, nil, fmt.Errorf("HTTP service %q uses a different design", transport.Name())
		}
		services = append(services, transport)
		for _, endpoint := range transport.HTTPEndpoints {
			if endpoint.Service != transport {
				return nil, nil, nil, fmt.Errorf(
					"HTTP endpoint %q belongs to a different HTTP service",
					endpoint.Name(),
				)
			}
			if !slices.Contains(transport.ServiceExpr.Methods, endpoint.MethodExpr) {
				return nil, nil, nil, fmt.Errorf(
					"HTTP endpoint %q uses a method outside service %q",
					endpoint.Name(),
					transport.Name(),
				)
			}
			endpoints = append(endpoints, endpoint)
		}
		for _, fileServer := range transport.FileServers {
			if fileServer.Service != transport {
				return nil, nil, nil, fmt.Errorf(
					"HTTP file server %q belongs to a different HTTP service",
					fileServer.FilePath,
				)
			}
			fileServers = append(fileServers, fileServer)
		}
	}
	return services, endpoints, fileServers, nil
}

// collectGRPCExpressions returns the selected gRPC services and endpoints and
// rejects any endpoint that points outside its service.
func collectGRPCExpressions(
	transports []*GRPCServiceExpr,
	selected map[*ServiceExpr]struct{},
) (eval.ExpressionSet, eval.ExpressionSet, error) {
	services := make(eval.ExpressionSet, 0)
	endpoints := make(eval.ExpressionSet, 0)
	for _, transport := range transports {
		if _, ok := selected[transport.ServiceExpr]; !ok {
			continue
		}
		services = append(services, transport)
		for _, endpoint := range transport.GRPCEndpoints {
			if endpoint.Service != transport {
				return nil, nil, fmt.Errorf(
					"gRPC endpoint %q belongs to a different gRPC service",
					endpoint.Name(),
				)
			}
			if !slices.Contains(transport.ServiceExpr.Methods, endpoint.MethodExpr) {
				return nil, nil, fmt.Errorf(
					"gRPC endpoint %q uses a method outside service %q",
					endpoint.Name(),
					transport.Name(),
				)
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return services, endpoints, nil
}

// prepareExpressions calls Prepare on each added expression in Goa's required
// order.
func prepareExpressions(sets []eval.ExpressionSet) {
	for _, set := range sets {
		for _, expression := range set {
			if preparer, ok := expression.(eval.Preparer); ok {
				preparer.Prepare()
			}
		}
	}
}

// validateExpressions checks the complete design and each added expression. It
// returns all errors together.
func validateExpressions(root *RootExpr, sets []eval.ExpressionSet) error {
	errors := new(eval.ValidationErrors)
	if err := root.Validate(); err != nil {
		errors.AddError(root, err)
	}
	for _, set := range sets {
		for _, expression := range set {
			if validator, ok := expression.(eval.Validator); ok {
				if err := validator.Validate(); err != nil {
					errors.AddError(expression, err)
				}
			}
		}
	}
	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}

// finalizeExpressions calls Finalize on each added expression after every
// check succeeds.
func finalizeExpressions(sets []eval.ExpressionSet) {
	for _, set := range sets {
		for _, expression := range set {
			if finalizer, ok := expression.(eval.Finalizer); ok {
				finalizer.Finalize()
			}
		}
	}
}
