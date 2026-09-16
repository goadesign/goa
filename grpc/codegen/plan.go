// This file stores one gRPC design, its chosen Go names, and every file built
// from it.
package codegen

import (
	"fmt"
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/example"
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

	// Plan stores one design and every gRPC file built from it.
	Plan struct {
		generation   *codegen.Generation
		root         *expr.RootExpr
		service      *service.Plan
		cli          *grpcCLIPlan
		protobuf     map[*expr.GRPCServiceExpr]*protobufServicePlan
		packages     map[*expr.GRPCServiceExpr]*grpcServicePackage
		tools        map[*expr.GRPCServiceExpr]*protobufToolPlan
		symbols      map[*expr.GRPCServiceExpr]*grpcSymbols
		fileImports  map[string]*codegen.GeneratedImportPlan
		expressions  []*expr.GRPCServiceExpr
		servicesPlan []*grpcServicePlan
		services     *ServicesData
		proto        []*codegen.File
		server       []*codegen.File
		client       []*codegen.File
		serverType   []*codegen.File
		clientType   []*codegen.File
		clientCLI    []*codegen.File
	}

	// ExamplePlan builds runnable gRPC programs from server data and generated
	// services that came from the same design.
	ExamplePlan struct {
		root      *example.Root
		transport *Plan
	}

	// grpcCLIPlan contains the command parser and payload function names for one
	// design.
	grpcCLIPlan struct {
		parsers  map[*expr.ServerExpr]*cli.ParserPlan
		builders map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration
		servers  []*grpcCLIServerPlan
	}

	// grpcCLIServerPlan stores one server name and the command parser declared
	// for that server before generated files are built.
	grpcCLIServerPlan struct {
		expression *expr.ServerExpr
		name       string
		parser     *cli.ParserPlan
	}

	// grpcServicePackage stores the generated service import and the directory
	// used by every gRPC package for that service.
	grpcServicePackage struct {
		service  *codegen.ImportSpec
		views    *codegen.ImportSpec
		pathName string
	}
)

// NewPlans reads every service design and stores one plan for each input. It
// chooses all shared package names before files are built.
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	return newPlans(generation, systemProtobufTools(), inputs...)
}

// NewExamplePlan returns an example renderer only when examples contains the
// server data copied from transport's service design.
func NewExamplePlan(transport *Plan, examples *example.Plan) (*ExamplePlan, error) {
	root, ok := examples.Root(transport.service)
	if !ok {
		return nil, fmt.Errorf("gRPC examples require server data created from the same service design")
	}
	if err := planGRPCExampleImports(transport.generation, transport, root); err != nil {
		return nil, err
	}
	return &ExamplePlan{root: root, transport: transport}, nil
}

// newPlans lets tests provide fixed protobuf executable paths and versions.
func newPlans(generation *codegen.Generation, resolver protobufToolResolver, inputs ...PlanInput) ([]*Plan, error) {
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
	toolPlans, err := planProtobufTools(inputs, resolver)
	if err != nil {
		return nil, err
	}
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		packages, err := planGRPCServicePackages(input)
		if err != nil {
			return nil, err
		}
		cliPlan, err := planGRPCCLI(generation, input, packages)
		if err != nil {
			return nil, err
		}
		tools := make(map[*expr.GRPCServiceExpr]*protobufToolPlan, len(input.Root.API.GRPC.Services))
		for _, grpcService := range input.Root.API.GRPC.Services {
			tools[grpcService] = toolPlans[grpcService]
		}
		plans[index] = &Plan{
			generation:  generation,
			root:        input.Root,
			service:     input.Service,
			cli:         cliPlan,
			protobuf:    make(map[*expr.GRPCServiceExpr]*protobufServicePlan),
			packages:    packages,
			tools:       tools,
			symbols:     make(map[*expr.GRPCServiceExpr]*grpcSymbols),
			fileImports: make(map[string]*codegen.GeneratedImportPlan),
			expressions: append([]*expr.GRPCServiceExpr(nil), input.Root.API.GRPC.Services...),
		}
	}
	if err := planProtobufServices(generation, plans); err != nil {
		return nil, err
	}
	conversions := make(map[grpcConversionKey]*grpcConversion)
	registries := make(map[*codegen.GeneratedPackage]*grpcTransformRegistry)
	for _, plan := range plans {
		input := PlanInput{Root: plan.root, Service: plan.service}
		for _, grpcService := range plan.expressions {
			pathName := plan.packages[grpcService].pathName
			symbols, err := collectGRPCSymbols(generation, input, grpcService, pathName)
			if err != nil {
				return nil, err
			}
			if err := planGRPCValidations(generation, input, grpcService, plan.protobuf[grpcService], pathName); err != nil {
				return nil, err
			}
			if err := planGRPCTransforms(generation, input, grpcService, plan.protobuf[grpcService], symbols, conversions, registries, pathName); err != nil {
				return nil, err
			}
			plan.symbols[grpcService] = symbols
		}
	}
	if err := declareGRPCTransforms(conversions, registries); err != nil {
		return nil, err
	}
	for _, plan := range plans {
		servicesPlan, err := collectGRPCServicePlans(plan)
		if err != nil {
			return nil, err
		}
		plan.servicesPlan = servicesPlan
		if err := planGRPCImports(generation, plan); err != nil {
			return nil, err
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

// ServiceData returns the finalized gRPC data for the exact service used to
// build this plan. Callers must call Link before reading the service data.
func (p *Plan) ServiceData(service *expr.GRPCServiceExpr) (*ServiceData, bool) {
	p.requireLinked()
	data, ok := p.services.serviceByExpr[service]
	return data, ok
}

// Link builds the gRPC files after all generated Go names are fixed. The
// service plan must already have built its files.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("gRPC plan cannot link before generation freeze")
	}
	if p.services != nil {
		return fmt.Errorf("gRPC plan is already linked")
	}
	for filePath, imports := range p.fileImports {
		if err := imports.Link(); err != nil {
			return fmt.Errorf("link gRPC imports for %q: %w", filePath, err)
		}
	}
	services := newServicesData(p.service.Services(), p)
	p.services = services
	p.proto = protoFiles(services)
	p.server = serverFiles(services)
	p.client = clientFiles(services)
	p.serverType = serverTypeFiles(services)
	p.clientType = clientTypeFiles(services)
	p.clientCLI = clientCLIFiles(services)
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

// ServerFiles builds runnable gRPC servers from the copied server data.
func (p *ExamplePlan) ServerFiles() []*codegen.File {
	p.transport.requireLinked()
	return exampleServerFiles(p.root, p.transport.services)
}

// CLIFiles builds runnable gRPC clients from the copied server data.
func (p *ExamplePlan) CLIFiles() []*codegen.File {
	p.transport.requireLinked()
	return exampleCLIFiles(p.root, p.transport.services)
}

// planGRPCCLI chooses parser and payload builder names for one design.
func planGRPCCLI(generation *codegen.Generation, input PlanInput, packages map[*expr.GRPCServiceExpr]*grpcServicePackage) (*grpcCLIPlan, error) {
	design := input.Root
	plan := &grpcCLIPlan{
		parsers:  make(map[*expr.ServerExpr]*cli.ParserPlan),
		builders: make(map[*expr.GRPCEndpointExpr]*codegen.NameDeclaration),
	}
	for _, grpcService := range design.API.GRPC.Services {
		pathName := packages[grpcService].pathName
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
		var commands []cli.CommandDeclarationInput
		for _, serviceName := range server.Services {
			grpcService := design.API.GRPC.Service(serviceName)
			if grpcService == nil {
				continue
			}
			if len(grpcService.GRPCEndpoints) == 0 {
				continue
			}
			command := cli.CommandDeclarationInput{Service: grpcService.Name()}
			for _, endpoint := range grpcService.GRPCEndpoints {
				command.Methods = append(command.Methods, endpoint.Name())
				command.NeedsFlagPresence = command.NeedsFlagPresence || grpcEndpointNeedsCLIFlagPresence(endpoint)
			}
			commands = append(commands, command)
		}
		if len(commands) == 0 {
			continue
		}
		serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
		serverPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "grpc", "cli", serverName))
		if err != nil {
			return nil, err
		}
		parser, err := cli.DeclareParser(serverPackage, "grpc", design.API.Name, server.Name, commands)
		if err != nil {
			return nil, err
		}
		plan.parsers[server] = parser
		plan.servers = append(plan.servers, &grpcCLIServerPlan{
			expression: server,
			name:       server.Name,
			parser:     parser,
		})
	}
	return plan, nil
}

// grpcEndpointNeedsCLIFlagPresence reports whether the generated command must
// distinguish a missing flag from a flag whose value is empty. Protobuf request
// messages have no CLI default. Metadata fields use the default from the design
// when one is present.
func grpcEndpointNeedsCLIFlagPresence(endpoint *expr.GRPCEndpointExpr) bool {
	if endpoint.Request.Type != expr.Empty {
		return true
	}
	for _, field := range *expr.AsObject(endpoint.Metadata.Type) {
		if field.Attribute.DefaultValue == nil {
			return true
		}
	}
	return false
}

// planGRPCServicePackages records the exact service imports before generated
// package names become final. Every later gRPC planning step uses these paths.
func planGRPCServicePackages(input PlanInput) (map[*expr.GRPCServiceExpr]*grpcServicePackage, error) {
	packages := make(map[*expr.GRPCServiceExpr]*grpcServicePackage, len(input.Root.API.GRPC.Services))
	for _, transportService := range input.Root.API.GRPC.Services {
		serviceImport, viewsImport, err := input.Service.ServicePackageImports(transportService.ServiceExpr)
		if err != nil {
			return nil, err
		}
		packages[transportService] = &grpcServicePackage{
			service:  serviceImport,
			views:    viewsImport,
			pathName: path.Base(serviceImport.Path),
		}
	}
	return packages, nil
}

// requireLinked stops file reads before Link stores the files.
func (p *Plan) requireLinked() {
	if p.services == nil {
		panic("gRPC files requested before plan linking")
	}
}
