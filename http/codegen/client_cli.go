// This file renders HTTP client command parsers and per-service payload
// builders, including imports for relocated payload types used by each builder.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/expr"
)

// commandData wraps the common CommandData and adds HTTP-specific fields.
type commandData struct {
	*cli.CommandData
	// Subcommands is the list of endpoint commands.
	Subcommands []*subcommandData
	// NeedDialer if true initializes the websocket dialer.
	NeedDialer bool
	// JSONRPC if true indicates the command targets a JSON-RPC service:
	// streaming endpoints are configured with a goahttp.ConnConfigureFunc
	// instead of a client package ConnConfigurer.
	JSONRPC bool
}

// commandData wraps the common SubcommandData and adds HTTP-specific fields.
type subcommandData struct {
	*cli.SubcommandData
	// MultipartFuncName is the name of the function used to render a multipart
	// request encoder.
	MultipartFuncName string
	// MultipartFuncName is the name of the variable used to render a multipart
	// request encoder.
	MultipartVarName string
	// StreamFlag is the flag used to identify the file to be streamed when
	// the endpoint uses SkipRequestBodyEncodeDecode.
	StreamFlag *cli.FlagData
	// BuildStreamPayload is the name of the generated function that builds the
	// request data structure that wraps the payload and the file stream for
	// endpoints that use SkipRequestBodyEncodeDecode.
	BuildStreamPayload string
}

// ClientCLIFiles returns the client HTTP CLI support file.
func ClientCLIFiles(data *ServicesData) []*codegen.File {
	if len(data.Expressions.Services) == 0 {
		return nil
	}
	var (
		cmds []*commandData
		svcs []*expr.HTTPServiceExpr
	)
	for _, svc := range data.Expressions.Services {
		sd := data.Get(svc.Name())
		if len(sd.Endpoints) > 0 {
			command := &commandData{
				CommandData: cli.BuildCommandData(sd.Service),
				NeedDialer:  HasWebSocket(sd),
				JSONRPC:     sd.Endpoints[0].IsJSONRPC,
			}

			for _, e := range sd.Endpoints {
				sub := buildSubcommandData(sd, e)
				command.Subcommands = append(command.Subcommands, sub)
				command.CommandData.Subcommands = append(command.CommandData.Subcommands, sub.SubcommandData)
			}

			command.Example = command.Subcommands[0].Example

			cmds = append(cmds, command)
			svcs = append(svcs, svc)
		}
	}
	files := make([]*codegen.File, 0, len(data.Root.API.Servers)*2) // preallocate for CLI files
	for _, svr := range data.Root.API.Servers {
		var svrData []*commandData
		for _, name := range svr.Services {
			for i, svc := range svcs {
				if svc.Name() == name {
					svrData = append(svrData, cmds[i])
				}
			}
		}
		files = append(files, endpointParser(data.Root, svr, svrData, data))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(svc, cmds[i].CommandData, data))
	}
	return files
}

func buildSubcommandData(sd *ServiceData, e *EndpointData) *subcommandData {
	flags, buildFunction := buildFlags(sd, e)

	sub := &subcommandData{
		SubcommandData: cli.BuildSubcommandData(sd.Service, e.Method, buildFunction, flags),
	}
	if e.MultipartRequestEncoder != nil {
		sub.MultipartVarName = e.MultipartRequestEncoder.VarName
		sub.MultipartFuncName = e.MultipartRequestEncoder.FuncName
	}
	if e.Method.SkipRequestBodyEncodeDecode {
		sub.StreamFlag = streamFlag(sd.Service.Name, e.Method.Name)
		sub.BuildStreamPayload = e.BuildStreamPayload
	}
	return sub
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(root *expr.RootExpr, svr *expr.ServerExpr, data []*commandData, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	path := filepath.Join(codegen.Gendir, services.dir(), "cli", pkg, "cli.go")
	title := fmt.Sprintf("%s %s client CLI support package", svr.Name, services.label())
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
	}
	for _, sv := range svr.Services {
		svc := root.Service(sv)
		sd := services.Get(svc.Name)
		if sd == nil {
			continue
		}
		specs = append(specs, &codegen.ImportSpec{
			Path: genpkg + "/" + services.dir() + "/" + sd.Service.PathName + "/client",
			Name: sd.Service.PkgName + "c",
		})
		// Add interceptors import if service has client interceptors
		if len(sd.Service.ClientInterceptors) > 0 {
			specs = append(specs, services.ServiceImport(svc.Name))
		}
	}

	cliData := make([]*cli.CommandData, len(data))
	for i, cmd := range data {
		cliData[i] = cmd.CommandData
	}

	parseSection := &codegen.SectionTemplate{
		Name:   "parse-endpoint",
		Source: httpTemplates.Read(parseEndpointT),
		Data: struct {
			FlagsCode string
			Commands  []*commandData
		}{
			cli.FlagsCode(cliData),
			data,
		},
		FuncMap: map[string]any{"streamingCmdExists": streamingCmdExists},
	}
	return cli.EndpointParserFile(path, title, specs, cliData, parseSection)
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(svc *expr.HTTPServiceExpr, data *cli.CommandData, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, services.dir(), sd.Service.PathName, "client", "cli.go")
	title := fmt.Sprintf("%s %s client CLI support package", svc.Name(), services.label())
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		services.ServiceImport(svc.Name()),
	}
	return addEndpointImports(cli.PayloadBuildersFile(path, title, specs, data), services, svc.HTTPEndpoints...)
}

// buildFlags builds the flag data and build function for an endpoint.
func buildFlags(svc *ServiceData, e *EndpointData) ([]*cli.FlagData, *cli.BuildFunctionData) {
	var (
		flags         []*cli.FlagData
		buildFunction *cli.BuildFunctionData
	)

	svcn := svc.Service.Name
	en := e.Method.Name
	if e.Payload.Request.PayloadInit != nil {
		args := e.Payload.Request.PayloadInit.ClientArgs
		args = append(args, e.Payload.Request.PayloadInit.CLIArgs...)
		flags, buildFunction = makeFlags(e, args, e.Payload.Request.PayloadType)
	} else if e.Payload.Ref != "" {
		flags = append(flags, cli.NewFlagData(svcn, en, "p", e.Method.PayloadRef, e.Method.PayloadDesc, true, e.Method.PayloadEx, e.Method.PayloadDefault))
	}
	if e.Method.SkipRequestBodyEncodeDecode {
		flags = append(flags, streamFlag(svcn, en))
	}

	return flags, buildFunction
}

// makeFlags creates flag data and build function from endpoint arguments.
func makeFlags(e *EndpointData, args []*InitArgData, payload expr.DataType) ([]*cli.FlagData, *cli.BuildFunctionData) {
	fargs := make([]*cli.FlagArgData, len(args))
	pInitArgs := make([]*codegen.InitArgData, len(args))
	for i, arg := range args {
		pInitArgs[i] = &codegen.InitArgData{
			Name:         arg.VarName,
			Pointer:      arg.Pointer,
			FieldName:    arg.FieldName,
			FieldPointer: arg.FieldPointer,
			FieldType:    arg.FieldType,
			Type:         arg.Type,
		}
		fargs[i] = &cli.FlagArgData{
			Name:         arg.VarName,
			TypeName:     arg.TypeName,
			TypeRef:      arg.TypeRef,
			FieldName:    arg.FieldName,
			Description:  arg.Description,
			Required:     arg.Required,
			Example:      arg.Example,
			DefaultValue: arg.DefaultValue,
			Validate:     arg.Validate,
			OmitField:    arg.FieldName == "" && arg.VarName != "body",
		}
	}

	pInit := &cli.PayloadInitData{
		Code:                       e.Payload.Request.PayloadInit.ClientCode,
		ReturnTypeAttribute:        e.Payload.Request.PayloadInit.ReturnTypeAttribute,
		ReturnTypeAttributePointer: e.Payload.Request.PayloadInit.ReturnIsPrimitivePointer,
		ReturnIsStruct:             e.Payload.Request.PayloadInit.ReturnIsStruct,
		ReturnTypeName:             e.Payload.Request.PayloadInit.ReturnTypeName,
		ReturnTypePkg:              e.Payload.Request.PayloadInit.ReturnTypePkg,
		Args:                       pInitArgs,
	}

	return cli.MakeFlags(e.ServiceName, e.Method, fargs, payload, e.Payload.Ref, pInit)
}

// streamFlag returns the flag used to specify the upload file for endpoints
// that use SkipRequestBodyEncodeDecode.
func streamFlag(svcn, en string) *cli.FlagData {
	return cli.NewFlagData(svcn, en, "stream", "string", "path to file containing the streamed request body", true, "goa.png", nil)
}

// streamingCmdExists returns true if at least one command in the list of commands
// uses stream for sending payload/result.
func streamingCmdExists(data []*commandData) bool {
	for _, c := range data {
		if c.NeedDialer {
			return true
		}
	}
	return false
}
