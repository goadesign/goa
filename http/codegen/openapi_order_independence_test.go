// This file verifies HTTP and OpenAPI analysis produce identical examples
// regardless of which transport representation is analyzed first.
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

// TestOpenAPIOrderIndependence verifies that the OpenAPI specifications do not
// depend on whether the HTTP transport data was computed first: the HTTP
// analyze pass must treat the design expression tree as read-only so the
// OpenAPI generators always see the pristine design. The production "goa gen"
// flow runs the transport generators before the OpenAPI one while the OpenAPI
// golden tests run on pristine roots; any difference between the two
// generations is output that production emits but no golden test covers.
func TestOpenAPIOrderIndependence(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		// Aliased payload/result attributes: makeHTTPType used to flatten
		// the aliases in place which changed the schemas OpenAPI generated.
		{"alias-type", testdata.AliasTypeDSL},
		{"result-body-multiple-views", testdata.ResultBodyMultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"error-response", testdata.PrimitiveErrorResponseDSL},
		{"streaming-result", testdata.StreamingResultDSL},
		{"streaming-payload", testdata.StreamingPayloadDSL},
		// NOTE: methods declaring anonymous object results (e.g.
		// testdata.SSEObjectDSL) only pass this check because the raw
		// object wrapping moved out of the service analyze pass into
		// codegen.NewGeneration, which CreateHTTPServices constructs before
		// computing the transport data. The pristine root below is rendered
		// without generation ownership, so designs whose OpenAPI output depends
		// on the wrapping must prepare both roots (see
		// TestGeneratorsTreatDesignAsReadOnly in codegen/generator for the
		// full read-only guarantee).
		{"sse", testdata.SSEStringDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Golden test order: generate the OpenAPI specifications from a
			// pristine root.
			pristine := renderOpenAPI(t, expr.RunDSL(t, c.DSL))

			// Production order: compute the HTTP transport data first, then
			// generate the OpenAPI specifications from the same root.
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			for _, svc := range root.API.HTTP.Services {
				require.NotNil(t, services.Get(svc.Name()))
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
