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
	for _, res := range design.Design.Resources {
		for _, a := range res.Actions {
			ctxName := inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Context"
			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: res.Name,
				ActionName:   a.Name,
				Params:       a.Params,
				Payload:      a.Payload,
				Headers:      a.Headers,
				Responses:    a.Responses,
			}
			if err := w.contextsWriter.Write(&ctxData); err != nil {
				return nil, err
			}
		}
	}
	if err := w.contextsWriter.FormatCode(); err != nil {
		return nil, err
	}

	imports = []string{}
	title = fmt.Sprintf("%s: Application Handlers", w.apiName)
	w.handlersWriter.WriteHeader(title, w.targetPackage, imports)
	if err := w.handlersWriter.Write(w.targetPackage); err != nil {
		return nil, err
	}
	if err := w.handlersWriter.FormatCode(); err != nil {
		return nil, err
	}

	imports = []string{}
	title = fmt.Sprintf("%s: Application Resources", w.apiName)
	w.contextsWriter.WriteHeader(title, w.targetPackage, imports)
	for _, res := range design.Design.Resources {
		m, ok := design.Design.MediaTypes[res.MediaType]
		var identifier string
		var resType *design.UserTypeDefinition
		if ok {
			identifier = m.Identifier
			resType = m.UserTypeDefinition
		} else {
			identifier = "application/text"
		}
		canoTemplate, canoParams := res.CanonicalPathAndParams()
		canoTemplate = design.ParamsRegex.ReplaceAllLiteralString(canoTemplate, "%s")

		data := ResourceTemplateData{
			Name:              res.Name,
			Identifier:        identifier,
			Description:       res.Description,
			Type:              resType,
			CanonicalTemplate: canoTemplate,
			CanonicalParams:   canoParams,
		}
		if err := w.resourcesWriter.Write(&data); err != nil {
			return nil, err
		}
	}
	if err := w.resourcesWriter.FormatCode(); err != nil {
		return nil, err
	}

	return []string{w.contextsFilename, w.handlersFilename, w.resourcesFilename}, nil
}
