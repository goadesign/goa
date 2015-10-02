package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/pkg/inflect"

	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/design"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Generator is the application code generator.
type Generator struct {
	*codegen.GoGenerator
	ContextsWriter     *ContextsWriter
	HandlersWriter     *HandlersWriter
	ResourcesWriter    *ResourcesWriter
	MediaTypesWriter   *MediaTypesWriter
	UserTypesWriter    *UserTypesWriter
	contextsFilename   string
	handlersFilename   string
	resourcesFilename  string
	mediaTypesFilename string
	userTypesFilename  string
	canDeleteOutputDir bool
	genfiles           []string
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
	canDeleteDir := false
	if _, err = os.Stat(outdir); err == nil {
		if !codegen.Force {
			if cwd, err := os.Getwd(); err == nil {
				outdir, _ = filepath.Rel(cwd, outdir)
			}
			return nil, fmt.Errorf("directory %#v already exists, use --force to overwrite", outdir)
		}
	} else {
		canDeleteDir = true
	}
	os.RemoveAll(outdir)
	if err = os.MkdirAll(outdir, 0777); err != nil {
		return nil, err
	}
	ctxFile := filepath.Join(outdir, "contexts.go")
	hdlFile := filepath.Join(outdir, "handlers.go")
	resFile := filepath.Join(outdir, "resources.go")
	mtFile := filepath.Join(outdir, "media_types.go")
	utFile := filepath.Join(outdir, "user_types.go")

	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		panic(err) // bug
	}
	hdlWr, err := NewHandlersWriter(hdlFile)
	if err != nil {
		panic(err) // bug
	}
	resWr, err := NewResourcesWriter(resFile)
	if err != nil {
		panic(err) // bug
	}
	mtWr, err := NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	utWr, err := NewUserTypesWriter(utFile)
	if err != nil {
		panic(err) // bug
	}
	return &Generator{
		GoGenerator:        codegen.NewGoGenerator(outdir),
		ContextsWriter:     ctxWr,
		HandlersWriter:     hdlWr,
		ResourcesWriter:    resWr,
		MediaTypesWriter:   mtWr,
		UserTypesWriter:    utWr,
		contextsFilename:   ctxFile,
		handlersFilename:   hdlFile,
		resourcesFilename:  resFile,
		mediaTypesFilename: mtFile,
		userTypesFilename:  utFile,
		canDeleteOutputDir: canDeleteDir,
	}, nil
}

// AppOutputDir returns the directory containing the generated files.
func AppOutputDir() string {
	return filepath.Join(codegen.OutputDir, AppSubDir)
}

// Generate the application code, implement codegen.Generator.
func (g *Generator) Generate(api *design.APIDefinition) ([]string, error) {
	if api == nil {
		return nil, fmt.Errorf("missing API definition, make sure design.Design is properly initialized")
	}
	title := fmt.Sprintf("%s: Application Contexts", api.Name)
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/raphael/goa"),
		codegen.SimpleImport("strconv"),
	}
	g.ContextsWriter.WriteHeader(title, TargetPackage, imports)
	err := api.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := inflect.Camelize(a.Name) + inflect.Camelize(a.Parent.Name) + "Context"
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Payload:      a.Payload,
				Params:       r.Params.Merge(a.Params),
				Headers:      r.Headers.Merge(a.Headers),
				Responses:    MergeResponses(r.Responses, a.Responses),
				MediaTypes:   api.MediaTypes,
				Types:        api.Types,
			}
			return g.ContextsWriter.Execute(&ctxData)
		})
	})
	g.genfiles = append(g.genfiles, g.contextsFilename)
	if err != nil {
		g.Cleanup()
		return nil, err
	}
	if err := g.ContextsWriter.FormatCode(); err != nil {
		g.Cleanup()
		return nil, err
	}

	title = fmt.Sprintf("%s: Application Handlers", api.Name)
	imports = []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("github.com/raphael/goa"),
	}
	g.HandlersWriter.WriteHeader(title, TargetPackage, imports)
	var handlersData []*HandlerTemplateData
	api.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			if len(a.Routes) > 0 {
				name := fmt.Sprintf("%s%sHandler", a.FormatName(true), r.FormatName(false, true))
				context := fmt.Sprintf("%s%sContext", a.FormatName(false), r.FormatName(false, false))
				handlersData = append(handlersData, &HandlerTemplateData{
					Resource: r.FormatName(true, true),
					Action:   a.Name,
					Verb:     a.Routes[0].Verb,
					Path:     a.Routes[0].Path,
					Name:     name,
					Context:  context,
				})
			}
			return nil
		})
	})
	g.genfiles = append(g.genfiles, g.handlersFilename)
	if err := g.HandlersWriter.Execute(handlersData); err != nil {
		g.Cleanup()
		return nil, err
	}
	if err := g.HandlersWriter.FormatCode(); err != nil {
		g.Cleanup()
		return nil, err
	}

	title = fmt.Sprintf("%s: Application Resources", api.Name)
	g.ResourcesWriter.WriteHeader(title, TargetPackage, nil)
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		m, ok := api.MediaTypes[r.MediaType]
		var identifier string
		var resType *design.UserTypeDefinition
		if ok {
			identifier = m.Identifier
			resType = m.UserTypeDefinition
		} else {
			identifier = "application/text"
		}
		canoTemplate, canoParams := r.CanonicalPathAndParams()
		canoTemplate = design.ParamsRegex.ReplaceAllLiteralString(canoTemplate, "%s")

		data := ResourceData{
			Name:              r.FormatName(false, false),
			Identifier:        identifier,
			Description:       r.Description,
			Type:              resType,
			CanonicalTemplate: canoTemplate,
			CanonicalParams:   canoParams,
		}
		return g.ResourcesWriter.Execute(&data)
	})
	g.genfiles = append(g.genfiles, g.resourcesFilename)
	if err != nil {
		g.Cleanup()
		return nil, err
	}
	if err := g.ResourcesWriter.FormatCode(); err != nil {
		g.Cleanup()
		return nil, err
	}

	title = fmt.Sprintf("%s: Application Media Types", api.Name)
	g.MediaTypesWriter.WriteHeader(title, TargetPackage, nil)
	err = api.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.Type.IsObject() {
			return g.MediaTypesWriter.Execute(mt)
		}
		return nil
	})
	g.genfiles = append(g.genfiles, g.mediaTypesFilename)
	if err != nil {
		g.Cleanup()
		return nil, err
	}
	if err := g.MediaTypesWriter.FormatCode(); err != nil {
		g.Cleanup()
		return nil, err
	}

	return g.genfiles, nil

	title = fmt.Sprintf("%s: Application User Types", api.Name)
	g.UserTypesWriter.WriteHeader(title, TargetPackage, nil)
	err = api.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		return g.UserTypesWriter.Execute(t)
	})
	g.genfiles = append(g.genfiles, g.userTypesFilename)
	if err != nil {
		g.Cleanup()
		return nil, err
	}
	if err := g.UserTypesWriter.FormatCode(); err != nil {
		g.Cleanup()
		return nil, err
	}

	return g.genfiles, nil
}

// Cleanup removes the entire "app" directory if it was created by this generator.
func (g *Generator) Cleanup() {
	if len(g.genfiles) == 0 {
		return
	}
	if g.canDeleteOutputDir {
		os.RemoveAll(AppOutputDir())
	} else {
		for _, f := range g.genfiles {
			os.Remove(f)
		}
	}
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
