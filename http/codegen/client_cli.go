package codegen

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

type (
	buildFunctionData struct {
		// Name is the build payload function name.
		Name string
		// ActualParams is the list of passed build function parameters.
		ActualParams []string
		// FormalParams is the list of build function formal parameter
		// names.
		FormalParams []string
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// ResultType is the fully qualified payload type name.
		ResultType string
		// Fields describes the payload fields.
		Fields []*fieldData
		// PayloadInit contains the data needed to render the function
		// body.
		PayloadInit *InitData
		// CheckErr is true if the payload initialization code requires
		// an "err error" variable that must be checked.
		CheckErr bool
		// Args contains the data needed to build payload.
		Args []*InitArgData
	}

	fieldData struct {
		// Name is the field name, e.g. "Vintage"
		Name string
		// VarName is the name of the local variable holding the field
		// value, e.g. "vintage"
		VarName string
		// TypeName is the name of the type.
		TypeName string
		// Init is the code initializing the variable.
		Init string
		// Pointer is true if the variable needs to be declared as a
		// pointer.
		Pointer bool
	}
)


// ClientCLIFiles returns the client HTTP CLI support file.
func ClientCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var (
		data []*codegen.CommandData
		svcs []*expr.HTTPServiceExpr
		buildFunctionsByService = map[string][]*buildFunctionData{}
	)
	for _, svc := range root.API.HTTP.Services {
		sd := HTTPServices.Get(svc.Name())
		if len(sd.Endpoints) > 0 {

			endpoints := make([]*codegen.EndpointInfo, len(sd.Endpoints))
			for i, e := range sd.Endpoints {
				flags, buildFunction := buildFlags(sd, e)

				endpoints[i] = &codegen.EndpointInfo {
					Name: e.Method.Name,
					Description: e.Method.Description,
					Payload: e.Method.Payload,
					VarName: e.Method.VarName,
					Flags: flags,
				}
				if buildFunction != nil {
					endpoints[i].BuildFunction = &codegen.BuildFunctionInfo{
						Name: buildFunction.Name,
						ActualParams: buildFunction.ActualParams,
					}

					// keep track of any buildFunctions so we can create the payload file
					buildFunctionsByService[svc.Name()] = append(buildFunctionsByService[svc.Name()], buildFunction)
				}
				if e.MultipartRequestEncoder != nil {
					endpoints[i].MultipartRequestEncoder = &codegen.MultipartInfo {
						FuncName: e.MultipartRequestEncoder.FuncName,
						VarName: e.MultipartRequestEncoder.VarName,
					}
				}
			}

			si := codegen.ServiceInfo {
				Name: sd.Service.Name,
				Description: sd.Service.Description,
				PkgName: sd.Service.PkgName,
				Endpoints: endpoints,
			}

			data = append(data, codegen.BuildCLICommandData(si, streamingEndpointExists(sd)))
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
	for svcName, buildFunctions := range buildFunctionsByService {
		files = append(files, payloadBuilders(genpkg, svcName, buildFunctions))
	}
	return files
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, data []*codegen.CommandData) *codegen.File {
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

	return codegen.EndpointParser(title, path, specs, data, "HTTP")
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svcName string, buildFunctions []*buildFunctionData) *codegen.File {
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
	for _, buildFunction := range buildFunctions {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "cli-build-payload",
			Source: buildPayloadT,
			Data:   buildFunction,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func buildFlags(svc *ServiceData, e *EndpointData) ([]*codegen.FlagData, *buildFunctionData) {
	var (
		flags []*codegen.FlagData
		buildFunction *buildFunctionData
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
			fn := codegen.GoifyTerms(svcn, en, "p")
			flags = append(flags, &codegen.FlagData{
				Name:        "p",
				Type:        codegen.FlagType(e.Method.PayloadRef),
				FullName:    fn,
				Description: e.Method.PayloadDesc,
				Required:    true,
				Example:     ex,
			})
		}
	}

	return flags, buildFunction
}

func makeFlags(e *EndpointData, args []*InitArgData) ([]*codegen.FlagData, *buildFunctionData) {
	var (
		fdata  []*fieldData
		flags  = make([]*codegen.FlagData, len(args))
		params = make([]string, len(args))
		check  bool
	)
	for i, arg := range args {
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
		fdata = append(fdata, &fieldData{
			Name:     arg.Name,
			VarName:  arg.Name,
			TypeName: tn,
			Init:     code,
			Pointer:  arg.Pointer,
		})
	}
	return flags, &buildFunctionData{
		Name:         "Build" + e.Method.VarName + "Payload",
		ActualParams: params,
		FormalParams: params,
		ServiceName:  e.ServiceName,
		MethodName:   e.Method.Name,
		ResultType:   e.Payload.Ref,
		Fields:       fdata,
		PayloadInit:  e.Payload.Request.PayloadInit,
		CheckErr:     check,
		Args:         args,
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
			code, check = codegen.ConversionCode(actual, arg.Name, arg.TypeName, !arg.Required && arg.DefaultValue == nil)
			if check {
				code += "\nif err != nil {\n"
				if codegen.FlagType(arg.TypeName) == "JSON" {
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

func argToFlag(svcn, en string, arg *InitArgData) *codegen.FlagData {
	ex := jsonExample(arg.Example)
	fn := codegen.GoifyTerms(svcn, en, arg.Name)
	return &codegen.FlagData{
		Name:        codegen.KebabCase(arg.Name),
		VarName:     codegen.Goify(arg.Name, false),
		Type:        codegen.FlagType(arg.TypeName),
		FullName:    fn,
		Description: arg.Description,
		Required:    arg.Required,
		Example:     ex,
	}
}

// input: buildFunctionData
const buildPayloadT = `{{ printf "%s builds the payload for the %s %s endpoint from CLI flags." .Name .ServiceName .MethodName | comment }}
func {{ .Name }}({{ range .FormalParams }}{{ . }} string, {{ end }}) ({{ .ResultType }}, error) {
	{{- if .CheckErr }}
	var err error
	{{- end }}
	{{- range .Fields }}
		{{- if .VarName }}
	var {{ .VarName }} {{ if .Pointer }}*{{ end }}{{ .TypeName }}
	{
		{{ .Init }}
	}
		{{- end }}
	{{- end }}
	{{- if .CheckErr }}
	if err != nil {
		return nil, err
	}
	{{- end }}
	{{- with .PayloadInit }}

		{{- if .ClientCode }}
	{{ .ClientCode }}
			{{- if .ReturnTypeAttribute }}
	res := &{{ .ReturnTypeName }}{
		{{ .ReturnTypeAttribute }}: v,
	}
			{{- end }}
			{{- if .ReturnIsStruct }}
				{{- range $.Args }}
					{{- if .FieldName }}
	{{ if $.PayloadInit.ReturnTypeAttribute }}res{{ else }}v{{ end }}.{{ .FieldName }} = {{ .Name }}
				{{- end }}
			{{- end }}
		{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}, nil

		{{- else }}

			{{- if .ReturnIsStruct }}
	payload := &{{ .ReturnTypeName }}{
				{{- range $.Args }}
					{{- if .FieldName }}
		{{ .FieldName }}: {{ .Name }},
					{{- end }}
				{{- end }}
	}
	return payload, nil
			{{-  end }}

		{{- end }}
	{{- end }}
}
`