// This file retains one gRPC design, its chosen Go names, and every generated
// file from planning through rendering.
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

	// Plan retains one design root and every gRPC file built from it.
	Plan struct {
		generation *codegen.Generation
		root       *expr.RootExpr
		service    *service.Plan
		cli        *grpcCLIPlan
		services   *ServicesData
		proto      []*codegen.File
		server     []*codegen.File
		client     []*codegen.File
		serverType []*codegen.File
		clientType []*codegen.File
		clientCLI  []*codegen.File
		example    []*codegen.File
		exampleCLI []*codegen.File
	}

	// grpcCLIPlan contains the command parser and payload function names for one
	// design.
	grpcCLIPlan struct {
		parsers  map[*expr.ServerExpr]*cli.ParserPlan
		builders map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration
	}
)

// NewPlans reads every service design in generation and retains one plan for
// each input. It chooses shared package names before generation freezes.
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	owned := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		if root, ok := candidate.(*expr.RootExpr); ok && len(root.API.GRPC.Services) > 0 {
			owned[root] = struct{}{}
		}
	}
	if len(inputs) != len(owned) {
		return nil, fmt.Errorf("gRPC planning requires all %d gRPC roots, got %d", len(owned), len(inputs))
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
	if err := requireGRPCImports(generation); err != nil {
		return nil, err
	}
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		cliPlan, err := planGRPCCLI(generation, input)
		if err != nil {
			return nil, err
		}
		plans[index] = &Plan{
			generation: generation,
			root:       input.Root,
			service:    input.Service,
			cli:        cliPlan,
		}
	}
	return plans, nil
}

// Generation returns the generation that owns this plan's package names.
func (p *Plan) Generation() *codegen.Generation {
	return p.generation
}

// Root returns the exact design supplied to NewPlans.
func (p *Plan) Root() *expr.RootExpr {
	return p.root
}

// Service returns the exact service plan supplied to NewPlans.
func (p *Plan) Service() *service.Plan {
	return p.service
}

// Link builds the gRPC render data and files after all names are frozen. The
// service plan must already be linked.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("gRPC plan cannot link before generation freeze")
	}
	if p.services != nil {
		return fmt.Errorf("gRPC plan is already linked")
	}
	services := newServicesData(p.service.Services(), p)
	for _, grpcService := range p.root.API.GRPC.Services {
		services.Get(grpcService.Name())
	}
	p.services = services
	p.proto = ProtoFiles(services)
	p.server = ServerFiles(services)
	p.client = ClientFiles(services)
	p.serverType = ServerTypeFiles(services)
	p.clientType = ClientTypeFiles(services)
	p.clientCLI = ClientCLIFiles(services)
	p.example = ExampleServerFiles(services)
	p.exampleCLI = ExampleCLIFiles(services)
	return nil
}

// ProtoFiles returns the protobuf schemas built by Link.
func (p *Plan) ProtoFiles() []*codegen.File {
	p.requireLinked()
	return p.proto
}

// ServerFiles returns the gRPC server files built by Link.
func (p *Plan) ServerFiles() []*codegen.File {
	p.requireLinked()
	return p.server
}

// ClientFiles returns the gRPC client files built by Link.
func (p *Plan) ClientFiles() []*codegen.File {
	p.requireLinked()
	return p.client
}

// ServerTypeFiles returns the server transport type files built by Link.
func (p *Plan) ServerTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.serverType
}

// ClientTypeFiles returns the client transport type files built by Link.
func (p *Plan) ClientTypeFiles() []*codegen.File {
	p.requireLinked()
	return p.clientType
}

// ClientCLIFiles returns the command-line client files built by Link.
func (p *Plan) ClientCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.clientCLI
}

// ExampleServerFiles returns the runnable gRPC server files built by Link.
func (p *Plan) ExampleServerFiles() []*codegen.File {
	p.requireLinked()
	return p.example
}

// ExampleCLIFiles returns the runnable gRPC client files built by Link.
func (p *Plan) ExampleCLIFiles() []*codegen.File {
	p.requireLinked()
	return p.exampleCLI
}

// requireGRPCImports records packages used by gRPC files before names freeze.
func requireGRPCImports(generation *codegen.Generation) error {
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
			return err
		}
	}
	return nil
}

// planGRPCCLI chooses parser and payload builder names for one design.
func planGRPCCLI(generation *codegen.Generation, input PlanInput) (*grpcCLIPlan, error) {
	design := input.Root
	plan := &grpcCLIPlan{
		parsers:  make(map[*expr.ServerExpr]*cli.ParserPlan),
		builders: make(map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration),
	}
	for _, grpcService := range design.API.GRPC.Services {
		pathName := codegen.SnakeCase(codegen.Goify(grpcService.Name(), false))
		packageName := strings.ToLower(codegen.Goify(grpcService.Name(), false))
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
		for _, endpoint := range grpcService.GRPCEndpoints {
			if endpoint.MethodExpr.Payload.Type == expr.Empty {
				continue
			}
			names, err := input.Service.HTTPMethodNames(endpoint.MethodExpr)
			if err != nil {
				return nil, err
			}
			declaration, err := cli.DeclarePayloadBuilder(clientPackage, "grpc", design.API.Name, grpcService.Name(), endpoint.Name(), "Build"+names.Method+"Payload")
			if err != nil {
				return nil, err
			}
			plan.builders[endpoint] = declaration
		}
	}
	if len(design.API.GRPC.Services) == 0 {
		return plan, nil
	}
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
		plan.parsers[server] = parser
	}
	return plan, nil
}

// requireLinked stops file reads before Link stores the files.
func (p *Plan) requireLinked() {
	if p.services == nil {
		panic("gRPC files requested before plan linking")
	}
}
