// This file builds the HTTP, gRPC, and JSON-RPC files for every service. Each
// generated file lists the Go packages that it uses.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// transportFiles returns all HTTP, gRPC, and JSON-RPC files for one run.
func transportFiles(plan *Plan) ([]*codegen.File, error) {
	var files []*codegen.File
	generation := plan.Generation()
	designRoots := serviceRoots(generation.Roots())
	for _, r := range designRoots {
		services := plan.Service(r).Services()
		// HTTP
		if httpPlan := plan.http[r]; httpPlan != nil {
			files = append(files, httpPlan.ServerFiles()...)
			files = append(files, httpPlan.ClientFiles()...)
			files = append(files, httpPlan.ServerTypeFiles()...)
			files = append(files, httpPlan.ClientTypeFiles()...)
			files = append(files, httpPlan.PathFiles()...)
			files = append(files, httpPlan.ClientCLIFiles()...)
		}

		// GRPC
		if plan.grpc != nil {
			grpcServices := grpccodegen.NewServicesData(services, plan.grpc)
			files = append(files, grpccodegen.ProtoFiles(grpcServices)...)
			files = append(files, grpccodegen.ServerFiles(grpcServices)...)
			files = append(files, grpccodegen.ClientFiles(grpcServices)...)
			files = append(files, grpccodegen.ServerTypeFiles(grpcServices)...)
			files = append(files, grpccodegen.ClientTypeFiles(grpcServices)...)
			files = append(files, grpccodegen.ClientCLIFiles(grpcServices)...)
		}

		// JSON-RPC
		if jsonrpcPlan := plan.jsonrpc[r]; jsonrpcPlan != nil {
			files = append(files, jsonrpcPlan.ServerFiles()...)
			files = append(files, jsonrpcPlan.ClientFiles()...)
			files = append(files, jsonrpcPlan.ServerTypeFiles()...)
			files = append(files, jsonrpcPlan.ClientTypeFiles()...)
			files = append(files, jsonrpcPlan.PathFiles()...)
			files = append(files, jsonrpcPlan.ClientCLIFiles()...)
		}
	}
	return files, nil
}

// planTransportData chooses all Go package, import, type, and function names
// before generated files use them.
func planTransportData(plan *Plan) error {
	if err := planServiceData(plan); err != nil {
		return err
	}
	if plan.transportDone {
		return nil
	}
	generation := plan.Generation()
	if err := example.Plan(generation); err != nil {
		return err
	}
	roots := serviceRoots(generation.Roots())
	if err := planHTTPTransports(plan, roots); err != nil {
		return err
	}
	if err := planJSONRPCTransports(plan, roots); err != nil {
		return err
	}
	var hasGRPC bool
	for _, root := range roots {
		hasGRPC = hasGRPC || len(root.API.GRPC.Services) > 0
	}
	if hasGRPC {
		inputs := make([]grpccodegen.PlanInput, len(roots))
		for index, root := range roots {
			inputs[index] = grpccodegen.PlanInput{Root: root, Service: plan.Service(root)}
		}
		grpcPlan, err := grpccodegen.Plan(generation, inputs...)
		if err != nil {
			return err
		}
		plan.grpc = grpcPlan
	}
	plan.transportDone = true
	return nil
}

// planHTTPTransports prepares every HTTP service together. When two services
// write to the same Go package, Goa gives their types and functions different names.
func planHTTPTransports(plan *Plan, roots []*expr.RootExpr) error {
	var inputs []httpcodegen.PlanInput
	var plannedRoots []*expr.RootExpr
	for _, root := range roots {
		if len(root.API.HTTP.Services) == 0 {
			continue
		}
		inputs = append(inputs, httpcodegen.PlanInput{Root: root, Service: plan.Service(root)})
		plannedRoots = append(plannedRoots, root)
	}
	if len(inputs) == 0 {
		return nil
	}
	plans, err := httpcodegen.NewPlans(plan.Generation(), inputs...)
	if err != nil {
		return err
	}
	plan.http = make(map[*expr.RootExpr]*httpcodegen.Plan, len(plans))
	for index, root := range plannedRoots {
		plan.http[root] = plans[index]
	}
	return nil
}

// planJSONRPCTransports prepares the HTTP request and response types used by
// JSON-RPC. It then gives those values to the JSON-RPC generator so function
// definitions and calls use the same Go names.
func planJSONRPCTransports(plan *Plan, roots []*expr.RootExpr) error {
	var inputs []httpcodegen.PlanInput
	var plannedRoots []*expr.RootExpr
	for _, root := range roots {
		if len(root.API.JSONRPC.Services) == 0 {
			continue
		}
		inputs = append(inputs, httpcodegen.PlanInput{Root: root, Service: plan.Service(root)})
		plannedRoots = append(plannedRoots, root)
	}
	if len(inputs) == 0 {
		return nil
	}
	httpPlans, err := httpcodegen.NewJSONRPCPlans(plan.Generation(), inputs...)
	if err != nil {
		return err
	}
	jsonrpcInputs := make([]jsonrpccodegen.PlanInput, len(inputs))
	plan.jsonrpcHTTP = make(map[*expr.RootExpr]*httpcodegen.Plan, len(httpPlans))
	for index, root := range plannedRoots {
		plan.jsonrpcHTTP[root] = httpPlans[index]
		jsonrpcInputs[index] = jsonrpccodegen.PlanInput{
			Root:            root,
			Service:         plan.Service(root),
			HTTP:            httpPlans[index],
			ApplicationHTTP: plan.http[root],
		}
	}
	jsonrpcPlans, err := jsonrpccodegen.NewPlans(plan.Generation(), jsonrpcInputs...)
	if err != nil {
		return err
	}
	plan.jsonrpc = make(map[*expr.RootExpr]*jsonrpccodegen.Plan, len(jsonrpcPlans))
	for index, root := range plannedRoots {
		plan.jsonrpc[root] = jsonrpcPlans[index]
	}
	return nil
}
