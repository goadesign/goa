package generator

import (
	"go/build"
	"os"
	"path/filepath"
	"sort"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
)

// Generate runs the code generation algorithms.
func Generate(dir, cmd string) ([]string, error) {
	// 1. Compute design roots.
	var roots []eval.Root
	{
		rs, err := eval.Context.Roots()
		if err != nil {
			return nil, err
		}
		roots = rs
	}

	// 2. Compute "gen" package import path.
	var genpkg string
	{
		base, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(base, codegen.Gendir)
		if err := os.MkdirAll(path, 0777); err != nil {
			return nil, err
		}
		pkg, err := build.ImportDir(path, build.FindOnly)
		if err != nil {
			return nil, err
		}
		genpkg = pkg.ImportPath
	}

	// 3. Retrieve goa generators for given command.
	var genfuncs []Genfunc
	{
		gs, err := Generators(cmd)
		if err != nil {
			return nil, err
		}
		genfuncs = gs
	}

	// 4. Generate initial set of files produced by goa code generators.
	var genfiles []*codegen.File
	for _, gen := range genfuncs {
		fs, err := gen(genpkg, roots)
		if err != nil {
			return nil, err
		}
		genfiles = append(genfiles, fs...)
	}

	// 5. Run the code generation plugins.
	genfiles, err := codegen.RunPlugins(cmd, genpkg, roots, genfiles)
	if err != nil {
		return nil, err
	}

	// 6. Write the files.
	written := make(map[string]struct{})
	for _, f := range genfiles {
		filename, err := f.Render(dir)
		if err != nil {
			return nil, err
		}
		written[filename] = struct{}{}
	}

	// 7. Compute all output filenames.
	var outputs []string
	{
		outputs = make([]string, len(written))
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		i := 0
		for o := range written {
			rel, err := filepath.Rel(cwd, o)
			if err != nil {
				rel = o
			}
			outputs[i] = rel
			i++
		}
	}
	sort.Strings(outputs)

	return outputs, nil
}
