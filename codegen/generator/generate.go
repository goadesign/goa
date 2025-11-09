package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"golang.org/x/tools/go/packages"
)

// Generate runs the code generation algorithms.
func Generate(dir, cmd string, debug bool) (outputs []string, err1 error) {
	startGenerate := time.Now()
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Starting generator.Generate()\n")
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
		pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, path)
		if err != nil {
			return nil, err
		}
		// In temporary workspaces (e.g., tests) and on Windows, PkgPath may resolve
		// to an absolute filesystem path which is not a valid Go import path and
		// would produce invalid imports (e.g., backslashes). Fall back to the
		// relative generated package import path in that case.
		if filepath.IsAbs(pkgs[0].PkgPath) {
			genpkg = codegen.Gendir
		} else {
			genpkg = pkgs[0].PkgPath
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate]   packages.Load took %v\n", time.Since(startPkgLoad))
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 2: Compute gen package import path took %v\n", time.Since(start))
		}
	}

	// 3. Retrieve goa generators for given command.
	var genfuncs []Genfunc
	{
		start := time.Now()
		gs, err := Generators(cmd)
		if err != nil {
			return nil, err
		}
		genfuncs = gs
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 3: Retrieve goa generators took %v (%d generators)\n", time.Since(start), len(genfuncs))
		}
	}

	// 4. Run the code pre generation plugins.
	{
		start := time.Now()
		err := codegen.RunPluginsPrepare(cmd, genpkg, roots)
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 4: Run pre-generation plugins took %v\n", time.Since(start))
		}
	}

	// 5. Generate initial set of files produced by goa code generators.
	// NOTE: Parallelization causes infinite recursion in AsObject() for circular type references
	var genfiles []*codegen.File
	{
		start := time.Now()
		for i, gen := range genfuncs {
			genStart := time.Now()
			fs, err := gen(genpkg, roots)
			if err != nil {
				return nil, err
			}
			genfiles = append(genfiles, fs...)
			if debug {
				fmt.Fprintf(os.Stderr, "[TIMING]     [generate]   Generator %d produced %d files in %v\n", i, len(fs), time.Since(genStart))
			}
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 5: Generate initial files took %v (total %d files)\n", time.Since(start), len(genfiles))
		}
	}

	// 6. Run the code generation plugins.
	{
		start := time.Now()
		var err error
		genfiles, err = codegen.RunPlugins(cmd, genpkg, roots, genfiles)
		if err != nil {
			return nil, err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 6: Run post-generation plugins took %v (now %d files)\n", time.Since(start), len(genfiles))
		}
	}

	// 7. Merge files that target the same path to avoid overwriting content when
	// multiple generators (or services) emit sections for the same file.
	{
		start := time.Now()
		genfiles = mergeFilesByPath(genfiles)
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 7: Merging files by path took %v (now %d files)\n", time.Since(start), len(genfiles))
		}
	}

	// 8. Write the files (in parallel).
	written := make(map[string]struct{})
	{
		start := time.Now()
		numWorkers := runtime.NumCPU()
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 8: Starting parallel file writing with %d workers\n", numWorkers)
		}

		// Channel for work items
		type workItem struct {
			index int
			file  *codegen.File
		}
		workChan := make(chan workItem, len(genfiles))

		// Channel for results
		type result struct {
			index    int
			filename string
			duration time.Duration
			err      error
		}
		resultChan := make(chan result, len(genfiles))

		// Start worker pool
		var wg sync.WaitGroup
		for range numWorkers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for work := range workChan {
					renderStart := time.Now()
					filename, err := work.file.Render(dir)
					resultChan <- result{
						index:    work.index,
						filename: filename,
						duration: time.Since(renderStart),
						err:      err,
					}
				}
			}()
		}

		// Send all files to work channel
		for i, f := range genfiles {
			workChan <- workItem{index: i, file: f}
		}
		close(workChan)

		// Wait for all workers to finish in a separate goroutine
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		var firstErr error
		slowRenders := 0
		for res := range resultChan {
			if res.err != nil && firstErr == nil {
				firstErr = res.err
			}
			if res.filename != "" {
				written[res.filename] = struct{}{}
			}
			// Only log slow renders (>100ms) to avoid spam
			if debug && res.duration > 100*time.Millisecond {
				fmt.Fprintf(os.Stderr, "[TIMING]     [generate]   File %d (%s) render took %v\n", res.index, res.filename, res.duration)
				slowRenders++
			}
		}

		if firstErr != nil {
			return nil, firstErr
		}

		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 8: Write files took %v (%d files written, %d slow renders)\n", time.Since(start), len(written), slowRenders)
		}
	}

	// 9. Compute all output filenames.
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
			fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Stage 9: Compute output filenames took %v\n", time.Since(start))
		}
	}
	sort.Strings(outputs)

	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]     [generate] Total generator.Generate() took %v\n", time.Since(startGenerate))
	}
	return outputs, nil
}

// mergeFilesByPath coalesces files that share the same output path by
// concatenating their non-header sections and merging header imports. This
// prevents later renders from truncating earlier content when multiple
// services contribute sections to the same file (e.g., shared user types with
// union value methods).
func mergeFilesByPath(files []*codegen.File) []*codegen.File {
	if len(files) <= 1 {
		return files
	}

	byPath := make(map[string]*codegen.File)
	namesByPath := make(map[string]map[string]struct{})

	// First pass: build merged file per path
	for _, f := range files {
		if f == nil {
			continue
		}
		path := f.Path
		if existing, ok := byPath[path]; ok {
			// Merge headers (index 0) imports
			if len(existing.SectionTemplates) > 0 && len(f.SectionTemplates) > 0 {
				mergeHeaderImports(existing.SectionTemplates[0], f.SectionTemplates[0])
			}
			// Initialize seen section names for this path
			if namesByPath[path] == nil {
				namesByPath[path] = make(map[string]struct{})
				for _, st := range existing.SectionTemplates {
					namesByPath[path][st.Name] = struct{}{}
				}
			}
			// Append unique sections (skip header at index 0)
			for i, st := range f.SectionTemplates {
				if i == 0 {
					continue
				}
				if _, seen := namesByPath[path][st.Name]; seen {
					continue
				}
				existing.SectionTemplates = append(existing.SectionTemplates, st)
				namesByPath[path][st.Name] = struct{}{}
			}
			// Preserve a finalize function if destination does not have one
			if existing.FinalizeFunc == nil && f.FinalizeFunc != nil {
				existing.FinalizeFunc = f.FinalizeFunc
			}
			// Skip adding a duplicate File entry
			continue
		}

		// New path: record and initialize seen names
		byPath[path] = f
		m := make(map[string]struct{})
		for _, st := range f.SectionTemplates {
			m[st.Name] = struct{}{}
		}
		namesByPath[path] = m
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
	return merged
}

// mergeHeaderImports merges the import specs from src header into dst header,
// deduplicating by (Name, Path). If either section is not a header produced by
// codegen.Header, this function is a no-op.
func mergeHeaderImports(dst, src *codegen.SectionTemplate) {
	if dst == nil || src == nil {
		return
	}
	dmap, dok := dst.Data.(map[string]any)
	smap, sok := src.Data.(map[string]any)
	if !dok || !sok {
		return
	}
	dlist, _ := dmap["Imports"].([]*codegen.ImportSpec)
	slist, _ := smap["Imports"].([]*codegen.ImportSpec)
	if len(slist) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(dlist))
	for _, imp := range dlist {
		if imp == nil {
			continue
		}
		seen[imp.Name+"|"+imp.Path] = struct{}{}
	}
	for _, imp := range slist {
		if imp == nil {
			continue
		}
		key := imp.Name + "|" + imp.Path
		if _, ok := seen[key]; ok {
			continue
		}
		dlist = append(dlist, imp)
		seen[key] = struct{}{}
	}
	dmap["Imports"] = dlist
}
