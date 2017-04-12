package genclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

func (g *Generator) generateMain(mainFile string, clientPkg, cliPkg string, funcs template.FuncMap) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io/ioutil"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport(clientPkg),
		codegen.SimpleImport(cliPkg),
		codegen.SimpleImport("github.com/spf13/cobra"),
		codegen.NewImport("goaclient", "github.com/goadesign/goa/client"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}

	funcs["defaultRouteParams"] = defaultRouteParams
	funcs["defaultRouteTemplate"] = defaultRouteTemplate
	funcs["joinNames"] = joinNames
	funcs["signerSignature"] = signerSignature
	funcs["signerArgs"] = signerArgs

	file, err := codegen.SourceFileFor(mainFile)
	if err != nil {
		return err
	}
	if err := file.WriteHeader("", "main", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, mainFile)
	version := design.Design.Version
	if version == "" {
		version = "0"
	}

	hasSigners := false
	hasBasicAuthSigners := false
	hasAPIKeySigners := false
	hasTokenSigners := false
	for _, s := range g.API.SecuritySchemes {
		if signerType(s) != "" {
			hasSigners = true
			switch s.Type {
			case "basic":
				hasBasicAuthSigners = true
			case "apiKey":
				hasAPIKeySigners = true
			case "jwt", "oauth2":
				hasTokenSigners = true
			}
		}
	}

	data := struct {
		API                 *design.APIDefinition
		Version             string
		Package             string
		HasSigners          bool
		HasBasicAuthSigners bool
		HasAPIKeySigners    bool
		HasTokenSigners     bool
	}{
		API:                 g.API,
		Version:             version,
		Package:             g.Target,
		HasSigners:          hasSigners,
		HasBasicAuthSigners: hasBasicAuthSigners,
		HasAPIKeySigners:    hasAPIKeySigners,
		HasTokenSigners:     hasTokenSigners,
	}
	if err := file.ExecuteTemplate("main", mainTmpl, funcs, data); err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateCommands(commandsFile string, clientPkg string, funcs template.FuncMap) error {
	file, err := codegen.SourceFileFor(commandsFile)
	if err != nil {
		return err
	}

	funcs["defaultRouteParams"] = defaultRouteParams
	funcs["defaultRouteTemplate"] = defaultRouteTemplate
	funcs["joinNames"] = joinNames
	funcs["joinRouteParams"] = joinRouteParams
	funcs["routes"] = routes
	funcs["flagType"] = flagType
	funcs["cmdFieldType"] = cmdFieldTypeString
	funcs["formatExample"] = formatExample
	funcs["shouldAddExample"] = shouldAddExample
	funcs["kebabCase"] = codegen.KebabCase

	commandTypesTmpl := template.Must(template.New("commandTypes").Funcs(funcs).Parse(commandTypesTmpl))
	commandsTmpl := template.Must(template.New("commands").Funcs(funcs).Parse(commandsTmpl))
	commandsTmplWS := template.Must(template.New("commandsWS").Funcs(funcs).Parse(commandsTmplWS))
	downloadCommandTmpl := template.Must(template.New("download").Funcs(funcs).Parse(downloadCommandTmpl))
	registerTmpl := template.Must(template.New("register").Funcs(funcs).Parse(registerTmpl))

	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("log"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("path/filepath"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/spf13/cobra"),
		codegen.SimpleImport(clientPkg),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("golang.org/x/net/websocket"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}
	if len(g.API.Resources) > 0 {
		imports = append(imports, codegen.NewImport("goaclient", "github.com/goadesign/goa/client"))
	}
	title := fmt.Sprintf("%s: CLI Commands", g.API.Context())
	if err := file.WriteHeader(title, "cli", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, commandsFile)

	file.Write([]byte("type (\n"))
	var fs []*design.FileServerDefinition
	if err := g.API.IterateResources(func(res *design.ResourceDefinition) error {
		fs = append(fs, res.FileServers...)
		return res.IterateActions(func(action *design.ActionDefinition) error {
			return commandTypesTmpl.Execute(file, action)
		})
	}); err != nil {
		return err
	}
	if len(fs) > 0 {
		file.Write([]byte(downloadCommandType))
	}
	file.Write([]byte(")\n\n"))

	actions := make(map[string][]*design.ActionDefinition)
	hasDownloads := false
	g.API.IterateResources(func(res *design.ResourceDefinition) error {
		if len(res.FileServers) > 0 {
			hasDownloads = true
		}
		return res.IterateActions(func(action *design.ActionDefinition) error {
			name := codegen.Goify(action.Name, false)
			if as, ok := actions[name]; ok {
				actions[name] = append(as, action)
			} else {
				actions[name] = []*design.ActionDefinition{action}
			}
			return nil
		})
	})
	data := struct {
		Actions      map[string][]*design.ActionDefinition
		Package      string
		HasDownloads bool
	}{
		Actions:      actions,
		Package:      g.Target,
		HasDownloads: hasDownloads,
	}
	if err := file.ExecuteTemplate("registerCmds", registerCmdsT, funcs, data); err != nil {
		return err
	}

	var fsdata []map[string]interface{}
	g.API.IterateResources(func(res *design.ResourceDefinition) error {
		if res.FileServers != nil {
			res.IterateFileServers(func(fs *design.FileServerDefinition) error {
				wcs := design.ExtractWildcards(fs.RequestPath)
				isDir := len(wcs) > 0
				var reqDir, filename string
				if isDir {
					reqDir, _ = path.Split(fs.RequestPath)
				} else {
					_, filename = filepath.Split(fs.FilePath)
				}
				fsdata = append(fsdata, map[string]interface{}{
					"IsDir":       isDir,
					"RequestPath": fs.RequestPath,
					"FilePath":    fs.FilePath,
					"FileName":    filename,
					"Name":        g.fileServerMethod(fs),
					"RequestDir":  reqDir,
				})
				return nil
			})
		}
		return nil
	})
	if fsdata != nil {
		data := struct {
			Package     string
			FileServers []map[string]interface{}
		}{
			Package:     g.Target,
			FileServers: fsdata,
		}
		if err := downloadCommandTmpl.Execute(file, data); err != nil {
			return err
		}
	}
	err = g.API.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(action *design.ActionDefinition) error {
			data := map[string]interface{}{
				"Action":          action,
				"Resource":        action.Parent,
				"Package":         g.Target,
				"HasMultiContent": len(g.API.Consumes) > 1,
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
	nz := make(map[string]bool, len(params))
	pparams := a.PathParams()
	for _, p := range params {
		o[p] = pparams.Type.ToObject()[p]
		nz[p] = true
	}
	return &design.AttributeDefinition{Type: o, NonZeroAttributes: nz}
}

// produces a fmt template to render the first route of action.
func defaultRouteTemplate(a *design.ActionDefinition) string {
	return design.WildcardRegex.ReplaceAllLiteralString(a.Routes[0].FullPath(), "/%v")
}

// return a ',' joined list of Params as a reference to cmd.XFieldName
// ordered by the required first rules.
func joinRouteParams(action *design.ActionDefinition, att *design.AttributeDefinition) string {
	var (
		params = action.Routes[0].Params()
		elems  = make([]string, len(params))
	)
	for i, p := range params {
		patt, ok := att.Type.ToObject()[p]
		if !ok {
			continue
		}
		pf := "cmd.%s"
		if patt.Type.Kind() == design.StringKind {
			pf = "url.QueryEscape(cmd.%s)"
		}
		field := fmt.Sprintf(pf, codegen.Goify(p, true))
		elems[i] = field
	}
	return strings.Join(elems, ", ")
}

// joinNames is a code generation helper function that generates a string built from concatenating
// the keys of the given attribute type (assuming it's an object).
func joinNames(useNil bool, atts ...*design.AttributeDefinition) string {
	var elems []string
	for _, att := range atts {
		if att == nil {
			continue
		}
		obj := att.Type.ToObject()
		var names, optNames []string

		keys := make([]string, len(obj))
		i := 0
		for n := range obj {
			keys[i] = n
			i++
		}
		sort.Strings(keys)

		for _, n := range keys {
			a := obj[n]
			field := fmt.Sprintf("cmd.%s", codegen.Goify(n, true))
			if !a.Type.IsArray() && !att.IsRequired(n) && !att.IsNonZero(n) {
				if useNil {
					field = flagTypeVal(a, n, field)
				} else {
					field = "&" + field
				}
			} else if a.Type.IsArray() {
				field = flagTypeArrayVal(a, field)
			} else {
				field = flagRequiredTypeVal(a, field)
			}
			if att.IsRequired(n) {
				names = append(names, field)
			} else {
				optNames = append(optNames, field)
			}
		}
		elems = append(elems, names...)
		elems = append(elems, optNames...)
	}
	return strings.Join(elems, ", ")
}

// resolve non required, non array Param/QueryParam for access via CII flags.
// Some types need convertion from string to 'Type' before calling rich client Commands.
func flagTypeVal(a *design.AttributeDefinition, key string, field string) string {
	switch a.Type {
	case design.Integer:
		return `intFlagVal("` + key + `", ` + field + ")"
	case design.String:
		return `stringFlagVal("` + key + `", ` + field + ")"
	case design.Number, design.Boolean, design.UUID, design.DateTime, design.Any:
		return "%s"
	default:
		return "&" + field
	}
}

// resolve required Param/QueryParam for access via CII flags.
// Required Params are not generated as pointers
// Special types like Number/UUID need to be converted from String
// %s maps to specialTypeResult.Temps
func flagRequiredTypeVal(a *design.AttributeDefinition, field string) string {
	switch a.Type {
	case design.Number, design.Boolean, design.UUID, design.DateTime, design.Any:
		return "*%s"
	default:
		return field
	}
}

// resolve required Param/QueryParam for access via CII flags.
// Special types like Number/UUID need to be converted from String
// %s maps to specialTypeResult.Temps
func flagTypeArrayVal(a *design.AttributeDefinition, field string) string {
	switch a.Type.ToArray().ElemType.Type {
	case design.Number, design.Boolean, design.UUID, design.DateTime, design.Any:
		return "%s"
	}
	return field
}

// format a stirng format("%s") with the given vars as argument
func format(format string, vars []string) string {
	new := make([]interface{}, len(vars))
	for i, v := range vars {
		new[i] = v
	}
	return fmt.Sprintf(format, new...)
}

// temp structure to describe the relationship between XParams
// and their tmp var as generated in the Output. See handleSpecialTypes
type specialTypeResult struct {
	Temps  []string
	Output string
}

// generate the relation and output of specially typed Params that need
// custom convertion from String Flags to Rich objects in Client action
//
// tmp, err := uuidVal(cmd.X)
// if err != nil {
//        goa.LogError(ctx, "argument parse failed", "err", err)
//        return err
// }
// resp, err := c.ShowX(ctx, path, tmp)
//
func handleSpecialTypes(atts ...*design.AttributeDefinition) specialTypeResult {
	result := specialTypeResult{}
	for _, att := range atts {
		if att == nil {
			continue
		}
		obj := att.Type.ToObject()
		var names, optNames []string

		keys := make([]string, len(obj))
		i := 0
		for n := range obj {
			keys[i] = n
			i++
		}
		sort.Strings(keys)
		for _, n := range keys {
			a := obj[n]
			field := fmt.Sprintf("cmd.%s", codegen.Goify(n, true))
			typ := cmdFieldType(a.Type, true)
			var typeHandler, nilVal string
			if !a.Type.IsArray() {
				nilVal = `""`
				switch a.Type {
				case design.Number:
					typeHandler = "float64Val"
				case design.Boolean:
					typeHandler = "boolVal"
				case design.UUID:
					typeHandler = "uuidVal"
				case design.DateTime:
					typeHandler = "timeVal"
				case design.Any:
					typeHandler = "jsonVal"
				}

			} else if a.Type.IsArray() {
				nilVal = "nil"
				switch a.Type.ToArray().ElemType.Type {
				case design.Number:
					typeHandler = "float64Array"
				case design.Boolean:
					typeHandler = "boolArray"
				case design.UUID:
					typeHandler = "uuidArray"
				case design.DateTime:
					typeHandler = "timeArray"
				case design.Any:
					typeHandler = "jsonArray"
				}
			}
			if typeHandler != "" {
				tmpVar := codegen.Tempvar()
				if att.IsRequired(n) {
					names = append(names, tmpVar)
				} else {
					optNames = append(optNames, tmpVar)
				}

				//result.Temps = append(result.Temps, tmpVar)
				result.Output += fmt.Sprintf(`
	var %s %s
	if %s != %s {
		var err error
		%s, err = %s(%s)
		if err != nil {
			goa.LogError(ctx, "failed to parse flag into %s value", "flag", "--%s", "err", err)
			return err
		}
	}`, tmpVar, typ, field, nilVal, tmpVar, typeHandler, field, typ, n)
				if att.IsRequired(n) {
					result.Output += fmt.Sprintf(`
	if %s == nil {
		goa.LogError(ctx, "required flag is missing", "flag", "--%s")
		return fmt.Errorf("required flag %s is missing")
	}`, tmpVar, n, n)
				}
			}
		}
		result.Temps = append(result.Temps, names...)
		result.Temps = append(result.Temps, optNames...)
	}
	return result
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
		path := r.FullPath()
		matches := design.WildcardRegex.FindAllStringSubmatch(path, -1)
		for _, match := range matches {
			paramName := match[1]
			path = strings.Replace(path, ":"+paramName, strings.ToUpper(paramName), 1)
		}
		paths[i] = fmt.Sprintf("%q", path)
	}
	buf.WriteString(strings.Join(paths, "|"))
	if len(routes) > 1 {
		buf.WriteRune(')')
	}
	buf.WriteRune(']')
	return buf.String()
}

// signerSignature returns the callee signature for the signer factory function for the given security
// scheme.
func signerSignature(sec *design.SecuritySchemeDefinition) string {
	switch sec.Type {
	case "basic":
		return "user, pass string"
	case "apiKey":
		return "key, format string"
	case "jwt":
		return "source goaclient.TokenSource"
	case "oauth2":
		return "source goaclient.TokenSource"
	default:
		return ""
	}
}

// signerArgs returns the caller signature for the signer factory function for the given security
// scheme.
func signerArgs(sec *design.SecuritySchemeDefinition) string {
	switch sec.Type {
	case "basic":
		return "user, pass"
	case "apiKey":
		return "key, format"
	case "jwt":
		return "source"
	case "oauth2":
		return "source"
	default:
		return ""
	}
}

// flagType returns the flag type for the given (basic type) attribute definition.
func flagType(att *design.AttributeDefinition) string {
	switch att.Type.Kind() {
	case design.IntegerKind:
		return "Int"
	case design.NumberKind:
		return "String"
	case design.BooleanKind:
		return "String"
	case design.StringKind:
		return "String"
	case design.DateTimeKind:
		return "String"
	case design.UUIDKind:
		return "String"
	case design.AnyKind:
		return "String"
	case design.ArrayKind:
		switch att.Type.ToArray().ElemType.Type.Kind() {
		case design.NumberKind:
			return "StringSlice"
		case design.BooleanKind:
			return "StringSlice"
		default:
			return flagType(att.Type.(*design.Array).ElemType) + "Slice"
		}
	case design.UserTypeKind:
		return flagType(att.Type.(*design.UserTypeDefinition).AttributeDefinition)
	case design.MediaTypeKind:
		return flagType(att.Type.(*design.MediaTypeDefinition).AttributeDefinition)
	default:
		panic("invalid flag attribute type " + att.Type.Name())
	}
}

func shouldAddExample(ut *design.UserTypeDefinition) bool {
	if ut == nil {
		return false
	}
	return ut.Example != nil
}

func formatExample(example interface{}) string {
	if example == nil {
		return ""
	}
	data, _ := json.MarshalIndent(example, "", "   ")
	return string(data)
}

const mainTmpl = `
func main() {
	// Create command line parser
	app := &cobra.Command{
		Use: "{{ .API.Name }}-cli",
		Short: ` + "`" + `CLI client for the {{ .API.Name }} service{{ if .API.Docs }} ({{ escapeBackticks .API.Docs.URL }}){{ end }}` + "`" + `,
	}

	// Create client struct
	httpClient := newHTTPClient()
	c := {{ .Package }}.New(goaclient.HTTPClientDoer(httpClient))

	// Register global flags
	app.PersistentFlags().StringVarP(&c.Scheme, "scheme", "s", "", "Set the requests scheme")
	app.PersistentFlags().StringVarP(&c.Host, "host", "H", "{{ .API.Host }}", "API hostname")
	app.PersistentFlags().DurationVarP(&httpClient.Timeout, "timeout", "t", time.Duration(20) * time.Second, "Set the request timeout")
	app.PersistentFlags().BoolVar(&c.Dump, "dump", false, "Dump HTTP request and response.")

{{ if .HasSigners }}	// Register signer flags
{{ if .HasBasicAuthSigners }} var user, pass string
	app.PersistentFlags().StringVar(&user, "user", "", "Username used for authentication")
	app.PersistentFlags().StringVar(&pass, "pass", "", "Password used for authentication")
{{ end }}{{ if .HasAPIKeySigners }} var key, format string
	app.PersistentFlags().StringVar(&key, "key", "", "API key used for authentication")
	app.PersistentFlags().StringVar(&format, "format", "Bearer %s", "Format used to create auth header or query from key")
{{ end }}{{ if .HasTokenSigners }} var token, typ string
	app.PersistentFlags().StringVar(&token, "token", "", "Token used for authentication")
	app.PersistentFlags().StringVar(&typ, "token-type", "Bearer", "Token type used for authentication")
{{ end }}
	// Parse flags and setup signers
	app.ParseFlags(os.Args)
{{ if .HasTokenSigners }}	source := &goaclient.StaticTokenSource{
		StaticToken: &goaclient.StaticToken{Type: typ, Value: token},
	}
{{ end }}{{ end }}{{ range $security := .API.SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}{{/*
*/}}	{{ goify $security.SchemeName false }}Signer := new{{ goify $security.SchemeName true }}Signer({{ signerArgs $security }}){{ end }}
{{ end }}

	// Initialize API client
{{ range $security := .API.SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}{{/*
*/}}	c.Set{{ goify $security.SchemeName true }}Signer({{ goify $security.SchemeName false }}Signer)
{{ end }}{{ end }} c.UserAgent = "{{ .API.Name }}-cli/{{ .Version }}"

	// Register API commands
	cli.RegisterCommands(app, c)

	// Execute!
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(-1)
	}
}

// newHTTPClient returns the HTTP client used by the API client to make requests to the service.
func newHTTPClient() *http.Client {
	// TBD: Change as needed (e.g. to use a different transport to control redirection policy or
	// disable cert validation or...)
	return http.DefaultClient
}

{{ range $security := .API.SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}
// new{{ goify $security.SchemeName true }}Signer returns the request signer used for authenticating
// against the {{ $security.SchemeName }} security scheme.
func new{{ goify $security.SchemeName true }}Signer({{ signerSignature $security }}) goaclient.Signer {
{{ if eq .Type "basic" }}	return &goaclient.BasicSigner{
		Username: user,
		Password: pass,
	}
{{ else if eq .Type "apiKey" }}	return &goaclient.APIKeySigner{
		SignQuery: {{ if eq $security.In "query" }}true{{ else }}false{{ end }},
		KeyName: "{{ $security.Name }}",
		KeyValue: key,
		Format: {{ if eq $security.In "query" }}"%s"{{ else }}format{{ end }},
	}
{{ else if eq .Type "jwt" }}	return &goaclient.JWTSigner{
		TokenSource: source,
	}
{{ else if eq .Type "oauth2" }}	return &goaclient.OAuth2Signer{
		TokenSource: source,
	}
{{ end }}
}
{{ end }}{{ end }}
`

const commandTypesTmpl = `{{ $cmdName := goify (printf "%s%sCommand" .Name (title (kebabCase .Parent.Name))) true }}	// {{ $cmdName }} is the command line data structure for the {{ .Name }} action of {{ .Parent.Name }}
	{{ $cmdName }} struct {
{{ if .Payload }}		Payload string
		ContentType string
{{ end }}{{ $params := defaultRouteParams . }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type false }}
{{ end }}{{ end }}{{ $params := .QueryParams }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type false}}
{{ end }}{{ end }}{{ $headers := .Headers }}{{ if $headers }}{{ range $name, $att := $headers.Type.ToObject }}{{ if $att.Description }}		{{ multiComment $att.Description }}
{{ end }}		{{ goify $name true }} {{ cmdFieldType $att.Type false}}
{{ end }}{{ end }}		PrettyPrint bool
	}

`

const downloadCommandType = `// DownloadCommand is the command line data structure for the download command.
	DownloadCommand struct {
		// OutFile is the path to the download output file.
		OutFile string
	}

`

const commandsTmplWS = `
{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title (kebabCase .Resource.Name))) true }}// Run establishes a websocket connection for the {{ $cmdName }} command.
func (cmd *{{ $cmdName }}) Run(c *{{ .Package }}.Client, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
{{ $default := defaultPath .Action }}{{ if $default }}	path = "{{ $default }}"
{{ else }}{{ $pparams := defaultRouteParams .Action }}	path = fmt.Sprintf({{ printf "%q" (defaultRouteTemplate .Action)}}, {{ joinRouteParams .Action $pparams }})
{{ end }}	}
	logger := goa.NewLogger(log.New(os.Stderr, "", log.LstdFlags))
	ctx := goa.WithLogger(context.Background(), logger){{ $specialTypeResult := handleSpecialTypes .Action.QueryParams .Action.Headers }}{{ $specialTypeResult.Output }}
	ws, err := c.{{ goify (printf "%s%s" .Action.Name (title .Resource.Name)) true }}(ctx, path{{/*
	*/}}{{ $params := joinNames true .Action.QueryParams .Action.Headers }}{{ if $params }}, {{ format $params $specialTypeResult.Temps }}{{ end }})
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return err
	}
	go goaclient.WSWrite(ws)
	goaclient.WSRead(ws)

	return nil
}
`

const downloadCommandTmpl = `
// Run downloads files with given paths.
func (cmd *DownloadCommand) Run(c *{{ .Package }}.Client, args []string) error {
	var (
		fnf func (context.Context, string) (int64, error)
		fnd func (context.Context, string, string) (int64, error)

		rpath = args[0]
		outfile = cmd.OutFile
		logger = goa.NewLogger(log.New(os.Stderr, "", log.LstdFlags))
		ctx = goa.WithLogger(context.Background(), logger)
		err error
	)

	if rpath[0] != '/' {
		rpath = "/" + rpath
	}
{{ range .FileServers }}{{ if not .IsDir }}	if rpath == "{{ .RequestPath }}" {
		fnf = c.{{ .Name }}
		if outfile == "" {
			outfile = "{{ .FileName }}"
		}
		goto found
	}
{{ end }}{{ end }}{{ range .FileServers }}{{ if .IsDir }}	if strings.HasPrefix(rpath, "{{ .RequestDir }}") {
		fnd = c.{{ .Name }}
		rpath = rpath[{{ len .RequestDir }}:]
		if outfile == "" {
			_, outfile = path.Split(rpath)
		}
		goto found
	}
{{ end }}{{ end }}	return fmt.Errorf("don't know how to download %s", rpath)
found:
	ctx = goa.WithLogContext(ctx, "file", outfile)
	if fnf != nil {
		_, err = fnf(ctx, outfile)
	} else {
		_, err = fnd(ctx, rpath, outfile)
	}
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return err
	}

	return nil
}
`

const registerTmpl = `{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title (kebabCase .Resource.Name))) true }}// RegisterFlags registers the command flags with the command line.
func (cmd *{{ $cmdName }}) RegisterFlags(cc *cobra.Command, c *{{ .Package }}.Client) {
{{ if .Action.Payload }}	cc.Flags().StringVar(&cmd.Payload, "payload", "", "Request body encoded in JSON")
	cc.Flags().StringVar(&cmd.ContentType, "content", "", "Request content type override, e.g. 'application/x-www-form-urlencoded'")
{{ end }}{{ $pparams := defaultRouteParams .Action }}{{ if $pparams }}{{ range $pname, $pparam := $pparams.Type.ToObject }}{{ $tmp := goify $pname false }}{{/*
*/}}{{ if not $pparam.DefaultValue }}	var {{ $tmp }} {{ cmdFieldType $pparam.Type false }}
{{ end }}	cc.Flags().{{ flagType $pparam }}Var(&cmd.{{ goify $pname true }}, "{{ $pname }}", {{/*
*/}}{{ if $pparam.DefaultValue }}{{ printf "%#v" $pparam.DefaultValue }}{{ else }}{{ $tmp }}{{ end }}, ` + "`" + `{{ escapeBackticks $pparam.Description }}` + "`" + `)
{{ end }}{{ end }}{{ $params := .Action.QueryParams }}{{ if $params }}{{ range $name, $param := $params.Type.ToObject }}{{ $tmp := goify $name false }}{{/*
*/}}{{ if not $param.DefaultValue }}	var {{ $tmp }} {{ cmdFieldType $param.Type false }}
{{ end }}	cc.Flags().{{ flagType $param }}Var(&cmd.{{ goify $name true }}, "{{ $name }}", {{/*
*/}}{{ if $param.DefaultValue }}{{ printf "%#v" $param.DefaultValue }}{{ else }}{{ $tmp }}{{ end }}, ` + "`" + `{{ escapeBackticks $param.Description }}` + "`" + `)
{{ end }}{{ end }}{{ $headers := .Action.Headers }}{{ if $headers }}{{ range $name, $header := $headers.Type.ToObject }}{{/*
*/}} cc.Flags().StringVar(&cmd.{{ goify $name true }}, "{{ $name }}", {{/*
*/}}{{ if $header.DefaultValue }}{{ printf "%q" $header.DefaultValue }}{{ else }}""{{ end }}, ` + "`" + `{{ escapeBackticks $header.Description }}` + "`" + `)
{{ end }}{{ end }}}`

const commandsTmpl = `
{{ $cmdName := goify (printf "%s%sCommand" .Action.Name (title (kebabCase .Resource.Name))) true }}// Run makes the HTTP request corresponding to the {{ $cmdName }} command.
func (cmd *{{ $cmdName }}) Run(c *{{ .Package }}.Client, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
{{ $default := defaultPath .Action }}{{ if $default }}	path = "{{ $default }}"
{{ else }}{{ $pparams := defaultRouteParams .Action }}	path = fmt.Sprintf({{ printf "%q" (defaultRouteTemplate .Action) }}, {{ joinRouteParams .Action $pparams }})
{{ end }}	}
{{ if .Action.Payload }}var payload {{ gotyperefext .Action.Payload 2 .Package }}
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
{{ if eq .Action.Payload.Type.Kind 4 }}	payload = cmd.Payload
{{ else }}			return fmt.Errorf("failed to deserialize payload: %s", err)
{{ end }}		}
	}
{{ end }}	logger := goa.NewLogger(log.New(os.Stderr, "", log.LstdFlags))
	ctx := goa.WithLogger(context.Background(), logger){{ $specialTypeResult := handleSpecialTypes .Action.QueryParams .Action.Headers }}{{ $specialTypeResult.Output }}
	resp, err := c.{{ goify (printf "%s%s" .Action.Name (title .Resource.Name)) true }}(ctx, path{{ if .Action.Payload }}, {{/*
	*/}}{{ if or .Action.Payload.Type.IsObject .Action.Payload.IsPrimitive }}&{{ end }}payload{{ else }}{{ end }}{{/*
	*/}}{{ $params := joinNames true .Action.QueryParams .Action.Headers }}{{ if $params }}, {{ format $params $specialTypeResult.Temps }}{{ end }}{{/*
	*/}}{{ if and .Action.Payload .HasMultiContent }}, cmd.ContentType{{ end }})
	if err != nil {
		goa.LogError(ctx, "failed", "err", err)
		return err
	}

	goaclient.HandleResponse(c.Client, resp, cmd.PrettyPrint)
	return nil
}
`

// Takes map[string][]*design.ActionDefinition as input
const registerCmdsT = `// RegisterCommands registers the resource action CLI commands.
func RegisterCommands(app *cobra.Command, c *{{ .Package }}.Client) {
{{ with .Actions }}{{ if gt (len .) 0 }}	var command, sub *cobra.Command
{{ end }}{{ range $name, $actions := . }}	command = &cobra.Command{
		Use:   "{{ kebabCase $name }}",
		Short: ` + "`" + `{{ if eq (len $actions) 1 }}{{ $a := index $actions 0 }}{{ escapeBackticks $a.Description }}{{ else }}{{ $name }} action{{ end }}` + "`" + `,
	}
{{ range $action := $actions }}{{ $cmdName := goify (printf "%s%sCommand" $action.Name (title (kebabCase $action.Parent.Name))) true }}{{/*
*/}}{{ $tmp := tempvar }}	{{ $tmp }} := new({{ $cmdName }})
	sub = &cobra.Command{
		Use:   ` + "`" + `{{ kebabCase $action.Parent.Name }} {{ routes $action }}` + "`" + `,
		Short: ` + "`" + `{{ escapeBackticks $action.Parent.Description }}` + "`" + `,{{ if shouldAddExample $action.Payload }}
		Long:  ` + "`" + `{{ escapeBackticks $action.Parent.Description }}

Payload example:

{{ formatExample $action.Payload.Example }}` + "`" + `,{{ end }}
		RunE:  func(cmd *cobra.Command, args []string) error { return {{ $tmp }}.Run(c, args) },
	}
	{{ $tmp }}.RegisterFlags(sub, c)
	sub.PersistentFlags().BoolVar(&{{ $tmp }}.PrettyPrint, "pp", false, "Pretty print response body")
	command.AddCommand(sub)
{{ end }}app.AddCommand(command)
{{ end }}{{ end }}{{ if .HasDownloads }}
	dl := new(DownloadCommand)
	dlc := &cobra.Command{
		Use:	"download [PATH]",
		Short: "Download file with given path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return dl.Run(c, args)
		},
	}
	dlc.Flags().StringVar(&dl.OutFile, "out", "", "Output file")
	app.AddCommand(dlc)
{{ end }}}

func intFlagVal(name string, parsed int) *int {
	if hasFlag(name) {
		return &parsed
	}
	return nil
}

func float64FlagVal(name string, parsed float64) *float64 {
	if hasFlag(name) {
		return &parsed
	}
	return nil
}

func boolFlagVal(name string, parsed bool) *bool {
	if hasFlag(name) {
		return &parsed
	}
	return nil
}

func stringFlagVal(name string, parsed string) *string {
	if hasFlag(name) {
		return &parsed
	}
	return nil
}

func hasFlag(name string) bool {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--"+name) {
			return true
		}
	}
	return false
}

func jsonVal(val string) (*interface{}, error) {
	var t interface{}
	err := json.Unmarshal([]byte(val), &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func jsonArray(ins []string) ([]interface{}, error) {
	if ins == nil {
		return nil, nil
	}
	var vals []interface{}
	for _, id := range ins {
		val, err := jsonVal(id)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
	return vals, nil
}

func timeVal(val string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func timeArray(ins []string) ([]time.Time, error) {
	if ins == nil {
		return nil, nil
	}
	var vals []time.Time
	for _, id := range ins {
		val, err := timeVal(id)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *val)
	}
	return vals, nil
}

func uuidVal(val string) (*uuid.UUID, error) {
	t, err := uuid.FromString(val)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func uuidArray(ins []string) ([]uuid.UUID, error) {
	if ins == nil {
		return nil, nil
	}
	var vals []uuid.UUID
	for _, id := range ins {
		val, err := uuidVal(id)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *val)
	}
	return vals, nil
}

func float64Val(val string) (*float64, error) {
	t, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func float64Array(ins []string) ([]float64, error) {
	if ins == nil {
		return nil, nil
	}
	var vals []float64
	for _, id := range ins {
		val, err := float64Val(id)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *val)
	}
	return vals, nil
}

func boolVal(val string) (*bool, error) {
	t, err := strconv.ParseBool(val)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func boolArray(ins []string) ([]bool, error) {
	if ins == nil {
		return nil, nil
	}
	var vals []bool
	for _, id := range ins {
		val, err := boolVal(id)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *val)
	}
	return vals, nil
}`
