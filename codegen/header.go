package codegen

import (
	"encoding/json"
	"path/filepath"

	goa "goa.design/goa/v3/pkg"
)

// Header returns a Go source file header section template.
func Header(title, pack string, imports []*ImportSpec) *SectionTemplate {
	return &SectionTemplate{
		Name:   "source-header",
		Source: codegenTemplates.Read(headerT),
		Data: map[string]any{
			"Title":   title,
			"Pkg":     pack,
			"Imports": imports,
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

// AddImport adds imports to a section template that was generated with
// Header.
func AddImport(section *SectionTemplate, imprts ...*ImportSpec) {
	if len(imprts) == 0 {
		return
	}
	var specs []*ImportSpec
	if data, ok := section.Data.(map[string]any); ok {
		if imports, ok := data["Imports"]; ok {
			specs = imports.([]*ImportSpec)
		}
		data["Imports"] = append(specs, imprts...)
	}
}
