package example

import (
	"embed"
	"path/filepath"
)

//go:embed templates/*
var tmplFS embed.FS

// template returns the example template with the given name.
func template(name string) string {
	content, err := tmplFS.ReadFile(filepath.Join("templates/", name) + ".go.tpl")
	if err != nil {
		panic("failed to load template " + name + ": " + err.Error()) // Should never happen, bug if it does
	}
	return string(content)
}
