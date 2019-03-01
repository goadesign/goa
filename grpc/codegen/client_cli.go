package codegen

import (
	"encoding/json"
	"fmt"
	"path"
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
		// ResultType is the fully qualified result type name.
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

// ClientCLIFiles returns the client gRPC CLI support file.
func ClientCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var (
		data []*codegen.CommandData
		svcs []*expr.GRPCServiceExpr
		buildFunctionsByService = map[string][]*buildFunctionData{}
	)
	for _, svc := range root.API.GRPC.Services {
		sd := GRPCServices.Get(svc.Name())
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
			}

			si := codegen.ServiceInfo {
				Name: sd.Service.Name,
				Description: sd.Service.Description,
				PkgName: sd.Service.PkgName,
				Endpoints: endpoints,
			}

			data = append(data, codegen.BuildCLICommandData(si, false))
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
	fpath := filepath.Join(codegen.Gendir, "grpc", "cli", pkg, "cli.go")
	title := svr.Name+" gRPC client CLI support package"
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

	return codegen.EndpointParser(title, fpath, specs, data, "GRPC")
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svcName string, buildFunctions []*buildFunctionData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		sd = GRPCServices.Get(svcName)
	)
	{
		svcNameSnakeCase := codegen.SnakeCase(svcName)
		fpath = filepath.Join(codegen.Gendir, "grpc", svcNameSnakeCase, "client", "cli.go")
		sections = []*codegen.SectionTemplate{
			codegen.Header(svcName+" gRPC client CLI support package", "client", []*codegen.ImportSpec{
				{Path: "encoding/json"},
				{Path: "fmt"},
				{Path: path.Join(genpkg, svcNameSnakeCase), Name: sd.Service.PkgName},
				{Path: path.Join(genpkg, "grpc", svcNameSnakeCase, pbPkgName), Name: sd.PkgName},
			}),
		}
		for _, buildFunction := range buildFunctions {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "cli-build-payload",
				Source: buildPayloadT,
				Data:   buildFunction,
			})
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func buildFlags(svc *ServiceData, e *EndpointData) ([]*codegen.FlagData, *buildFunctionData) {
	var (
		flags []*codegen.FlagData
		buildFunction *buildFunctionData
	)

	if e.Request != nil {
		args := e.Request.CLIArgs
		flags, buildFunction = makeFlags(e, args)
	}

	return flags, buildFunction
}

func makeFlags(e *EndpointData, args []*InitArgData) ([]*codegen.FlagData, *buildFunctionData) {
	var (
		fdata  []*fieldData
		flags  = make([]*codegen.FlagData, len(args))
		params = make([]string, len(args))
		check  bool
		pinit  *InitData
	)
	for i, arg := range args {
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
		fdata = append(fdata, &fieldData{
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
		pinit = e.Request.ServerConvert.Init
	}
	return flags, &buildFunctionData{
		Name:         "Build" + e.Method.VarName + "Payload",
		ActualParams: params,
		FormalParams: params,
		ServiceName:  e.ServiceName,
		MethodName:   e.Method.Name,
		ResultType:   e.PayloadRef,
		Fields:       fdata,
		PayloadInit:  pinit,
		CheckErr:     check,
		Args:         args,
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
		var {{ .VarName }} {{ .TypeName }}
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
{{- if .PayloadInit }}
	{{- with .PayloadInit }}
	{{ .Code }}
	{{- if .ReturnIsStruct }}
		{{- range .Args }}
			{{- if .FieldName }}
				payload.{{ .FieldName }} = {{ .Name }}
			{{- end }}
		{{- end }}
	{{- end }}
	return payload, nil
	{{- end }}
{{- end }}
}
`