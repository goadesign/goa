package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestHandlerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no payload no result", testdata.ServerNoPayloadNoResultDSL},
		{"no payload no result with a redirect", testdata.ServerNoPayloadNoResultWithRedirectDSL},
		{"payload no result", testdata.ServerPayloadNoResultDSL},
		{"payload no result with a redirect", testdata.ServerPayloadNoResultWithRedirectDSL},
		{"no payload result", testdata.ServerNoPayloadResultDSL},
		{"payload result", testdata.ServerPayloadResultDSL},
		{"payload result error", testdata.ServerPayloadResultErrorDSL},
		{"skip response body encode decode", testdata.ServerSkipResponseBodyEncodeDecodeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ServerFiles()
			sections := codegentest.Sections(fs, "server.go", "server-handler-init")
			require.Greater(t, len(sections), 0)
			code := codegen.SectionCode(t, sections[0])
			testutil.AssertGo(t, "testdata/golden/handler_"+c.Name+".go.golden", code)
		})
	}
}
