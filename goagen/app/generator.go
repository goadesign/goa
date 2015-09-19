package app

import (
	"fmt"
	"os"
	"path/filepath"

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
	apiName           string
}

// NewGenerator returns the application code generator.
func NewGenerator() (*Generator, error) {
	app := kingpin.New("Code generator", "application code generator")
	goagen.RegisterFlags(app)
	NewCommand().RegisterFlags(app)
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	outdir := goagen.OutputDir
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
	var name string
	if design.Design == nil {
		name = "<missing API definition>"
	} else {
		name = design.Design.Name
	}
	return &Generator{
		GoGenerator:       goagen.NewGoGenerator(outdir),
		ContextsWriter:    ctxWr,
		HandlersWriter:    hdlWr,
		ResourcesWriter:   resWr,
		contextsFilename:  ctxFile,
		handlersFilename:  hdlFile,
		resourcesFilename: resFile,
		apiName:           name,
	}, nil
}

// Generate the application code, implement goagen.Generator.
func (g *Generator) Generate() ([]string, error) {
	if design.Design == nil {
		return nil, fmt.Errorf("missing API definition, make sure design.Design is properly initialized")
	}
	imports := []string{}
	title := fmt.Sprintf("%s: Application Contexts", g.apiName)
	g.ContextsWriter.WriteHeader(title, TargetPackage, imports)
	err := design.Design.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Context"
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Params:       a.Params,
				Payload:      a.Payload,
				Headers:      a.Headers,
				Responses:    a.Responses,
				MediaTypes:   design.Design.MediaTypes,
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

	imports = []string{}
	title = fmt.Sprintf("%s: Application Handlers", g.apiName)
	g.HandlersWriter.WriteHeader(title, TargetPackage, imports)
	var handlersData []*HandlerTemplateData
	design.Design.IterateResources(func(r *design.ResourceDefinition) error {
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

	imports = []string{}
	title = fmt.Sprintf("%s: Application Resources", g.apiName)
	g.ResourcesWriter.WriteHeader(title, TargetPackage, imports)
	err = design.Design.IterateResources(func(r *design.ResourceDefinition) error {
		m, ok := design.Design.MediaTypes[r.MediaType]
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
