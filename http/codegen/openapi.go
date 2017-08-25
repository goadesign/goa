package codegen

import (
	"encoding/json"
	"path/filepath"
	"text/template"

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

// OpenAPIFile returns the file for the OpenAPIFile spec of the given HTTP API.
func OpenAPIFile(root *httpdesign.RootExpr) (*codegen.File, error) {
	path := filepath.Join(codegen.Gendir, "http", "openapi.json")
	var section *codegen.SectionTemplate
	{
		spec, err := openapi.NewV2(root)
		if err != nil {
			return nil, err
		}
		section = &codegen.SectionTemplate{
			Name:    "openapi",
			FuncMap: template.FuncMap{"toJSON": toJSON},
			Source:  "{{ toJSON .}}",
			Data:    spec,
		}
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: []*codegen.SectionTemplate{section},
	}, nil
}

func toJSON(d interface{}) string {
	b, err := json.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}
