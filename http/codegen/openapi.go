package codegen

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	"goa.design/goa/codegen"
	"goa.design/goa/http/codegen/openapi"
	httpdesign "goa.design/goa/http/design"
)

type (
	// openAPI is the OpenAPI spec file implementation.
	openAPI struct {
		spec *openapi.V2
	}
)

// OpenAPIFiles returns the files for the OpenAPIFile spec of the given HTTP API.
func OpenAPIFiles(root *httpdesign.RootExpr) ([]*codegen.File, error) {
	jsonPath := filepath.Join(codegen.Gendir, "http", "openapi.json")
	yamlPath := filepath.Join(codegen.Gendir, "http", "openapi.yaml")
	var (
		jsonSection *codegen.SectionTemplate
		yamlSection *codegen.SectionTemplate
	)
	{
		spec, err := openapi.NewV2(root)
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
		&codegen.File{
			Path:             jsonPath,
			SectionTemplates: []*codegen.SectionTemplate{jsonSection},
		},
		&codegen.File{
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
