package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server"
	"goa.design/goa/expr"
)

// ClientCLIFiles returns the client gRPC CLI support file.
func ClientCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var (
		data []*server.CommandData
		svcs []*expr.GRPCServiceExpr
	)
	for _, svc := range root.API.GRPC.Services {
		sd := GRPCServices.Get(svc.Name())
		if len(sd.Endpoints) > 0 {
			command := buildCommandData(sd)

			for _, e := range sd.Endpoints {
				command.Subcommands = append(command.Subcommands, buildSubcommandData(sd, e))
			}

			command.Example = command.Subcommands[0].Example

			data = append(data, command)
			svcs = append(svcs, svc)
		}
	}
	if len(svcs) == 0 {
		return nil
	}

	var files []*codegen.File
	for _, svr := range root.API.Servers {
		files = append(files, endpointParser(genpkg, root, svr, data))
	}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(genpkg, svc.Name(), data[i]))
	}
	return files
}

func buildCommandData(sd *ServiceData) *server.CommandData {
	return server.BuildCommandData(sd.Service.Name, sd.Service.Description, sd.Service.PkgName, false)
}

func buildSubcommandData(sd *ServiceData, e *EndpointData) *server.SubcommandData {
	flags, buildFunction := buildFlags(sd, e)

	return server.BuildSubcommandData(sd.Service.Name, e.Method, buildFunction, flags)
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, data []*server.CommandData) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	fpath := filepath.Join(codegen.Gendir, "grpc", "cli", pkg, "cli.go")
	title := svr.Name + " gRPC client CLI support package"
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "os"},
		{Path: "goa.design/goa", Name: "goa"},
		{Path: "goa.design/goa/grpc", Name: "goagrpc"},
		{Path: "google.golang.org/grpc", Name: "grpc"},
	}
	for _, svc := range root.API.GRPC.Services {
		sd := GRPCServices.Get(svc.Name())
		if sd == nil {
			continue
		}
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, "grpc", codegen.SnakeCase(sd.Service.Name), "client"),
			Name: sd.Service.PkgName + "c",
		})
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, "grpc", codegen.SnakeCase(sd.Service.Name), pbPkgName),
			Name: codegen.SnakeCase(svc.Name()) + pbPkgName,
		})
	}

	var flagsCode bytes.Buffer
	err := server.EndpointParserFlagsSection(data).Write(&flagsCode)
	if err != nil {
		panic(err)
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "cli", specs),
		server.EndpointParserUsagesSection(data),
		server.EndpointParserExamplesSection(data),
		{
			Name:   "parse-endpoint",
			Source: parseEndpointT,
			Data: struct {
				FlagsCode string
				Commands  []*server.CommandData
			}{
				flagsCode.String(),
				data,
			},
		},
	}
	sections = append(sections, server.EndpointParserCommandUsageSections(data)...)

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svcName string, data *server.CommandData) *codegen.File {
	svcNameSnakeCase := codegen.SnakeCase(svcName)
	fpath := filepath.Join(codegen.Gendir, "grpc", svcNameSnakeCase, "client", "cli.go")
	title := svcName + " gRPC client CLI support package"
	sd := GRPCServices.Get(svcName)
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: path.Join(genpkg, svcNameSnakeCase), Name: sd.Service.PkgName},
		{Path: path.Join(genpkg, "grpc", svcNameSnakeCase, pbPkgName), Name: sd.PkgName},
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", specs),
	}
	for _, sub := range data.Subcommands {
		if sub.BuildFunction != nil {
			sections = append(sections, server.PayloadBuilderSection(sub.BuildFunction))
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func buildFlags(svc *ServiceData, e *EndpointData) ([]*server.FlagData, *server.BuildFunctionData) {
	var (
		flags         []*server.FlagData
		buildFunction *server.BuildFunctionData
	)

	if e.Request != nil {
		args := e.Request.CLIArgs
		flags, buildFunction = makeFlags(e, args)
	}

	return flags, buildFunction
}

func makeFlags(e *EndpointData, args []*InitArgData) ([]*server.FlagData, *server.BuildFunctionData) {
	var (
		fdata     []*server.FieldData
		flags     = make([]*server.FlagData, len(args))
		params    = make([]string, len(args))
		pInitArgs = make([]*server.PayloadInitArgData, len(args))
		check     bool
		pinit     *server.PayloadInitData
	)
	for i, arg := range args {
		pInitArgs[i] = &server.PayloadInitArgData{
			Name:      arg.Name,
			FieldName: arg.FieldName,
		}

		f := argToFlag(e.ServiceName, e.Method.Name, arg)
		flags[i] = f
		params[i] = f.FullName
		code, chek := fieldLoadCode(f.FullName, f.Type, arg)
		check = check || chek
		tn := arg.TypeRef
		if f.Type == "JSON" {
			// We need to declare the variable without
			// a pointer to be able to unmarshal the JSON
			// using its address.
			tn = arg.TypeName
		}
		fdata = append(fdata, &server.FieldData{
			Name:     arg.Name,
			VarName:  arg.Name,
			TypeName: tn,
			Init:     code,
			Pointer:  arg.Pointer,
		})
	}
	if e.Method.PayloadRef == "" {
		return flags, nil
	}
	if e.Request.ServerConvert != nil {
		pinit = &server.PayloadInitData{
			Code:                e.Request.ServerConvert.Init.Code,
			ReturnTypeAttribute: e.Request.ServerConvert.Init.ReturnTypeRef, // TODO::TIM is this right
			ReturnIsStruct:      e.Request.ServerConvert.Init.ReturnIsStruct,
			ReturnTypeName:      e.Request.ServerConvert.Init.ReturnVarName, // TODO::TIM is this right
			Args:                pInitArgs,
		}
	}

	return flags, &server.BuildFunctionData{
		Name:         "Build" + e.Method.VarName + "Payload",
		ActualParams: params,
		FormalParams: params,
		ServiceName:  e.ServiceName,
		MethodName:   e.Method.Name,
		ResultType:   e.PayloadRef,
		Fields:       fdata,
		PayloadInit:  pinit,
		CheckErr:     check,
	}
}

func jsonExample(v interface{}) string {
	// In JSON, keys must be a string. But goa allows map keys to be anything.
	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Map {
		keys := r.MapKeys()
		if keys[0].Kind() != reflect.String {
			a := make(map[string]interface{}, len(keys))
			var kstr string
			for _, k := range keys {
				switch t := k.Interface().(type) {
				case bool:
					kstr = strconv.FormatBool(t)
				case int32:
					kstr = strconv.FormatInt(int64(t), 10)
				case int64:
					kstr = strconv.FormatInt(t, 10)
				case int:
					kstr = strconv.Itoa(t)
				case float32:
					kstr = strconv.FormatFloat(float64(t), 'f', -1, 32)
				case float64:
					kstr = strconv.FormatFloat(t, 'f', -1, 64)
				default:
					kstr = k.String()
				}
				a[kstr] = r.MapIndex(k).Interface()
			}
			v = a
		}
	}
	b, err := json.MarshalIndent(v, "   ", "   ")
	ex := "?"
	if err == nil {
		ex = string(b)
	}
	if strings.Contains(ex, "\n") {
		ex = "'" + strings.Replace(ex, "'", "\\'", -1) + "'"
	}
	return ex
}

// fieldLoadCode returns the code of the build payload function that initializes
// one of the payload object fields. It returns the initialization code and a
// boolean indicating whether the code requires an "err" variable.
func fieldLoadCode(actual, fType string, arg *InitArgData) (string, bool) {
	var (
		code    string
		check   bool
		startIf string
		endIf   string
	)
	{
		if !arg.Required {
			startIf = fmt.Sprintf("if %s != \"\" {\n", actual)
			endIf = "\n}"
		}
		if arg.TypeName == codegen.GoNativeTypeName(expr.String) {
			ref := "&"
			if arg.Required || arg.DefaultValue != nil {
				ref = ""
			}
			code = arg.Name + " = " + ref + actual
		} else {
			ex := jsonExample(arg.Example)
			code, check = server.ConversionCode(actual, arg.Name, arg.TypeName, !arg.Required && arg.DefaultValue == nil)
			if check {
				code += "\nif err != nil {\n"
				if server.FlagType(arg.TypeName) == "JSON" {
					code += fmt.Sprintf(`return nil, fmt.Errorf("invalid JSON for %s, example of valid JSON:\n%%s", %q)`,
						arg.Name, ex)
				} else {
					code += fmt.Sprintf(`err = fmt.Errorf("invalid value for %s, must be %s")`,
						arg.Name, fType)
				}
				code += "\n}"
			}
			if arg.Validate != "" {
				code += "\n" + arg.Validate + "\n" + "if err != nil {\n\treturn nil, err\n}"
			}
		}
	}
	return fmt.Sprintf("%s%s%s", startIf, code, endIf), check
}

func argToFlag(svcn, en string, arg *InitArgData) *server.FlagData {
	ex := jsonExample(arg.Example)
	fn := server.GoifyTerms(svcn, en, arg.Name)
	return &server.FlagData{
		Name:        codegen.KebabCase(arg.Name),
		VarName:     codegen.Goify(arg.Name, false),
		Type:        server.FlagType(arg.TypeName),
		FullName:    fn,
		Description: arg.Description,
		Required:    arg.Required,
		Example:     ex,
	}
}

const parseEndpointT = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(cc *grpc.ClientConn, opts ...grpc.CallOption) (goa.Endpoint, interface{}, error) {
	{{ .FlagsCode }}
	var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
	{{- range .Commands }}
		case "{{ .Name }}":
			c := {{ .PkgName }}.NewClient(cc, opts...)	
			switch epn {
		{{- $pkgName := .PkgName }}{{ range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}()
			{{- if .BuildFunction }}
				data, err = {{ $pkgName}}.{{ .BuildFunction.Name }}({{ range .BuildFunction.ActualParams }}*{{ . }}Flag, {{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- else }}
				data = nil
			{{- end }}
		{{- end }}
			}
	{{- end }}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`
