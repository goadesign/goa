// This file reads one HTTP design and builds the requested OpenAPI files.
// The plan builds every file immediately and returns those files later without
// reading the design again.
package codegen

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

type (
	// OpenAPIPlan stores the OpenAPI files built from one HTTP design.
	OpenAPIPlan struct {
		files []*codegen.File
	}
)

// NewOpenAPIPlan builds the OpenAPI files for root. Later calls to Files
// return these same files without reading root again.
// The "openapi:versions" API meta selects the generated specification
// versions and the "openapi:path:<version>" API meta overrides their output
// paths, see openapi.Specs.
func NewOpenAPIPlan(root *expr.RootExpr, generator *expr.ExampleGenerator) (*OpenAPIPlan, error) {
	return NewOpenAPIPlanWithValues(root, generator, openapi.Values{})
}

// NewOpenAPIPlanWithValues builds OpenAPI files using values in place of
// matching titles, descriptions, and examples from the evaluated design.
func NewOpenAPIPlanWithValues(root *expr.RootExpr, generator *expr.ExampleGenerator, values openapi.Values) (*OpenAPIPlan, error) {
	specs, err := openapi.Specs(root.API.Meta)
	if err != nil {
		return nil, err
	}
	return NewOpenAPIPlanFromSpecs(root, generator, specs, values)
}

// NewOpenAPIPlanFromSpecs builds the exact OpenAPI versions and paths in specs,
// using values in place of matching design text and examples. Paths are
// relative to the gen directory and omit the JSON or YAML extension because
// Goa writes both formats.
func NewOpenAPIPlanFromSpecs(root *expr.RootExpr, generator *expr.ExampleGenerator, specs []openapi.Spec, values openapi.Values) (*OpenAPIPlan, error) {
	if err := validateOpenAPISpecs(specs); err != nil {
		return nil, err
	}
	// Only create a OpenAPI specification if there are HTTP services.
	if len(root.API.HTTP.Services) == 0 {
		return &OpenAPIPlan{}, nil
	}

	var files []*codegen.File
	for _, spec := range specs {
		specGenerator := generator
		if examplesDisabled(root.API.Meta) {
			specGenerator = &expr.ExampleGenerator{}
		}
		var (
			fs  []*codegen.File
			err error
		)
		switch spec.Version {
		case openapi.Version20:
			fs, err = openapiv2.FilesWithValues(root, spec.Path, specGenerator, values)
			if err != nil {
				return nil, err
			}
		default: // Version30, Version32
			fs = openapiv3.FilesWithValues(root, spec.Version, spec.Path, specGenerator, values)
		}
		files = append(files, fs...)
	}
	return &OpenAPIPlan{files: files}, nil
}

// Files returns the OpenAPI files built when the plan was created.
func (p *OpenAPIPlan) Files() []*codegen.File {
	return p.files
}

// examplesDisabled reports whether API metadata suppresses examples from
// generated OpenAPI documents.
func examplesDisabled(meta expr.MetaExpr) bool {
	value, ok := meta.Last("openapi:example")
	if !ok {
		value, ok = meta.Last("swagger:example")
	}
	return ok && value == "false"
}

// validateOpenAPISpecs checks the complete version and path list before any
// file is built.
func validateOpenAPISpecs(specs []openapi.Spec) error {
	versions := make(map[openapi.Version]struct{}, len(specs))
	paths := make(map[string]openapi.Spec, len(specs))
	for _, spec := range specs {
		switch spec.Version {
		case openapi.Version20, openapi.Version30, openapi.Version32:
		default:
			return fmt.Errorf("unsupported OpenAPI version %q", spec.Version)
		}
		if _, ok := versions[spec.Version]; ok {
			return fmt.Errorf("OpenAPI version %q appears more than once", spec.Version)
		}
		versions[spec.Version] = struct{}{}
		if err := validateOpenAPIPath(spec.Path); err != nil {
			return fmt.Errorf("invalid OpenAPI %s path %q: %w", spec.Version, spec.Path, err)
		}
		for existingPath, existing := range paths {
			if existingPath == spec.Path {
				return fmt.Errorf("OpenAPI versions %s and %s use the same output path %q", existing.Version, spec.Version, spec.Path)
			}
			if strings.EqualFold(existingPath, spec.Path) {
				return fmt.Errorf(
					"OpenAPI paths %q and %q collide on a case-insensitive filesystem",
					existingPath,
					spec.Path,
				)
			}
		}
		paths[spec.Path] = spec
	}
	return nil
}

// validateOpenAPIPath checks one extension-less path relative to gen.
func validateOpenAPIPath(outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.Contains(outputPath, "\\") {
		return fmt.Errorf("path cannot contain a backslash")
	}
	if strings.HasPrefix(outputPath, "/") {
		return fmt.Errorf("path must be relative to the gen directory")
	}
	cleaned := path.Clean(outputPath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("path must not escape the gen directory")
	}
	if cleaned != outputPath {
		return fmt.Errorf("path must be clean; use %q", cleaned)
	}
	switch path.Ext(outputPath) {
	case ".json", ".yaml", ".yml":
		return fmt.Errorf("path must not include an extension")
	}
	return nil
}
