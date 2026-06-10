package openapi

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Files returns the codegen files rendering the given OpenAPI document in
// JSON and YAML at gen/<path>.json and gen/<path>.yaml. section names the
// generated section templates and meta provides the JSON formatting options
// ("openapi:json:prefix" and "openapi:json:indent").
func Files(spec any, meta expr.MetaExpr, section, path string) []*codegen.File {
	jsonSection := &codegen.SectionTemplate{
		Name:    section,
		FuncMap: template.FuncMap{"toJSON": toJSON(meta)},
		Source:  "{{ toJSON .}}",
		Data:    spec,
	}
	yamlSection := &codegen.SectionTemplate{
		Name:    section,
		FuncMap: template.FuncMap{"toYAML": toYAML},
		Source:  "{{ toYAML .}}",
		Data:    spec,
	}
	return []*codegen.File{
		{
			Path:             filepath.Join(codegen.Gendir, filepath.FromSlash(path)+".json"),
			SectionTemplates: []*codegen.SectionTemplate{jsonSection},
		},
		{
			Path:             filepath.Join(codegen.Gendir, filepath.FromSlash(path)+".yaml"),
			SectionTemplates: []*codegen.SectionTemplate{yamlSection},
		},
	}
}

// toJSON returns a template function encoding its argument in JSON, honoring
// the "openapi:json:prefix" and "openapi:json:indent" formatting meta.
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

// toYAML encodes its argument in YAML.
func toYAML(d any) string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}
