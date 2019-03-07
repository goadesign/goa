package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/server"
	"goa.design/goa/expr"
)

// ClientCLIFiles returns the client HTTP CLI support file.
func ClientCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var (
		data []*server.CommandData
		svcs []*expr.HTTPServiceExpr
	)
	for _, svc := range root.API.HTTP.Services {
		sd := HTTPServices.Get(svc.Name())
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
	return server.BuildCommandData(sd.Service.Name, sd.Service.Description, sd.Service.PkgName, streamingEndpointExists(sd))
}

func buildSubcommandData(sd *ServiceData, e *EndpointData) *server.SubcommandData {
	flags, buildFunction := buildFlags(sd, e)

	sub := server.BuildSubcommandData(sd.Service.Name, e.Method, buildFunction, flags)
	if e.MultipartRequestEncoder != nil {
		sub.MultipartVarName = e.MultipartRequestEncoder.VarName
		sub.MultipartFuncName = e.MultipartRequestEncoder.FuncName
	}

	return sub
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, data []*server.CommandData) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	path := filepath.Join(codegen.Gendir, "http", "cli", pkg, "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svr.Name)
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		{Path: "goa.design/goa", Name: "goa"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
	}
	for _, sv := range svr.Services {
		svc := root.Service(sv)
		sd := HTTPServices.Get(svc.Name)
		if sd == nil {
			continue
		}
		specs = append(specs, &codegen.ImportSpec{
			Path: genpkg + "/http/" + codegen.SnakeCase(sd.Service.Name) + "/client",
			Name: sd.Service.PkgName + "c",
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
			FuncMap: map[string]interface{}{
				"streamingCmdExists": streamingCmdExists,
			},
		},
	}
	sections = append(sections, server.EndpointParserCommandUsageSections(data)...)
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svcName string, data *server.CommandData) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svcName), "client", "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svcName)
	sd := HTTPServices.Get(svcName)
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		{Path: "goa.design/goa", Name: "goa"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
		{Path: genpkg + "/" + codegen.SnakeCase(svcName), Name: sd.Service.PkgName},
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", specs),
	}
	for _, sub := range data.Subcommands {
		if sub.BuildFunction != nil {
			sections = append(sections, server.PayloadBuilderSection(sub.BuildFunction))
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func buildFlags(svc *ServiceData, e *EndpointData) ([]*server.FlagData, *server.BuildFunctionData) {
	var (
		flags         []*server.FlagData
		buildFunction *server.BuildFunctionData
	)

	svcn := svc.Service.Name
	en := e.Method.Name
	if e.Payload != nil {
		if e.Payload.Request.PayloadInit != nil {
			args := e.Payload.Request.PayloadInit.ClientArgs
			args = append(args, e.Payload.Request.PayloadInit.CLIArgs...)
			flags, buildFunction = makeFlags(e, args)
		} else if e.Payload.Ref != "" {
			ex := jsonExample(e.Method.PayloadEx)
			fn := server.GoifyTerms(svcn, en, "p")
			flags = append(flags, &server.FlagData{
				Name:        "p",
				Type:        server.FlagType(e.Method.PayloadRef),
				FullName:    fn,
				Description: e.Method.PayloadDesc,
				Required:    true,
				Example:     ex,
			})
		}
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
	)
	for i, arg := range args {
		pInitArgs[i] = &server.PayloadInitArgData{
			Name:      arg.Name,
			FieldName: arg.FieldName,
		}

		f := argToFlag(e.ServiceName, e.Method.Name, arg)
		flags[i] = f
		params[i] = f.FullName
		if arg.FieldName == "" && arg.Name != "body" {
			continue
		}
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

	pInit := server.PayloadInitData{
		Code:                e.Payload.Request.PayloadInit.ClientCode,
		ReturnTypeAttribute: e.Payload.Request.PayloadInit.ReturnTypeAttribute,
		ReturnIsStruct:      e.Payload.Request.PayloadInit.ReturnIsStruct,
		ReturnTypeName:      e.Payload.Request.PayloadInit.ReturnTypeName,
		Args:                pInitArgs,
	}

	return flags, &server.BuildFunctionData{
		Name:           "Build" + e.Method.VarName + "Payload",
		ActualParams:   params,
		FormalParams:   params,
		ServiceName:    e.ServiceName,
		MethodName:     e.Method.Name,
		ResultType:     e.Payload.Ref,
		ReturnTypeName: e.Payload.Ref, // TODO:TIM pick one or the other.
		Fields:         fdata,
		PayloadInit:    &pInit,
		CheckErr:       check,
	}
}

func jsonExample(v interface{}) string {
	b, err := json.MarshalIndent(jsonify(v), "   ", "   ")
	ex := "?"
	if err == nil {
		ex = string(b)
	}
	if strings.Contains(ex, "\n") {
		ex = "'" + strings.Replace(ex, "'", "\\'", -1) + "'"
	}
	return ex
}

func jsonify(v interface{}) interface{} {
	r := reflect.ValueOf(v)
	// In JSON, keys must be a string. But goa allows map keys to be anything.
	if r.Kind() == reflect.Map {
		keys := r.MapKeys()
		a := make(map[string]interface{}, len(keys))
		for _, k := range keys {
			kstr := k.String()
			if k.Kind() != reflect.String {
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
				}
			}
			a[kstr] = r.MapIndex(k).Interface()
		}
		if r.MapIndex(keys[0]).Kind() == reflect.Map {
			// if nested map, jsonify inner map
			for key, val := range a {
				a[key] = jsonify(val)
			}
		}
		v = a
	}
	return v
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

// streamingCmdExists returns true if at least one command in the list of commands
// uses stream for sending payload/result.
func streamingCmdExists(data []*server.CommandData) bool {
	for _, c := range data {
		if c.NeedStream {
			return true
		}
	}
	return false
}

const parseEndpointT = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	{{- if streamingCmdExists .Commands }}
	dialer goahttp.Dialer,
	connConfigFn goahttp.ConnConfigureFunc,
	{{- end }}
	{{- range $c := .Commands }}
	{{- range .Subcommands }}
		{{- if .MultipartVarName }}
	{{ .MultipartVarName }} {{ $c.PkgName }}.{{ .MultipartFuncName }},
		{{- end }}
	{{- end }}
	{{- end }}
) (goa.Endpoint, interface{}, error) {
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
			c := {{ .PkgName }}.NewClient(scheme, host, doer, enc, dec, restore{{ if .NeedStream }}, dialer, connConfigFn{{- end }})	
			switch epn {
		{{- $pkgName := .PkgName }}{{ range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}({{ if .MultipartVarName }}{{ .MultipartVarName }}{{ end }})
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
