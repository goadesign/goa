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
	serviceName string
	// Subcommands is the list of endpoint commands.
	Subcommands []*subcommandData
	// NeedDialer if true initializes the websocket dialer.
	NeedDialer bool
	// JSONRPC if true indicates the command targets a JSON-RPC service:
	// streaming endpoints are configured with a goahttp.ConnConfigureFunc
	// instead of a client package ConnConfigurer.
	JSONRPC bool
	// ClientInit is the client constructor called by ParseEndpoint.
	ClientInit *codegen.NameDeclaration
	// Configurer is the WebSocket configuration type accepted by ParseEndpoint.
	Configurer *codegen.NameDeclaration
	// ConfigurerLocal is the exact ParseEndpoint parameter that receives the
	// WebSocket or JSON-RPC connection configuration.
	ConfigurerLocal *cli.ParserLocalData
}

// subcommandData wraps the common SubcommandData and adds HTTP-specific fields.
type subcommandData struct {
	*cli.SubcommandData
	methodName string
	// MultipartFuncDeclaration supplies the multipart request encoder type name.
	MultipartFuncDeclaration *codegen.NameDeclaration
	// MultipartFuncName is the final multipart request encoder name kept for
	// existing plugin templates.
	//
	// Deprecated: Use MultipartFuncDeclaration.Name() after planning.
	MultipartFuncName string
	// MultipartVarName is the variable that holds the multipart request encoder.
	MultipartVarName string
	// MultipartLocal is the exact ParseEndpoint parameter that receives the
	// multipart request encoder.
	MultipartLocal *cli.ParserLocalData
	// StreamFlag is the flag used to identify the file to be streamed when
	// the endpoint uses SkipRequestBodyEncodeDecode.
	StreamFlag *cli.FlagData
	// StreamPointerVar is the exact parser variable passed to the stream payload builder.
	StreamPointerVar string
	// BuildStreamPayloadDeclaration is the generated function that builds the
	// request containing the payload and file stream.
	BuildStreamPayloadDeclaration *codegen.NameDeclaration
	// BuildStreamPayload is the final stream payload helper name kept for
	// existing plugin templates.
	//
	// Deprecated: Use BuildStreamPayloadDeclaration.Name() after planning.
	BuildStreamPayload string
}

// ClientCLIFiles returns the client HTTP CLI support files. genpkg must match
// the package used to create data.
func ClientCLIFiles(genpkg string, data *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, data)
	return clientCLIFiles(data)
}

// clientCLIFiles builds the client command file read by Plan.Link.
func clientCLIFiles(data *ServicesData) []*codegen.File {
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
				serviceName: sd.Service.Name,
				NeedDialer:  HasWebSocket(sd),
				JSONRPC:     sd.Endpoints[0].IsJSONRPC,
				ClientInit:  sd.ClientInitDeclaration,
				Configurer:  sd.ClientConnConfigurerDeclaration,
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
		files = append(files, endpointParser(svr, svrData, data))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(svc, cmds[i].CommandData, data))
	}
	return files
}

func buildSubcommandData(sd *ServiceData, e *EndpointData) *subcommandData {
	flags, buildFunction := buildFlags(sd, e)
	if buildFunction != nil {
		buildFunction.Name = e.CLIPayloadDeclaration.Name()
	}

	sub := &subcommandData{
		SubcommandData: cli.BuildSubcommandData(sd.Service, e.Method, buildFunction, flags),
		methodName:     e.Method.Name,
	}
	if e.MultipartRequestEncoder != nil {
		sub.MultipartVarName = e.MultipartRequestEncoder.VarName
		sub.MultipartFuncDeclaration = e.MultipartRequestEncoder.FuncDeclaration
		sub.MultipartFuncName = e.MultipartRequestEncoder.FuncDeclaration.Name()
	}
	if e.Method.SkipRequestBodyEncodeDecode {
		sub.StreamFlag = flags[len(flags)-1]
		sub.BuildStreamPayloadDeclaration = e.BuildStreamPayloadDeclaration
		sub.BuildStreamPayload = e.BuildStreamPayloadDeclaration.Name()
	}
	return sub
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(svr *expr.ServerExpr, data []*commandData, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	path := filepath.Join(codegen.Gendir, services.dir(), "cli", pkg, "cli.go")
	outputPackage := generatedFileOutputPackage(services, path)
	title := fmt.Sprintf("%s %s client CLI support package", svr.Name, services.label())
	specs := services.fileImports[filepathKey(path)]

	parser := services.cliParsers[svr.Name]
	if parser == nil {
		panic(fmt.Sprintf("HTTP CLI parser names are missing for server %q", svr.Name))
	}
	plannedData := make([]*commandData, len(data))
	cliData := make([]*cli.CommandData, len(data))
	var parserLocals []*cli.ParserLocalData
	for i, command := range data {
		commandNames := parser.Commands[command.serviceName]
		if commandNames == nil {
			panic(fmt.Sprintf("HTTP CLI command names are missing for service %q", command.serviceName))
		}
		commandCopy := *command
		commonCommand := *command.CommandData
		clientImport := services.PackageImport(
			outputPackage,
			genpkg+"/"+services.dir()+"/"+services.Get(command.serviceName).Service.PathName+"/client",
		)
		commonCommand.PkgName = clientImport.Name
		if commonCommand.Interceptors != nil {
			interceptors := *commonCommand.Interceptors
			interceptors.PkgName = services.ServiceImport(outputPackage, command.serviceName).Name
			commonCommand.Interceptors = &interceptors
		}
		commonCommand.UsageDeclaration = commandNames.Usage
		commandCopy.CommandData = &commonCommand
		if commandCopy.NeedDialer {
			suffix := "Configurer"
			use := "websocket configurer"
			if commandCopy.JSONRPC {
				suffix = "ConfigFn"
				use = "JSON-RPC connection configurer"
			}
			commandCopy.ConfigurerLocal = &cli.ParserLocalData{
				ServiceName:   command.serviceName,
				Use:           use,
				PreferredName: commonCommand.VarName + suffix,
			}
			parserLocals = append(parserLocals, commandCopy.ConfigurerLocal)
		}
		commandCopy.Subcommands = make([]*subcommandData, len(command.Subcommands))
		commonCommand.Subcommands = make([]*cli.SubcommandData, len(command.Subcommands))
		for j, subcommand := range command.Subcommands {
			usage := commandNames.Methods[subcommand.methodName]
			if usage == nil {
				panic(fmt.Sprintf("HTTP CLI method help name is missing for %q.%q", command.serviceName, subcommand.methodName))
			}
			subcommandCopy := *subcommand
			commonSubcommand := *subcommand.SubcommandData
			if commonSubcommand.Interceptors != nil {
				interceptors := *commonSubcommand.Interceptors
				interceptors.PkgName = services.ServiceImport(outputPackage, command.serviceName).Name
				commonSubcommand.Interceptors = &interceptors
			}
			commonSubcommand.UsageDeclaration = usage
			subcommandCopy.SubcommandData = &commonSubcommand
			if subcommandCopy.MultipartVarName != "" {
				subcommandCopy.MultipartLocal = &cli.ParserLocalData{
					ServiceName:   command.serviceName,
					MethodName:    subcommand.methodName,
					Use:           "multipart request encoder",
					PreferredName: subcommandCopy.MultipartVarName,
				}
				parserLocals = append(parserLocals, subcommandCopy.MultipartLocal)
			}
			commandCopy.Subcommands[j] = &subcommandCopy
			commonCommand.Subcommands[j] = &commonSubcommand
		}
		plannedData[i] = &commandCopy
		cliData[i] = &commonCommand
	}
	parser.PlanVariables(cliData, parserLocals)
	for _, command := range plannedData {
		for _, subcommand := range command.Subcommands {
			if subcommand.StreamFlag != nil {
				subcommand.StreamPointerVar = subcommand.StreamFlag.PointerVar
			}
		}
	}

	parseSection := &codegen.SectionTemplate{
		Name:   "parse-endpoint",
		Source: httpTemplates.Read(parseEndpointT),
		Data: struct {
			Declaration *codegen.NameDeclaration
			FlagsCode   string
			Commands    []*commandData
			Variables   *cli.ParserVariablesData
		}{
			parser.Declarations.ParseEndpoint,
			parser.FlagsCode(cliData),
			plannedData,
			parser.Variables,
		},
		FuncMap: map[string]any{"streamingCmdExists": streamingCmdExists},
	}
	return parser.EndpointParserFile(path, title, specs, cliData, parseSection)
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(svc *expr.HTTPServiceExpr, data *cli.CommandData, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, services.dir(), sd.Service.PathName, "client", "cli.go")
	title := fmt.Sprintf("%s %s client CLI support package", svc.Name(), services.label())
	return cli.PayloadBuildersFile(path, title, services.fileImports[filepathKey(path)], data)
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
		flags = append(flags, cli.NewFlagDataForPlan(svcn, en, "p", e.Payload.CLIPlan, e.Method.PayloadDesc, true, e.Method.PayloadEx, e.Method.PayloadDefault))
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
			FieldTypeRef: arg.ServiceTypeRef,
			Type:         arg.Type,
		}
		fargs[i] = &cli.FlagArgData{
			Name:         arg.VarName,
			TypeName:     arg.TypeName,
			Plan:         arg.CLIPlan,
			TypeRef:      arg.TypeRef,
			Pointer:      arg.Pointer,
			FieldName:    arg.FieldName,
			Description:  arg.Description,
			Required:     arg.Required,
			Example:      arg.Example,
			DefaultValue: arg.DefaultValue,
			OmitField:    arg.FieldName == "" && arg.VarName != "body",
		}
	}

	pInit := &cli.PayloadInitData{
		Code:                       e.Payload.Request.PayloadInit.ClientCode,
		ReturnTypeAttribute:        e.Payload.Request.PayloadInit.ReturnTypeAttribute,
		ReturnTypeAttributePointer: e.Payload.Request.PayloadInit.ReturnIsPrimitivePointer,
		ReturnTypeAttributeUnion:   e.Payload.Request.PayloadInit.ReturnIsUnion,
		OptionalBody:               e.Payload.Request.PayloadInit.ClientBodyOptional,
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
	plan := cli.NewFlagPlan(&expr.AttributeExpr{Type: expr.String}, "string", "string", nil)
	return cli.NewFlagDataForPlan(svcn, en, "stream", plan, "path to file containing the streamed request body", true, "goa.png", nil)
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
