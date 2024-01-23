package codegen

import (
	"embed"
	"fmt"
	"path"
	"strings"
)

//go:embed templates/*
var tmplFS embed.FS

// readTemplate returns the service template with the given name.
func readTemplate(name string, partials ...string) string {
	var prefix string
	{
		var partialDefs []string
		for _, partial := range partials {
			tmpl, err := tmplFS.ReadFile(path.Join("templates", "partial", partial+".go.tpl"))
			if err != nil {
				panic("failed to read partial template " + partial + ": " + err.Error()) // Should never happen, bug if it does
			}
			partialDefs = append(partialDefs,
				fmt.Sprintf("{{ define \"partial_%s\" }}\n%s{{ end }}", partial, string(tmpl)))
		}
		prefix = strings.Join(partialDefs, "\n")
	}
	content, err := tmplFS.ReadFile(path.Join("templates", name) + ".go.tpl")
	if err != nil {
		panic("failed to load template " + name + ": " + err.Error()) // Should never happen, bug if it does
	}
	return prefix + "\n" + string(content)
}
