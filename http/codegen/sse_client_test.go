package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSSEClient(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"string", testdata.SSEStringDSL},
		{"int", testdata.SSEIntDSL},
		{"bool", testdata.SSEBoolDSL},
		{"object", testdata.SSEObjectDSL},
		{"data-field", testdata.SSEDataFieldDSL},
		{"data-id-field", testdata.SSEDataIDFieldDSL},
		{"request-id", testdata.SSERequestIDDSL},
		{"all-fields", testdata.SSEAllFieldsDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 3)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[1])
			golden := filepath.Join("testdata", "golden", "sse-client-"+c.Name+".golden")
			testutil.CompareOrUpdateGolden(t, code, golden)
		})
	}
}

// TestSSEClientSpecializesDataAndRetryParsing checks that generated clients
// parse each designed field into its exact Go type.
func TestSSEClientSpecializesDataAndRetryParsing(t *testing.T) {
	tests := []struct {
		name     string
		design   func()
		contains []string
	}{
		{
			name:   "string alias",
			design: ssePrimitiveAliasDSL,
			contains: []string{
				"body = dataContent",
				"event = NewWatchEventTextOK(body)",
			},
		},
		{
			name:   "optional data field",
			design: testdata.SSEDataFieldDSL,
			contains: []string{
				"value := dataContent",
				"body.Data = &value",
				"event = NewSSEDataFieldMethodResultOK(&body)",
			},
		},
		{
			name:   "viewed data field",
			design: viewedSSEDataFieldDSL,
			contains: []string{
				"value := dataContent",
				"body.Data = &value",
			},
		},
		{
			name:   "viewed alias data field",
			design: viewedSSEPrimitiveAliasDataFieldDSL,
			contains: []string{
				"value := dataContent",
				"body.Data = &value",
			},
		},
		{
			name:   "retry",
			design: testdata.SSEAllFieldsDSL,
			contains: []string{
				`retryContent := string(value)`,
				`strconv.ParseInt(retryContent, 10, 0)`,
				"body.Retry = &value",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := expr.RunDSL(t, test.design)
			code := renderedFile(t, linkedHTTPPlanForRoot(t, root).ClientFiles())
			for _, expected := range test.contains {
				require.Contains(t, code, expected)
			}
			require.NotContains(t, code, "retry value parsing depends on the field type")
		})
	}
}
