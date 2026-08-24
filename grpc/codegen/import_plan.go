// This file records every import in the generated gRPC package that writes
// the reference. Package-local planning keeps an unrelated transport or
// executable from changing a gRPC qualifier.
package codegen

import (
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// planGRPCImports records imports for each generated client, server, and
// command-parser package before Generation.Freeze chooses their names.
func planGRPCImports(generation *codegen.Generation, plan *Plan) error {
	for _, servicePlan := range plan.servicesPlan {
		service := servicePlan.expression
		pathName := servicePlan.packages.pathName
		clientPath := path.Join(generation.GenPkg(), "grpc", pathName, "client")
		serverPath := path.Join(generation.GenPkg(), "grpc", pathName, "server")
		protobufPath := path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)

		client := generation.Package(clientPath)
		clientFixed := []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("strconv"),
			codegen.SimpleImport("unicode/utf8"),
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.GoaNamedImport("grpc/pb", "goapb"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/metadata"),
		}
		if len(service.GRPCEndpoints) > 0 {
			clientFixed = append(clientFixed,
				codegen.SimpleImport("encoding/json"),
				codegen.SimpleImport("fmt"),
			)
		}
		if servicePlan.usesAny {
			clientFixed = append(clientFixed, codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"))
		}
		if err := requirePackageImports(client, clientFixed); err != nil {
			return err
		}
		if err := reservePackageImports(client,
			servicePlan.packages.service,
			codegen.NewImport(pathName+"pb", protobufPath),
		); err != nil {
			return err
		}
		if grpcServiceHasViewedResult(service) {
			if err := client.ReserveGeneratedImport(servicePlan.packages.views); err != nil {
				return err
			}
		}
		if err := planGRPCAttributeImports(client, generation, grpcEndpointAttributes(service.GRPCEndpoints...)); err != nil {
			return err
		}
		if err := requirePackageImports(client, servicePlan.protoGoImports); err != nil {
			return err
		}

		server := generation.Package(serverPath)
		serverFixed := []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("errors"),
			codegen.SimpleImport("strconv"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("unicode/utf8"),
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/codes"),
			codegen.SimpleImport("google.golang.org/grpc/metadata"),
		}
		if grpcServiceStreamsPayload(service) {
			serverFixed = append(serverFixed, codegen.SimpleImport("io"))
		}
		if grpcResponseMetadataUsesAny(service) {
			serverFixed = append(serverFixed, codegen.SimpleImport("fmt"))
		}
		if servicePlan.usesAnyInErrors {
			serverFixed = append(serverFixed, codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"))
		}
		if err := requirePackageImports(server, serverFixed); err != nil {
			return err
		}
		if err := reservePackageImports(server,
			servicePlan.packages.service,
			codegen.NewImport(pathName+"pb", protobufPath),
		); err != nil {
			return err
		}
		if grpcServiceHasViewedResult(service) {
			if err := server.ReserveGeneratedImport(servicePlan.packages.views); err != nil {
				return err
			}
		}
		if err := planGRPCAttributeImports(server, generation, grpcEndpointAttributes(service.GRPCEndpoints...)); err != nil {
			return err
		}
		if err := requirePackageImports(server, servicePlan.protoGoImports); err != nil {
			return err
		}
	}

	for _, serverPlan := range plan.cli.servers {
		serverName := codegen.SnakeCase(codegen.Goify(serverPlan.name, true))
		outputPath := path.Join(generation.GenPkg(), "grpc", "cli", serverName)
		output := generation.Package(outputPath)
		fixed := []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("strconv"),
			codegen.SimpleImport("unicode/utf8"),
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.SimpleImport("google.golang.org/grpc"),
		}
		if grpcServerPlansUseAny(plan.servicesPlan, serverPlan.expression.Services) {
			fixed = append(fixed, codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"))
		}
		if err := requirePackageImports(output, fixed); err != nil {
			return err
		}
		for _, serviceName := range serverPlan.expression.Services {
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan == nil {
				continue
			}
			pathName := servicePlan.packages.pathName
			if err := reservePackageImports(output,
				codegen.NewImport(servicePlan.packages.service.Name+"c", path.Join(generation.GenPkg(), "grpc", pathName, "client")),
				codegen.NewImport(pathName+"pb", path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)),
			); err != nil {
				return err
			}
			if len(servicePlan.source.ServiceExpr.ClientInterceptors) > 0 {
				if err := output.ReserveGeneratedImport(servicePlan.packages.service); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// planGRPCExampleImports adds the gRPC files' imports to the executable
// packages already claimed by the shared example planner.
func planGRPCExampleImports(generation *codegen.Generation, plan *Plan, root *example.Root) error {
	rootPath := path.Dir(generation.GenPkg())
	for _, server := range root.Servers {
		serverPath := path.Join(rootPath, "cmd", server.Dir)
		serverPackage, err := generation.ClaimOutputPackage(serverPath, filepath.Join("cmd", server.Dir))
		if err != nil {
			return err
		}
		if err := requirePackageImports(serverPackage, []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("net"),
			codegen.SimpleImport("net/url"),
			codegen.SimpleImport("sync"),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.SimpleImport("goa.design/clue/debug"),
			codegen.SimpleImport("goa.design/clue/log"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/reflection"),
		}); err != nil {
			return err
		}
		for _, serviceName := range server.Services {
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan == nil {
				continue
			}
			pathName := servicePlan.packages.pathName
			if err := reservePackageImports(serverPackage,
				codegen.NewImport(servicePlan.packages.service.Name+"svr", path.Join(generation.GenPkg(), "grpc", pathName, "server")),
				servicePlan.packages.service,
				codegen.NewImport(pathName+"pb", path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName)),
			); err != nil {
				return err
			}
		}

		if server.DefaultTransport() == nil {
			continue
		}
		clientPath := path.Join(rootPath, "cmd", server.Dir+"-cli")
		clientPackage, err := generation.ClaimOutputPackage(clientPath, filepath.Join("cmd", server.Dir+"-cli"))
		if err != nil {
			return err
		}
		if err := requirePackageImports(clientPackage, []*codegen.ImportSpec{
			codegen.SimpleImport("context"),
			codegen.SimpleImport("errors"),
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("io"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("time"),
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.SimpleImport("google.golang.org/grpc"),
			codegen.SimpleImport("google.golang.org/grpc/credentials/insecure"),
		}); err != nil {
			return err
		}
		if err := clientPackage.ReserveGeneratedImport(codegen.NewImport(
			"cli",
			path.Join(generation.GenPkg(), "grpc", "cli", server.Dir),
		)); err != nil {
			return err
		}
		for _, serviceName := range server.Services {
			service := plan.root.API.GRPC.Service(serviceName)
			if service == nil {
				continue
			}
			if grpcServiceStreamsResult(service) {
				if err := clientPackage.ReserveGeneratedImport(plan.packages[service].service); err != nil {
					return err
				}
			}
			servicePlan := grpcServicePlanByName(plan.servicesPlan, serviceName)
			if servicePlan != nil && len(servicePlan.source.ServiceExpr.ClientInterceptors) > 0 {
				if err := clientPackage.ReserveGeneratedImport(codegen.NewImport("interceptors", rootPath+"/interceptors")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func requirePackageImports(output *codegen.GeneratedPackage, imports []*codegen.ImportSpec) error {
	for _, spec := range imports {
		if err := output.RequireImport(spec); err != nil {
			return err
		}
	}
	return nil
}

func reservePackageImports(output *codegen.GeneratedPackage, imports ...*codegen.ImportSpec) error {
	for _, spec := range imports {
		if err := output.ReserveGeneratedImport(spec); err != nil {
			return err
		}
	}
	return nil
}

// planGRPCAttributeImports preserves authored metadata aliases while generated
// service types use a package-local preferred name.
func planGRPCAttributeImports(output *codegen.GeneratedPackage, generation *codegen.Generation, attributes []*expr.AttributeExpr) error {
	seen := make(map[expr.UserType]struct{})
	var walk func(*expr.AttributeExpr) error
	walk = func(attribute *expr.AttributeExpr) error {
		if attribute == nil || attribute.Type == expr.Empty {
			return nil
		}
		if _, spec := codegen.GetMetaType(attribute); spec != nil && spec.Path != output.ImportPath() {
			if err := output.DeclareImport(spec); err != nil {
				return err
			}
		}
		switch actual := attribute.Type.(type) {
		case expr.UserType:
			if location := codegen.UserTypeLocation(actual); location != nil {
				importPath := path.Join(generation.GenPkg(), location.RelImportPath)
				if importPath != output.ImportPath() {
					preferred := strings.ToLower(codegen.Goify(path.Base(importPath), false))
					if err := output.ReserveGeneratedImport(codegen.NewImport(preferred, importPath)); err != nil {
						return err
					}
				}
			}
			origin := actual.Origin()
			if _, ok := seen[origin]; ok {
				return nil
			}
			seen[origin] = struct{}{}
			return walk(actual.Attribute())
		case *expr.Object:
			for _, named := range *actual {
				if err := walk(named.Attribute); err != nil {
					return err
				}
			}
		case *expr.Array:
			return walk(actual.ElemType)
		case *expr.Map:
			if err := walk(actual.KeyType); err != nil {
				return err
			}
			return walk(actual.ElemType)
		case *expr.Union:
			for _, named := range actual.Values {
				if err := walk(named.Attribute); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, attribute := range attributes {
		if err := walk(attribute); err != nil {
			return err
		}
	}
	return nil
}

func grpcServiceHasViewedResult(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if _, ok := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr); ok {
			return true
		}
	}
	return false
}

func grpcServiceStreamsPayload(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if endpoint.MethodExpr.IsPayloadStreaming() && !isEmpty(endpoint.Request.Type) {
			return true
		}
	}
	return false
}

func grpcServiceStreamsResult(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		if endpoint.MethodExpr.IsResultStreaming() && !endpoint.MethodExpr.IsPayloadStreaming() {
			return true
		}
	}
	return false
}

// grpcServerPlansUseAny reports whether one server's generated command parser
// handles a service with protobuf Any fields.
func grpcServerPlansUseAny(services []*grpcServicePlan, names []string) bool {
	for _, name := range names {
		service := grpcServicePlanByName(services, name)
		if service != nil && service.usesAny {
			return true
		}
	}
	return false
}

func grpcResponseMetadataUsesAny(service *expr.GRPCServiceExpr) bool {
	for _, endpoint := range service.GRPCEndpoints {
		for _, metadata := range []*expr.MappedAttributeExpr{endpoint.Response.Headers, endpoint.Response.Trailers} {
			if metadata == nil {
				continue
			}
			for _, named := range *expr.AsObject(metadata.Type) {
				typeKind := named.Attribute.Type.Kind()
				if array := expr.AsArray(named.Attribute.Type); array != nil {
					typeKind = array.ElemType.Type.Kind()
				}
				if typeKind == expr.AnyKind {
					return true
				}
			}
		}
	}
	return false
}

func grpcServicePlanByName(services []*grpcServicePlan, name string) *grpcServicePlan {
	for _, service := range services {
		if service.expression.Name() == name {
			return service
		}
	}
	return nil
}
