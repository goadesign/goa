package codegen

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	goa "goa.design/goa/v3/pkg"
)

// Header returns a Go source file header section template. It panics when the
// imports give one package path different explicit package names.
func Header(title, pack string, imports []*ImportSpec) *SectionTemplate {
	return &SectionTemplate{
		Name:   "source-header",
		Source: codegenTemplates.Read(headerT),
		Data: map[string]any{
			"Title":   title,
			"Pkg":     pack,
			"Imports": appendImports(nil, imports...),
		},
	}
}

// VersionFile returns a file that contains the goa version used to generate
// the code. The file is written to gen/goa.json.
func VersionFile() *File {
	data := map[string]string{"goa_version": goa.Version()}
	b, _ := json.MarshalIndent(data, "", "  ")
	return &File{
		Path: filepath.Join(Gendir, "goa.json"),
		SectionTemplates: []*SectionTemplate{{
			Name:   "goa-version",
			Source: string(b),
		}},
	}
}

// AddImport adds imports to a section template that was generated with Header.
// It panics when one package path is given different explicit package names.
func AddImport(section *SectionTemplate, imprts ...*ImportSpec) {
	if len(imprts) == 0 {
		return
	}
	var specs []*ImportSpec
	data := section.Data.(map[string]any)
	if imports, ok := data["Imports"]; ok {
		specs = imports.([]*ImportSpec)
	}
	data["Imports"] = appendImports(specs, imprts...)
}

// appendImports keeps one import for each package path. An explicit package
// name replaces an unspecified name. Different explicit names are a generator
// error because one Go file cannot use both names for the same package.
func appendImports(existing []*ImportSpec, additions ...*ImportSpec) []*ImportSpec {
	positions := make(map[string]int, len(existing)+len(additions))
	result := make([]*ImportSpec, 0, len(existing)+len(additions))
	appendImport := func(spec *ImportSpec) {
		position, ok := positions[spec.Path]
		if !ok {
			positions[spec.Path] = len(result)
			result = append(result, spec)
			return
		}
		current := result[position]
		switch {
		case current.Name == spec.Name, spec.Name == "":
			return
		case current.Name == "":
			result[position] = spec
		default:
			panic(fmt.Sprintf(
				"import path %q uses package names %q and %q",
				spec.Path,
				current.Name,
				spec.Name,
			))
		}
	}
	for _, spec := range existing {
		appendImport(spec)
	}
	for _, spec := range additions {
		appendImport(spec)
	}
	return result
}
