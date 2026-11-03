// This file produces the gen command's source without filesystem output or
// protobuf compilation. The same plugin lifecycle owns names and templates.
package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
)

// Sources returns the gen generators' relative filenames and source bytes for
// evaluated roots. genpkg is the known import path of the generated Go package.
// It runs the registered preparation and generation plugins, shared name
// planning, file merging, formatting, and prepared-design mutation audit.
//
// Sources does not discover import paths, write files, or invoke file
// finalizers. Protobuf output contains Goa's .proto source, not protoc's Go
// files. Installed protobuf programs are unnecessary. SkipExist is ignored
// because no output directory is inspected. Command-added goa.json version
// metadata is omitted; generator and plugin files are retained unchanged.
// Returned bytes belong to the caller.
//
// Registered callbacks retain their own behavior. Callers evaluating untrusted
// design or plugin code must enforce process, filesystem, and network isolation.
// Generation owns and prepares roots; concurrent runs require distinct graphs.
func Sources(genpkg string, roots []eval.Root) (map[string][]byte, error) {
	return sources(genpkg, roots, defaultRegistry)
}

// CompleteSources returns the gen generators' relative filenames and source
// bytes for evaluated roots, like Sources, but only when those bytes are the
// complete content the gen command writes: it fails before rendering anything
// if a planned file has a FinalizeFunc or SkipExist, naming each such path,
// because the gen command would rewrite or skip that file after rendering.
// This includes protobuf output, whose .proto files are compiled by a
// finalizer. Callers that publish generated files, rather than inspect them,
// use CompleteSources. Command-added goa.json version metadata is omitted, as
// in Sources. Returned bytes belong to the caller.
//
// Registered callbacks retain their own behavior. Callers evaluating untrusted
// design or plugin code must enforce process, filesystem, and network isolation.
// Generation owns and prepares roots; concurrent runs require distinct graphs.
func CompleteSources(genpkg string, roots []eval.Root) (map[string][]byte, error) {
	return completeSources(genpkg, roots, defaultRegistry)
}

// sources uses an explicit registry so tests can prove lifecycle parity and
// callback isolation without changing production registrations.
func sources(genpkg string, roots []eval.Root, registry *registry) (map[string][]byte, error) {
	files, result, err := plannedSources(genpkg, roots, registry)
	if err != nil {
		return nil, err
	}
	return renderAudited(files, result)
}

// completeSources is CompleteSources with an explicit registry, so tests can
// register files with finalizers without changing production registrations.
func completeSources(genpkg string, roots []eval.Root, registry *registry) (map[string][]byte, error) {
	files, result, err := plannedSources(genpkg, roots, registry)
	if err != nil {
		return nil, err
	}
	var deferred []string
	for _, file := range files {
		switch {
		case file.FinalizeFunc != nil:
			deferred = append(deferred, filepath.ToSlash(file.Path)+" (finalizer)")
		case file.SkipExist:
			deferred = append(deferred, filepath.ToSlash(file.Path)+" (skipped if it exists)")
		}
	}
	if len(deferred) > 0 {
		return nil, fmt.Errorf("generated files are not complete before the gen command writes them: %s", strings.Join(deferred, ", "))
	}
	return renderAudited(files, result)
}

// plannedSources runs the gen command's generation lifecycle without
// protobuf compilation plans and merges the planned files by path.
func plannedSources(genpkg string, roots []eval.Root, registry *registry) ([]*codegen.File, *generationResult, error) {
	run, err := newGenerationRun("gen", registry)
	if err != nil {
		return nil, nil, err
	}
	run.grpcPlans = grpccodegen.NewSourcePlans
	result, err := run.execute(genpkg, roots)
	if err != nil {
		return nil, nil, err
	}
	files, err := mergeFilesByPath(result.files)
	if err != nil {
		return nil, nil, err
	}
	return files, result, nil
}

// renderAudited renders the merged files and then audits the prepared design,
// so a template that mutates the design fails even when rendering succeeded.
func renderAudited(files []*codegen.File, result *generationResult) (map[string][]byte, error) {
	output, renderErr := renderSources(files)
	if err := result.plan.verifyPreparedDesign("generated source renders"); err != nil {
		return nil, err
	}
	return output, renderErr
}

// renderSources formats each merged file once without running finalizers. It
// retains one temporary syntax tree at a time while returning the source bytes.
func renderSources(files []*codegen.File) (map[string][]byte, error) {
	output := make(map[string][]byte, len(files))
	for _, file := range files {
		content, err := file.Source()
		if err != nil {
			return nil, fmt.Errorf("render source %s: %w", file.Path, err)
		}
		output[filepath.ToSlash(file.Path)] = content
	}
	return output, nil
}
