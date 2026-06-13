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
			services := CreateHTTPServices(root)
			fs := ServerFiles("", services)
			require.Len(t, fs, 3)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[1])
			golden := filepath.Join("testdata", "golden", "sse-"+c.Name+".golden")
			testutil.CompareOrUpdateGolden(t, code, golden)
		})
	}
}

func TestSSETransportDefaultsToStatusOK(t *testing.T) {
	root := expr.RunDSL(t, testdata.SSEStringDSL)
	services := CreateHTTPServices(root)
	fs := ServerFiles("", services)
	require.Len(t, fs, 3)

	sections := fs[1].SectionTemplates
	require.Greater(t, len(sections), 1)
	code := codegen.SectionCode(t, sections[1])
	require.Contains(t, code, "s.w.WriteHeader(http.StatusOK)")
	require.NotContains(t, code, "http.StatusSwitchingProtocols")
}
