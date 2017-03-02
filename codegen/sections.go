package codegen

import "goa.design/goa.v2/pkg"

// Header returns a header section.
func Header(title, pack string, imports []*ImportSpec) *Section {
	return &Section{
		Template: *headerTmpl,
		Data: map[string]interface{}{
			"Title":       title,
			"ToolVersion": pkg.Version(),
			"Pkg":         pack,
			"Imports":     imports,
		},
	}
}
