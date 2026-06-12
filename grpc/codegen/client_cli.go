package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/expr"
)

// ClientCLIFiles returns the CLI files to generate a command-line client that
// makes gRPC requests.
func ClientCLIFiles(genpkg string, services *ServicesData) []*codegen.File {
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
		command := cli.BuildCommandData(sd.Service)
		for _, e := range sd.Endpoints {
			flags, buildFunction := buildFlags(e)
			subcmd := cli.BuildSubcommandData(sd.Service, e.Method, buildFunction, flags)
			command.Subcommands = append(command.Subcommands, subcmd)
		}
		command.Example = command.Subcommands[0].Example
		data = append(data, command)
		svcs = append(svcs, svc)
	}
	files := make([]*codegen.File, 0, len(services.Root.API.Servers)+len(svcs))
	for _, svr := range services.Root.API.Servers {
		files = append(files, endpointParser(genpkg, services, svr, data))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(genpkg, svc, data[i], services))
	}
	return files
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, services *ServicesData, svr *expr.ServerExpr, data []*cli.CommandData) *codegen.File {
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
			&codegen.ImportSpec{Path: path.Join(genpkg, "grpc", svcName, "client"), Name: sd.Service.PkgName + "c"},
			&codegen.ImportSpec{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: svcName + pbPkgName})
		// Add interceptors import if service has client interceptors
		if len(sd.Service.ClientInterceptors) > 0 {
			specs = append(specs, &codegen.ImportSpec{
				Path: genpkg + "/" + sd.Service.PathName,
				Name: sd.Service.PkgName,
			})
		}
	}

	parseSection := &codegen.SectionTemplate{
		Name:   "parse-endpoint-grpc",
		Source: grpcTemplates.Read(grpcParseEndpointT),
		Data: struct {
			FlagsCode string
			Commands  []*cli.CommandData
		}{
			cli.FlagsCode(data),
			data,
		},
	}
	return cli.EndpointParserFile(fpath, title, specs, data, parseSection)
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svc *expr.GRPCServiceExpr, data *cli.CommandData, services *ServicesData) *codegen.File {
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
		{Path: path.Join(genpkg, svcName), Name: sd.Service.PkgName},
		{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: sd.PkgName},
	}
	// Add structpb import if Any type is used
	if usesAnyType(svc.GRPCEndpoints, false) {
		specs = append(specs,
			&codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"},
		)
	}
	return cli.PayloadBuildersFile(fpath, title, specs, data)
}

func buildFlags(e *EndpointData) ([]*cli.FlagData, *cli.BuildFunctionData) {
	if e.Request != nil {
		return makeFlags(e, e.Request.CLIArgs)
	}
	return nil, nil
}

func makeFlags(e *EndpointData, args []*InitArgData) ([]*cli.FlagData, *cli.BuildFunctionData) {
	fargs := make([]*cli.FlagArgData, len(args))
	pInitArgs := make([]*codegen.InitArgData, len(args))
	for i, arg := range args {
		pInitArgs[i] = &codegen.InitArgData{
			Name:      arg.Name,
			FieldName: arg.FieldName,
			FieldType: arg.FieldType,
			Type:      arg.Type,
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
