package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientInit(t *testing.T) {
	cases := []struct {
		Name       string
		DSL        func()
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, 2, 2},
		{"streaming", testdata.StreamingResultDSL, 3, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, c.FileCount)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), c.SectionNum)
			code := codegen.SectionCode(t, sections[c.SectionNum])
			testutil.AssertGo(t, "testdata/golden/client_init_"+c.Name+".go.golden", code)
		})
	}
}
