package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/pkg/inflect"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Generator is the application code generator.
type Generator struct {
	*goagen.GoGenerator
	ContextsWriter    *ContextsWriter
	HandlersWriter    *HandlersWriter
	ResourcesWriter   *ResourcesWriter
	contextsFilename  string
	handlersFilename  string
	resourcesFilename string
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
	goagen.RegisterFlags(app)
	NewCommand().RegisterFlags(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf(`invalid command line: %s. Command line was "%s"`,
			err, strings.Join(os.Args, " "))
	}
	outdir := filepath.Join(goagen.OutputDir, "app")
	if _, err = os.Stat(outdir); err == nil {
		if !goagen.Force {
			if cwd, err := os.Getwd(); err == nil {
				outdir, _ = filepath.Rel(cwd, outdir)
			}
			return nil, fmt.Errorf("directory %#v already exists, use --foce to overwrite", outdir)
		}
	}
	os.RemoveAll(outdir)
	if err = os.MkdirAll(outdir, 0777); err != nil {
		return nil, err
	}
	ctxFile := filepath.Join(outdir, "contexts.go")
	hdlFile := filepath.Join(outdir, "handlers.go")
	resFile := filepath.Join(outdir, "resources.go")

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
	return &Generator{
		GoGenerator:       goagen.NewGoGenerator(outdir),
		ContextsWriter:    ctxWr,
		HandlersWriter:    hdlWr,
		ResourcesWriter:   resWr,
		contextsFilename:  ctxFile,
		handlersFilename:  hdlFile,
		resourcesFilename: resFile,
	}, nil
}

// Generate the application code, implement goagen.Generator.
func (g *Generator) Generate(api *design.APIDefinition) ([]string, error) {
	if api == nil {
		return nil, fmt.Errorf("missing API definition, make sure design.Design is properly initialized")
	}
	title := fmt.Sprintf("%s: Application Contexts", api.Name)
	g.ContextsWriter.WriteHeader(title, TargetPackage, nil)
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
			}
			return g.ContextsWriter.Execute(&ctxData)
		})
	})
	if err != nil {
		return nil, err
	}
	if err := g.ContextsWriter.FormatCode(); err != nil {
		return nil, err
	}

	title = fmt.Sprintf("%s: Application Handlers", api.Name)
	g.HandlersWriter.WriteHeader(title, TargetPackage, nil)
	var handlersData []*HandlerTemplateData
	api.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			if len(a.Routes) > 0 {
				name := fmt.Sprintf("%s%sHandler", a.FormatName(true), r.FormatName(false, true))
				context := fmt.Sprintf("%s%sContext", a.FormatName(false), r.FormatName(false, false))
				handlersData = append(handlersData, &HandlerTemplateData{
					Resource: r.Name,
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
	if err := g.HandlersWriter.Execute(handlersData); err != nil {
		return nil, err
	}
	if err := g.HandlersWriter.FormatCode(); err != nil {
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

		data := ResourceTemplateData{
			Name:              r.Name,
			Identifier:        identifier,
			Description:       r.Description,
			Type:              resType,
			CanonicalTemplate: canoTemplate,
			CanonicalParams:   canoParams,
		}
		return g.ResourcesWriter.Execute(&data)
	})
	if err != nil {
		return nil, err
	}
	if err := g.ResourcesWriter.FormatCode(); err != nil {
		return nil, err
	}

	return []string{g.contextsFilename, g.handlersFilename, g.resourcesFilename}, nil
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
