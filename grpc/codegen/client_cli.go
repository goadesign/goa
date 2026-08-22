// This file renders gRPC command parsers and per-service payload builders,
// including relocated payload imports in the builder that references them.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/expr"
)

// ClientCLIFiles returns the CLI files to generate a command-line client that
// makes gRPC requests.
func ClientCLIFiles(services *ServicesData) []*codegen.File {
	if len(services.Root.API.GRPC.Services) == 0 {
		return nil
	}
	var (
		data = make([]*cli.CommandData, 0, len(services.Root.API.GRPC.Services))
		svcs = make([]*expr.GRPCServiceExpr, 0, len(services.Root.API.GRPC.Services))
	)
	for _, svc := range services.Root.API.GRPC.Services {
		if len(svc.GRPCEndpoints) == 0 {
			continue
		}
		sd := services.Get(svc.Name())
		command := cli.BuildCommandData(sd.Service, sd.ClientPkgName)
		for index, e := range sd.Endpoints {
			flags, buildFunction := buildFlags(e, services.cliPlan.builders[svc.GRPCEndpoints[index]])
			subcmd := cli.BuildSubcommandData(sd.Service, e.Method, buildFunction, flags)
			command.Subcommands = append(command.Subcommands, subcmd)
		}
		command.Example = command.Subcommands[0].Example
		data = append(data, command)
		svcs = append(svcs, svc)
	}
	files := make([]*codegen.File, 0, len(services.Root.API.Servers)+len(svcs))
	for _, svr := range services.Root.API.Servers {
		files = append(files, endpointParser(services, svr, data))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(svc, data[i], services))
	}
	return files
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(services *ServicesData, svr *expr.ServerExpr, data []*cli.CommandData) *codegen.File {
	genpkg := services.GenPkg()
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	fpath := filepath.Join(codegen.Gendir, "grpc", "cli", pkg, "cli.go")
	title := svr.Name + " gRPC client CLI support package"
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
	for _, svc := range services.Root.API.GRPC.Services {
		if usesAnyType(svc.GRPCEndpoints, false) {
			needsAnyPb = true
			break
		}
	}
	if needsAnyPb {
		specs = append(specs,
			&codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"},
		)
	}
	for _, svc := range services.Root.API.GRPC.Services {
		sd := services.Get(svc.Name())
		if sd == nil {
			continue
		}
		svcName := sd.Service.PathName
		specs = append(specs,
			services.PackageImport(path.Join(genpkg, "grpc", svcName, "client")),
			services.PackageImport(path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)))
		// Add interceptors import if service has client interceptors
		if len(sd.Service.ClientInterceptors) > 0 {
			specs = append(specs, services.ServiceImport(svc.Name()))
		}
	}

	parser := services.cliPlan.parsers[svr]
	if parser == nil {
		panic(fmt.Sprintf("gRPC command parser names are missing for server %q", svr.Name))
	}
	plannedData := make([]*cli.CommandData, len(data))
	for index, command := range data {
		commandNames := parser.Commands[command.ServiceName]
		if commandNames == nil {
			panic(fmt.Sprintf("gRPC command names are missing for service %q", command.ServiceName))
		}
		commandCopy := *command
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
		plannedData[index] = &commandCopy
	}
	parseSection := &codegen.SectionTemplate{
		Name:   "parse-endpoint-grpc",
		Source: grpcTemplates.Read(grpcParseEndpointT),
		Data: struct {
			Declaration *codegen.NameDeclaration
			FlagsCode   string
			Commands    []*cli.CommandData
		}{
			parser.Declarations.ParseEndpoint,
			cli.FlagsCode(plannedData),
			plannedData,
		},
	}
	return cli.EndpointParserFile(fpath, title, specs, plannedData, parser.Declarations, parseSection)
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(svc *expr.GRPCServiceExpr, data *cli.CommandData, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	svcName := sd.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "grpc", svcName, "client", "cli.go")
	title := svc.Name() + " gRPC client CLI support package"
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		services.ServiceImport(svc.Name()),
		services.PackageImport(path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
	}
	// Add structpb import if Any type is used
	if usesAnyType(svc.GRPCEndpoints, false) {
		specs = append(specs,
			&codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"},
		)
	}
	return addEndpointImports(cli.PayloadBuildersFile(fpath, title, specs, data), services, svc.GRPCEndpoints...)
}

func buildFlags(e *EndpointData, declaration *codegen.NameDeclaration) ([]*cli.FlagData, *cli.BuildFunctionData) {
	if e.Request != nil {
		flags, buildFunction := makeFlags(e, e.Request.CLIArgs)
		if buildFunction != nil {
			if declaration == nil {
				panic(fmt.Sprintf("gRPC payload builder name is missing for %q.%q", e.ServiceName, e.Method.Name))
			}
			buildFunction.Declaration = declaration
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
			TypeRef:      arg.TypeRef,
			FieldName:    arg.FieldName,
			Description:  arg.Description,
			Required:     arg.Required,
			Example:      arg.Example,
			DefaultValue: arg.DefaultValue,
			Validate:     arg.Validate,
		}
	}

	var pinit *cli.PayloadInitData
	if e.Method.PayloadRef != "" && e.Request.ServerConvert != nil {
		pinit = &cli.PayloadInitData{
			Code:           e.Request.ServerConvert.Init.Code,
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
