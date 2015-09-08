package app

import (
	"path/filepath"

	"bitbucket.org/pkg/inflect"
	"github.com/raphael/goa/design"
)

// Writer is the application code writer.
type Writer struct {
	contextsWriter    *ContextsWriter
	handlersWriter    *HandlersWriter
	resourcesWriter   *ResourcesWriter
	contextsFilename  string
	handlersFilename  string
	resourcesFilename string
	targetPackage     string
}

// NewWriter creates a new application code writer.
func NewWriter(outdir, target string) (*Writer, error) {
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
	}, nil
}

// Write writes the code and returns the list of generated files in case of success, an error.
func (w *Writer) Write() ([]string, error) {
	imports := []string{}
	w.contextsWriter.WriteHeader(w.targetPackage, imports)
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
	if err := w.handlersWriter.Write(w.targetPackage); err != nil {
		return nil, err
	}
	if err := w.resourcesWriter.Write(w.targetPackage); err != nil {
		return nil, err
	}
	return []string{w.contextsFilename, w.handlersFilename, w.resourcesFilename}, nil
}
