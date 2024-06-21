package codegen

import (
	"embed"
	"path"
	"strings"
)

//go:embed templates/*
var tmplFS embed.FS

// readTemplate returns the service template with the given name.
func readTemplate(name string, partials ...string) string {
	var tmpl strings.Builder
	{
		for _, partial := range partials {
			data, err := tmplFS.ReadFile(path.Join("templates", "partial", partial+".go.tpl"))
			if err != nil {
				panic("failed to read partial template " + partial + ": " + err.Error()) // Should never happen, bug if it does
			}
			tmpl.Write(data)
			tmpl.WriteByte('\n')
		}
	}
	data, err := tmplFS.ReadFile(path.Join("templates", name) + ".go.tpl")
	if err != nil {
		panic("failed to load template " + name + ": " + err.Error()) // Should never happen, bug if it does
	}
	tmpl.Write(data)
	return tmpl.String()
}
