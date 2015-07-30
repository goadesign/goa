package main

import (
	"os"
	"path/filepath"
)

// Generator is the goa code generator.
type Generator struct {
	Outdir        string   // Output directory
	TargetPackage string   // Target package name
	Files         []string // Generated files
}

// WriteCode writes the code and updates the generator Files field accordingly.
func (g *Generator) WriteCode() error {
	g.Files = nil
	c, err := NewContextsWriter()
	if err != nil {
		return err
	}
	ctxFile, err := os.Create(filepath.Join(g.Outdir, "contexts.go"))
	if err != nil {
		return err
	}
	err = c.Write(g.TargetPackage, ctxFile)
	if err != nil {
		return err
	}
	h, err := NewHandlersWriter()
	if err != nil {
		return err
	}
	hFile, err := os.Create(filepath.Join(g.Outdir, "handlers.go"))
	if err != nil {
		return err
	}
	err = h.Write(g.TargetPackage, hFile)
	if err != nil {
		return err
	}
	r, err := NewResourcesWriter()
	if err != nil {
		return err
	}
	resFile, err := os.Create(filepath.Join(g.Outdir, "resources.go"))
	if err != nil {
		return err
	}
	return r.Write(g.TargetPackage, resFile)
}
