package codegen

import (
	"fmt"
	"strings"

	"goa.design/goa/expr"
)

type (
	// ServiceInfo contains data about the service to be rendered
	ServiceInfo struct {
		Name string
		Description string
		PkgName string
		Endpoints []*EndpointInfo
	}

	// EndpointInfo contains data about the endpoint to be rendered
	EndpointInfo struct {
		Name string
		Description string
		Payload string
		VarName string
		Flags []*FlagData
		BuildFunction *BuildFunctionInfo
		MultipartRequestEncoder *MultipartInfo
	}

	// MultipartInfo contains the data needed to render multipart
	// encoder/decoder.
	MultipartInfo struct {
		// FuncName is the name used to generate function type.
		FuncName string
		// VarName is the name of the variable referring to the function.
		VarName string
	}

	// CommandData contains the data needed to render a command
	CommandData struct {
		// Name of command e.g. "cellar-storage"
		Name string
		// VarName is the name of the command variable e.g.
		// "cellarStorage"
		VarName string
		// Description is the help text.
		Description string
		// Subcommands is the list of endpoint commands.
		Subcommands []*SubcommandData
		// Example is a valid command invocation, starting with the
		// command name.
		Example string
		// PkgName is the service HTTP client package import name,
		// e.g. "storagec".
		PkgName string
		// NeedStream if true passes websocket specific arguments to the CLI.
		NeedStream bool
	}

	// SubcommandData contains the data needed to render a subcommand
	SubcommandData struct {
		// Name is the subcommand name e.g. "add"
		Name string
		// FullName is the subcommand full name e.g. "storageAdd"
		FullName string
		// Description is the help text.
		Description string
		// Flags is the list of flags supported by the subcommand.
		Flags []*FlagData
		// MethodVarName is the endpoint method name, e.g. "Add"
		MethodVarName string
		// BuildFunction contains the data for the payload build
		// function if any. Exclusive with Conversion.
		BuildFunction *BuildFunctionInfo
		// Conversion contains the flag value to payload conversion
		// function if any. Exclusive with BuildFunction.
		Conversion string
		// Example is a valid command invocation, starting with the
		// command name.
		Example string
		// MultipartRequestEncoder is the data necessary to render
		// multipart request encoder.
		MultipartRequestEncoder *MultipartInfo
	}

	// FlagData contains the data needed to render a flag
	FlagData struct {
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

	// BuildFunctionInfo contains the data needed to render a build function
	BuildFunctionInfo struct {
		// Name is the build payload function name.
		Name string
		// ActualParams is the list of passed build function parameters.
		ActualParams []string
	}
)

// streamingCmdExists returns true if at least one command in the list of commands
// uses stream for sending payload/result.
func streamingCmdExists(data []*CommandData) bool {
	for _, c := range data {
		if c.NeedStream {
			return true
		}
	}
	return false
}

// EndpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func EndpointParser(title, path string, specs []*ImportSpec, data []*CommandData, protocol string) *File {
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

	sections := []*SectionTemplate{
		Header(title, "cli", specs),
		{Source: usageT, Data: usages},
		{Source: exampleT, Data: examples},
	}
	sections = append(sections, &SectionTemplate{
		Name:   "parse-endpoint",
		Source: parseT,
		Data:   data,
		FuncMap: map[string]interface{}{
			"streamingCmdExists": streamingCmdExists,
			"protocolHTTP": func() bool { return protocol == "HTTP" },
			"protocolGRPC": func() bool { return protocol == "GRPC" },
		},
	})
	for _, cmd := range data {
		sections = append(sections, &SectionTemplate{
			Name:    "cli-command-usage",
			Source:  commandUsageT,
			Data:    cmd,
			FuncMap: map[string]interface{}{"printDescription": printDescription},
		})
	}

	return &File{Path: path, SectionTemplates: sections}
}

func printDescription(desc string) string {
	res := strings.Replace(desc, "`", "`+\"`\"+`", -1)
	res = strings.Replace(res, "\n", "\n\t", -1)
	return res
}

// BuildCLICommandData builds the data needed by the templates to render the CLI
// parsing of the service command.
func BuildCLICommandData(svc ServiceInfo, needsStream bool) *CommandData {
	var (
		name        string
		description string
		subcommands []*SubcommandData
		example     string
	)
	{
		name = svc.Name
		description = svc.Description
		if description == "" {
			description = fmt.Sprintf("Make requests to the %q service", name)
		}
		subcommands = make([]*SubcommandData, len(svc.Endpoints))
		for i, e := range svc.Endpoints {
			subcommands[i] = buildSubcommandData(svc.Name, e)
		}
		if len(subcommands) > 0 {
			example = subcommands[0].Example
		}
	}
	return &CommandData{
		Name:        KebabCase(name),
		VarName:     Goify(name, false),
		Description: description,
		Subcommands: subcommands,
		Example:     example,
		PkgName:     svc.PkgName + "c",
		NeedStream:  needsStream,
	}
}

func buildSubcommandData(svcName string, e *EndpointInfo) *SubcommandData {
	var (
		name          string
		fullName      string
		description   string

		conversion    string
	)
	{
		en := e.Name
		name = KebabCase(en)
		fullName = GoifyTerms(svcName, en)
		description = e.Description
		if description == "" {
			description = fmt.Sprintf("Make request to the %q endpoint", e.Name)
		}

		if e.BuildFunction == nil && len(e.Flags) > 0 {
			// No build function, just convert the arg to the body type
			var convPre, convSuff string
			target := "data"
			if FlagType(e.Payload) == "JSON" {
				target = "val"
				convPre = fmt.Sprintf("var val %s\n", e.Payload)
				convSuff = "\ndata = val"
			}
			conv, check := ConversionCode(
				"*"+e.Flags[0].FullName+"Flag",
				target,
				e.Payload,
				false,
			)
			conversion = convPre + conv + convSuff
			if check {
				conversion = "var err error\n" + conversion
				conversion += "\nif err != nil {\n"
				if FlagType(e.Payload) == "JSON" {
					conversion += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid JSON for %s, example of valid JSON:\n%%s", %q)`,
						e.Flags[0].FullName+"Flag", e.Flags[0].Example)
				} else {
					conversion += fmt.Sprintf(`return nil, nil, fmt.Errorf("invalid value for %s, must be %s")`,
						e.Flags[0].FullName+"Flag", e.Flags[0].Type)
				}
				conversion += "\n}"
			}
		}
	}
	sub := &SubcommandData{
		Name:          name,
		FullName:      fullName,
		Description:   description,
		Flags:         e.Flags,
		MethodVarName: e.VarName,
		BuildFunction: e.BuildFunction,
		Conversion:    conversion,
	}
	if e.MultipartRequestEncoder != nil {
		sub.MultipartRequestEncoder = e.MultipartRequestEncoder
	}
	generateExample(sub, svcName)

	return sub
}

func generateExample(sub *SubcommandData, svc string) {
	ex := KebabCase(svc) + " " + KebabCase(sub.Name)
	for _, f := range sub.Flags {
		ex += " --" + f.Name + " " + f.Example
	}
	sub.Example = ex
}

// GoifyTerms makes valid go identifiers out of the supplied terms
func GoifyTerms(terms ...string) string {
	res := Goify(terms[0], false)
	if len(terms) == 1 {
		return res
	}
	for _, t := range terms[1:] {
		res += Goify(t, true)
	}
	return res
}

var (
	boolN    = GoNativeTypeName(expr.Boolean)
	intN     = GoNativeTypeName(expr.Int)
	int32N   = GoNativeTypeName(expr.Int32)
	int64N   = GoNativeTypeName(expr.Int64)
	uintN    = GoNativeTypeName(expr.UInt)
	uint32N  = GoNativeTypeName(expr.UInt32)
	uint64N  = GoNativeTypeName(expr.UInt64)
	float32N = GoNativeTypeName(expr.Float32)
	float64N = GoNativeTypeName(expr.Float64)
	stringN  = GoNativeTypeName(expr.String)
	bytesN   = GoNativeTypeName(expr.Bytes)
)

// ConversionCode produces the code that converts the string stored in the
// variable "from" to the value stored in the variable "to" of type typeName.
func ConversionCode(from, to, typeName string, pointer bool) (string, bool) {
	var (
		parse    string
		cast     string
		checkErr bool
	)
	target := to
	needCast := typeName != stringN && typeName != bytesN && FlagType(typeName) != "JSON"
	decl := ""
	if needCast && pointer {
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
		if pointer {
			ref = "&"
		}
		parse = parse + fmt.Sprintf("\n%s = %s%s", to, ref, target)
	}
	return parse, checkErr
}

// FlagType calculates the type of a flag
func FlagType(tname string) string {
	switch tname {
	case boolN, intN, int32N, int64N, uintN, uint32N, uint64N, float32N, float64N, stringN:
		return strings.ToUpper(tname)
	case bytesN:
		return "STRING"
	default: // Any, Array, Map, Object, User
		return "JSON"
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
{{- if protocolHTTP }}
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	{{- if streamingCmdExists . }}
	dialer goahttp.Dialer,
	connConfigFn goahttp.ConnConfigureFunc,
	{{- end }}
	{{- range $c := . }}
	{{- range .Subcommands }}
		{{- if .MultipartRequestEncoder }}
	{{ .MultipartRequestEncoder.VarName }} {{ $c.PkgName }}.{{ .MultipartRequestEncoder.FuncName }},
		{{- end }}
	{{- end }}
{{- end }}
) (goa.Endpoint, interface{}, error) {
{{- else if protocolGRPC }}
func ParseEndpoint(cc *grpc.ClientConn, opts ...grpc.CallOption) (goa.Endpoint, interface{}, error) {
{{- end }}
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

	if flag.NArg() < 2 { // two non flag args are required: SERVICE and ENDPOINT (aka COMMAND)
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = flag.Arg(0)
		switch svcn {
	{{- range . }}
		case "{{ .Name }}":
			svcf = {{ .VarName }}Flags
	{{- end }}
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(flag.Args()[1:]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = svcf.Arg(0)
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
	if svcf.NArg() > 1 {
		if err := epf.Parse(svcf.Args()[1:]); err != nil {
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
			{{- if protocolHTTP }}
			c := {{ .PkgName }}.NewClient(scheme, host, doer, enc, dec, restore{{ if .NeedStream }}, dialer, connConfigFn{{- end }})
			{{- else if protocolGRPC }}
			c := {{ .PkgName }}.NewClient(cc, opts...)
			{{- end }}
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
