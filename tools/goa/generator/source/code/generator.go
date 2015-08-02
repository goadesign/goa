package main

import "path/filepath"

// Generator is the goa code generator.
type Generator struct {
	Outdir        string   // Output directory
	TargetPackage string   // Target package name
	Files         []string // Generated files
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

	if err = ctxWr.Write(g.TargetPackage); err != nil {
		return err
	}
	if err = hdlWr.Write(g.TargetPackage); err != nil {
		return err
	}
	if err = resWr.Write(g.TargetPackage); err != nil {
		return err
	}
	return nil
}
