package generator

import (
	"os"
	"path/filepath"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"golang.org/x/tools/go/packages"
)

// Generate runs the code generation algorithms.
func Generate(dir, cmd string) (outputs []string, err1 error) {
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

		// We create a temporary Go file to make sure the directory is a valid Go package
		dummy, err := os.CreateTemp(path, "temp.*.go")
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := os.Remove(dummy.Name()); err != nil {
				outputs = nil
				err1 = err
			}
		}()
		if _, err = dummy.Write([]byte("package gen")); err != nil {
			return nil, err
		}
		if err = dummy.Close(); err != nil {
			return nil, err
		}

		pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, path)
		if err != nil {
			return nil, err
		}
		genpkg = pkgs[0].PkgPath
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

	// 4. Run the code pre generation plugins.
	err := codegen.RunPluginsPrepare(cmd, genpkg, roots)
	if err != nil {
		return nil, err
	}

	// 5. Generate initial set of files produced by goa code generators.
	var genfiles []*codegen.File
	for _, gen := range genfuncs {
		fs, err := gen(genpkg, roots)
		if err != nil {
			return nil, err
		}
		genfiles = append(genfiles, fs...)
	}

	// 6. Run the code generation plugins.
	genfiles, err = codegen.RunPlugins(cmd, genpkg, roots, genfiles)
	if err != nil {
		return nil, err
	}

	// 7. Write the files.
	written := make(map[string]struct{})
	for _, f := range genfiles {
		filename, err := f.Render(dir)
		if err != nil {
			return nil, err
		}
		if filename != "" {
			written[filename] = struct{}{}
		}
	}

	// 8. Compute all output filenames.
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
