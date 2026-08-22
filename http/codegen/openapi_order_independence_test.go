// This file verifies that HTTP and OpenAPI generation produce the same examples
// regardless of which one reads the design first.
package codegen

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

// TestOpenAPIOrderIndependence verifies that building HTTP files first does not
// change the OpenAPI documents. HTTP generation must not change the design that
// the OpenAPI generator reads afterward.
func TestOpenAPIOrderIndependence(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		// Named request and result fields used to be flattened in place, which
		// changed the OpenAPI schemas generated afterward.
		{"alias-type", testdata.AliasTypeDSL},
		{"result-body-multiple-views", testdata.ResultBodyMultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"error-response", testdata.PrimitiveErrorResponseDSL},
		{"streaming-result", testdata.StreamingResultDSL},
		{"streaming-payload", testdata.StreamingPayloadDSL},
		// Anonymous object results pass because codegen.NewGeneration prepares
		// those result objects before HTTP generation starts. See
		// TestGeneratorsTreatDesignAsReadOnly for the check that covers the whole
		// generation run.
		{"sse", testdata.SSEStringDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// First build OpenAPI documents from an untouched design.
			pristine := renderOpenAPI(t, expr.RunDSL(t, c.DSL))

			// Then build HTTP files first and OpenAPI documents second from the
			// same design.
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			for _, svc := range root.API.HTTP.Services {
				require.NotNil(t, plan.services.Get(svc.Name()))
			}
			produced := renderOpenAPI(t, root)

			require.Equal(t, len(pristine), len(produced))
			for path, content := range pristine {
				assert.Equal(t, content, produced[path], "OpenAPI file %q changed after the HTTP transport data was computed", path)
			}
		})
	}
}

// renderOpenAPI generates and renders all the OpenAPI specification files for
// the given root and returns their content indexed by file path. The global
// schema registry is reset first and the call receives a fresh example
// generator so two identical design trees yield identical documents.
func renderOpenAPI(t *testing.T, root *expr.RootExpr) map[string]string {
	t.Helper()
	openapi.Definitions = make(map[string]*openapi.Schema)
	files, err := OpenAPIFiles(root, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	out := make(map[string]string, len(files))
	for _, f := range files {
		require.Len(t, f.SectionTemplates, 1)
		s := f.SectionTemplates[0]
		var buf bytes.Buffer
		tmpl := template.Must(template.New("openapi").Funcs(s.FuncMap).Parse(s.Source))
		require.NoError(t, tmpl.Execute(&buf, s.Data))
		out[f.Path] = buf.String()
	}
	return out
}
