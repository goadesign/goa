package main

// Generator is the goa code generator.
type Generator struct {
	Outdir        string   // Output directory
	TargetPackage string   // Target package name
	Files         []string // Generated files
}

// WriteCode writes the code and updates the generator Files field accordingly.
func (g *Generator) WriteCode() error {
	g.Files = nil
	w := ContextsWriter()
	file, err := w.Write(g.TargetPackage, g.Outdir)
	if err != nil {
		return err
	}
	g.Files = append(g.Files, file)
	w := HandlersWriter()
	file, err = w.Write(g.TargetPackage, g.Outdir)
	if err != nil {
		return err
	}
	g.Files = append(g.Files, file)
	w := ResourcesWriter()
	file, err = w.Write(g.TargetPackage, g.Outdir)
	if err != nil {
		return err
	}
	g.Files = append(g.Files, file)
	return nil
}
