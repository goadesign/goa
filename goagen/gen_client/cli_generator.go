package genclient

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

func makeToolDir(g *Generator, apiName string) (toolDir string, err error) {
	codegen.OutputDir = filepath.Join(codegen.OutputDir, "client")
	if err = os.RemoveAll(codegen.OutputDir); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, codegen.OutputDir)
	apiName = strings.Replace(apiName, " ", "-", -1)
	toolDir = filepath.Join(codegen.OutputDir, fmt.Sprintf("%s-cli", codegen.SnakeCase(apiName)))
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
	funcs["defaultRouteParams"] = defaultRouteParams
	funcs["defaultRouteTemplate"] = defaultRouteTemplate
	funcs["joinNames"] = joinNames
	funcs["routes"] = routes
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

	err = api.IterateResources(func(res *design.ResourceDefinition) error {
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
			err = registerTmpl.Execute(file, data)
			return err
		})
	})
	if err != nil {
		return err
	}

	return file.FormatCode()
}

// defaultRouteParams returns the parameters needed to build the first route of the given action.
func defaultRouteParams(a *design.ActionDefinition) *design.AttributeDefinition {
	r := a.Routes[0]
	params := r.Params()
	o := make(design.Object, len(params))
	pparams := a.PathParams()
	for _, p := range params {
		o[p] = pparams.Type.ToObject()[p]
	}
	return &design.AttributeDefinition{Type: o}
}

// produces a fmt template to render the first route of action.
func defaultRouteTemplate(a *design.ActionDefinition) string {
	return design.WildcardRegex.ReplaceAllLiteralString(a.Routes[0].FullPath(), "/%v")
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

// routes create the action command "Use" suffix.
func routes(action *design.ActionDefinition) string {
	var buf bytes.Buffer
	routes := action.Routes
	buf.WriteRune('[')
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
	buf.WriteRune(']')
	return buf.String()
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

const commandTypesTmpl = `{{ $cmdName := goify (printf "%s%s%s" .Name (title .Parent.Name) "Command") true }}	// {{ $cmdName }} is the command line data structure for the {{ .Name }} action of {{ .Parent.Name }}
	{{ $cmdName }} struct {
{{ if .Payload }}		Payload string
{{ end }}{{ $params := defaultRouteParams . }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type }}
{{ end }}{{ end }}{{ $params := .QueryParams }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
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
{{ else }}{{ $pparams := defaultRouteParams .Action }}	path = fmt.Sprintf("{{ defaultRouteTemplate .Action}}", {{ joinNames $pparams }})
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
{{ end }}{{ $pparams := defaultRouteParams .Action }}{{ if $pparams }}{{ range $pname, $pparam := $pparams.Type.ToObject }}{{ $tmp := goify $pname false }}{{/*
*/}}{{ if not $pparam.DefaultValue }}	var {{ $tmp }} {{ cmdFieldType $pparam.Type }}
{{ end }}	cc.Flags().{{ flagType $pparam }}Var(&cmd.{{ goify $pname true }}, "{{ $pname }}", {{/*
*/}}{{ if $pparam.DefaultValue }}{{ printf "%#v" $pparam.DefaultValue }}{{ else }}{{ $tmp }}{{ end }}, ` + "`" + `{{ escapeBackticks $pparam.Description }}` + "`" + `)
{{ end }}{{ end }}{{ $params := .Action.QueryParams }}{{ if $params }}{{ range $name, $param := $params.Type.ToObject }}{{ $tmp := goify $name false }}{{/*
*/}}{{ if not $param.DefaultValue }}	var {{ $tmp }} {{ cmdFieldType $param.Type }}
{{ end }}	cc.Flags().{{ flagType $param }}Var(&cmd.{{ goify $name true }}, "{{ $name }}", {{/*
*/}}{{ if $param.DefaultValue }}{{ printf "%#v" $param.DefaultValue }}{{ else }}{{ $tmp }}{{ end }}, ` + "`" + `{{ escapeBackticks $param.Description }}` + "`" + `)
{{ end }}{{ end }}{{ $headers := .Action.Headers }}{{ if $headers }}{{ range $name, $header := $headers.Type.ToObject }}{{/*
*/}} cc.Flags().StringVar(&cmd.{{ goify $name true }}, "{{ $name }}", {{/*
*/}}{{ if $header.DefaultValue }}{{ printf "%q" $header.DefaultValue }}{{ else }}""{{ end }}, ` + "`" + `{{ escapeBackticks $header.Description }}` + "`" + `)
{{ end }}{{ end }}{{ if .Action.Security }}   c.Signer{{ goify .Action.Security.Scheme.SchemeName true }}.RegisterFlags(cc){{ end }}}`

const commandsTmpl = `
{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title .Resource.Name)) true }}// Run makes the HTTP request corresponding to the {{ $cmdName }} command.
func (cmd *{{ $cmdName }}) Run(c *client.Client, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
{{ $default := defaultPath .Action }}{{ if $default }}	path = "{{ $default }}"
{{ else }}{{ $pparams := defaultRouteParams .Action }}	path = fmt.Sprintf("{{ defaultRouteTemplate .Action }}", {{ joinNames $pparams }})
{{ end }}	}
{{ if .Action.Payload }}var payload {{ gotyperefext .Action.Payload 2 "client" }}
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
