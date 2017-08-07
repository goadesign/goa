package genapp

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

//NewGenerator returns an initialized instance of an Application Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}
	g.validator = codegen.NewValidator()

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API       *design.APIDefinition // The API definition
	OutDir    string                // Path to output directory
	Target    string                // Name of generated package
	NoTest    bool                  // Whether to skip test generation
	genfiles  []string              // Generated files
	validator *codegen.Validator    // Validation code generator
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, toolDir, target, ver string
		notest, regen                bool
	)

	set := flag.NewFlagSet("app", flag.PanicOnError)
	set.String("design", "", "")
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&target, "pkg", "app", "")
	set.StringVar(&ver, "version", "", "")
	set.StringVar(&toolDir, "tooldir", "tool", "")
	set.BoolVar(&notest, "notest", false, "")
	set.BoolVar(&regen, "regen", false, "")
	set.Bool("force", false, "")
	set.Parse(os.Args[1:])
	outDir = filepath.Join(outDir, target)

	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	target = codegen.Goify(target, false)
	g := &Generator{OutDir: outDir, Target: target, NoTest: notest, API: design.Design, validator: codegen.NewValidator()}

	return g.Generate()
}

// Generate the application code, implement codegen.Generator.
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

	codegen.Reserved[g.Target] = true

	os.RemoveAll(g.OutDir)

	if err := os.MkdirAll(g.OutDir, 0755); err != nil {
		return nil, err
	}
	g.genfiles = []string{g.OutDir}
	if err := g.generateContexts(); err != nil {
		return nil, err
	}
	if err := g.generateControllers(); err != nil {
		return nil, err
	}
	if err := g.generateSecurity(); err != nil {
		return nil, err
	}
	if err := g.generateHrefs(); err != nil {
		return nil, err
	}
	if err := g.generateMediaTypes(); err != nil {
		return nil, err
	}
	if err := g.generateUserTypes(); err != nil {
		return nil, err
	}
	if !g.NoTest {
		if err := g.generateResourceTest(); err != nil {
			return nil, err
		}
	}

	return g.genfiles, nil
}

// Cleanup removes the entire "app" directory if it was created by this generator.
func (g *Generator) Cleanup() {
	if len(g.genfiles) == 0 {
		return
	}
	os.RemoveAll(g.OutDir)
	g.genfiles = nil
}

// generateContexts iterates through the API resources and actions and generates the action
// contexts.
func (g *Generator) generateContexts() (err error) {
	var (
		ctxFile string
		ctxWr   *ContextsWriter
	)
	{
		ctxFile = filepath.Join(g.OutDir, "contexts.go")
		ctxWr, err = NewContextsWriter(ctxFile)
		if err != nil {
			return
		}
	}
	defer func() {
		ctxWr.Close()
		if err == nil {
			err = ctxWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application Contexts", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
		codegen.SimpleImport("context"),
	}
	g.API.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			if a.Payload != nil {
				imports = codegen.AttributeImports(a.Payload.AttributeDefinition, imports, nil)
			}
			return nil
		})
	})

	g.genfiles = append(g.genfiles, ctxFile)
	if err = ctxWr.WriteHeader(title, g.Target, imports); err != nil {
		return
	}
	err = g.API.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := codegen.Goify(a.Name, true) + codegen.Goify(a.Parent.Name, true) + "Context"
			headers := &design.AttributeDefinition{
				Type: design.Object{},
			}
			if r.Headers != nil {
				headers.Merge(r.Headers)
				headers.Validation = r.Headers.Validation
			}
			if a.Headers != nil {
				headers.Merge(a.Headers)
				headers.Validation = a.Headers.Validation
			}
			if headers != nil && len(headers.Type.ToObject()) == 0 {
				headers = nil // So that {{if .Headers}} returns false in templates
			}
			params := a.AllParams()
			if params != nil && len(params.Type.ToObject()) == 0 {
				params = nil // So that {{if .Params}} returns false in templates
			}

			non101 := make(map[string]*design.ResponseDefinition)
			for k, v := range a.Responses {
				if v.Status != 101 {
					non101[k] = v
				}
			}
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Payload:      a.Payload,
				Params:       params,
				Headers:      headers,
				Routes:       a.Routes,
				Responses:    non101,
				API:          g.API,
				DefaultPkg:   g.Target,
				Security:     a.Security,
			}
			return ctxWr.Execute(&ctxData)
		})
	})
	return
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateControllers() (err error) {
	var (
		ctlFile string
		ctlWr   *ControllersWriter
	)
	{
		ctlFile = filepath.Join(g.OutDir, "controllers.go")
		ctlWr, err = NewControllersWriter(ctlFile)
		if err != nil {
			return
		}
	}
	defer func() {
		ctlWr.Close()
		if err == nil {
			err = ctlWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application Controllers", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
		codegen.SimpleImport("regexp"),
	}
	encoders, err := BuildEncoders(g.API.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := BuildEncoders(g.API.Consumes, false)
	if err != nil {
		return err
	}
	encoderImports := make(map[string]bool)
	for _, data := range encoders {
		encoderImports[data.PackagePath] = true
	}
	for _, data := range decoders {
		encoderImports[data.PackagePath] = true
	}
	var packagePaths []string
	for packagePath := range encoderImports {
		if packagePath != "github.com/goadesign/goa" {
			packagePaths = append(packagePaths, packagePath)
		}
	}
	sort.Strings(packagePaths)
	for _, packagePath := range packagePaths {
		imports = append(imports, codegen.SimpleImport(packagePath))
	}
	if err = ctlWr.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	if err = ctlWr.WriteInitService(encoders, decoders); err != nil {
		return err
	}

	g.genfiles = append(g.genfiles, ctlFile)
	var controllersData []*ControllerTemplateData
	g.API.IterateResources(func(r *design.ResourceDefinition) error {
		// Create file servers for all directory file servers that serve index.html.
		fileServers := r.FileServers
		for _, fs := range r.FileServers {
			if fs.IsDir() {
				rpath := design.WildcardRegex.ReplaceAllLiteralString(fs.RequestPath, "")
				rpath += "/"
				fileServers = append(fileServers, &design.FileServerDefinition{
					Parent:      fs.Parent,
					Description: fs.Description,
					Docs:        fs.Docs,
					FilePath:    filepath.Join(fs.FilePath, "index.html"),
					RequestPath: rpath,
					Metadata:    fs.Metadata,
					Security:    fs.Security,
				})
			}
		}
		data := &ControllerTemplateData{
			API:            g.API,
			Resource:       codegen.Goify(r.Name, true),
			PreflightPaths: r.PreflightPaths(),
			FileServers:    fileServers,
		}
		r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			unmarshal := fmt.Sprintf("unmarshal%s%sPayload", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":            codegen.Goify(a.Name, true),
				"DesignName":      a.Name,
				"Routes":          a.Routes,
				"Context":         context,
				"Unmarshal":       unmarshal,
				"Payload":         a.Payload,
				"PayloadOptional": a.PayloadOptional,
				"Security":        a.Security,
			}
			data.Actions = append(data.Actions, action)
			return nil
		})
		if len(data.Actions) > 0 || len(data.FileServers) > 0 {
			data.Encoders = encoders
			data.Decoders = decoders
			data.Origins = r.AllOrigins()
			controllersData = append(controllersData, data)
		}
		return nil
	})
	err = ctlWr.Execute(controllersData)
	return
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateSecurity() (err error) {
	if len(g.API.SecuritySchemes) == 0 {
		return nil
	}

	var (
		secFile string
		secWr   *SecurityWriter
	)
	{
		secFile = filepath.Join(g.OutDir, "security.go")
		secWr, err = NewSecurityWriter(secFile)
		if err != nil {
			return
		}
	}
	defer func() {
		secWr.Close()
		if err == nil {
			err = secWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application Security", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	if err = secWr.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, secFile)
	err = secWr.Execute(design.Design.SecuritySchemes)

	return
}

// generateHrefs iterates through the API resources and generates the href factory methods.
func (g *Generator) generateHrefs() (err error) {
	var (
		hrefFile string
		resWr    *ResourcesWriter
	)
	{
		hrefFile = filepath.Join(g.OutDir, "hrefs.go")
		resWr, err = NewResourcesWriter(hrefFile)
		if err != nil {
			return
		}
	}
	defer func() {
		resWr.Close()
		if err == nil {
			err = resWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application Resource Href Factories", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("strings"),
	}
	if err = resWr.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, hrefFile)
	err = g.API.IterateResources(func(r *design.ResourceDefinition) error {
		m := g.API.MediaTypeWithIdentifier(r.MediaType)
		var identifier string
		if m != nil {
			identifier = m.Identifier
		} else {
			identifier = "text/plain"
		}
		data := ResourceData{
			Name:              codegen.Goify(r.Name, true),
			Identifier:        identifier,
			Description:       r.Description,
			Type:              m,
			CanonicalTemplate: codegen.CanonicalTemplate(r),
			CanonicalParams:   codegen.CanonicalParams(r),
		}
		return resWr.Execute(&data)
	})
	return
}

// generateMediaTypes iterates through the media types and generate the data structures and
// marshaling code.
func (g *Generator) generateMediaTypes() (err error) {
	var (
		mtFile string
		mtWr   *MediaTypesWriter
	)
	{
		mtFile = filepath.Join(g.OutDir, "media_types.go")
		mtWr, err = NewMediaTypesWriter(mtFile)
		if err != nil {
			return
		}
	}
	defer func() {
		mtWr.Close()
		if err == nil {
			err = mtWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application Media Types", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}
	for _, v := range g.API.MediaTypes {
		imports = codegen.AttributeImports(v.AttributeDefinition, imports, nil)
	}
	if err = mtWr.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, mtFile)
	err = g.API.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.IsError() {
			return nil
		}
		if mt.Type.IsObject() || mt.Type.IsArray() {
			return mtWr.Execute(mt)
		}
		return nil
	})
	return
}

// generateUserTypes iterates through the user types and generates the data structures and
// marshaling code.
func (g *Generator) generateUserTypes() (err error) {
	var (
		utFile string
		utWr   *UserTypesWriter
	)
	{
		utFile = filepath.Join(g.OutDir, "user_types.go")
		utWr, err = NewUserTypesWriter(utFile)
		if err != nil {
			return
		}
	}
	defer func() {
		utWr.Close()
		if err == nil {
			err = utWr.FormatCode()
		}
	}()
	title := fmt.Sprintf("%s: Application User Types", g.API.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}
	for _, v := range g.API.Types {
		imports = codegen.AttributeImports(v.AttributeDefinition, imports, nil)
	}
	if err = utWr.WriteHeader(title, g.Target, imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, utFile)
	err = g.API.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		return utWr.Execute(t)
	})
	return
}
