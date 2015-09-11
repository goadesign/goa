package app

import (
	"fmt"
	"path/filepath"
	"regexp"

	"bitbucket.org/pkg/inflect"
	"github.com/raphael/goa/design"
)

// ParamsRegex is the regex used to capture path parameters.
var ParamsRegex = regexp.MustCompile("(?:[^/]*/:([^/]+))+")

// Writer is the application code writer.
type Writer struct {
	contextsWriter    *ContextsWriter
	handlersWriter    *HandlersWriter
	resourcesWriter   *ResourcesWriter
	contextsFilename  string
	handlersFilename  string
	resourcesFilename string
	targetPackage     string
	apiName           string
}

// NewWriter creates a new application code writer.
func NewWriter(apiName, outdir, target string) (*Writer, error) {
	ctxFile := filepath.Join(outdir, "contexts.go")
	hdlFile := filepath.Join(outdir, "handlers.go")
	resFile := filepath.Join(outdir, "resources.go")

	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		return nil, err
	}
	hdlWr, err := NewHandlersWriter(hdlFile)
	if err != nil {
		return nil, err
	}
	resWr, err := NewResourcesWriter(resFile)
	if err != nil {
		return nil, err
	}
	return &Writer{
		contextsWriter:    ctxWr,
		handlersWriter:    hdlWr,
		resourcesWriter:   resWr,
		contextsFilename:  ctxFile,
		handlersFilename:  hdlFile,
		resourcesFilename: resFile,
		targetPackage:     target,
		apiName:           apiName,
	}, nil
}

// Write writes the code and returns the list of generated files in case of success, an error.
func (w *Writer) Write() ([]string, error) {
	imports := []string{}
	title := fmt.Sprintf("%s: Application Contexts", w.apiName)
	w.contextsWriter.WriteHeader(title, w.targetPackage, imports)
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
			}
			return w.contextsWriter.Write(&ctxData)
		})
	})
	if err != nil {
		return nil, err
	}
	if err := w.contextsWriter.FormatCode(); err != nil {
		return nil, err
	}

	imports = []string{}
	title = fmt.Sprintf("%s: Application Handlers", w.apiName)
	w.handlersWriter.WriteHeader(title, w.targetPackage, imports)
	var handlersData []*ActionHandlerTemplateData
	design.Design.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			if len(a.Routes) > 0 {
				name := fmt.Sprintf("%s%sHandler", a.FormatName(true), r.FormatName(false, true))
				context := fmt.Sprintf("%s%sContext", a.FormatName(false), r.FormatName(false, false))
				handlersData = append(handlersData, &ActionHandlerTemplateData{
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
	if err := w.handlersWriter.Write(handlersData); err != nil {
		return nil, err
	}
	if err := w.handlersWriter.FormatCode(); err != nil {
		return nil, err
	}

	imports = []string{}
	title = fmt.Sprintf("%s: Application Resources", w.apiName)
	w.contextsWriter.WriteHeader(title, w.targetPackage, imports)
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
		return w.resourcesWriter.Write(&data)
	})
	if err != nil {
		return nil, err
	}
	if err := w.resourcesWriter.FormatCode(); err != nil {
		return nil, err
	}

	return []string{w.contextsFilename, w.handlersFilename, w.resourcesFilename}, nil
}
