package openapiv3

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Files returns the OpenAPI v3 specification files in JSON and YAML formats.
func Files(root *expr.RootExpr) ([]*codegen.File, error) {
	spec := New(root)
	jsonSection := &codegen.SectionTemplate{
		Name:    "openapi_v3",
		FuncMap: template.FuncMap{"toJSON": toJSON},
		Source:  "{{ toJSON .}}",
		Data:    spec,
	}
	yamlSection := &codegen.SectionTemplate{
		Name:    "openapi_v3",
		FuncMap: template.FuncMap{"toYAML": toYAML},
		Source:  "{{ toYAML .}}",
		Data:    spec,
	}

	return []*codegen.File{
		{
			Path:             filepath.Join(codegen.Gendir, "http", "openapi3.json"),
			SectionTemplates: []*codegen.SectionTemplate{jsonSection},
		},
		{
			Path:             filepath.Join(codegen.Gendir, "http", "openapi3.yaml"),
			SectionTemplates: []*codegen.SectionTemplate{yamlSection},
		},
	}, nil
}

func toJSON(d interface{}) string {
	b, err := json.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}

func toYAML(d interface{}) string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}
