package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Generator is the application code generator.
type Generator struct {
	*codegen.GoGenerator
	genfiles []string
}

// Generate is the generator entry point called by the meta generator.
func Generate(api *design.APIDefinition) ([]string, error) {
	g, err := NewGenerator()
	if err != nil {
		return nil, err
	}
	return g.Generate(api)
}

// NewGenerator returns the application code generator.
func NewGenerator() (*Generator, error) {
	app := kingpin.New("Code generator", "application code generator")
	codegen.RegisterFlags(app)
	NewCommand().RegisterFlags(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf(`invalid command line: %s. Command line was "%s"`,
			err, strings.Join(os.Args, " "))
	}
	outdir := AppOutputDir()
	os.RemoveAll(outdir)
	if err = os.MkdirAll(outdir, 0777); err != nil {
		return nil, err
	}
	return &Generator{
		GoGenerator: codegen.NewGoGenerator(outdir),
		genfiles:    []string{outdir},
	}, nil
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
		return nil, fmt.Errorf("missing API definition, make sure design.Design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	outdir := AppOutputDir()
	err = api.IterateVersions(func(v *design.APIVersionDefinition) error {
		verdir := outdir
		if v.Version != "" {
			verdir = filepath.Join(verdir, codegen.VersionPackage(v.Version))
		}
		if err := os.MkdirAll(verdir, 0755); err != nil {
			return err
		}
		if err := g.generateContexts(verdir, api, v); err != nil {
			return err
		}
		if err := g.generateControllers(verdir, v); err != nil {
			return err
		}
		if err := g.generateHrefs(verdir, v); err != nil {
			return err
		}
		if err := g.generateMediaTypes(verdir, v); err != nil {
			return err
		}
		if err := g.generateUserTypes(verdir, v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
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

// Generated package name for resources supporting the given version.
func packageName(version *design.APIVersionDefinition) (pack string) {
	pack = TargetPackage
	if version.Version != "" {
		pack = codegen.Goify(codegen.VersionPackage(version.Version), false)
	}
	return
}

// generateContexts iterates through the version resources and actions and generates the action
// contexts.
func (g *Generator) generateContexts(verdir string, api *design.APIDefinition, version *design.APIVersionDefinition) error {
	ctxFile := filepath.Join(verdir, "contexts.go")
	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Contexts", version.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	if !version.IsDefault() {
		appPkg, err := AppPackagePath()
		if err != nil {
			return err
		}
		imports = append(imports, codegen.SimpleImport(appPkg))
	}
	ctxWr.WriteHeader(title, packageName(version), imports)
	err = version.IterateResources(func(r *design.ResourceDefinition) error {
		if !r.SupportsVersion(version.Version) {
			return nil
		}
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := codegen.Goify(a.Name, true) + codegen.Goify(a.Parent.Name, true) + "Context"
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Payload:      a.Payload,
				Params:       a.AllParams(),
				Headers:      r.Headers.Merge(a.Headers),
				Routes:       a.Routes,
				Responses:    MergeResponses(r.Responses, a.Responses),
				API:          api,
				Versioned:    version.Version != "",
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

// generateControllers iterates through the version resources and generates the low level
// controllers.
func (g *Generator) generateControllers(verdir string, version *design.APIVersionDefinition) error {
	ctlFile := filepath.Join(verdir, "controllers.go")
	ctlWr, err := NewControllersWriter(ctlFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Controllers", version.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/julienschmidt/httprouter"),
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	if !version.IsDefault() {
		appPkg, err := AppPackagePath()
		if err != nil {
			return err
		}
		imports = append(imports, codegen.SimpleImport(appPkg))
	}
	ctlWr.WriteHeader(title, packageName(version), imports)
	var controllersData []*ControllerTemplateData
	version.IterateResources(func(r *design.ResourceDefinition) error {
		if !r.SupportsVersion(version.Version) {
			return nil
		}
		data := &ControllerTemplateData{Resource: codegen.Goify(r.Name, true)}
		err := r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":    codegen.Goify(a.Name, true),
				"Routes":  a.Routes,
				"Context": context,
			}
			data.Actions = append(data.Actions, action)
			return nil
		})
		if err != nil {
			return err
		}
		if len(data.Actions) > 0 {
			data.Version = version.Version
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

// generateHrefs iterates through the version resources and generates the href factory methods.
func (g *Generator) generateHrefs(verdir string, version *design.APIVersionDefinition) error {
	hrefFile := filepath.Join(verdir, "hrefs.go")
	resWr, err := NewResourcesWriter(hrefFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Resource Href Factories", version.Context())
	resWr.WriteHeader(title, packageName(version), nil)
	err = version.IterateResources(func(r *design.ResourceDefinition) error {
		if !r.SupportsVersion(version.Version) {
			return nil
		}
		m := design.Design.MediaTypeWithIdentifier(r.MediaType)
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
func (g *Generator) generateMediaTypes(verdir string, version *design.APIVersionDefinition) error {
	mtFile := filepath.Join(verdir, "media_types.go")
	mtWr, err := NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Media Types", version.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/raphael/goa"),
		codegen.SimpleImport("fmt"),
	}
	mtWr.WriteHeader(title, packageName(version), imports)
	err = version.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		data := &MediaTypeTemplateData{
			MediaType:  mt,
			Versioned:  version.Version != "",
			DefaultPkg: TargetPackage,
		}
		if mt.Type.IsObject() || mt.Type.IsArray() {
			return mtWr.Execute(data)
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
func (g *Generator) generateUserTypes(verdir string, version *design.APIVersionDefinition) error {
	utFile := filepath.Join(verdir, "user_types.go")
	utWr, err := NewUserTypesWriter(utFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application User Types", version.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/raphael/goa"),
		codegen.SimpleImport("fmt"),
	}
	utWr.WriteHeader(title, packageName(version), imports)
	err = version.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		data := &UserTypeTemplateData{
			UserType:   t,
			Versioned:  version.Version != "",
			DefaultPkg: TargetPackage,
		}
		return utWr.Execute(data)
	})
	g.genfiles = append(g.genfiles, utFile)
	if err != nil {
		return err
	}
	return utWr.FormatCode()
}
