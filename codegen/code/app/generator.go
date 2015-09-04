package app

import (
	"path/filepath"

	"bitbucket.org/pkg/inflect"

	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/design"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Generator is the goa code generator.
type Generator struct {
	Outdir        string   // Output directory
	TargetPackage string   // Target package name
	Files         []string // Generated files
}

// NewGenerator returns the application code generator instance.
func NewGenerator() codegen.Generator {
	return new(Generator)
}

// RegisterFlags initializes the generator command line flags in the given kingpin command.
func RegisterFlags(cmd *kingpin.CmdClause) {
}

// WriteCode writes the code and updates the generator Files field accordingly.
func (g *Generator) WriteCode() error {
	g.Files = nil
	ctxFile := filepath.Join(g.Outdir, "contexts.go")
	hdlFile := filepath.Join(g.Outdir, "handlers.go")
	resFile := filepath.Join(g.Outdir, "resources.go")

	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		return err
	}
	hdlWr, err := NewHandlersWriter(hdlFile)
	if err != nil {
		return err
	}
	resWr, err := NewResourcesWriter(resFile)
	if err != nil {
		return err
	}

	ctxWr.WriteHeader(g.TargetPackage)
	for _, res := range design.Design.Resources {
		for _, a := range res.Actions {
			ctxName := inflect.Camelize(a.Name) + inflect.Camelize(a.Resource.Name) + "Context"
			ctxData := ContextData{
				Name:         ctxName,
				ResourceName: res.Name,
				ActionName:   a.Name,
				Params:       a.Params,
				Payload:      a.Payload,
				Headers:      a.Headers,
				Responses:    a.Responses,
			}
			if err = ctxWr.Write(&ctxData); err != nil {
				return err
			}
		}
	}
	if err = hdlWr.Write(g.TargetPackage); err != nil {
		return err
	}
	if err = resWr.Write(g.TargetPackage); err != nil {
		return err
	}
	return nil
}
