package rest

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design/rest"
)

type (
	// openAPI is the OpenAPI spec file implementation.
	openAPI struct {
		spec *OpenAPIV2
	}
)

// OpenAPI returns the file for the OpenAPI spec of the given HTTP API.
func OpenAPI(root *rest.RootExpr) (codegen.File, error) {
	spec, err := makeOpenAPIV2(root)
	if err != nil {
		return nil, err
	}
	return &openAPI{spec}, nil
}

// Sections is the list of file sections.
func (w *openAPI) Sections(_ string) []*codegen.Section {
	funcs := template.FuncMap{"toJSON": toJSON}
	tmpl := template.Must(template.New("openapiV2").Funcs(funcs).Parse(openapiTmpl))
	return []*codegen.Section{&codegen.Section{
		Template: tmpl,
		Data:     w.spec,
	}}
}

// OutputPath is the relative path to the output file.
func (w *openAPI) OutputPath() string {
	return filepath.Join("openapi", "swagger.json")
}

// Finalize is a no-op for this file.
func (w *openAPI) Finalize(_ string) error { return nil }

func toJSON(d interface{}) string {
	b, err := json.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}

// Dummy template
const openapiTmpl = `{{ toJSON . }}`
