// This file renders gRPC command parsers and per-service payload builders,
// including relocated payload imports in the builder that references them.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
)

type (
	// commandData adds the exact gRPC client constructor to the shared command
	// data used by transport command-line generators.
	commandData struct {
		*cli.CommandData
		// ClientInit is the client constructor called by ParseEndpoint.
		ClientInit *codegen.NameDeclaration
	}
)

// clientCLIFiles returns the planned command-line client files.
func clientCLIFiles(services *ServicesData) []*codegen.File {
	if len(services.servicePlans) == 0 {
		return nil
	}
	var (
		data = make([]*commandData, 0, len(services.servicePlans))
		svcs = make([]*grpcServicePlan, 0, len(services.servicePlans))
	)
	for _, servicePlan := range services.servicePlans {
		svc := servicePlan.expression
		if len(svc.GRPCEndpoints) == 0 {
			continue
		}
		sd := services.Get(svc.Name())
		command := &commandData{
			CommandData: cli.BuildCommandData(sd.Service),
			ClientInit:  sd.ClientInitDeclaration,
		}
		for index, e := range sd.Endpoints {
			flags, buildFunction := buildFlags(e, services.cliPlan.builders[svc.GRPCEndpoints[index]])
			subcmd := cli.BuildSubcommandData(sd.Service, e.Method, buildFunction, flags)
			command.CommandData.Subcommands = append(command.CommandData.Subcommands, subcmd)
		}
		command.Example = command.Subcommands[0].Example
		data = append(data, command)
		svcs = append(svcs, servicePlan)
	}
	files := make([]*codegen.File, 0, len(services.cliPlan.servers)+len(svcs))
	for _, serverPlan := range services.cliPlan.servers {
		serverData := make([]*commandData, 0, len(serverPlan.expression.Services))
		for _, serviceName := range serverPlan.expression.Services {
			for _, command := range data {
				if command.ServiceName == serviceName {
					serverData = append(serverData, command)
					break
				}
			}
		}
		files = append(files, endpointParser(services, serverPlan, serverData))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(svc, data[i].CommandData, services))
	}
	return files
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(services *ServicesData, serverPlan *grpcCLIServerPlan, data []*commandData) *codegen.File {
	genpkg := services.GenPkg()
	pkg := codegen.SnakeCase(codegen.Goify(serverPlan.name, true))
	outputPackage := path.Join(genpkg, "grpc", "cli", pkg)
	fpath := filepath.Join(codegen.Gendir, "grpc", "cli", pkg, "cli.go")
	title := serverPlan.name + " gRPC client CLI support package"
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("grpc", "goagrpc"),
		{Path: "google.golang.org/grpc", Name: "grpc"},
	}
	// Add structpb import if Any type is used
	needsAnyPb := false
	for _, serviceName := range serverPlan.expression.Services {
		servicePlan := grpcServicePlanByName(services.servicePlans, serviceName)
		if servicePlan != nil && servicePlan.usesAny {
			needsAnyPb = true
			break
		}
	}
	if needsAnyPb {
		specs = append(specs,
			&codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"},
		)
	}
	for _, serviceName := range serverPlan.expression.Services {
		servicePlan := grpcServicePlanByName(services.servicePlans, serviceName)
		if servicePlan == nil {
			continue
		}
		svc := servicePlan.expression
		sd := services.Get(svc.Name())
		if sd == nil {
			continue
		}
		svcName := sd.Service.PathName
		specs = append(specs,
			services.PackageImport(outputPackage, path.Join(genpkg, "grpc", svcName, "client")),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)))
		// Add interceptors import if service has client interceptors
		if len(sd.Service.ClientInterceptors) > 0 {
			specs = append(specs, services.ServiceImport(outputPackage, svc.Name()))
		}
	}

	parser := serverPlan.parser
	if parser == nil {
		panic(fmt.Sprintf("gRPC command parser names are missing for server %q", serverPlan.name))
	}
	plannedData := make([]*commandData, len(data))
	plannedCommands := make([]*cli.CommandData, len(data))
	for index, command := range data {
		commandNames := parser.Commands[command.ServiceName]
		if commandNames == nil {
			panic(fmt.Sprintf("gRPC command names are missing for service %q", command.ServiceName))
		}
		commandCopy := *command.CommandData
		sd := services.Get(command.ServiceName)
		clientPath := path.Join(genpkg, "grpc", sd.Service.PathName, "client")
		commandCopy.PkgName = services.PackageImport(outputPackage, clientPath).Name
		if command.Interceptors != nil {
			interceptors := *command.Interceptors
			interceptors.PkgName = services.ServiceImport(outputPackage, command.ServiceName).Name
			commandCopy.Interceptors = &interceptors
		}
		commandCopy.UsageDeclaration = commandNames.Usage
		commandCopy.Subcommands = make([]*cli.SubcommandData, len(command.Subcommands))
		for methodIndex, subcommand := range command.Subcommands {
			usage := commandNames.Methods[subcommand.MethodName]
			if usage == nil {
				panic(fmt.Sprintf("gRPC method help name is missing for %q.%q", command.ServiceName, subcommand.Name))
			}
			subcommandCopy := *subcommand
			subcommandCopy.UsageDeclaration = usage
			commandCopy.Subcommands[methodIndex] = &subcommandCopy
		}
		plannedData[index] = &commandData{
			CommandData: &commandCopy,
			ClientInit:  command.ClientInit,
		}
		plannedCommands[index] = &commandCopy
	}
	parser.PlanVariables(plannedCommands, nil)
	parseSection := &codegen.SectionTemplate{
		Name:   "parse-endpoint-grpc",
		Source: grpcTemplates.Read(grpcParseEndpointT),
		Data: struct {
			Declaration *codegen.NameDeclaration
			FlagsCode   string
			Commands    []*commandData
			Variables   *cli.ParserVariablesData
		}{
			parser.Declarations.ParseEndpoint,
			parser.FlagsCode(plannedCommands),
			plannedData,
			parser.Variables,
		},
	}
	return parser.EndpointParserFile(fpath, title, specs, plannedCommands, parseSection)
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(servicePlan *grpcServicePlan, data *cli.CommandData, services *ServicesData) *codegen.File {
	svc := servicePlan.expression
	sd := services.Get(svc.Name())
	svcName := sd.Service.PathName
	outputPackage := path.Join(services.GenPkg(), "grpc", svcName, "client")
	fpath := filepath.Join(codegen.Gendir, "grpc", svcName, "client", "cli.go")
	title := svc.Name() + " gRPC client CLI support package"
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		services.ServiceImport(outputPackage, svc.Name()),
		services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
	}
	// Add structpb import if Any type is used
	if servicePlan.usesAny {
		specs = append(specs,
			&codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"},
		)
	}
	return addEndpointImports(cli.PayloadBuildersFile(fpath, title, specs, data), services, servicePlan)
}

func buildFlags(e *EndpointData, declaration *codegen.NameDeclaration) ([]*cli.FlagData, *cli.BuildFunctionData) {
	if e.Request != nil {
		flags, buildFunction := makeFlags(e, e.Request.CLIArgs)
		if buildFunction != nil {
			if declaration == nil {
				panic(fmt.Sprintf("gRPC payload builder name is missing for %q.%q", e.ServiceName, e.Method.Name))
			}
			buildFunction.Name = declaration.Name()
		}
		return flags, buildFunction
	}
	return nil, nil
}

func makeFlags(e *EndpointData, args []*InitArgData) ([]*cli.FlagData, *cli.BuildFunctionData) {
	fargs := make([]*cli.FlagArgData, len(args))
	pInitArgs := make([]*codegen.InitArgData, len(args))
	for i, arg := range args {
		pInitArgs[i] = &codegen.InitArgData{
			Name:         arg.Name,
			FieldName:    arg.FieldName,
			FieldType:    arg.FieldType,
			Type:         arg.Type,
			Pointer:      arg.Pointer,
			FieldPointer: arg.Pointer,
		}
		fargs[i] = &cli.FlagArgData{
			Name:         arg.Name,
			TypeName:     arg.TypeName,
			Plan:         arg.CLIPlan,
			TypeRef:      arg.TypeRef,
			FieldName:    arg.FieldName,
			Description:  arg.Description,
			Required:     arg.Required,
			Example:      arg.Example,
			DefaultValue: arg.DefaultValue,
		}
	}

	var pinit *cli.PayloadInitData
	if e.Method.PayloadRef != "" && e.Request.ServerConvert != nil {
		pinit = &cli.PayloadInitData{
			Code:           e.Request.CLIInitCode,
			ReturnIsStruct: e.Request.ServerConvert.Init.ReturnIsStruct,
			ReturnTypePkg:  e.Request.ServerConvert.Init.ReturnTypePkg,
			Args:           pInitArgs,
		}
	}

	flags, buildFunction := cli.MakeFlags(e.ServiceName, e.Method, fargs, e.PayloadType, e.PayloadRef, pinit)
	if e.Method.PayloadRef == "" {
		return flags, nil
	}
	return flags, buildFunction
}
