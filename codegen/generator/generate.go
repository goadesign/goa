// The goa command calls this file with an output directory, command, and debug
// flag; it reads the evaluated design roots and returns the files it wrote.
// Every core generator and plugin plans against one Generation. Before any
// callback renders files, Goa chooses every package and declaration name and
// rejects attempts to add another declaration.
package generator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"golang.org/x/mod/module"
	"golang.org/x/tools/go/packages"
)

// Generate runs the code generation algorithms.
func Generate(dir, cmd string, debug bool) (outputs []string, err1 error) {
	return generate(dir, cmd, debug, defaultRegistry)
}

// generate runs code generation with an explicit registry so package tests can
// use isolated factories without replacing production globals.
func generate(dir, cmd string, debug bool, registry *registry) (outputs []string, err1 error) {
	startGenerate := time.Now()
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Starting generator.Generate()\n")
	}
	run, err := newGenerationRun(cmd, registry)
	if err != nil {
		return nil, err
	}

	// 1. Compute design roots.
	var roots []eval.Root
	{
		start := time.Now()
		rs, err := eval.Context.Roots()
		if err != nil {
			return nil, err
		}
		roots = rs
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 1: Compute design roots took %v\n", time.Since(start))
		}
	}

	// 2. Compute "gen" package import path.
	var genpkg string
	{
		start := time.Now()
		base, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(base, codegen.Gendir)
		if err := os.MkdirAll(path, 0750); err != nil {
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

		startPkgLoad := time.Now()
		genpkg, err = generatedPackageImportPath(path)
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate]   packages.Load took %v\n", time.Since(startPkgLoad))
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 2: Compute gen package import path took %v\n", time.Since(start))
		}
	}

	// 3. Prepare roots and build one plan. Choose every package and declaration
	// name, then render core and plugin files through the fresh run objects
	// created before root evaluation.
	startLifecycle := time.Now()
	result, err := run.execute(genpkg, roots)
	if err != nil {
		return nil, err
	}
	genfiles := result.files
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 3: Lifecycle produced %d files in %v\n", len(genfiles), time.Since(startLifecycle))
	}

	// 8. Merge files that target the same path to avoid overwriting content when
	// multiple generators (or services) emit sections for the same file.
	{
		start := time.Now()
		genfiles, err = mergeFilesByPath(genfiles)
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 8: Merging files by path took %v (now %d files)\n", time.Since(start), len(genfiles))
		}
	}

	// 9. Emit goa.json version file (gen command only).
	if cmd == "gen" {
		genfiles = append(genfiles, codegen.VersionFile())
	}

	// 10. Write the files in parallel, then audit the prepared design after all
	// templates and file finalizers have completed.
	written := make(map[string]struct{})
	{
		start := time.Now()
		numWorkers := runtime.NumCPU()
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 10: Starting parallel file writing with %d workers\n", numWorkers)
		}

		type workItem struct {
			index int
			file  *codegen.File
		}
		workChan := make(chan workItem, len(genfiles))

		type renderResult struct {
			index    int
			filename string
			duration time.Duration
			err      error
		}
		resultChan := make(chan renderResult, len(genfiles))

		var workers sync.WaitGroup
		for range numWorkers {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for work := range workChan {
					renderStart := time.Now()
					filename, err := work.file.Render(dir)
					if err != nil {
						err = fmt.Errorf("render %s: %w", work.file.Path, err)
					}
					resultChan <- renderResult{
						index:    work.index,
						filename: filename,
						duration: time.Since(renderStart),
						err:      err,
					}
				}
			}()
		}

		for i, file := range genfiles {
			workChan <- workItem{index: i, file: file}
		}
		close(workChan)

		go func() {
			workers.Wait()
			close(resultChan)
		}()

		firstErrorIndex := len(genfiles)
		var firstErr error
		slowRenders := 0
		for render := range resultChan {
			if render.err != nil && render.index < firstErrorIndex {
				firstErrorIndex = render.index
				firstErr = render.err
			}
			if render.filename != "" {
				written[render.filename] = struct{}{}
			}
			if debug && render.duration > 100*time.Millisecond {
				fmt.Fprintf(os.Stderr, "[TIMING]     [generate]   File %d (%s) render took %v\n", render.index, render.filename, render.duration)
				slowRenders++
			}
		}
		if err := result.plan.verifyPreparedDesign("generated file renders"); err != nil {
			return nil, err
		}
		if firstErr != nil {
			return nil, firstErr
		}

		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 10: Write files took %v (%d files written, %d slow renders)\n", time.Since(start), len(written), slowRenders)
		}
	}

	// 11. Compute all output filenames.
	{
		start := time.Now()
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
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 11: Compute output filenames took %v\n", time.Since(start))
		}
	}
	sort.Strings(outputs)

	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Total generator.Generate() took %v\n", time.Since(startGenerate))
	}
	return outputs, nil
}

// generatedPackageImportPath asks Go to identify the package in dir and
// returns its cleaned import path for generated files.
func generatedPackageImportPath(dir string) (string, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedModule | packages.NeedFiles,
		Dir:  dir,
	}, ".")
	if err != nil {
		return "", fmt.Errorf("load generated Go package in %q: %w", dir, err)
	}
	if len(pkgs) != 1 {
		return "", fmt.Errorf("load generated Go package in %q: expected exactly one package, got %d", dir, len(pkgs))
	}
	pkg := pkgs[0]
	if len(pkg.Errors) != 0 {
		packageErrors := make([]error, len(pkg.Errors))
		for i, packageError := range pkg.Errors {
			packageErrors[i] = packageError
		}
		return "", fmt.Errorf("load generated Go package in %q: %w", dir, errors.Join(packageErrors...))
	}
	importPath := pkg.PkgPath
	if pkg.Module == nil && strings.HasPrefix(importPath, "_/") {
		owned, err := gopathOwnsImportPath(pkg.Dir, importPath)
		if err != nil {
			return "", err
		}
		if !owned {
			return "", fmt.Errorf("generated Go package in %q has synthetic import path %q", dir, importPath)
		}
	}
	if path.Clean(importPath) != importPath {
		return "", fmt.Errorf("generated Go package in %q has noncanonical import path %q", dir, importPath)
	}
	if err := module.CheckImportPath(importPath); err != nil {
		return "", fmt.Errorf("generated Go package in %q has invalid import path %q: %w", dir, importPath, err)
	}
	return importPath, nil
}

// gopathOwnsImportPath asks the Go command for its effective GOPATH and reports
// whether dir appears at importPath beneath one of GOPATH's source directories.
func gopathOwnsImportPath(dir, importPath string) (bool, error) {
	output, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return false, fmt.Errorf("read effective GOPATH: %w", err)
	}
	gopath := string(output)
	if strings.HasSuffix(gopath, "\r\n") {
		gopath = strings.TrimSuffix(gopath, "\r\n")
	} else {
		gopath = strings.TrimSuffix(gopath, "\n")
	}
	roots := filepath.SplitList(gopath)
	for _, root := range roots {
		if gopathSourceOwnsImportPath(filepath.Join(root, "src"), dir, importPath) {
			return true, nil
		}
	}

	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return false, fmt.Errorf("resolve generated Go package directory %q: %w", dir, err)
	}
	var resolutionErrors []error
	for _, root := range roots {
		packagePath := filepath.Join(root, "src", filepath.FromSlash(importPath))
		resolvedPackagePath, err := filepath.EvalSymlinks(packagePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			resolutionErrors = append(resolutionErrors, fmt.Errorf("resolve %q: %w", packagePath, err))
			continue
		}
		if resolvedPackagePath == resolvedDir {
			return true, nil
		}
	}
	if len(resolutionErrors) != 0 {
		return false, fmt.Errorf("resolve GOPATH package path for %q: %w", importPath, errors.Join(resolutionErrors...))
	}
	return false, nil
}

// gopathSourceOwnsImportPath reports whether dir is lexically beneath source
// with the exact slash-separated relative import path.
func gopathSourceOwnsImportPath(source, dir, importPath string) bool {
	relative, err := filepath.Rel(source, dir)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	return filepath.ToSlash(relative) == importPath
}

// mergeFilesByPath coalesces files that share the same output path by
// concatenating their non-header sections and merging header imports. This
// prevents later renders from truncating earlier content when multiple
// services contribute sections to the same file (e.g., shared user types with
// union value methods).
func mergeFilesByPath(files []*codegen.File) ([]*codegen.File, error) {
	if len(files) == 0 {
		return files, nil
	}

	byPath := make(map[string]*codegen.File)
	portablePaths := make(map[string]string)

	// First pass: build one complete file per path.
	for _, f := range files {
		if f == nil {
			continue
		}
		canonicalPath, portablePath, err := canonicalOutputFilePath(f.Path)
		if err != nil {
			return nil, err
		}
		if claimedPath, ok := portablePaths[portablePath]; ok && claimedPath != canonicalPath {
			return nil, fmt.Errorf(
				"generated file paths %q and %q collide on a case-insensitive filesystem",
				claimedPath,
				canonicalPath,
			)
		}
		portablePaths[portablePath] = canonicalPath
		f.Path = canonicalPath
		if existing, ok := byPath[canonicalPath]; ok {
			if existing.SkipExist != f.SkipExist {
				return nil, fmt.Errorf("generated file %q has conflicting SkipExist settings", canonicalPath)
			}
			existingHeader, existingHasHeader := firstHeader(existing)
			contributorHeader, contributorHasHeader := firstHeader(f)
			if existingHasHeader != contributorHasHeader {
				return nil, fmt.Errorf("generated file %q mixes header and headerless contributions", canonicalPath)
			}
			sectionStart := 0
			if existingHasHeader {
				if err := mergeHeaderImports(existingHeader, contributorHeader); err != nil {
					return nil, fmt.Errorf("merge generated file %q: %w", canonicalPath, err)
				}
				sectionStart = 1
			}
			existing.SectionTemplates = append(existing.SectionTemplates, f.SectionTemplates[sectionStart:]...)
			existing.FinalizeFunc = composeFinalizers(existing.FinalizeFunc, f.FinalizeFunc)
			continue
		}

		byPath[canonicalPath] = f
	}

	// Second pass: preserve original order by first occurrence of each path
	merged := make([]*codegen.File, 0, len(byPath))
	seenPaths := make(map[string]struct{})
	for _, f := range files {
		if f == nil {
			continue
		}
		if _, ok := seenPaths[f.Path]; ok {
			continue
		}
		if mf, ok := byPath[f.Path]; ok {
			merged = append(merged, mf)
			seenPaths[f.Path] = struct{}{}
		}
	}
	return merged, nil
}

// mergeHeaderImports merges the import specs from src header into dst header,
// rejecting package and alias conflicts rather than producing invalid Go.
func mergeHeaderImports(dst, src *codegen.SectionTemplate) error {
	dmap, _ := dst.Data.(map[string]any)
	smap, _ := src.Data.(map[string]any)
	dpkg, _ := dmap["Pkg"].(string)
	spkg, _ := smap["Pkg"].(string)
	if dpkg != spkg {
		return fmt.Errorf("header packages %q and %q conflict", dpkg, spkg)
	}
	dlist, _ := dmap["Imports"].([]*codegen.ImportSpec)
	slist, _ := smap["Imports"].([]*codegen.ImportSpec)
	paths := make(map[string]string, len(dlist)+len(slist))
	aliases := make(map[string]string, len(dlist)+len(slist))
	for _, imp := range dlist {
		if _, err := recordImportSpec(paths, aliases, imp); err != nil {
			return err
		}
	}
	for _, imp := range slist {
		duplicate, err := recordImportSpec(paths, aliases, imp)
		if err != nil {
			return err
		}
		if !duplicate {
			dlist = append(dlist, imp)
		}
	}
	dmap["Imports"] = dlist
	return nil
}

// recordImportSpec validates one import against the complete merged header and
// reports whether the exact path and alias were already present.
func recordImportSpec(paths, names map[string]string, spec *codegen.ImportSpec) (bool, error) {
	if spec == nil {
		return true, nil
	}
	if alias, ok := paths[spec.Path]; ok {
		if alias != spec.Name {
			return false, fmt.Errorf("import path %q uses aliases %q and %q", spec.Path, alias, spec.Name)
		}
		return true, nil
	}
	localName := spec.Name
	if localName != "" && localName != "_" && localName != "." {
		if importPath, ok := names[localName]; ok {
			return false, fmt.Errorf("import name %q refers to paths %q and %q", localName, importPath, spec.Path)
		}
		names[localName] = spec.Path
	}
	paths[spec.Path] = spec.Name
	return false, nil
}

// canonicalOutputFilePath cleans rawPath into the portable relative path used
// to group and render a generated file. The second result is case-folded so
// two paths cannot overwrite one another on a case-insensitive filesystem.
func canonicalOutputFilePath(rawPath string) (string, string, error) {
	portable := filepath.ToSlash(rawPath)
	portable = strings.ReplaceAll(portable, `\`, "/")
	canonical := path.Clean(portable)
	if canonical == "." ||
		canonical == ".." ||
		strings.HasPrefix(canonical, "../") ||
		strings.HasPrefix(canonical, "/") {
		return "", "", fmt.Errorf("generated file path %q must stay within the output directory", rawPath)
	}
	if strings.Contains(canonical, ":") {
		return "", "", fmt.Errorf("generated file path %q is not portable", rawPath)
	}
	return filepath.FromSlash(canonical), strings.ToLower(canonical), nil
}

// firstHeader reports the header produced by codegen.Header when it is the
// first section of file.
func firstHeader(file *codegen.File) (*codegen.SectionTemplate, bool) {
	if len(file.SectionTemplates) == 0 {
		return nil, false
	}
	header := file.SectionTemplates[0]
	data, ok := header.Data.(map[string]any)
	if !ok {
		return nil, false
	}
	_, hasPackage := data["Pkg"].(string)
	_, hasImports := data["Imports"].([]*codegen.ImportSpec)
	return header, hasPackage && hasImports
}

// composeFinalizers preserves every same-path contributor's post-render work
// in contributor order and stops at the first error.
func composeFinalizers(first, second func(string) error) func(string) error {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func(path string) error {
		if err := first(path); err != nil {
			return err
		}
		return second(path)
	}
}
