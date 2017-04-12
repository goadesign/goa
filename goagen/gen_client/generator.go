package genclient

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	"github.com/goadesign/goa/goagen/utils"
)

// Filename used to generate all data types (without the ".go" extension)
const typesFileName = "datatypes"

//NewGenerator returns an initialized instance of a Go Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API            *design.APIDefinition // The API definition
	OutDir         string                // Path to output directory
	Target         string                // Name of generated package
	ToolDirName    string                // Name of tool directory where CLI main is generated once
	Tool           string                // Name of CLI tool
	NoTool         bool                  // Whether to skip tool generation
	genfiles       []string
	encoders       []*genapp.EncoderTemplateData
	decoders       []*genapp.EncoderTemplateData
	encoderImports []string
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, target, toolDir, tool, ver string
		notool                             bool
	)
	dtool := defaultToolName(design.Design)

	set := flag.NewFlagSet("client", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&target, "pkg", "client", "")
	set.StringVar(&toolDir, "tooldir", "tool", "")
	set.StringVar(&tool, "tool", dtool, "")
	set.StringVar(&ver, "version", "", "")
	set.BoolVar(&notool, "notool", false, "")
	set.String("design", "", "")
	set.Bool("force", false, "")
	set.Bool("notest", false, "")
	set.Parse(os.Args[1:])

	// First check compatibility
	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	// Now proceed
	target = codegen.Goify(target, false)
	g := &Generator{OutDir: outDir, Target: target, ToolDirName: toolDir, Tool: tool, NoTool: notool, API: design.Design}

	return g.Generate()
}

// Generate generats the client package and CLI.
func (g *Generator) Generate() (_ []string, err error) {
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	firstNonEmpty := func(args ...string) string {
		for _, value := range args {
			if len(value) > 0 {
				return value
			}
		}
		return ""
	}

	g.Target = firstNonEmpty(g.Target, "client")
	g.ToolDirName = firstNonEmpty(g.ToolDirName, "tool")
	g.Tool = firstNonEmpty(g.Tool, defaultToolName(g.API))

	codegen.Reserved[g.Target] = true

	// Setup output directories as needed
	var pkgDir, toolDir, cliDir string
	{
		if !g.NoTool {
			toolDir = filepath.Join(g.OutDir, g.ToolDirName, g.Tool)
			if _, err = os.Stat(toolDir); err != nil {
				if err = os.MkdirAll(toolDir, 0755); err != nil {
					return
				}
			}

			cliDir = filepath.Join(g.OutDir, g.ToolDirName, "cli")
			if err = os.RemoveAll(cliDir); err != nil {
				return
			}
			if err = os.MkdirAll(cliDir, 0755); err != nil {
				return
			}
		}

		pkgDir = filepath.Join(g.OutDir, g.Target)
		if err = os.RemoveAll(pkgDir); err != nil {
			return
		}
		if err = os.MkdirAll(pkgDir, 0755); err != nil {
			return
		}
	}

	// Setup generation
	var funcs template.FuncMap
	var clientPkg string
	{
		funcs = template.FuncMap{
			"add":                func(a, b int) int { return a + b },
			"cmdFieldType":       cmdFieldType,
			"defaultPath":        defaultPath,
			"escapeBackticks":    escapeBackticks,
			"goify":              codegen.Goify,
			"gotypedef":          codegen.GoTypeDef,
			"gotypedesc":         codegen.GoTypeDesc,
			"gotypename":         codegen.GoTypeName,
			"gotyperef":          codegen.GoTypeRef,
			"gotyperefext":       goTypeRefExt,
			"join":               join,
			"joinStrings":        strings.Join,
			"multiComment":       multiComment,
			"pathParams":         pathParams,
			"pathTemplate":       pathTemplate,
			"signerType":         signerType,
			"tempvar":            codegen.Tempvar,
			"title":              strings.Title,
			"toString":           toString,
			"typeName":           typeName,
			"format":             format,
			"handleSpecialTypes": handleSpecialTypes,
		}
		clientPkg, err = codegen.PackagePath(pkgDir)
		if err != nil {
			return
		}
		arrayToStringTmpl = template.Must(template.New("client").Funcs(funcs).Parse(arrayToStringT))
	}

	if !g.NoTool {
		var cliPkg string
		cliPkg, err = codegen.PackagePath(cliDir)
		if err != nil {
			return
		}

		// Generate tool/main.go (only once)
		mainFile := filepath.Join(toolDir, "main.go")
		if _, err := os.Stat(mainFile); err != nil {
			g.genfiles = append(g.genfiles, toolDir)
			if err = g.generateMain(mainFile, clientPkg, cliPkg, funcs); err != nil {
				return nil, err
			}
		}

		// Generate tool/cli/commands.go
		g.genfiles = append(g.genfiles, cliDir)
		if err = g.generateCommands(filepath.Join(cliDir, "commands.go"), clientPkg, funcs); err != nil {
			return
		}
	}

	// Generate client/client.go
	g.genfiles = append(g.genfiles, pkgDir)
	if err = g.generateClient(filepath.Join(pkgDir, "client.go"), clientPkg, funcs); err != nil {
		return
	}

	// Generate client/$res.go and types.go
	if err = g.generateClientResources(pkgDir, clientPkg, funcs); err != nil {
		return
	}

	return g.genfiles, nil
}

func defaultToolName(api *design.APIDefinition) string {
	if api == nil {
		return ""
	}
	return strings.Replace(strings.ToLower(api.Name), " ", "-", -1) + "-cli"
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}

func (g *Generator) generateClient(clientFile string, clientPkg string, funcs template.FuncMap) error {
	file, err := codegen.SourceFileFor(clientFile)
	if err != nil {
		return err
	}
	clientTmpl := template.Must(template.New("client").Funcs(funcs).Parse(clientTmpl))

	// Compute list of encoders and decoders
	encoders, err := genapp.BuildEncoders(g.API.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := genapp.BuildEncoders(g.API.Consumes, false)
	if err != nil {
		return err
	}
	im := make(map[string]bool)
	for _, data := range encoders {
		im[data.PackagePath] = true
	}
	for _, data := range decoders {
		im[data.PackagePath] = true
	}
	var packagePaths []string
	for packagePath := range im {
		if packagePath != "github.com/goadesign/goa" {
			packagePaths = append(packagePaths, packagePath)
		}
	}
	sort.Strings(packagePaths)

	// Setup codegen
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.NewImport("goaclient", "github.com/goadesign/goa/client"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}
	for _, packagePath := range packagePaths {
		imports = append(imports, codegen.SimpleImport(packagePath))
	}
	title := fmt.Sprintf("%s: Client", g.API.Context())
	if err := file.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, clientFile)

	// Generate
	data := struct {
		API      *design.APIDefinition
		Encoders []*genapp.EncoderTemplateData
		Decoders []*genapp.EncoderTemplateData
	}{
		API:      g.API,
		Encoders: encoders,
		Decoders: decoders,
	}
	if err := clientTmpl.Execute(file, data); err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateClientResources(pkgDir, clientPkg string, funcs template.FuncMap) error {
	err := g.API.IterateResources(func(res *design.ResourceDefinition) error {
		return g.generateResourceClient(pkgDir, res, funcs)
	})
	if err != nil {
		return err
	}
	if err := g.generateUserTypes(pkgDir); err != nil {
		return err
	}

	return g.generateMediaTypes(pkgDir, funcs)
}

func (g *Generator) generateResourceClient(pkgDir string, res *design.ResourceDefinition, funcs template.FuncMap) error {
	payloadTmpl := template.Must(template.New("payload").Funcs(funcs).Parse(payloadTmpl))
	pathTmpl := template.Must(template.New("pathTemplate").Funcs(funcs).Parse(pathTmpl))

	resFilename := codegen.SnakeCase(res.Name)
	if resFilename == typesFileName {
		// Avoid clash with datatypes.go
		resFilename += "_client"
	}
	filename := filepath.Join(pkgDir, resFilename+".go")
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return err
	}
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("io/ioutil"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("path"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("golang.org/x/net/websocket"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}
	title := fmt.Sprintf("%s: %s Resource Client", g.API.Context(), res.Name)
	if err := file.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, filename)

	err = res.IterateFileServers(func(fs *design.FileServerDefinition) error {
		return g.generateFileServer(file, fs, funcs)
	})

	err = res.IterateActions(func(action *design.ActionDefinition) error {
		if action.Payload != nil {
			found := false
			typeName := action.Payload.TypeName
			for _, t := range design.Design.Types {
				if t.TypeName == typeName {
					found = true
					break
				}
			}
			if !found {
				if err := payloadTmpl.Execute(file, action); err != nil {
					return err
				}
			}
		}
		for i, r := range action.Routes {
			routeParams := r.Params()
			var pd []*paramData

			for _, p := range routeParams {
				requiredParams, _ := initParams(&design.AttributeDefinition{
					Type: &design.Object{
						p: action.Params.Type.ToObject()[p],
					},
					Validation: &dslengine.ValidationDefinition{
						Required: routeParams,
					},
				})
				pd = append(pd, requiredParams...)
			}

			data := struct {
				Route  *design.RouteDefinition
				Index  int
				Params []*paramData
			}{
				Route:  r,
				Index:  i,
				Params: pd,
			}
			if err := pathTmpl.Execute(file, data); err != nil {
				return err
			}
		}
		return g.generateActionClient(action, file, funcs)
	})
	if err != nil {
		return err
	}

	return file.FormatCode()
}

func (g *Generator) generateFileServer(file *codegen.SourceFile, fs *design.FileServerDefinition, funcs template.FuncMap) error {
	var (
		dir string

		fsTmpl = template.Must(template.New("fileserver").Funcs(funcs).Parse(fsTmpl))
		name   = g.fileServerMethod(fs)
		wcs    = design.ExtractWildcards(fs.RequestPath)
		scheme = "http"
	)

	if len(wcs) > 0 {
		dir = "/"
		fileElems := filepath.SplitList(fs.FilePath)
		if len(fileElems) > 1 {
			dir = fileElems[len(fileElems)-2]
		}
	}
	if len(design.Design.Schemes) > 0 {
		scheme = design.Design.Schemes[0]
	}
	requestDir, _ := path.Split(fs.RequestPath)

	data := struct {
		Name            string // Download functionn name
		RequestPath     string // File server request path
		FilePath        string // File server file path
		FileName        string // Filename being download if request path has no wildcard
		DirName         string // Parent directory name if request path has wildcard
		RequestDir      string // Request path without wildcard suffix
		CanonicalScheme string // HTTP scheme
	}{
		Name:            name,
		RequestPath:     fs.RequestPath,
		FilePath:        fs.FilePath,
		FileName:        filepath.Base(fs.FilePath),
		DirName:         dir,
		RequestDir:      requestDir,
		CanonicalScheme: scheme,
	}
	return fsTmpl.Execute(file, data)
}

func (g *Generator) generateActionClient(action *design.ActionDefinition, file *codegen.SourceFile, funcs template.FuncMap) error {
	var (
		params        []string
		names         []string
		queryParams   []*paramData
		headers       []*paramData
		signer        string
		clientsTmpl   = template.Must(template.New("clients").Funcs(funcs).Parse(clientsTmpl))
		requestsTmpl  = template.Must(template.New("requests").Funcs(funcs).Parse(requestsTmpl))
		clientsWSTmpl = template.Must(template.New("clientsws").Funcs(funcs).Parse(clientsWSTmpl))
	)
	if action.Payload != nil {
		params = append(params, "payload "+codegen.GoTypeRef(action.Payload, action.Payload.AllRequired(), 1, false))
		names = append(names, "payload")
	}

	initParamsScoped := func(att *design.AttributeDefinition) []*paramData {
		reqData, optData := initParams(att)

		sort.Sort(byParamName(reqData))
		sort.Sort(byParamName(optData))

		// Update closure
		for _, p := range reqData {
			names = append(names, p.VarName)
			params = append(params, p.VarName+" "+cmdFieldType(p.Attribute.Type, false))
		}
		for _, p := range optData {
			names = append(names, p.VarName)
			params = append(params, p.VarName+" "+cmdFieldType(p.Attribute.Type, p.Attribute.Type.IsPrimitive()))
		}
		return append(reqData, optData...)
	}
	queryParams = initParamsScoped(action.QueryParams)
	headers = initParamsScoped(action.Headers)

	if action.Security != nil {
		signer = codegen.Goify(action.Security.Scheme.SchemeName, true)
	}
	data := struct {
		Name               string
		ResourceName       string
		Description        string
		Routes             []*design.RouteDefinition
		HasPayload         bool
		HasMultiContent    bool
		DefaultContentType string
		Params             string
		ParamNames         string
		CanonicalScheme    string
		Signer             string
		QueryParams        []*paramData
		Headers            []*paramData
	}{
		Name:               action.Name,
		ResourceName:       action.Parent.Name,
		Description:        action.Description,
		Routes:             action.Routes,
		HasPayload:         action.Payload != nil,
		HasMultiContent:    len(design.Design.Consumes) > 1,
		DefaultContentType: design.Design.Consumes[0].MIMETypes[0],
		Params:             strings.Join(params, ", "),
		ParamNames:         strings.Join(names, ", "),
		CanonicalScheme:    action.CanonicalScheme(),
		Signer:             signer,
		QueryParams:        queryParams,
		Headers:            headers,
	}
	if action.WebSocket() {
		return clientsWSTmpl.Execute(file, data)
	}
	if err := clientsTmpl.Execute(file, data); err != nil {
		return err
	}
	return requestsTmpl.Execute(file, data)
}

// fileServerMethod returns the name of the client method for downloading assets served by the given
// file server.
// Note: the implementation opts for generating good names rather than names that are guaranteed to
// be unique. This means that the generated code could be potentially incorrect in the rare cases
// where it produces the same names for two different file servers. This should be addressed later
// (when it comes up?) using metadata to let users override the default.
func (g *Generator) fileServerMethod(fs *design.FileServerDefinition) string {
	var (
		suffix string

		wcs      = design.ExtractWildcards(fs.RequestPath)
		reqElems = strings.Split(fs.RequestPath, "/")
	)

	if len(wcs) == 0 {
		suffix = path.Base(fs.RequestPath)
		ext := filepath.Ext(suffix)
		suffix = strings.TrimSuffix(suffix, ext)
		suffix += codegen.Goify(ext, true)
	} else {
		if len(reqElems) == 1 {
			suffix = filepath.Base(fs.RequestPath)
			suffix = suffix[1:] // remove "*" prefix
		} else {
			suffix = reqElems[len(reqElems)-2] // should work most of the time
		}
	}
	return "Download" + codegen.Goify(suffix, true)
}

// generateMediaTypes iterates through the media types and generate the data structures and
// marshaling code.
func (g *Generator) generateMediaTypes(pkgDir string, funcs template.FuncMap) error {
	funcs["decodegotyperef"] = decodeGoTypeRef
	funcs["decodegotypename"] = decodeGoTypeName
	typeDecodeTmpl := template.Must(template.New("typeDecode").Funcs(funcs).Parse(typeDecodeTmpl))
	mtFile := filepath.Join(pkgDir, "media_types.go")
	mtWr, err := genapp.NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Media Types", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}
	for _, v := range g.API.MediaTypes {
		imports = codegen.AttributeImports(v.AttributeDefinition, imports, nil)
	}
	mtWr.WriteHeader(title, g.Target, imports)
	err = g.API.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if (mt.Type.IsObject() || mt.Type.IsArray()) && !mt.IsError() {
			if err := mtWr.Execute(mt); err != nil {
				return err
			}
		}
		err := mt.IterateViews(func(view *design.ViewDefinition) error {
			p, _, err := mt.Project(view.Name)
			if err != nil {
				return err
			}
			if err := typeDecodeTmpl.Execute(mtWr.SourceFile, p); err != nil {
				return err
			}
			return nil
		})
		return err
	})
	g.genfiles = append(g.genfiles, mtFile)
	if err != nil {
		return err
	}
	return mtWr.FormatCode()
}

// generateUserTypes iterates through the user types and generates the data structures and
// marshaling code.
func (g *Generator) generateUserTypes(pkgDir string) error {
	utFile := filepath.Join(pkgDir, "user_types.go")
	utWr, err := genapp.NewUserTypesWriter(utFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application User Types", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.NewImport("uuid", "github.com/goadesign/goa/uuid"),
	}
	for _, v := range g.API.Types {
		imports = codegen.AttributeImports(v.AttributeDefinition, imports, nil)
	}
	utWr.WriteHeader(title, g.Target, imports)
	err = g.API.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		return utWr.Execute(t)
	})
	g.genfiles = append(g.genfiles, utFile)
	if err != nil {
		return err
	}
	return utWr.FormatCode()
}

// join is a code generation helper function that generates a function signature built from
// concatenating the properties (name type) of the given attribute type (assuming it's an object).
// join accepts an optional slice of strings which indicates the order in which the parameters
// should appear in the signature. If pos is specified then it must list all the parameters. If
// it's not specified then parameters are sorted alphabetically.
func join(att *design.AttributeDefinition, usePointers bool, pos ...[]string) string {
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
		elems[i] = fmt.Sprintf("%s %s", codegen.Goify(n, false),
			cmdFieldType(a.Type, usePointers && !a.IsRequired(n)))
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

// decodeGoTypeRef handles the case where the type being decoded is a error response media type.
func decodeGoTypeRef(t design.DataType, required []string, tabs int, private bool) string {
	mt, ok := t.(*design.MediaTypeDefinition)
	if ok && mt.IsError() {
		return "*goa.ErrorResponse"
	}
	return codegen.GoTypeRef(t, required, tabs, private)
}

// decodeGoTypeName handles the case where the type being decoded is a error response media type.
func decodeGoTypeName(t design.DataType, required []string, tabs int, private bool) string {
	mt, ok := t.(*design.MediaTypeDefinition)
	if ok && mt.IsError() {
		return "goa.ErrorResponse"
	}
	return codegen.GoTypeName(t, required, tabs, private)
}

// cmdFieldType computes the Go type name used to store command flags of the given design type.
func cmdFieldType(t design.DataType, point bool) string {
	var pointer, suffix string
	if point && !t.IsArray() {
		pointer = "*"
	}
	suffix = codegen.GoNativeType(t)
	return pointer + suffix
}

// cmdFieldTypeString computes the Go type name used to store command flags of the given design type. Complex types are String
func cmdFieldTypeString(t design.DataType, point bool) string {
	var pointer, suffix string
	if point && !t.IsArray() {
		pointer = "*"
	}
	if t.Kind() == design.UUIDKind || t.Kind() == design.DateTimeKind || t.Kind() == design.AnyKind || t.Kind() == design.NumberKind || t.Kind() == design.BooleanKind {
		suffix = "string"
	} else if isArrayOfType(t, design.UUIDKind, design.DateTimeKind, design.AnyKind, design.NumberKind, design.BooleanKind) {
		suffix = "[]string"
	} else {
		suffix = codegen.GoNativeType(t)
	}
	return pointer + suffix
}

func isArrayOfType(array design.DataType, kinds ...design.Kind) bool {
	if !array.IsArray() {
		return false
	}
	kind := array.ToArray().ElemType.Type.Kind()
	for _, t := range kinds {
		if t == kind {
			return true
		}
	}
	return false
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
		case design.StringKind:
			return fmt.Sprintf("%s := %s", target, name)
		case design.DateTimeKind:
			return fmt.Sprintf("%s := %s.Format(time.RFC3339)", target, strings.Replace(name, "*", "", -1)) // remove pointer if present
		case design.UUIDKind:
			return fmt.Sprintf("%s := %s.String()", target, strings.Replace(name, "*", "", -1)) // remove pointer if present
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

// pathTemplate returns a fmt format suitable to build a request path to the route.
func pathTemplate(r *design.RouteDefinition) string {
	return design.WildcardRegex.ReplaceAllLiteralString(r.FullPath(), "/%s")
}

// pathParams return the function signature of the path factory function for the given route.
func pathParams(r *design.RouteDefinition) string {
	pnames := r.Params()
	params := make(design.Object, len(pnames))
	for _, p := range pnames {
		params[p] = r.Parent.Params.Type.ToObject()[p]
	}
	return join(&design.AttributeDefinition{Type: params}, false, pnames)
}

// typeName returns Go type name of given MediaType definition.
func typeName(mt *design.MediaTypeDefinition) string {
	if mt.IsError() {
		return "ErrorResponse"
	}
	return codegen.GoTypeName(mt, mt.AllRequired(), 1, false)
}

// initParams returns required and optional paramData extracted from given attribute definition.
func initParams(att *design.AttributeDefinition) ([]*paramData, []*paramData) {
	if att == nil {
		return nil, nil
	}
	obj := att.Type.ToObject()
	var reqParamData []*paramData
	var optParamData []*paramData
	for n, q := range obj {
		varName := codegen.Goify(n, false)
		param := &paramData{
			Name:      n,
			VarName:   varName,
			Attribute: q,
		}
		if q.Type.IsPrimitive() {
			param.MustToString = q.Type.Kind() != design.StringKind
			if att.IsRequired(n) {
				param.ValueName = varName
				reqParamData = append(reqParamData, param)
			} else {
				param.ValueName = "*" + varName
				param.CheckNil = true
				optParamData = append(optParamData, param)
			}
		} else {
			if q.Type.IsArray() {
				param.IsArray = true
				param.ElemAttribute = q.Type.ToArray().ElemType
			}
			param.MustToString = true
			param.ValueName = varName
			param.CheckNil = true
			if att.IsRequired(n) {
				reqParamData = append(reqParamData, param)
			} else {
				optParamData = append(optParamData, param)
			}
		}
	}

	return reqParamData, optParamData
}

// paramData is the data structure holding the information needed to generate query params and
// headers handling code.
type paramData struct {
	Name          string
	VarName       string
	ValueName     string
	Attribute     *design.AttributeDefinition
	ElemAttribute *design.AttributeDefinition
	MustToString  bool
	IsArray       bool
	CheckNil      bool
}

type byParamName []*paramData

func (b byParamName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byParamName) Less(i, j int) bool { return b[i].Name < b[j].Name }
func (b byParamName) Len() int           { return len(b) }

const (
	arrayToStringT = `	{{ $tmp := tempvar }}{{ $tmp }} := make([]string, len({{ .Name }}))
	for i, e := range {{ .Name }} {
		{{ $tmp2 := tempvar }}{{ toString "e" $tmp2 .ElemType }}
		{{ $tmp }}[i] = {{ $tmp2 }}
	}
	{{ .Target }} := strings.Join({{ $tmp }}, ",")`

	payloadTmpl = `// {{ gotypename .Payload nil 0 false }} is the {{ .Parent.Name }} {{ .Name }} action payload.
type {{ gotypename .Payload nil 1 false }} {{ gotypedef .Payload 0 true false }}
`

	typeDecodeTmpl = `{{ $typeName := typeName . }}{{ $funcName := printf "Decode%s" $typeName }}// {{ $funcName }} decodes the {{ $typeName }} instance encoded in resp body.
func (c *Client) {{ $funcName }}(resp *http.Response) ({{ decodegotyperef . .AllRequired 0 false }}, error) {
	var decoded {{ decodegotypename . .AllRequired 0 false }}
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return {{ if .IsObject }}&{{ end }}decoded, err
}
`

	pathTmpl = `{{ $funcName := printf "%sPath%s" (goify (printf "%s%s" .Route.Parent.Name (title .Route.Parent.Parent.Name)) true) ((or (and .Index (add .Index 1)) "") | printf "%v") }}{{/*
*/}}// {{ $funcName }} computes a request path to the {{ .Route.Parent.Name }} action of {{ .Route.Parent.Parent.Name }}.
func {{ $funcName }}({{ pathParams .Route }}) string {
	{{ range $i, $param := .Params }}{{/*
*/}}{{ toString $param.VarName (printf "param%d" $i) $param.Attribute }}
	{{ end }}
	return fmt.Sprintf({{ printf "%q" (pathTemplate .Route) }}{{ range $i, $param := .Params }}, {{ printf "param%d" $i }}{{ end }})
}
`

	clientsTmpl = `{{ $funcName := goify (printf "%s%s" .Name (title .ResourceName)) true }}{{ $desc := .Description }}{{/*
*/}}{{ if $desc }}{{ multiComment $desc }}{{ else }}{{/*
*/}}// {{ $funcName }} makes a request to the {{ .Name }} action endpoint of the {{ .ResourceName }} resource{{ end }}
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Params }}, {{ .Params }}{{ end }}{{ if and .HasPayload .HasMultiContent }}, contentType string{{ end }}) (*http.Response, error) {
	req, err := c.New{{ $funcName }}Request(ctx, path{{ if .ParamNames }}, {{ .ParamNames }}{{ end }}{{ if and .HasPayload .HasMultiContent }}, contentType{{ end }})
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}
`

	clientsWSTmpl = `{{ $funcName := goify (printf "%s%s" .Name (title .ResourceName)) true }}{{ $desc := .Description }}{{/*
*/}}{{ if $desc }}{{ multiComment $desc }}{{ else }}// {{ $funcName }} establishes a websocket connection to the {{ .Name }} action endpoint of the {{ .ResourceName }} resource{{ end }}
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Params }}, {{ .Params }}{{ end }}) (*websocket.Conn, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "{{ .CanonicalScheme }}"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
{{ if .QueryParams }}	values := u.Query()
{{ range .QueryParams }}{{ if .CheckNil }}	if {{ .VarName }} != nil {
	{{ end }}{{/*

// ARRAY
*/}}{{ if .IsArray }}		for _, p := range {{ .VarName }} {
{{ if .MustToString }}{{ $tmp := tempvar }}			{{ toString "p" $tmp .ElemAttribute }}
			values.Add("{{ .Name }}", {{ $tmp }})
{{ else }}			values.Add("{{ .Name }}", {{ .ValueName }})
{{ end }}}{{/*

// NON STRING
*/}}{{ else if .MustToString }}{{ $tmp := tempvar }}	{{ toString .ValueName $tmp .Attribute }}
	values.Set("{{ .Name }}", {{ $tmp }}){{/*

// STRING
*/}}{{ else }}	values.Set("{{ .Name }}", {{ .ValueName }})
{{ end }}{{ if .CheckNil }}	}
{{ end }}{{ end }}	u.RawQuery = values.Encode()
{{ end }}	url_ := u.String()
	cfg, err := websocket.NewConfig(url_, url_)
	if err != nil {
		return nil, err
	}
{{ range $header := .Headers }}{{ $tmp := tempvar }}	{{ toString $header.VarName $tmp $header.Attribute }}
	cfg.Header["{{ $header.Name }}"] = []string{ {{ $tmp }} }
{{ end }}	return websocket.DialConfig(cfg)
}
`

	fsTmpl = `// {{ .Name }} downloads {{ if .DirName }}{{ .DirName }}files with the given filename{{ else }}{{ .FileName }}{{ end }} and writes it to the file dest.
// It returns the number of bytes downloaded in case of success.
func (c * Client) {{ .Name }}(ctx context.Context, {{ if .DirName }}filename, {{ end }}dest string) (int64, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "{{ .CanonicalScheme }}"
	}
{{ if .DirName }}	p := path.Join("{{ .RequestDir }}", filename)
{{ end }}	u := url.URL{Host: c.Host, Scheme: scheme, Path: {{ if .DirName }}p{{ else }}"{{ .RequestPath }}"{{ end }}}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		var body string
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			if len(b) > 0 {
				body = ": "+ string(b)
			}
		}
		return 0, fmt.Errorf("%s%s", resp.Status, body)
	}
	defer resp.Body.Close()
	out, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	return io.Copy(out, resp.Body)
}
`

	requestsTmpl = `{{ $funcName := goify (printf "New%s%sRequest" (title .Name) (title .ResourceName)) true }}{{/*
*/}}// {{ $funcName }} create the request corresponding to the {{ .Name }} action endpoint of the {{ .ResourceName }} resource.
func (c *Client) {{ $funcName }}(ctx context.Context, path string{{ if .Params }}, {{ .Params }}{{ end }}{{ if .HasPayload }}{{ if .HasMultiContent }}, contentType string{{ end }}{{ end }}) (*http.Request, error) {
{{ if .HasPayload }}	var body bytes.Buffer
{{ if .HasMultiContent }}	if contentType == "" {
		contentType = "*/*" // Use default encoder
	}
{{ end }}	err := c.Encoder.Encode(payload, &body, {{ if .HasMultiContent }}contentType{{ else }}"*/*"{{ end }})
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
{{ end }}	scheme := c.Scheme
	if scheme == "" {
		scheme = "{{ .CanonicalScheme }}"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
{{ if .QueryParams }}	values := u.Query()
{{ range .QueryParams }}{{/*

// ARRAY
*/}}{{ if .IsArray }}		for _, p := range {{ .VarName }} {
{{ if .MustToString }}{{ $tmp := tempvar }}			{{ toString "p" $tmp .ElemAttribute }}
			values.Add("{{ .Name }}", {{ $tmp }})
{{ else }}			values.Add("{{ .Name }}", {{ .ValueName }})
{{ end }}	 }
{{/*

// NON STRING
*/}}{{ else if .MustToString }}{{ if .CheckNil }}	if {{ .VarName }} != nil {
	{{ end }}{{ $tmp := tempvar }}	{{ toString .ValueName $tmp .Attribute }}
	values.Set("{{ .Name }}", {{ $tmp }})
{{ if .CheckNil }}	}
{{ end }}{{/*

// STRING
*/}}{{ else }}{{ if .CheckNil }}	if {{ .VarName }} != nil {
	{{ end }}	values.Set("{{ .Name }}", {{ .ValueName }})
{{ if .CheckNil }}	}
{{ end }}{{ end }}{{ end }}	u.RawQuery = values.Encode()
{{ end }}{{ if .HasPayload }}	req, err := http.NewRequest({{ $route := index .Routes 0 }}"{{ $route.Verb }}", u.String(), &body)
{{ else }}	req, err := http.NewRequest({{ $route := index .Routes 0 }}"{{ $route.Verb }}", u.String(), nil)
{{ end }}	if err != nil {
		return nil, err
	}
{{ if or .HasPayload .Headers }}	header := req.Header
{{ if .HasPayload }}{{ if .HasMultiContent }}	if contentType == "*/*" {
		header.Set("Content-Type", "{{ .DefaultContentType }}")
	} else {
		header.Set("Content-Type", contentType)
	}
{{ else }}	header.Set("Content-Type", "{{ .DefaultContentType }}")
{{ end }}{{ end }}{{ range .Headers }}{{ if .CheckNil }}	if {{ .VarName }} != nil {
{{ end }}{{ if .MustToString }}{{ $tmp := tempvar }}	{{ toString .ValueName $tmp .Attribute }}
	header.Set("{{ .Name }}", {{ $tmp }}){{ else }}
	header.Set("{{ .Name }}", {{ .ValueName }})
{{ end }}{{ if .CheckNil }}	}{{ end }}
{{ end }}{{ end }}{{ if .Signer }}	if c.{{ .Signer }}Signer != nil {
		c.{{ .Signer }}Signer.Sign(req)
	}
{{ end }}	return req, nil
}
`

	clientTmpl = `// Client is the {{ .API.Name }} service client.
type Client struct {
	*goaclient.Client{{range $security := .API.SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}
	{{ goify $security.SchemeName true }}Signer goaclient.Signer{{ end }}{{ end }}
	Encoder *goa.HTTPEncoder
	Decoder *goa.HTTPDecoder
}

// New instantiates the client.
func New(c goaclient.Doer) *Client {
	client := &Client{
		Client: goaclient.New(c),
		Encoder: goa.NewHTTPEncoder(),
		Decoder: goa.NewHTTPDecoder(),
	}

{{ if .Encoders }}	// Setup encoders and decoders
{{ range .Encoders }}{{/*
*/}}	client.Encoder.Register({{ .PackageName }}.{{ .Function }}, "{{ joinStrings .MIMETypes "\", \"" }}")
{{ end }}{{ range .Decoders }}{{/*
*/}}	client.Decoder.Register({{ .PackageName }}.{{ .Function }}, "{{ joinStrings .MIMETypes "\", \"" }}")
{{ end }}

	// Setup default encoder and decoder
{{ range .Encoders }}{{ if .Default }}{{/*
*/}}	client.Encoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}{{ range .Decoders }}{{ if .Default }}{{/*
*/}}	client.Decoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}
{{ end }}	return client
}

{{range $security := .API.SecuritySchemes }}{{ $signer := signerType $security }}{{ if $signer }}{{/*
*/}}{{ $name := printf "%sSigner" (goify $security.SchemeName true) }}{{/*
*/}}// Set{{ $name }} sets the request signer for the {{ $security.SchemeName }} security scheme.
func (c *Client) Set{{ $name }}(signer goaclient.Signer) {
	c.{{ $name }} = signer
}
{{ end }}{{ end }}
`
)
