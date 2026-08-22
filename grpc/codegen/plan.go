// This file records gRPC imports, generated protobuf package names, and command
// functions before generated files are written.
package codegen

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// PlanInput pairs one design with the generated service names chosen for it.
	PlanInput struct {
		// Root is the design that contains the gRPC services.
		Root *expr.RootExpr
		// Service provides the method names selected for Root.
		Service *service.Plan
	}

	// PreparedPlan contains the command-line function names requested before Go
	// assigns their final spellings.
	PreparedPlan struct {
		roots map[*expr.RootExpr]*grpcCLIPlan
	}

	// grpcCLIPlan contains the command parser and payload function names for one
	// design.
	grpcCLIPlan struct {
		parsers  map[*expr.ServerExpr]*cli.ParserPlan
		builders map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration
	}
)

// Plan requests the import names and command-line functions used by the gRPC
// files for inputs. ClientCLIFiles reads the returned names after Goa has made
// every name unique within its Go package.
func Plan(generation *codegen.Generation, inputs ...PlanInput) (*PreparedPlan, error) {
	owned := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		if root, ok := candidate.(*expr.RootExpr); ok {
			owned[root] = struct{}{}
		}
	}
	if len(inputs) != len(owned) {
		return nil, fmt.Errorf("gRPC planning requires all %d service roots, got %d", len(owned), len(inputs))
	}
	seen := make(map[*expr.RootExpr]struct{}, len(inputs))
	for _, input := range inputs {
		if input.Service == nil || input.Service.Root() != input.Root {
			return nil, fmt.Errorf("gRPC plan input does not pair a design with its service plan")
		}
		if _, ok := owned[input.Root]; !ok {
			return nil, fmt.Errorf("gRPC root %p is not part of this generation", input.Root)
		}
		if _, ok := seen[input.Root]; ok {
			return nil, fmt.Errorf("gRPC root %p is planned more than once", input.Root)
		}
		seen[input.Root] = struct{}{}
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("flag"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("net"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("goa.design/clue/debug"),
		codegen.SimpleImport("goa.design/clue/log"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("grpc", "goagrpc"),
		codegen.GoaNamedImport("grpc/pb", "goapb"),
		codegen.SimpleImport("google.golang.org/grpc"),
		codegen.SimpleImport("google.golang.org/grpc/codes"),
		codegen.SimpleImport("google.golang.org/grpc/credentials/insecure"),
		codegen.SimpleImport("google.golang.org/grpc/metadata"),
		codegen.SimpleImport("google.golang.org/grpc/reflection"),
		codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return nil, err
		}
	}
	plan := &PreparedPlan{roots: make(map[*expr.RootExpr]*grpcCLIPlan, len(inputs))}
	for _, input := range inputs {
		design := input.Root
		rootPlan := &grpcCLIPlan{
			parsers:  make(map[*expr.ServerExpr]*cli.ParserPlan),
			builders: make(map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration),
		}
		for _, service := range design.API.GRPC.Services {
			pathName := codegen.SnakeCase(codegen.Goify(service.Name(), false))
			packageName := strings.ToLower(codegen.Goify(service.Name(), false))
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"c", path.Join(generation.GenPkg(), "grpc", pathName, "client"))); err != nil {
				return nil, err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"svr", path.Join(generation.GenPkg(), "grpc", pathName, "server"))); err != nil {
				return nil, err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(pathName+"pb", path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName))); err != nil {
				return nil, err
			}
			clientPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "grpc", pathName, "client"))
			if err != nil {
				return nil, err
			}
			for _, endpoint := range service.GRPCEndpoints {
				if endpoint.MethodExpr.Payload.Type == expr.Empty {
					continue
				}
				names, err := input.Service.HTTPMethodNames(endpoint.MethodExpr)
				if err != nil {
					return nil, err
				}
				declaration, err := cli.DeclarePayloadBuilder(clientPackage, "grpc", design.API.Name, service.Name(), endpoint.Name(), "Build"+names.Method+"Payload")
				if err != nil {
					return nil, err
				}
				rootPlan.builders[endpoint] = declaration
			}
		}
		if len(design.API.GRPC.Services) > 0 {
			for _, server := range design.API.Servers {
				serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
				if err := generation.ReserveGeneratedImport(codegen.NewImport("cli", path.Join(generation.GenPkg(), "grpc", "cli", serverName))); err != nil {
					return nil, err
				}
				serverPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "grpc", "cli", serverName))
				if err != nil {
					return nil, err
				}
				var commands []cli.CommandDeclarationInput
				for _, grpcService := range design.API.GRPC.Services {
					if len(grpcService.GRPCEndpoints) == 0 {
						continue
					}
					command := cli.CommandDeclarationInput{Service: grpcService.Name()}
					for _, endpoint := range grpcService.GRPCEndpoints {
						command.Methods = append(command.Methods, endpoint.Name())
					}
					commands = append(commands, command)
				}
				parser, err := cli.DeclareParser(serverPackage, "grpc", design.API.Name, server.Name, commands)
				if err != nil {
					return nil, err
				}
				rootPlan.parsers[server] = parser
			}
		}
		plan.roots[design] = rootPlan
	}
	return plan, nil
}
