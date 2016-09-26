package genserver

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/goadesign/goa/codegen"
	"github.com/goadesign/goa/codegen/gen_server"
	"github.com/goadesign/goa/http/design"
)

// Generator is the application code generator.
type Generator struct {
	Root   *design.RootExpr // The design root expression
	OutDir string           // Path to output directory
	OutPkg string           // Name of generated package

	genfiles []string // Generated files so we know what to cleanup in case of error
}

// Generate is the generator entry point.
func Generate() ([]string, error) {
	var (
		outDir, outPkg, version string
	)

	set := flag.NewFlagSet("server", flag.PanicOnError)
	set.String("design", "", "")
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&outPkg, "pkg", "server", "")
	set.StringVar(&version, "version", "", "")
	set.Parse(os.Args[1:])
	outDir = filepath.Join(outDir, outPkg)

	if err := codegen.CheckVersion(version); err != nil {
		return nil, err
	}

	outPkg = codegen.Goify(outPkg, false)
	g := &Generator{OutDir: outDir, OutPkg: outPkg, Root: design.Root}

	return g.Generate()
}

// Generate the server code.
func (g *Generator) Generate() ([]string, error) {
	if g.Root == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go codegen.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	codegen.Reserved[g.OutPkg] = true

	os.RemoveAll(g.OutDir)

	if err := os.MkdirAll(g.OutDir, 0755); err != nil {
		return nil, err
	}
	g.genfiles = []string{g.OutDir}
	if err := g.generateInit(); err != nil {
		return nil, err
	}
	if err := g.generateControllers(); err != nil {
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

// generateInit generates the service setup code.
func (g *Generator) generateInit() error {
	initFile := filepath.Join(g.OutDir, "init.go")
	initWr, err := NewInitWriter(initFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Service Setup", g.Root.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
		codegen.SimpleImport("regexp"),
	}
	encoders, err := BuildEncoders(g.Root.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := BuildEncoders(g.Root.Consumes, false)
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
	initWr.WriteHeader(title, g.OutPkg, imports)
	initWr.Write(encoders, decoders)
	g.genfiles = append(g.genfiles, initFile)
	return initWr.FormatCode()
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateControllers() error {
	ctlFile := filepath.Join(g.OutDir, "controllers.go")
	ctlWr, err := NewControllersWriter(ctlFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Controllers", g.Root.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
		codegen.SimpleImport("regexp"),
	}
	encoders, err := BuildEncoders(g.Root.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := BuildEncoders(g.Root.Consumes, false)
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
	ctlWr.WriteHeader(title, g.OutPkg, imports)
	ctlWr.WriteInitService(encoders, decoders)

	var controllersData []*ControllerTemplateData
	err = g.Root.IterateResources(func(r *design.ResourceDefinition) error {
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
			API:            g.Root,
			Resource:       codegen.Goify(r.Name, true),
			PreflightPaths: r.PreflightPaths(),
			FileServers:    fileServers,
		}
		ierr := r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			unmarshal := fmt.Sprintf("unmarshal%s%sPayload", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":            codegen.Goify(a.Name, true),
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
		if ierr != nil {
			return ierr
		}
		if len(data.Actions) > 0 || len(data.FileServers) > 0 {
			data.Encoders = encoders
			data.Decoders = decoders
			data.Origins = r.AllOrigins()
			controllersData = append(controllersData, data)
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, ctlFile)
	if err = ctlWr.Execute(controllersData); err != nil {
		return err
	}
	return ctlWr.FormatCode()
}

// generateHrefs iterates through the API resources and generates the href factory methods.
func (g *Generator) generateHrefs() error {
	hrefFile := filepath.Join(g.OutDir, "hrefs.go")
	resWr, err := NewResourcesWriter(hrefFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Resource Href Factories", g.Root.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("strings"),
	}
	resWr.WriteHeader(title, g.OutPkg, imports)
	err = g.Root.IterateResources(func(r *design.ResourceDefinition) error {
		m := g.Root.MediaTypeWithIdentifier(r.MediaType)
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
	g.genfiles = append(g.genfiles, hrefFile)
	if err != nil {
		return err
	}
	return resWr.FormatCode()
}

// generateMediaTypes iterates through the media types and generate the data structures and
// marshaling code.
func (g *Generator) generateMediaTypes() error {
	mtFile := filepath.Join(g.OutDir, "media_types.go")
	mtWr, err := NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Media Types", g.Root.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.NewImport("uuid", "github.com/satori/go.uuid"),
	}
	mtWr.WriteHeader(title, g.OutPkg, imports)
	err = g.Root.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.IsError() {
			return nil
		}
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
func (g *Generator) generateUserTypes() error {
	file := filepath.Join(g.OutDir, "user_types.go")
	if err := genserver.GenerateUserTypes(g.OutPkg, file, g.Root.UserTypes); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, file)
	return utWr.FormatCode()
}
