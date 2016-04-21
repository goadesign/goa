package genclient

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
)

// Generator is the application code generator.
type Generator struct {
	genfiles []string
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	api := design.Design
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	root := &cobra.Command{
		Use:   "goagen",
		Short: "Client generator",
		Long:  "client tool and package generator",
		Run:   func(*cobra.Command, []string) { files, err = g.Generate(api) },
	}
	codegen.RegisterFlags(root)
	NewCommand().RegisterFlags(root)
	root.Execute()
	return
}

func makeToolDir(g *Generator, apiName string) (toolDir string, err error) {
	codegen.OutputDir = filepath.Join(codegen.OutputDir, "client")
	if err = os.RemoveAll(codegen.OutputDir); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, codegen.OutputDir)
	toolDir = filepath.Join(codegen.OutputDir, fmt.Sprintf("%s-cli", apiName))
	if err = os.MkdirAll(toolDir, 0755); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, toolDir)
	return
}

func (g *Generator) generateMain(mainFile string, clientPkg string, funcs template.FuncMap, api *design.APIDefinition) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io/ioutil"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport(clientPkg),
		codegen.SimpleImport("github.com/spf13/cobra"),
	}
	file, err := codegen.SourceFileFor(mainFile)
	if err != nil {
		return err
	}
	if err := file.WriteHeader("", "main", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, mainFile)

	data := map[string]interface{}{
		"API":     api,
		"Version": Version,
	}
	if err := file.ExecuteTemplate("main", mainTmpl, funcs, data); err != nil {
		return err
	}

	actions := make(map[string][]*design.ActionDefinition)
	api.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(action *design.ActionDefinition) error {
			if as, ok := actions[action.Name]; ok {
				actions[action.Name] = append(as, action)
			} else {
				actions[action.Name] = []*design.ActionDefinition{action}
			}
			return nil
		})
	})
	if err := file.ExecuteTemplate("registerCmds", registerCmdsT, funcs, actions); err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateCommands(commandsFile string, clientPkg string, funcs template.FuncMap, api *design.APIDefinition) error {
	file, err := codegen.SourceFileFor(commandsFile)
	if err != nil {
		return err
	}
	commandTypesTmpl := template.Must(template.New("commandTypes").Funcs(funcs).Parse(commandTypesTmpl))
	commandsTmpl := template.Must(template.New("commands").Funcs(funcs).Parse(commandsTmpl))
	commandsTmplWS := template.Must(template.New("commandsWS").Funcs(funcs).Parse(commandsTmplWS))
	registerTmpl := template.Must(template.New("register").Funcs(funcs).Parse(registerTmpl))

	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("log"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/spf13/cobra"),
		codegen.SimpleImport(AppPkg),
		codegen.SimpleImport(clientPkg),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("golang.org/x/net/websocket"),
	}
	if len(api.Resources) > 0 {
		imports = append(imports, codegen.NewImport("goaclient", "github.com/goadesign/goa/client"))
	}
	if err := file.WriteHeader("", "main", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, commandsFile)

	file.Write([]byte("type (\n"))
	if err := api.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(action *design.ActionDefinition) error {
			return commandTypesTmpl.Execute(file, action)
		})
	}); err != nil {
		return err
	}
	file.Write([]byte(")\n\n"))

	if err := api.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(action *design.ActionDefinition) error {
			data := map[string]interface{}{
				"Action":   action,
				"Resource": action.Parent,
			}
			var err error
			if action.WebSocket() {
				err = commandsTmplWS.Execute(file, data)
			} else {
				err = commandsTmpl.Execute(file, data)
			}
			if err != nil {
				return err
			}
			return registerTmpl.Execute(file, data)
		})
	}); err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateClient(clientFile string, clientPkg string, funcs template.FuncMap, api *design.APIDefinition) error {
	file, err := codegen.SourceFileFor(clientFile)
	if err != nil {
		return err
	}
	clientTmpl := template.Must(template.New("client").Funcs(funcs).Parse(clientTmpl))

	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.NewImport("goaclient", "github.com/goadesign/goa/client"),
	}
	if err := file.WriteHeader("", "client", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, clientFile)

	if err := clientTmpl.Execute(file, api); err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateClientResources(clientPkg string, funcs template.FuncMap, api *design.APIDefinition) error {
	clientsTmpl := template.Must(template.New("clients").Funcs(funcs).Parse(clientsTmpl))
	clientsWSTmpl := template.Must(template.New("clients").Funcs(funcs).Parse(clientsWSTmpl))
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport(AppPkg),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("golang.org/x/net/websocket"),
	}

	return api.IterateResources(func(res *design.ResourceDefinition) error {
		filename := filepath.Join(codegen.OutputDir, snakeCase(res.Name)+".go")
		file, err := codegen.SourceFileFor(filename)
		if err != nil {
			return err
		}
		if err := file.WriteHeader("", "client", imports); err != nil {
			return err
		}
		g.genfiles = append(g.genfiles, filename)

		if err := res.IterateActions(func(action *design.ActionDefinition) error {
			if action.Params != nil {
				params := make(design.Object, len(action.QueryParams.Type.ToObject()))
				for n, param := range action.QueryParams.Type.ToObject() {
					name := codegen.Goify(n, false)
					params[name] = param
				}
				action.QueryParams.Type = params
			}
			if action.Headers != nil {
				headers := make(design.Object, len(action.Headers.Type.ToObject()))
				for n, header := range action.Headers.Type.ToObject() {
					name := codegen.Goify(n, false)
					headers[name] = header
				}
				action.Headers.Type = headers
			}
			if action.WebSocket() {
				return clientsWSTmpl.Execute(file, action)
			}
			return clientsTmpl.Execute(file, action)
		}); err != nil {
			return err
		}

		return file.FormatCode()
	})
}

// Generate produces the skeleton main.
func (g *Generator) Generate(api *design.APIDefinition) (_ []string, err error) {
	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	// Make tool directory
	toolDir, err := makeToolDir(g, api.Name)
	if err != nil {
		return
	}

	funcs := template.FuncMap{
		"appPkg":          appPkg,
		"cmdFieldType":    cmdFieldType,
		"defaultPath":     defaultPath,
		"escapeBackticks": escapeBackticks,
		"flagType":        flagType,
		"goify":           codegen.Goify,
		"gotypedef":       codegen.GoTypeDef,
		"gotyperefext":    goTypeRefExt,
		"join":            join,
		"joinNames":       joinNames,
		"multiComment":    multiComment,
		"routes":          routes,
		"tempvar":         codegen.Tempvar,
		"title":           strings.Title,
		"toString":        toString,
		"signerType":      signerType,
	}
	clientPkg, err := codegen.PackagePath(codegen.OutputDir)
	if err != nil {
		return
	}
	arrayToStringTmpl = template.Must(template.New("client").Funcs(funcs).Parse(arrayToStringT))

	// Generate client/client-cli/main.go
	if err = g.generateMain(filepath.Join(toolDir, "main.go"), clientPkg, funcs, api); err != nil {
		return
	}

	// Generate client/client-cli/commands.go
	if err = g.generateCommands(filepath.Join(toolDir, "commands.go"), clientPkg, funcs, api); err != nil {
		return
	}

	// Generate client/client.go
	if err = g.generateClient(filepath.Join(codegen.OutputDir, "client.go"), clientPkg, funcs, api); err != nil {
		return
	}

	// Generate client/$res.go
	if err = g.generateClientResources(clientPkg, funcs, api); err != nil {
		return
	}

	return g.genfiles, nil
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}

// snakeCase produces the snake_case version of the given CamelCase string.
func snakeCase(name string) string {
	var b bytes.Buffer
	var lastUnderscore bool
	ln := len(name)
	if ln == 0 {
		return ""
	}
	b.WriteRune(unicode.ToLower(rune(name[0])))
	for i := 1; i < ln; i++ {
		r := rune(name[i])
		if unicode.IsUpper(r) {
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
			lastUnderscore = false
		}
	}
	return b.String()
}

// joinNames is a code generation helper function that generates a string built from concatenating
// the keys of the given attribute type (assuming it's an object).
func joinNames(att *design.AttributeDefinition) string {
	if att == nil {
		return ""
	}
	obj := att.Type.ToObject()
	names := make([]string, len(obj))
	i := 0
	for n := range obj {
		names[i] = fmt.Sprintf("cmd.%s", codegen.Goify(n, true))
		i++
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// join is a code generation helper function that generates a function signature built from
// concatenating the properties (name type) of the given attribute type (assuming it's an object).
func join(att *design.AttributeDefinition) string {
	if att == nil {
		return ""
	}
	obj := att.Type.ToObject()
	elems := make([]string, len(obj))
	i := 0
	for n, a := range obj {
		elems[i] = fmt.Sprintf("%s %s", codegen.Goify(n, false), cmdFieldType(a.Type))
		i++
	}
	sort.Strings(elems)
	return strings.Join(elems, ", ")
}

// escapeBackticks is a code generation helper that escapes backticks in a string.
func escapeBackticks(text string) string {
	return strings.Replace(text, "`", "`+\"`\"+`", -1)
}

// multiComment produces a Go comment containing the given string taking into account newlines.
func multiComment(text string) string {
	lines := strings.Split(text, "\n")
	nl := make([]string, len(lines))
	for i, l := range lines {
		nl[i] = "// " + strings.TrimSpace(l)
	}
	return strings.Join(nl, "\n")
}

// routes create the action command "Use" suffix.
func routes(action *design.ActionDefinition) string {
	var buf bytes.Buffer
	routes := action.Routes
	if len(routes) > 1 {
		buf.WriteRune('(')
	}
	paths := make([]string, len(routes))
	for i, r := range routes {
		paths[i] = fmt.Sprintf("%q", r.FullPath())
	}
	buf.WriteString(strings.Join(paths, "|"))
	if len(routes) > 1 {
		buf.WriteRune(')')
	}
	return buf.String()
}

// gotTypeRefExt computes the type reference for a type in a different package.
func goTypeRefExt(t design.DataType, tabs int, pkg string) string {
	ref := codegen.GoTypeRef(t, nil, tabs, false)
	if strings.HasPrefix(ref, "*") {
		return fmt.Sprintf("%s.%s", pkg, ref[1:])
	}
	return fmt.Sprintf("%s.%s", pkg, ref)
}

// cmdFieldType computes the Go type name used to store command flags of the given design type.
func cmdFieldType(t design.DataType) string {
	if t.Kind() == design.DateTimeKind {
		return "string"
	}
	return codegen.GoNativeType(t)
}

// template used to produce code that serializes arrays of simple values into comma separated
// strings.
var arrayToStringTmpl *template.Template

// toString generates Go code that converts the given simple type attribute into a string.
func toString(name, target string, att *design.AttributeDefinition) string {
	switch actual := att.Type.(type) {
	case design.Primitive:
		switch actual.Kind() {
		case design.IntegerKind:
			return fmt.Sprintf("%s := strconv.Itoa(%s)", target, name)
		case design.BooleanKind:
			return fmt.Sprintf("%s := strconv.FormatBool(%s)", target, name)
		case design.NumberKind:
			return fmt.Sprintf("%s := strconv.FormatFloat(%s, 'f', -1, 64)", target, name)
		case design.StringKind, design.DateTimeKind:
			return fmt.Sprintf("%s := %s", target, name)
		case design.AnyKind:
			return fmt.Sprintf("%s := fmt.Sprintf(\"%%v\", %s)", target, name)
		default:
			panic("unknown primitive type")
		}
	case *design.Array:
		data := map[string]interface{}{
			"Name":     name,
			"Target":   target,
			"ElemType": actual.ElemType,
		}
		return codegen.RunTemplate(arrayToStringTmpl, data)
	default:
		panic("cannot convert non simple type " + att.Type.Name() + " to string") // bug
	}
}

// flagType returns the flag type for the given (basic type) attribute definition.
func flagType(att *design.AttributeDefinition) string {
	switch att.Type.Kind() {
	case design.IntegerKind:
		return "Int"
	case design.NumberKind:
		return "Float64"
	case design.BooleanKind:
		return "Bool"
	case design.StringKind:
		return "String"
	case design.DateTimeKind:
		return "String"
	case design.AnyKind:
		return "String"
	case design.ArrayKind:
		return flagType(att.Type.(*design.Array).ElemType) + "Slice"
	case design.UserTypeKind:
		return flagType(att.Type.(*design.UserTypeDefinition).AttributeDefinition)
	case design.MediaTypeKind:
		return flagType(att.Type.(*design.MediaTypeDefinition).AttributeDefinition)
	default:
		panic("invalid flag attribute type " + att.Type.Name())
	}
}

// defaultPath returns the first route path for the given action that does not take any wildcard,
// empty string if none.
func defaultPath(action *design.ActionDefinition) string {
	for _, r := range action.Routes {
		candidate := r.FullPath()
		if !strings.ContainsRune(candidate, ':') {
			return candidate
		}
	}
	return ""
}

// appPkg returns the name of the generated application package
func appPkg() string {
	return path.Base(AppPkg)
}

// signerType returns the name of the client signer used for the defined security model on the Action
func signerType(scheme *design.SecuritySchemeDefinition) string {
	switch scheme.Kind {
	case design.JWTSecurityKind:
		return "goaclient.JWTSigner" // goa client package imported under goaclient
	case design.OAuth2SecurityKind:
		return "goaclient.OAuth2Signer"
	case design.BasicAuthSecurityKind:
		return "goaclient.BasicSigner"
	}
	return ""
}

const mainTmpl = `
// PrettyPrint is true if the tool output should be formatted for human consumption.
var PrettyPrint bool

func main() {
	// Create command line parser
	app := &cobra.Command{
		Use: "{{ .API.Name }}-cli",
		Short: ` + "`" + `CLI client for the {{ .API.Name }} service{{ if .API.Docs }} ({{ escapeBackticks .API.Docs.URL }}){{ end }}` + "`" + `,
	}
	c := client.New(nil)
	c.UserAgent = "{{ .API.Name }}-cli/{{ .Version }}"
	app.PersistentFlags().StringVarP(&c.Scheme, "scheme", "s", "", "Set the requests scheme")
	app.PersistentFlags().StringVarP(&c.Host, "host", "H", "{{ .API.Host }}", "API hostname")
	app.PersistentFlags().DurationVarP(&c.Timeout, "timeout", "t", time.Duration(20) * time.Second, "Set the request timeout")
	app.PersistentFlags().BoolVar(&c.Dump, "dump", false, "Dump HTTP request and response.")
	app.PersistentFlags().BoolVar(&PrettyPrint, "pp", false, "Pretty print response body")
	RegisterCommands(app, c)
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %s", err)
		os.Exit(-1)
	}
}
`

const arrayToStringT = `	{{ $tmp := tempvar }}{{ $tmp }} := make([]string, len({{ .Name }}))
	for i, e := range {{ .Name }} {
		{{ $tmp2 := tempvar }}{{ toString "e" $tmp2 .ElemType }}
		{{ $tmp }}[i] = {{ $tmp2 }}
	}
	{{ .Target }} := strings.Join({{ $tmp }}, ",")`

const commandTypesTmpl = `{{ $cmdName := goify (printf "%s%s%s" .Name (title .Parent.Name) "Command") true }}	// {{ $cmdName }} is the command line data structure for the {{ .Name }} action of {{ .Parent.Name }}
	{{ $cmdName }} struct {
{{ if .Payload }}		Payload string
{{ end }}{{ $params := .QueryParams }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type }}
{{ end }}{{ end }}{{ $headers := .Headers }}{{ if $headers }}{{ range $name, $att := $headers.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type }}
{{ end }}{{ end }}	}
`

const commandsTmplWS = `
{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title .Resource.Name)) true }}// Run establishes a websocket connection for the {{ $cmdName }} command.
func (cmd *{{ $cmdName }}) Run(c *client.Client, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
{{ $default := defaultPath .Action }}{{ if $default }}	path = "{{ $default }}"
{{ else }}	return fmt.Errorf("missing path argument")
{{ end }}	}
	logger := goa.NewLogger(log.New(os.Stderr, "", log.LstdFlags))
	ctx := goa.WithLogger(context.Background(), logger)
	ws, err := c.{{ goify (printf "%s%s" .Action.Name (title .Resource.Name)) true }}(ctx, path{{/*
	*/}}{{ $params := joinNames .Action.QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := joinNames .Action.Headers }}{{ if $headers }}, {{ $headers }}{{ end }})
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return err
	}
	go goaclient.WSWrite(ws)
	goaclient.WSRead(ws)

	return nil
}
`

const registerTmpl = `{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title .Resource.Name)) true }}// RegisterFlags registers the command flags with the command line.
func (cmd *{{ $cmdName }}) RegisterFlags(cc *cobra.Command, c *client.Client) {
{{ if .Action.Payload }}	cc.Flags().StringVar(&cmd.Payload, "payload", "", "Request JSON body")
{{ end }}{{ $params := .Action.QueryParams }}{{ if $params }}{{ range $name, $param := $params.Type.ToObject }}{{ $tmp := tempvar }}{{/*
*/}}{{ if not $param.DefaultValue }}	var {{ $tmp }} {{ cmdFieldType $param.Type }}
{{ end }}	cc.Flags().{{ flagType $param }}Var(&cmd.{{ goify $name true }}, "{{ $name }}", {{ if $param.DefaultValue }}{{ printf "%#v" $param.DefaultValue }}{{ else }}{{ $tmp }}{{ end }}, ` + "`" + `{{ escapeBackticks $param.Description }}` + "`" + `)
{{ end }}{{ end }}{{/*
*/}}{{ $headers := .Action.Headers }}{{ if $headers }}{{ range $name, $header := $headers.Type.ToObject }}{{/*
*/}}
	cc.Flags().StringVar(&cmd.{{ goify $name true }}, "{{ $name }}", {{ if $header.DefaultValue }}"{{ $header.DefaultValue }}"{{ else }}""{{ end }}, ` + "`" + `{{ escapeBackticks $header.Description }}` + "`" + `)
{{ end }}{{ end }}{{ if .Action.Security }}   c.Signer{{ goify .Action.Security.Scheme.SchemeName true }}.RegisterFlags(cc){{ end }}}`

const commandsTmpl = `
{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title .Resource.Name)) true }}// Run makes the HTTP request corresponding to the {{ $cmdName }} command.
func (cmd *{{ $cmdName }}) Run(c *client.Client, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
{{ $default := defaultPath .Action }}{{ if $default }}	path = "{{ $default }}"
{{ else }}	return fmt.Errorf("missing path argument")
{{ end }}	}
{{ if .Action.Payload }}var payload {{ gotyperefext .Action.Payload 2 appPkg }}
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
{{ if eq .Action.Payload.Type.Kind 4 }}	payload = cmd.Payload
{{ else }}			return fmt.Errorf("failed to deserialize payload: %s", err)
{{ end }}		}
	}
{{ end }}	logger := goa.NewLogger(log.New(os.Stderr, "", log.LstdFlags))
	ctx := goa.WithLogger(context.Background(), logger)
	resp, err := c.{{ goify (printf "%s%s" .Action.Name (title .Resource.Name)) true }}(ctx, path{{ if .Action.Payload }}, {{ if or .Action.Payload.Type.IsObject .Action.Payload.IsPrimitive }}&{{ end }}payload{{ else }}{{ end }}{{/*
	*/}}{{ $params := joinNames .Action.QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := joinNames .Action.Headers }}{{ if $headers }}, {{ $headers }}{{ end }})
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return err
	}

	goaclient.HandleResponse(c.Client, resp, PrettyPrint)
	return nil
}
`

const clientsWSTmpl = `{{ $funcName := goify (printf "%s%s" .Name (title .Parent.Name)) true }}{{/*
*/}}{{ $desc := .Description }}{{ if $desc }}{{ multiComment $desc }}{{ else }}// {{ $funcName }} establishes a websocket connection to the {{ .Name }} action endpoint of the {{ .Parent.Name }} resource{{ end }}
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{/*
	*/}}{{ $params := join .QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := join .Headers }}{{ if $headers }}, {{ $headers }}{{ end }}) (*websocket.Conn, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "{{ .CanonicalScheme }}"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
{{ $params := .QueryParams }}{{ if $params }}{{ if gt (len $params.Type.ToObject) 0 }}	values := u.Query()
{{ range $name, $att := $params.Type.ToObject }}{{ if (eq $att.Type.Kind 4) }}	values.Set("{{ $name }}", {{ goify $name false }})
{{ else }}{{ $tmp := tempvar }}{{ toString (goify $name false) $tmp $att }}
	values.Set("{{ $name }}", {{ $tmp }})
{{ end }}{{ end }}	u.RawQuery = values.Encode()
{{ end }}{{ end }}	return websocket.Dial(u.String(), "", u.String())
}
`

const clientsTmpl = `{{ $funcName := goify (printf "%s%s" .Name (title .Parent.Name)) true }}{{ $desc := .Description }}{{ if $desc }}{{ multiComment $desc }}{{ else }}// {{ $funcName }} makes a request to the {{ .Name }} action endpoint of the {{ .Parent.Name }} resource{{ end }}
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Payload }}, payload {{ if .Payload.Type.IsObject }}*{{ end }}{{ gotyperefext .Payload 1 appPkg }}{{ end }}{{/*
	*/}}{{ $params := join .QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := join .Headers }}{{ if $headers }}, {{ $headers }}{{ end }}) (*http.Response, error) {
	var body io.Reader
{{ if .Payload }}	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize body: %s", err)
	}
	body = bytes.NewBuffer(b)
{{ end }}	scheme := c.Scheme
	if scheme == "" {
		scheme = "{{ .CanonicalScheme }}"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
{{ $params := .QueryParams }}{{ if $params }}{{ if gt (len $params.Type.ToObject) 0 }}	values := u.Query()
{{ range $name, $att := $params.Type.ToObject }}{{ if (eq $att.Type.Kind 4) }}	values.Set("{{ $name }}", {{ goify $name false }})
{{ else }}{{ $tmp := tempvar }}{{ toString (goify $name false) $tmp $att }}
	values.Set("{{ $name }}", {{ $tmp }})
{{ end }}{{ end }}	u.RawQuery = values.Encode()
{{ end }}{{ end }}	req, err := http.NewRequest({{ $route := index .Routes 0 }}"{{ $route.Verb }}", u.String(), body)
	if err != nil {
		return nil, err
	}
{{ $headers := .Headers }}	header := req.Header
{{ if $headers }}{{ range $name, $att := $params.Type.ToObject }}{{ if (eq $att.Type.Kind 4) }}	header.Set("{{ $name }}", {{ goify $name false }})
{{ else }}{{ $tmp := tempvar }}{{ toString (goify $name false) $tmp $att }}
	header.Set("{{ $name }}", {{ $tmp }})
{{ end }}{{ end }}{{ end }}	header.Set("Content-Type", "application/json"){{ if .Security }}
    c.Signer{{ goify .Security.Scheme.SchemeName true }}.Sign(ctx, req){{ end }}
	return c.Client.Do(ctx, req)
}
`

const clientTmpl = `// Client is the {{ .Name }} service client.
type Client struct {
	*goaclient.Client{{range $security := .SecuritySchemes }}
    Signer{{ goify $security.SchemeName true }} goaclient.Signer{{ end }}
}

// New instantiates the client.
func New(c *http.Client) *Client {
	return &Client{
        Client: goaclient.New(c),{{range $security := .SecuritySchemes }}
        Signer{{ goify $security.SchemeName true }}: &{{ signerType ($security) }}{},{{ end }}
    }
}
`

// Takes map[string][]*design.ActionDefinition as input
const registerCmdsT = `// RegisterCommands all the resource action subcommands to the application command line.
func RegisterCommands(app *cobra.Command, c *client.Client) {
{{ if gt (len .) 0 }}	var command, sub *cobra.Command
{{ end }}{{ range $name, $actions := . }}	command = &cobra.Command{
		Use:   "{{ $name }}",
		Short: ` + "`" + `{{ if eq (len $actions) 1 }}{{ $a := index $actions 0 }}{{ escapeBackticks $a.Description }}{{ else }}{{ $name }} action{{ end }}` + "`" + `,
	}
{{ range $action := $actions }}{{ $cmdName := goify (printf "%s%sCommand" $action.Name (title $action.Parent.Name)) true }}{{/*
*/}}{{ $tmp := tempvar }}	{{ $tmp }} := new({{ $cmdName }})
	sub = &cobra.Command{
		Use:   ` + "`" + `{{ $action.Parent.Name }} {{ routes $action }}` + "`" + `,
		Short: ` + "`" + `{{ escapeBackticks $action.Parent.Description }}` + "`" + `,
		RunE:  func(cmd *cobra.Command, args []string) error { return {{ $tmp }}.Run(c, args) },
	}
	{{ $tmp }}.RegisterFlags(sub, c)
	command.AddCommand(sub)
{{ end }}app.AddCommand(command)
{{ end }}
}`
