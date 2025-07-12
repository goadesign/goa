package template

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// TemplateReader reads templates and partials from a provided filesystem.
type TemplateReader struct {
	FS fs.FS
}

// Read returns the template with the given name, optionally including partials.
// Partials are loaded from the 'partial' subdirectory and defined as named blocks.
func (tr *TemplateReader) Read(name string, partials ...string) string {
	var prefix string
	if len(partials) > 0 {
		var partialDefs []string
		for _, partial := range partials {
			content, err := fs.ReadFile(tr.FS, path.Join("templates", "partial", partial+".go.tpl"))
			if err != nil {
				panic(fmt.Sprintf("failed to read partial template %s: %v", partial, err))
			}
			partialDefs = append(partialDefs,
				fmt.Sprintf("{{- define \"partial_%s\" }}\n%s{{- end }}", partial, string(content)))
		}
		prefix = strings.Join(partialDefs, "\n")
	}
	content, err := fs.ReadFile(tr.FS, path.Join("templates", name)+".go.tpl")
	if err != nil {
		panic(fmt.Sprintf("failed to load template %s: %v", name, err))
	}
	if prefix != "" {
		return prefix + "\n" + string(content)
	}
	return string(content)
}
