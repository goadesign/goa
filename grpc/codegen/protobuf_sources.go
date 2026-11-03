// This file separates protobuf source planning from installed tool execution.
// Both constructors supply real finalizers to the same retained wire plan.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/expr"
)

// protobufFinalizer compiles one completed protobuf source file with the
// command and include paths captured by its planning constructor.
type protobufFinalizer func(string) error

// planProtobufSourceFinalizers copies and validates compiler declarations
// without inspecting the filesystem or executing installed programs.
func planProtobufSourceFinalizers(inputs []PlanInput) (map[*expr.GRPCServiceExpr]protobufFinalizer, error) {
	finalizers := make(map[*expr.GRPCServiceExpr]protobufFinalizer)
	for _, input := range inputs {
		for _, service := range input.Root.API.GRPC.Services {
			command := protobufCompilerCommand(input.Root, service)
			if err := validateProtobufCommand(command); err != nil {
				return nil, err
			}
			includes := append([]string{}, service.ServiceExpr.Meta["protoc:include"]...)
			includes = append(includes, input.Root.API.Meta["protoc:include"]...)
			finalizers[service] = sourceProtobufFinalizer(command, includes)
		}
	}
	return finalizers, nil
}

// capturedProtobufFinalizer retains the eagerly checked absolute programs used
// by NewPlans, so later PATH changes cannot choose another compiler or plugin.
func capturedProtobufFinalizer(tools *protobufToolPlan) protobufFinalizer {
	return func(path string) error {
		return runProtoc(tools, path)
	}
}

// sourceProtobufFinalizer resolves programs only when the caller explicitly
// renders a source-planned file. Source-only inspection never invokes it.
func sourceProtobufFinalizer(command, includes []string) protobufFinalizer {
	return func(path string) error {
		resolver := systemProtobufTools()
		goPlugin, err := resolveProtobufPlugin(resolver, protocGenGoName, protocGenGoVersion)
		if err != nil {
			return err
		}
		grpcPlugin, err := resolveProtobufPlugin(resolver, protocGenGoGRPCName, protocGenGoGRPCVersion)
		if err != nil {
			return err
		}
		compiler, err := resolver.resolve(command[0])
		if err != nil {
			return fmt.Errorf("resolve protobuf compiler %q: %w", command[0], err)
		}
		resolvedCommand := append([]string{compiler}, command[1:]...)
		return runProtoc(&protobufToolPlan{
			command: resolvedCommand, includes: includes, goPlugin: goPlugin, goGRPCPlugin: grpcPlugin,
		}, path)
	}
}

// validateProtobufCommand enforces authored command shape independently of
// whether an operation needs installed protobuf programs.
func validateProtobufCommand(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf(`Meta("protoc:cmd"): must be given arguments`)
	}
	if plugin := replacedGoPlugin(command[1:]); plugin != "" {
		return fmt.Errorf(`Meta("protoc:cmd") cannot replace required plugin %q`, plugin)
	}
	return nil
}
