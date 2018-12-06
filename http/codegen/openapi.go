package codegen

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/openapi"
)

type (
	// openAPI is the OpenAPI spec file implementation.
	openAPI struct {
		spec *openapi.V2
	}
)

// OpenAPIFiles returns the files for the OpenAPIFile spec of the given HTTP API.
func OpenAPIFiles(root *expr.RootExpr) ([]*codegen.File, error) {
	jsonPath := filepath.Join(codegen.Gendir, "http", "openapi.json")
	yamlPath := filepath.Join(codegen.Gendir, "http", "openapi.yaml")
	var (
		jsonSection *codegen.SectionTemplate
		yamlSection *codegen.SectionTemplate
	)
	{
		spec, err := openapi.NewV2(root, root.API.Servers[0].Hosts[0])
		if err != nil {
			return nil, err
		}
		jsonSection = &codegen.SectionTemplate{
			Name:    "openapi",
			FuncMap: template.FuncMap{"toJSON": toJSON},
			Source:  "{{ toJSON .}}",
			Data:    spec,
		}
		yamlSection = &codegen.SectionTemplate{
			Name:    "openapi",
			FuncMap: template.FuncMap{"toYAML": toYAML},
			Source:  "{{ toYAML .}}",
			Data:    spec,
		}
	}

	return []*codegen.File{
		{
			Path:             jsonPath,
			SectionTemplates: []*codegen.SectionTemplate{jsonSection},
		},
		{
			Path:             yamlPath,
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
