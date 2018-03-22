package codegen

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

type (
	commandData struct {
		// Name of command e.g. "cellar-storage"
		Name string
		// VarName is the name of the command variable e.g.
		// "cellarStorage"
		VarName string
		// Description is the help text.
		Description string
		// Subcommands is the list of endpoint commands.
		Subcommands []*subcommandData
		// Example is a valid command invocation, starting with the
		// command name.
		Example string
		// PkgName is the service HTTP client package import name,
		// e.g. "storagec".
		PkgName string
	}

	subcommandData struct {
		// Name is the subcommand name e.g. "add"
		Name string
		// FullName is the subcommand full name e.g. "storageAdd"
		FullName string
		// Description is the help text.
		Description string
		// Flags is the list of flags supported by the subcommand.
		Flags []*flagData
		// MethodVarName is the endpoint method name, e.g. "Add"
		MethodVarName string
		// BuildFunction contains the data for the payload build
		// function if any. Exclusive with Conversion.
		BuildFunction *buildFunctionData
		// Conversion contains the flag value to payload conversion
		// function if any. Exclusive with BuildFunction.
		Conversion string
		// Example is a valid command invocation, starting with the
		// command name.
		Example string
		// MultipartRequestEncoder is the data necessary to render
		// multipart request encoder.
		MultipartRequestEncoder *MultipartData
	}

	flagData struct {
		// Name is the name of the flag, e.g. "list-vintage"
		Name string
		// VarName is the name of the flag variable, e.g. "listVintage"
		VarName string
		// Type is the type of the flag, e.g. INT
		Type string
		// FullName is the flag full name e.g. "storageAddVintage"
		FullName string
		// Description is the flag help text.
		Description string
		// Required is true if the flag is required.
		Required bool
		// Example returns a JSON serialized example value.
		Example string
	}

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
func ClientCLIFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	var (
		data []*commandData
		svcs []*httpdesign.ServiceExpr
	)
	for _, svc := range root.HTTPServices {
		sd := HTTPServices.Get(svc.Name())
		if len(sd.Endpoints) > 0 {
			data = append(data, buildCommandData(sd))
			svcs = append(svcs, svc)
		}
	}

	files := []*codegen.File{endpointParser(genpkg, root, data)}
	for i, svc := range svcs {
		files = append(files, payloadBuilders(genpkg, svc, data[i]))
	}
	return files
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *httpdesign.RootExpr, data []*commandData) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", "cli", "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", root.Design.API.Name)
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
	for _, svc := range root.HTTPServices {
		sd := HTTPServices.Get(svc.Name())
		specs = append(specs, &codegen.ImportSpec{
			Path: genpkg + "/http/" + codegen.SnakeCase(sd.Service.Name) + "/client",
			Name: sd.Service.PkgName + "c",
		})
	}
	usages := make([]string, len(data))
	var examples []string
	for i, cmd := range data {
		subs := make([]string, len(cmd.Subcommands))
		for i, s := range cmd.Subcommands {
			subs[i] = s.Name
		}
		var lp, rp string
		if len(subs) > 1 {
			lp = "("
			rp = ")"
		}
		usages[i] = fmt.Sprintf("%s %s%s%s", cmd.Name, lp, strings.Join(subs, "|"), rp)
		if i < 5 {
			examples = append(examples, cmd.Example)
		}
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "cli", specs),
		{Source: usageT, Data: usages},
		{Source: exampleT, Data: examples},
		{Source: parseT, Data: data},
	}
	for _, cmd := range data {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "cli-command-usage",
			Source:  commandUsageT,
			Data:    cmd,
			FuncMap: map[string]interface{}{"printDescription": printDescription},
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func printDescription(desc string) string {
	res := strings.Replace(desc, "`", "`+\"`\"+`", -1)
	res = strings.Replace(res, "\n", "\n\t", -1)
	return res
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svc *httpdesign.ServiceExpr, data *commandData) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svc.Name())
	sd := HTTPServices.Get(svc.Name())
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		{Path: "goa.design/goa", Name: "goa"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
		{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: sd.Service.PkgName},
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", specs),
	}
	for _, sub := range data.Subcommands {
		if sub.BuildFunction != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "cli-build-payload",
				Source: buildPayloadT,
				Data:   sub.BuildFunction,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// buildCommandData builds the data needed by the templates to render the CLI
// parsing of the service command.
func buildCommandData(svc *ServiceData) *commandData {
	var (
		name        string
		description string
		subcommands []*subcommandData
		example     string
	)
	{
		name = svc.Service.Name
		description = svc.Service.Description
		if description == "" {
			description = fmt.Sprintf("Make requests to the %q service", name)
		}
		subcommands = make([]*subcommandData, len(svc.Endpoints))
		for i, e := range svc.Endpoints {
			subcommands[i] = buildSubcommandData(svc, e)
		}
		if len(subcommands) > 0 {
			example = subcommands[0].Example
		}
	}
	return &commandData{
		Name:        codegen.KebabCase(name),
		VarName:     codegen.Goify(name, false),
		Description: description,
		Subcommands: subcommands,
		Example:     example,
		PkgName:     svc.Service.PkgName + "c",
	}
}

func buildSubcommandData(svc *ServiceData, e *EndpointData) *subcommandData {
	var (
		name          string
		fullName      string
		description   string
		flags         []*flagData
		buildFunction *buildFunctionData
		conversion    string
	)
	{
		svcn := svc.Service.Name
		en := e.Method.Name
		name = codegen.KebabCase(en)
		fullName = goify(svcn, en)
		description = e.Method.Description
		if description == "" {
			description = fmt.Sprintf("Make request to the %q endpoint", e.Method.Name)
		}
		if e.Payload != nil {
			var actuals []string
			var args []*InitArgData
			if e.Payload.Request.PayloadInit != nil {
				for _, arg := range e.Payload.Request.PayloadInit.ClientArgs {
					f := argToFlag(svcn, en, arg)
					flags = append(flags, f)
					actuals = append(actuals, f.FullName)
					args = append(args, arg)
				}
			} else if e.Payload.Ref != "" {
				ex := jsonExample(e.Method.PayloadEx)
				fn := goify(svcn, en, "p")
				flags = append(flags, &flagData{
					Name:        "p",
					Type:        flagType(e.Method.PayloadRef),
					FullName:    fn,
					Description: e.Method.PayloadDesc,
					Required:    true,
					Example:     ex,
				})
			}
			if len(actuals) > 0 {
				// We need a build function
				var (
					fdata   []*fieldData
					formals []string
					check   bool
				)
				{
					formals = make([]string, len(args))
					for i, arg := range args {
						formals[i] = flags[i].FullName
						if arg.FieldName == "" && arg.Name != "body" {
							continue
						}
						code, chek := fieldLoadCode(flags[i].FullName, flags[i].Type, arg)
						check = check || chek
						tn := arg.TypeRef
						if flags[i].Type == "JSON" {
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
				}
				buildFunction = &buildFunctionData{
					Name:         "Build" + e.Method.VarName + "Payload",
					ActualParams: actuals,
					FormalParams: formals,
					ServiceName:  svcn,
					MethodName:   en,
					ResultType:   e.Payload.Ref,
					Fields:       fdata,
					PayloadInit:  e.Payload.Request.PayloadInit,
					CheckErr:     check,
				}
			} else if len(flags) > 0 {
				// No build function, just convert the arg to the body type
				var convPre, convSuff string
				target := "data"
				if flagType(e.Method.Payload) == "JSON" {
					target = "val"
					convPre = fmt.Sprintf("var val %s\n", e.Method.Payload)
					convSuff = "\ndata = val"
				}
				conv, check := conversionCode(
					"*"+flags[0].FullName+"Flag",
					target,
					e.Method.Payload,
					true,
				)
				conversion = convPre + conv + convSuff
				if check {
					conversion = "var err error\n" + conversion
					conversion += "\nif err != nil {\n"
					if flagType(e.Method.Payload) == "JSON" {
						conversion += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid JSON for %s, example of valid JSON:\n%%s", %q)`,
							flags[0].FullName+"Flag", flags[0].Example)
					} else {
						conversion += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid value for %s, must be %s")`,
							flags[0].FullName+"Flag", flags[0].Type)
					}
					conversion += "\n}"
				}
			}
		}
	}
	sub := &subcommandData{
		Name:          name,
		FullName:      fullName,
		Description:   description,
		Flags:         flags,
		MethodVarName: e.Method.VarName,
		BuildFunction: buildFunction,
		Conversion:    conversion,
	}
	if e.MultipartRequestEncoder != nil {
		sub.MultipartRequestEncoder = e.MultipartRequestEncoder
	}
	ex := svc.Service.Name + " " + codegen.KebabCase(sub.Name)
	for _, f := range sub.Flags {
		ex += " --" + f.Name + " " + f.Example
	}
	sub.Example = ex

	return sub
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

func goify(terms ...string) string {
	res := codegen.Goify(terms[0], false)
	if len(terms) == 1 {
		return res
	}
	for _, t := range terms[1:] {
		res += codegen.Goify(t, true)
	}
	return res
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
		if arg.TypeName == stringN {
			var ref string
			if !arg.Required {
				ref = "&"
			}
			code = arg.Name + " = " + ref + actual
		} else {
			ex := jsonExample(arg.Example)
			code, check = conversionCode(actual, arg.Name, arg.TypeName, arg.Required)
			if check {
				code += "\nif err != nil {\n"
				if flagType(arg.TypeName) == "JSON" {
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

var (
	boolN    = codegen.GoNativeTypeName(design.Boolean)
	intN     = codegen.GoNativeTypeName(design.Int)
	int32N   = codegen.GoNativeTypeName(design.Int32)
	int64N   = codegen.GoNativeTypeName(design.Int64)
	uintN    = codegen.GoNativeTypeName(design.UInt)
	uint32N  = codegen.GoNativeTypeName(design.UInt32)
	uint64N  = codegen.GoNativeTypeName(design.UInt64)
	float32N = codegen.GoNativeTypeName(design.Float32)
	float64N = codegen.GoNativeTypeName(design.Float64)
	stringN  = codegen.GoNativeTypeName(design.String)
	bytesN   = codegen.GoNativeTypeName(design.Bytes)
)

// conversionCode produces the code that converts the string stored in the
// variable "from" to the value stored in the variable "to" of type typeName.
func conversionCode(from, to, typeName string, required bool) (string, bool) {
	var (
		parse    string
		cast     string
		checkErr bool
	)
	target := to
	needCast := typeName != stringN && typeName != bytesN && flagType(typeName) != "JSON"
	decl := ""
	if needCast && !required {
		target = "val"
		decl = ":"
	}
	switch typeName {
	case boolN:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseBool(%s)", target, decl, from)
		checkErr = true
	case intN:
		parse = fmt.Sprintf("var v int64\nv, err = strconv.ParseInt(%s, 10, 64)", from)
		cast = fmt.Sprintf("%s %s= int(v)", target, decl)
		checkErr = true
	case int32N:
		parse = fmt.Sprintf("var v int64\nv, err = strconv.ParseInt(%s, 10, 32)", from)
		cast = fmt.Sprintf("%s %s= int32(v)", target, decl)
		checkErr = true
	case int64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseInt(%s, 10, 64)", target, decl, from)
	case uintN:
		parse = fmt.Sprintf("var v uint64\nv, err = strconv.ParseUint(%s, 10, 64)", from)
		cast = fmt.Sprintf("%s %s= uint(v)", target, decl)
		checkErr = true
	case uint32N:
		parse = fmt.Sprintf("var v uint64\nv, err = strconv.ParseUint(%s, 10, 32)", from)
		cast = fmt.Sprintf("%s %s= uint32(v)", target, decl)
		checkErr = true
	case uint64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseUint(%s, 10, 64)", target, decl, from)
		checkErr = true
	case float32N:
		parse = fmt.Sprintf("var v float64\nv, err = strconv.ParseFloat(%s, 32)", from)
		cast = fmt.Sprintf("%s %s= float32(v)", target, decl)
		checkErr = true
	case float64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseFloat(%s, 64)", target, decl, from)
		checkErr = true
	case stringN:
		parse = fmt.Sprintf("%s %s= %s", target, decl, from)
	case bytesN:
		parse = fmt.Sprintf("%s %s= string(%s)", target, decl, from)
	default:
		parse = fmt.Sprintf("err = json.Unmarshal([]byte(%s), &%s)", from, target)
		checkErr = true
	}
	if !needCast {
		return parse, checkErr
	}
	if cast != "" {
		parse = parse + "\n" + cast
	}
	if to != target {
		ref := ""
		if !required {
			ref = "&"
		}
		parse = parse + fmt.Sprintf("\n%s = %s%s", to, ref, target)
	}
	return parse, checkErr
}

func flagType(tname string) string {
	switch tname {
	case boolN, intN, int32N, int64N, uintN, uint32N, uint64N, float32N, float64N, stringN:
		return strings.ToUpper(tname)
	case bytesN:
		return "STRING"
	default: // Any, Array, Map, Object, User
		return "JSON"
	}
}

func argToFlag(svcn, en string, arg *InitArgData) *flagData {
	ex := jsonExample(arg.Example)
	fn := goify(svcn, en, arg.Name)
	return &flagData{
		Name:        codegen.KebabCase(arg.Name),
		VarName:     codegen.Goify(arg.Name, false),
		Type:        flagType(arg.TypeName),
		FullName:    fn,
		Description: arg.Description,
		Required:    arg.Required,
		Example:     ex,
	}
}

// input: []string
const usageT = `// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return ` + "`" + `{{ range . }}{{ . }}
{{ end }}` + "`" + `
}
`

// input: []string
const exampleT = `// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return {{ range . }}os.Args[0] + ` + "`" + ` {{ . }}` + "`" + ` + "\n" +
	{{ end }}""
}
`

// input: []commandData
const parseT = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	{{- range $c := . }}
	{{- range .Subcommands }}
		{{- if .MultipartRequestEncoder }}
	{{ .MultipartRequestEncoder.VarName }} {{ $c.PkgName }}.{{ .MultipartRequestEncoder.FuncName }},
		{{- end }}
	{{- end }}
{{- end }}
) (goa.Endpoint, interface{}, error) {
	var (
		{{- range . }}
		{{ .VarName }}Flags = flag.NewFlagSet("{{ .Name }}", flag.ContinueOnError)
		{{ range .Subcommands }}
		{{ .FullName }}Flags = flag.NewFlagSet("{{ .Name }}", flag.ExitOnError)
		{{- $sub := . }}
		{{- range .Flags }}
		{{ .FullName }}Flag = {{ $sub.FullName }}Flags.String("{{ .Name }}", "{{ if .Required }}REQUIRED{{ end }}", {{ printf "%q" .Description }})
		{{- end }}
		{{ end }}
		{{- end }}
	)
	{{ range . -}}
	{{ $cmd := . -}}
	{{ .VarName }}Flags.Usage = {{ .VarName }}Usage
	{{ range .Subcommands -}}
	{{ .FullName }}Flags.Usage = {{ .FullName }}Usage
	{{ end }}
	{{ end }}
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if len(os.Args) < flag.NFlag()+3 {
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = os.Args[1+flag.NFlag()]
		switch svcn {
	{{- range . }}
		case "{{ .Name }}":
			svcf = {{ .VarName }}Flags
	{{- end }}
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(os.Args[2+flag.NFlag():]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = os.Args[2+flag.NFlag()+svcf.NFlag()]
		switch svcn {
	{{- range . }}
		case "{{ .Name }}":
			switch epn {
		{{- range .Subcommands }}
			case "{{ .Name }}":
				epf = {{ .FullName }}Flags
		{{ end }}
			}
	{{ end }}
		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if len(os.Args) > 2+flag.NFlag()+svcf.NFlag() {
		if err := epf.Parse(os.Args[3+flag.NFlag()+svcf.NFlag():]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
	{{- range . }}
		case "{{ .Name }}":
			c := {{ .PkgName }}.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
		{{- $pkgName := .PkgName }}{{ range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.VarName }}{{ end }})
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
				{{- range .ClientArgs }}
					{{- if .FieldName }}
	{{ if $.PayloadInit.ReturnTypeAttribute }}res{{ else }}v{{ end }}.{{ .FieldName }} = {{ .Name }}
       				{{- end }}
       			{{- end }}
       		{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}, nil

		{{- else }}

			{{- if .ReturnIsStruct }}
	payload := &{{ .ReturnTypeName }}{
				{{- range .ClientArgs }}
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

// input: commandData
const commandUsageT = `{{ printf "%sUsage displays the usage of the %s command and its subcommands." .Name .Name | comment }}
func {{ .VarName }}Usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `{{ printDescription .Description }}
Usage:
    %s [globalflags] {{ .Name }} COMMAND [flags]

COMMAND:
    {{- range .Subcommands }}
    {{ .Name }}: {{ printDescription .Description }}
    {{- end }}

Additional help:
    %s {{ .Name }} COMMAND --help
` + "`" + `, os.Args[0], os.Args[0])
}

{{- range .Subcommands }}
func {{ .FullName }}Usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s [flags] {{ $.Name }} {{ .Name }}{{range .Flags }} -{{ .Name }} {{ .Type }}{{ end }}

{{ printDescription .Description}}
	{{- range .Flags }}
    -{{ .Name }} {{ .Type }}: {{ .Description }}
	{{- end }}

Example:
    ` + "`+os.Args[0]+" + "`" + ` {{ .Example }}
` + "`" + `, os.Args[0])
}
{{ end }}
`
