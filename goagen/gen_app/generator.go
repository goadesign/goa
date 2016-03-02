package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
)

// Generator is the application code generator.
type Generator struct {
	genfiles []string
}

// Generate is the generator entry point called by the meta generator.
func Generate(roots []dslengine.Root) (files []string, err error) {
	api := design.Design
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	root := &cobra.Command{
		Use:   "goagen",
		Short: "Code generator",
		Long:  "application code generator",
		PreRunE: func(*cobra.Command, []string) error {
			outdir := AppOutputDir()
			os.RemoveAll(outdir)
			g.genfiles = []string{outdir}
			err = os.MkdirAll(outdir, 0777)
			return err
		},
		Run: func(*cobra.Command, []string) { files, err = g.Generate(api) },
	}
	codegen.RegisterFlags(root)
	NewCommand().RegisterFlags(root)
	root.Execute()
	return
}

// AppOutputDir returns the directory containing the generated files.
func AppOutputDir() string {
	return filepath.Join(codegen.OutputDir, TargetPackage)
}

// AppPackagePath returns the Go package path to the generated package.
func AppPackagePath() (string, error) {
	outputDir := AppOutputDir()
	gopaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, gopath := range gopaths {
		if strings.HasPrefix(outputDir, gopath) {
			path, err := filepath.Rel(filepath.Join(gopath, "src"), outputDir)
			if err != nil {
				return "", err
			}
			return filepath.ToSlash(path), nil
		}
	}
	return "", fmt.Errorf("output directory outside of Go workspace, make sure to define GOPATH correctly or change output directory")
}

// Generate the application code, implement codegen.Generator.
func (g *Generator) Generate(api *design.APIDefinition) (_ []string, err error) {
	if api == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	os.RemoveAll(AppOutputDir())

	if err := os.MkdirAll(AppOutputDir(), 0755); err != nil {
		return nil, err
	}
	if err := g.generateContexts(api); err != nil {
		return nil, err
	}
	if err := g.generateControllers(api); err != nil {
		return nil, err
	}
	if err := g.generateHrefs(api); err != nil {
		return nil, err
	}
	if err := g.generateMediaTypes(api); err != nil {
		return nil, err
	}
	if err := g.generateUserTypes(api); err != nil {
		return nil, err
	}

	return g.genfiles, nil
}

// Cleanup removes the entire "app" directory if it was created by this generator.
func (g *Generator) Cleanup() {
	if len(g.genfiles) == 0 {
		return
	}
	os.RemoveAll(AppOutputDir())
	g.genfiles = nil
}

// MergeResponses merge the response maps overriding the first argument map entries with the
// second argument map entries in case of collision.
func MergeResponses(l, r map[string]*design.ResponseDefinition) map[string]*design.ResponseDefinition {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	for n, r := range r {
		l[n] = r
	}
	return l
}

// generateContexts iterates through the API resources and actions and generates the action
// contexts.
func (g *Generator) generateContexts(api *design.APIDefinition) error {
	ctxFile := filepath.Join(AppOutputDir(), "contexts.go")
	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Contexts", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	ctxWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := codegen.Goify(a.Name, true) + codegen.Goify(a.Parent.Name, true) + "Context"
			headers := r.Headers.Merge(a.Headers)
			if headers != nil && len(headers.Type.ToObject()) == 0 {
				headers = nil // So that {{if .Headers}} returns false in templates
			}
			params := a.AllParams()
			if params != nil && len(params.Type.ToObject()) == 0 {
				params = nil // So that {{if .Params}} returns false in templates
			}
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Payload:      a.Payload,
				Params:       params,
				Headers:      headers,
				Routes:       a.Routes,
				Responses:    MergeResponses(r.Responses, a.Responses),
				API:          api,
				DefaultPkg:   TargetPackage,
			}
			return ctxWr.Execute(&ctxData)
		})
	})
	g.genfiles = append(g.genfiles, ctxFile)
	if err != nil {
		return err
	}
	return ctxWr.FormatCode()
}

// BuildEncoderMap builds the template data needed to render the given encoding definitions.
// This extra map is needed to handle the case where a single encoding definition maps to multiple
// encoding packages. The data is indexed by encoder Go package path.
func BuildEncoderMap(info []*design.EncodingDefinition, encoder bool) (map[string]*EncoderTemplateData, error) {
	if len(info) == 0 {
		return nil, nil
	}
	packages := make(map[string]map[string]bool)
	for _, enc := range info {
		supporting := enc.SupportingPackages()
		if supporting == nil {
			// shouldn't happen - DSL validation shouldn't allow it - be graceful
			continue
		}
		for ppath, mimeTypes := range supporting {
			if _, ok := packages[ppath]; !ok {
				packages[ppath] = make(map[string]bool)
			}
			for _, m := range mimeTypes {
				packages[ppath][m] = true
			}
		}
	}
	data := make(map[string]*EncoderTemplateData, len(packages))
	if len(info[0].MIMETypes) == 0 {
		return nil, fmt.Errorf("No mime type associated with encoding info for package %s", info[0].PackagePath)
	}
	defaultMediaType := info[0].MIMETypes[0]
	for p, ms := range packages {
		pkgName := "goa"
		if !design.IsGoaEncoder(p) {
			srcPath, err := codegen.PackageSourcePath(p)
			if err == nil {
				pkgName, err = codegen.PackageName(srcPath)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to load package %s (%s)", p, err)
			}
		}
		mimeTypes := make([]string, len(ms))
		isDefault := false
		i := 0
		for m := range ms {
			if m == defaultMediaType {
				isDefault = true
			}
			mimeTypes[i] = m
			i++
		}
		first := mimeTypes[0]
		sort.Strings(mimeTypes)
		var factory string
		if encoder {
			if !design.IsGoaEncoder(p) {
				factory = "EncoderFactory"
			} else {
				factory = design.KnownEncoders[first][1]
			}
		} else {
			if !design.IsGoaEncoder(p) {
				factory = "DecoderFactory"
			} else {
				factory = design.KnownEncoders[first][2]
			}
		}
		d := &EncoderTemplateData{
			PackagePath: p,
			PackageName: pkgName,
			Factory:     factory,
			MIMETypes:   mimeTypes,
			Default:     isDefault,
		}
		data[p] = d
	}
	return data, nil
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateControllers(api *design.APIDefinition) error {
	ctlFile := filepath.Join(AppOutputDir(), "controllers.go")
	ctlWr, err := NewControllersWriter(ctlFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Controllers", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	encoderMap, err := BuildEncoderMap(api.Produces, true)
	if err != nil {
		return err
	}
	decoderMap, err := BuildEncoderMap(api.Consumes, false)
	if err != nil {
		return err
	}
	encoderImports := make(map[string]bool)
	for _, data := range encoderMap {
		encoderImports[data.PackagePath] = true
	}
	for _, data := range decoderMap {
		encoderImports[data.PackagePath] = true
	}
	for packagePath := range encoderImports {
		if !design.IsGoaEncoder(packagePath) {
			imports = append(imports, codegen.SimpleImport(packagePath))
		}
	}
	ctlWr.WriteHeader(title, TargetPackage, imports)
	ctlWr.WriteInitService(encoderMap, decoderMap)
	var controllersData []*ControllerTemplateData
	api.IterateResources(func(r *design.ResourceDefinition) error {
		data := &ControllerTemplateData{API: api, Resource: codegen.Goify(r.Name, true)}
		err := r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			unmarshal := fmt.Sprintf("unmarshal%s%sPayload", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":      codegen.Goify(a.Name, true),
				"Routes":    a.Routes,
				"Context":   context,
				"Unmarshal": unmarshal,
				"Payload":   a.Payload,
			}
			data.Actions = append(data.Actions, action)
			return nil
		})
		if err != nil {
			return err
		}
		if len(data.Actions) > 0 {
			data.EncoderMap = encoderMap
			data.DecoderMap = decoderMap
			controllersData = append(controllersData, data)
		}
		return nil
	})
	g.genfiles = append(g.genfiles, ctlFile)
	if err = ctlWr.Execute(controllersData); err != nil {
		return err
	}
	return ctlWr.FormatCode()
}

// generateHrefs iterates through the API resources and generates the href factory methods.
func (g *Generator) generateHrefs(api *design.APIDefinition) error {
	hrefFile := filepath.Join(AppOutputDir(), "hrefs.go")
	resWr, err := NewResourcesWriter(hrefFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Resource Href Factories", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
	}
	resWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		m := api.MediaTypeWithIdentifier(r.MediaType)
		var identifier string
		if m != nil {
			identifier = m.Identifier
		} else {
			identifier = "plain/text"
		}
		canoTemplate := r.URITemplate()
		canoTemplate = design.WildcardRegex.ReplaceAllLiteralString(canoTemplate, "/%v")
		var canoParams []string
		if ca := r.CanonicalAction(); ca != nil {
			if len(ca.Routes) > 0 {
				canoParams = ca.Routes[0].Params()
			}
		}
		data := ResourceData{
			Name:              codegen.Goify(r.Name, true),
			Identifier:        identifier,
			Description:       r.Description,
			Type:              m,
			CanonicalTemplate: canoTemplate,
			CanonicalParams:   canoParams,
		}
		return resWr.Execute(&data)
	})
	g.genfiles = append(g.genfiles, hrefFile)
	if err != nil {
		return err
	}
	return resWr.FormatCode()
}

// generateMediaTypes iterates through the media types and generate the data structures and
// marshaling code.
func (g *Generator) generateMediaTypes(api *design.APIDefinition) error {
	mtFile := filepath.Join(AppOutputDir(), "media_types.go")
	mtWr, err := NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Media Types", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
	}
	mtWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.Type.IsObject() || mt.Type.IsArray() {
			return mtWr.Execute(mt)
		}
		return nil
	})
	g.genfiles = append(g.genfiles, mtFile)
	if err != nil {
		return err
	}
	return mtWr.FormatCode()
}

// generateUserTypes iterates through the user types and generates the data structures and
// marshaling code.
func (g *Generator) generateUserTypes(api *design.APIDefinition) error {
	utFile := filepath.Join(AppOutputDir(), "user_types.go")
	utWr, err := NewUserTypesWriter(utFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application User Types", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
	}
	utWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		return utWr.Execute(t)
	})
	g.genfiles = append(g.genfiles, utFile)
	if err != nil {
		return err
	}
	return utWr.FormatCode()
}
