package example

import (
	"embed"
	"path"
)

//go:embed templates/*
var tmplFS embed.FS

// readTemplate returns the example template with the given name.
func readTemplate(name string) string {
	content, err := tmplFS.ReadFile(path.Join("templates", name) + ".go.tpl")
	if err != nil {
		panic("failed to load template " + name + ": " + err.Error()) // Should never happen, bug if it does
	}
	return string(content)
}
