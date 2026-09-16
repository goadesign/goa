package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSSE(t *testing.T) {
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
			fs := plan.ServerFiles()
			require.Len(t, fs, 3)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[1])
			golden := filepath.Join("testdata", "golden", "sse-"+c.Name+".golden")
			testutil.CompareOrUpdateGolden(t, code, golden)
		})
	}
}

// TestSSEServerSpecializesDataEncoding checks that generated send methods use
// the designed data type directly, including named primitive types.
func TestSSEServerSpecializesDataEncoding(t *testing.T) {
	tests := []struct {
		name     string
		design   func()
		contains string
	}{
		{name: "string", design: testdata.SSEStringDSL, contains: "data = string(body)"},
		{name: "string alias", design: ssePrimitiveAliasDSL, contains: "data = string(body)"},
		{name: "object", design: testdata.SSEObjectDSL, contains: "json.Marshal(body)"},
		{name: "optional data field", design: testdata.SSEDataFieldDSL, contains: "data = string(*body.Data)"},
		{name: "viewed data field", design: viewedSSEDataFieldDSL, contains: "data = string(body.Data)"},
		{name: "viewed alias data field", design: viewedSSEPrimitiveAliasDataFieldDSL, contains: "data = string(body.Data)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := expr.RunDSL(t, test.design)
			code := renderedFile(t, linkedHTTPPlanForRoot(t, root).ServerFiles())

			require.Contains(t, code, test.contains)
			require.NotContains(t, code, "var payload any")
			require.NotContains(t, code, "payload.(type)")
			if test.name == "optional data field" {
				require.Contains(t, code, "if body.Data != nil")
				require.NotContains(t, code, "json.Marshal(body.Data)")
			}
		})
	}
}

// TestSSEServerWritesOptionalRetryValue checks that a selected zero retry is
// written while an absent value is omitted.
func TestSSEServerWritesOptionalRetryValue(t *testing.T) {
	root := expr.RunDSL(t, testdata.SSEAllFieldsDSL)
	code := renderedFile(t, linkedHTTPPlanForRoot(t, root).ServerFiles())
	require.Contains(t, code, "if retry != nil {")
	require.Contains(t, code, `fmt.Fprintf(s.w, "retry: %d\n", *retry)`)
}

func TestSSETransportDefaultsToStatusOK(t *testing.T) {
	root := expr.RunDSL(t, testdata.SSEStringDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	fs := plan.ServerFiles()
	require.Len(t, fs, 3)

	sections := fs[1].SectionTemplates
	require.Greater(t, len(sections), 1)
	code := codegen.SectionCode(t, sections[1])
	require.Contains(t, code, "s.w.WriteHeader(http.StatusOK)")
	require.NotContains(t, code, "http.StatusSwitchingProtocols")
}

// ssePrimitiveAliasDSL streams a named string so generated SSE code must use
// its known underlying string representation.
func ssePrimitiveAliasDSL() {
	text := dsl.Type("EventText", dsl.String)
	dsl.Service("SSE Primitive Alias", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(text)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}
