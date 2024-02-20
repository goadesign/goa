package openapiv2

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Files returns the OpenAPI v2 specification files in JSON and YAML formats.
func Files(root *expr.RootExpr) ([]*codegen.File, error) {
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	if err != nil {
		return nil, err
	}
	jsonSection := &codegen.SectionTemplate{
		Name:    "openapi",
		FuncMap: template.FuncMap{"toJSON": toJSON(root.API.Meta)},
		Source:  "{{ toJSON .}}",
		Data:    spec,
	}
	yamlSection := &codegen.SectionTemplate{
		Name:    "openapi",
		FuncMap: template.FuncMap{"toYAML": toYAML},
		Source:  "{{ toYAML .}}",
		Data:    spec,
	}
	return []*codegen.File{
		{
			Path:             filepath.Join(codegen.Gendir, "http", "openapi.json"),
			SectionTemplates: []*codegen.SectionTemplate{jsonSection},
		},
		{
			Path:             filepath.Join(codegen.Gendir, "http", "openapi.yaml"),
			SectionTemplates: []*codegen.SectionTemplate{yamlSection},
		},
	}, nil
}

func toJSON(meta expr.MetaExpr) func(any) string {
	prefix, p := meta.Last("openapi:json:prefix")
	indent, i := meta.Last("openapi:json:indent")
	marshal := json.Marshal
	if p || i {
		marshal = func(v any) ([]byte, error) {
			return json.MarshalIndent(v, prefix, indent)
		}
	}
	return func(d any) string {
		b, err := marshal(d)
		if err != nil {
			panic("openapi: " + err.Error()) // bug
		}
		return string(b)
	}
}

func toYAML(d any) string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}
