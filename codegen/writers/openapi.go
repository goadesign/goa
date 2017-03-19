package writers

import (
	"encoding/json"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/openapi"
	rest "goa.design/goa.v2/rest/design"
)

type (
	// openAPIWriter is the OpenAPI spec writer implementation.
	openAPIWriter struct {
		spec *openapi.V2
	}
)

// OpenAPI returns the codegen.FileWriter for the OpenAPI spec of the given
// HTTP API.
func OpenAPI(root *rest.RootExpr) (codegen.FileWriter, error) {
	spec, err := openapi.MakeV2(root)
	if err != nil {
		return nil, err
	}
	return &openAPIWriter{spec}, nil
}

// Sections is the list of file sections.
func (w *openAPIWriter) Sections() []*codegen.Section {
	funcs := template.FuncMap{"toJSON": toJSON}
	tmpl := template.Must(template.New("openapiV2").Funcs(funcs).Parse(openapiTmpl))
	return []*codegen.Section{&codegen.Section{
		Template: tmpl,
		Data:     w.spec,
	}}
}

// OutputPath is the relative path to the output file.
func (w *openAPIWriter) OutputPath() string {
	return "openapi/swagger.json"
}

func toJSON(d interface{}) string {
	b, err := json.Marshal(d)
	if err != nil {
		panic("openapi: " + err.Error()) // bug
	}
	return string(b)
}

// Dummy template
const openapiTmpl = `{{ toJSON . }}`
