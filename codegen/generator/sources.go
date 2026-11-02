// This file produces the gen command's source without filesystem output or
// protobuf compilation. The same plugin lifecycle owns names and templates.
package generator

import (
	"fmt"
	"path/filepath"

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

// sources uses an explicit registry so tests can prove lifecycle parity and
// callback isolation without changing production registrations.
func sources(genpkg string, roots []eval.Root, registry *registry) (map[string][]byte, error) {
	run, err := newGenerationRun("gen", registry)
	if err != nil {
		return nil, err
	}
	run.grpcPlans = grpccodegen.NewSourcePlans
	result, err := run.execute(genpkg, roots)
	if err != nil {
		return nil, err
	}
	files, err := mergeFilesByPath(result.files)
	if err != nil {
		return nil, err
	}
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
