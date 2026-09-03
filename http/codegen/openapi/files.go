// This file writes OpenAPI documents as JSON and YAML code generation files.
package openapi

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

var yamlDatePrefix = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}(?:$|[Tt ].*)`)

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
	return quoteDateShapedYAMLStrings(string(b))
}

// quoteDateShapedYAMLStrings keeps YAML readers from treating string examples
// as timestamps. Some readers reject date-shaped strings with invalid months
// or days before they can see that the OpenAPI schema declares a string.
func quoteDateShapedYAMLStrings(source string) string {
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		valueStart := strings.LastIndex(line, ": ")
		if valueStart >= 0 {
			valueStart += 2
		} else {
			trimmed := strings.TrimLeft(line, " ")
			if !strings.HasPrefix(trimmed, "- ") {
				continue
			}
			valueStart = len(line) - len(trimmed) + 2
		}
		value := line[valueStart:]
		if yamlDatePrefix.MatchString(value) {
			lines[index] = line[:valueStart] + strconv.Quote(value)
		}
	}
	return strings.Join(lines, "\n")
}
