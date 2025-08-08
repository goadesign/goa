package codegen

import (
	goa "goa.design/goa/v3/pkg"
)

// Header returns a Go source file header section template.
func Header(title, pack string, imports []*ImportSpec) *SectionTemplate {
	return &SectionTemplate{
		Name:   "source-header",
		Source: codegenTemplates.Read(headerT),
		Data: map[string]any{
			"Title":       title,
			"ToolVersion": goa.Version(),
			"Pkg":         pack,
			"Imports":     imports,
		},
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
