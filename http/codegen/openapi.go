package codegen

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

// OpenAPIFiles returns the files for the OpenAPIFile spec of the given HTTP API.
func OpenAPIFiles(root *expr.RootExpr) ([]*codegen.File, error) {
	// Only create a OpenAPI specification if there are HTTP services.
	if len(root.API.HTTP.Services) == 0 {
		return nil, nil
	}

	var files []*codegen.File
	{
		// OpenAPI v2
		{
			spec, err := openapi.NewV2(root, root.API.Servers[0].Hosts[0])
			if err != nil {
				return nil, err
			}
			jsonSection := &codegen.SectionTemplate{
				Name:    "openapi",
				FuncMap: template.FuncMap{"toJSON": toJSON},
				Source:  "{{ toJSON .}}",
				Data:    spec,
			}
			yamlSection := &codegen.SectionTemplate{
				Name:    "openapi",
				FuncMap: template.FuncMap{"toYAML": toYAML},
				Source:  "{{ toYAML .}}",
				Data:    spec,
			}
			files = []*codegen.File{
				{
					Path:             filepath.Join(codegen.Gendir, "http", "openapi.json"),
					SectionTemplates: []*codegen.SectionTemplate{jsonSection},
				},
				{
					Path:             filepath.Join(codegen.Gendir, "http", "openapi.yaml"),
					SectionTemplates: []*codegen.SectionTemplate{yamlSection},
				},
			}
		}
		// OpenAPI v3
		{
			spec, err := openapiv3.New(root)
			if err != nil {
				return nil, err
			}
			jsonSection := &codegen.SectionTemplate{
				Name:    "openapi",
				FuncMap: template.FuncMap{"toJSON": toJSON},
				Source:  "{{ toJSON .}}",
				Data:    spec,
			}
			yamlSection := &codegen.SectionTemplate{
				Name:    "openapi",
				FuncMap: template.FuncMap{"toYAML": toYAML},
				Source:  "{{ toYAML .}}",
				Data:    spec,
			}
			files = append(files, &codegen.File{
				Path:             filepath.Join(codegen.Gendir, "http", "openapi_v3.json"),
				SectionTemplates: []*codegen.SectionTemplate{jsonSection},
			})
			files = append(files, &codegen.File{
				Path:             filepath.Join(codegen.Gendir, "http", "openapi_v3.yaml"),
				SectionTemplates: []*codegen.SectionTemplate{yamlSection},
			})
		}
	}
	return files, nil
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
