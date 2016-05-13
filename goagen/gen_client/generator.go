package genclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

// Filename used to generate all data types (without the ".go" extension)
const typesFileName = "datatypes"

// Generator is the application code generator.
type Generator struct {
	outDir         string // Path to output directory
	genfiles       []string
	generatedTypes map[string]bool // Keeps track of names of user types that correspond to action payloads.
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var outDir string

	set := flag.NewFlagSet("client", flag.PanicOnError)
	set.String("design", "", "")
	set.StringVar(&outDir, "out", "", "")
	set.Parse(os.Args[2:])

	g := &Generator{outDir: outDir}

	return g.Generate(design.Design)
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
	userTypeTmpl := template.Must(template.New("userType").Funcs(funcs).Parse(userTypeTmpl))
	typeDecodeTmpl := template.Must(template.New("typeDecode").Funcs(funcs).Parse(typeDecodeTmpl))

	err := api.IterateResources(func(res *design.ResourceDefinition) error {
		return g.generateResourceClient(res, funcs)
	})
	if err != nil {
		return err
	}
	types := make(map[string]*design.UserTypeDefinition)
	for _, res := range api.Resources {
		for n, ut := range res.UserTypes() {
			types[n] = ut
		}
	}
	filename := filepath.Join(g.outDir, typesFileName+".go")
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("time"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}
	if err := file.WriteHeader("User Types", "client", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, filename)

	// Generate user and media types used by action payloads and parameters
	err = api.IterateUserTypes(func(userType *design.UserTypeDefinition) error {
		if _, ok := g.generatedTypes[userType.TypeName]; ok {
			return nil
		}
		if _, ok := types[userType.TypeName]; ok {
			g.generatedTypes[userType.TypeName] = true
			return userTypeTmpl.Execute(file, userType)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Generate media types used by action responses and their load helpers
	err = api.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(a *design.ActionDefinition) error {
			return a.IterateResponses(func(r *design.ResponseDefinition) error {
				if mt := api.MediaTypeWithIdentifier(r.MediaType); mt != nil {
					if _, ok := g.generatedTypes[mt.TypeName]; !ok {
						g.generatedTypes[mt.TypeName] = true
						if !mt.IsBuiltIn() {
							if err := userTypeTmpl.Execute(file, mt); err != nil {
								return err
							}
						}
						typeName := mt.TypeName
						if mt.IsBuiltIn() {
							elems := strings.Split(typeName, ".")
							typeName = elems[len(elems)-1]
						}
						if err := typeDecodeTmpl.Execute(file, mt); err != nil {
							return err
						}
					}
				}
				return nil
			})
		})
	})
	if err != nil {
		return err
	}

	// Generate media types used in payloads but not in responses
	err = api.IterateMediaTypes(func(mediaType *design.MediaTypeDefinition) error {
		if mediaType.IsBuiltIn() {
			return nil
		}
		if _, ok := g.generatedTypes[mediaType.TypeName]; ok {
			return nil
		}
		if _, ok := types[mediaType.TypeName]; ok {
			g.generatedTypes[mediaType.TypeName] = true
			return userTypeTmpl.Execute(file, mediaType)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateResourceClient(res *design.ResourceDefinition, funcs template.FuncMap) error {
	payloadTmpl := template.Must(template.New("payload").Funcs(funcs).Parse(payloadTmpl))
	clientsTmpl := template.Must(template.New("clients").Funcs(funcs).Parse(clientsTmpl))
	requestsTmpl := template.Must(template.New("clients").Funcs(funcs).Parse(requestsTmpl))
	clientsWSTmpl := template.Must(template.New("clients").Funcs(funcs).Parse(clientsWSTmpl))
	pathTmpl := template.Must(template.New("pathTemplate").Funcs(funcs).Parse(pathTmpl))

	resFilename := codegen.SnakeCase(res.Name)
	if resFilename == typesFileName {
		// Avoid clash with datatypes.go
		resFilename += "_client"
	}
	filename := filepath.Join(g.outDir, resFilename+".go")
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return err
	}
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
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("golang.org/x/net/websocket"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}
	if err := file.WriteHeader("", "client", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, filename)
	g.generatedTypes = make(map[string]bool)
	err = res.IterateActions(func(action *design.ActionDefinition) error {
		if action.Payload != nil {
			if err := payloadTmpl.Execute(file, action); err != nil {
				return err
			}
			g.generatedTypes[action.Payload.TypeName] = true
		}
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
		for i, r := range action.Routes {
			data := struct {
				Route *design.RouteDefinition
				Index int
			}{
				Route: r,
				Index: i,
			}
			if err := pathTmpl.Execute(file, data); err != nil {
				return err
			}
		}
		if err := clientsTmpl.Execute(file, action); err != nil {
			return err
		}
		return requestsTmpl.Execute(file, action)
	})
	if err != nil {
		return err
	}

	return file.FormatCode()
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
	var toolDir string
	toolDir, err = g.makeToolDir(api.Name)
	if err != nil {
		return
	}

	funcs := template.FuncMap{
		"add":             func(a, b int) int { return a + b },
		"cmdFieldType":    cmdFieldType,
		"defaultPath":     defaultPath,
		"escapeBackticks": escapeBackticks,
		"flagType":        flagType,
		"goify":           codegen.Goify,
		"gotypedef":       codegen.GoTypeDef,
		"gotypedesc":      codegen.GoTypeDesc,
		"gotyperef":       codegen.GoTypeRef,
		"gotypename":      codegen.GoTypeName,
		"gotyperefext":    goTypeRefExt,
		"join":            join,
		"multiComment":    multiComment,
		"pathParams":      pathParams,
		"pathParamNames":  pathParamNames,
		"pathTemplate":    pathTemplate,
		"tempvar":         codegen.Tempvar,
		"title":           strings.Title,
		"toString":        toString,
		"typeName":        typeName,
		"signerType":      signerType,
	}
	clientPkg, err := codegen.PackagePath(g.outDir)
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
	if err = g.generateClient(filepath.Join(g.outDir, "client.go"), clientPkg, funcs, api); err != nil {
		return
	}

	// Generate client/$res.go and types.go
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

// join is a code generation helper function that generates a function signature built from
// concatenating the properties (name type) of the given attribute type (assuming it's an object).
// join accepts an optional slice of strings which indicates the order in which the parameters
// should appear in the signature. If pos is specified then it must list all the parameters. If
// it's not specified then parameters are sorted alphabetically.
func join(att *design.AttributeDefinition, pos ...[]string) string {
	if att == nil {
		return ""
	}
	obj := att.Type.ToObject()
	elems := make([]string, len(obj))
	var keys []string
	if len(pos) > 0 {
		keys = pos[0]
		if len(keys) != len(obj) {
			panic("invalid position slice, lenght does not match attribute field count") // bug
		}
	} else {
		keys = make([]string, len(obj))
		i := 0
		for n := range obj {
			keys[i] = n
			i++
		}
		sort.Strings(keys)
	}
	for i, n := range keys {
		a := obj[n]
		elems[i] = fmt.Sprintf("%s %s", codegen.Goify(n, false), cmdFieldType(a.Type))
	}
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
	if t.Kind() == design.DateTimeKind || t.Kind() == design.UUIDKind {
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
		case design.StringKind, design.DateTimeKind, design.UUIDKind:
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
	case design.UUIDKind:
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

// signerType returns the name of the client signer used for the defined security model on the Action
func signerType(scheme *design.SecuritySchemeDefinition) string {
	switch scheme.Kind {
	case design.JWTSecurityKind:
		return "goaclient.JWTSigner" // goa client package imported under goaclient
	case design.OAuth2SecurityKind:
		return "goaclient.OAuth2Signer"
	case design.APIKeySecurityKind:
		return "goaclient.APIKeySigner"
	case design.BasicAuthSecurityKind:
		return "goaclient.BasicSigner"
	}
	return ""
}

// pathTemplate returns a fmt format suitable to build a request path to the reoute.
func pathTemplate(r *design.RouteDefinition) string {
	return design.WildcardRegex.ReplaceAllLiteralString(r.FullPath(), "/%v")
}

// pathParams return the function signature of the path factory function for the given route.
func pathParams(r *design.RouteDefinition) string {
	pnames := r.Params()
	params := make(design.Object, len(pnames))
	for _, p := range pnames {
		params[p] = r.Parent.Params.Type.ToObject()[p]
	}
	return join(&design.AttributeDefinition{Type: params}, pnames)
}

// pathParamNames return the names of the parameters of the path factory function for the given route.
func pathParamNames(r *design.RouteDefinition) string {
	params := r.Params()
	goified := make([]string, len(params))
	for i, p := range params {
		goified[i] = codegen.Goify(p, false)
	}
	return strings.Join(goified, ", ")
}

func typeName(mt *design.MediaTypeDefinition) string {
	name := codegen.GoTypeName(mt, mt.AllRequired(), 1, false)
	if mt.IsBuiltIn() {
		return strings.Split(name, ".")[1]
	}
	return name
}

const arrayToStringT = `	{{ $tmp := tempvar }}{{ $tmp }} := make([]string, len({{ .Name }}))
	for i, e := range {{ .Name }} {
		{{ $tmp2 := tempvar }}{{ toString "e" $tmp2 .ElemType }}
		{{ $tmp }}[i] = {{ $tmp2 }}
	}
	{{ .Target }} := strings.Join({{ $tmp }}, ",")`

const payloadTmpl = `// {{ gotypename .Payload nil 0 false }} is the {{ .Parent.Name }} {{ .Name }} action payload.
type {{ gotypename .Payload nil 1 false }} {{ gotypedef .Payload 0 true false }}
`

const userTypeTmpl = `// {{ gotypedesc . true }}
type {{ gotypename . .AllRequired 1 false }} {{ gotypedef . 0 true false }}
`

const typeDecodeTmpl = `{{ $typeName := typeName . }}{{ $funcName := printf "Decode%s" $typeName }}// {{ $funcName }} decodes the {{ $typeName }} instance encoded in r.
func {{ $funcName }}(r io.Reader, decoderFn goa.DecoderFunc) ({{ gotyperef . .AllRequired 0 false }}, error) {
	var decoded {{ gotypename . .AllRequired 0 false }}
	err := decoderFn(r).Decode(&decoded)
	return {{ if .IsObject }}&{{ end }}decoded, err
}
`

const pathTmpl = `{{ $funcName := printf "%sPath%s" (goify (printf "%s%s" .Route.Parent.Name (title .Route.Parent.Parent.Name)) true) ((or (and .Index (add .Index 1)) "") | printf "%v") }}{{/*
*/}}{{ with .Route }}// {{ $funcName }} computes a request path to the {{ .Parent.Name }} action of {{ .Parent.Parent.Name }}.
func {{ $funcName }}({{ pathParams . }}) string {
	return fmt.Sprintf("{{ pathTemplate . }}", {{ pathParamNames . }})
}
{{ end }}`

const clientsTmpl = `{{ $funcName := goify (printf "%s%s" .Name (title .Parent.Name)) true }}{{ $desc := .Description }}{{ if $desc }}{{ multiComment $desc }}{{ else }}// {{ $funcName }} makes a request to the {{ .Name }} action endpoint of the {{ .Parent.Name }} resource{{ end }}
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Payload }}, payload {{ gotyperef .Payload .Payload.AllRequired 1 false }}{{ end }}{{/*
	*/}}{{ $params := join .QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := join .Headers }}{{ if $headers }}, {{ $headers }}{{ end }}) (*http.Response, error) {
	req, err := c.New{{ $funcName }}Request(ctx, path{{ if .Payload }}, payload {{ end }}{{/*
*/}}{{ $params := .QueryParams }}{{ if $params }}{{ range $name, $att := $params.Type.ToObject }}, {{ goify $name false }}{{ end }}{{ end }}{{/*
*/}}{{ $headers := join .Headers }}{{ if $headers }}, {{ $headers }}{{ end }})
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}
`

const requestsTmpl = `{{ $funcName := goify (printf "New%s%sRequest" (title .Name) (title .Parent.Name)) true }}{{/*
*/}}// {{ $funcName }} create the request corresponding to the {{ .Name }} action endpoint of the {{ .Parent.Name }} resource
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Payload }}, payload {{ gotyperef .Payload .Payload.AllRequired 1 false }}{{ end }}{{/*
	*/}}{{ $params := join .QueryParams }}{{ if $params }}, {{ $params }}{{ end }}{{/*
	*/}}{{ $headers := join .Headers }}{{ if $headers }}, {{ $headers }}{{ end }}) (*http.Request, error) {
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
{{ else }} {{ if $att.Type.IsArray }}	if {{ goify $name false }} != nil {
	{{ end }}{{ $tmp := tempvar }}{{ toString (goify $name false) $tmp $att }}
	values.Set("{{ $name }}", {{ $tmp }})
{{ if $att.Type.IsArray }}	}
{{ end }}{{ end }}{{ end }}	u.RawQuery = values.Encode()
{{ end }}{{ end }}	req, err := http.NewRequest({{ $route := index .Routes 0 }}"{{ $route.Verb }}", u.String(), body)
	if err != nil {
		return nil, err
	}
{{ $headers := .Headers }}	header := req.Header
{{ if $headers }}{{ range $name, $att := $params.Type.ToObject }}{{ if (eq $att.Type.Kind 4) }}	header.Set("{{ $name }}", {{ goify $name false }})
{{ else }}{{ $tmp := tempvar }}{{ toString (goify $name false) $tmp $att }}
	header.Set("{{ $name }}", {{ $tmp }})
{{ end }}{{ end }}{{ end }}	header.Set("Content-Type", "application/json"){{ if .Security }}
	c.{{ goify .Security.Scheme.SchemeName true }}Signer.Sign(ctx, req){{ end }}
	return req, nil
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

const clientTmpl = `// Client is the {{ .Name }} service client.
type Client struct {
	*goaclient.Client{{range $security := .SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}
	{{ goify $security.SchemeName true }}Signer *{{ $signer }}{{ end }}{{ end }}
}

// New instantiates the client.
func New(c *http.Client) *Client {
	return &Client{
		Client: goaclient.New(c),{{range $security := .SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}
		{{ goify $security.SchemeName true }}Signer: &{{ $signer }}{},{{ end }}{{ end }}
	}
}
`
