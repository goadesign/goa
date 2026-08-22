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

func TestServerHandler(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"server simple routing", testdata.ServerSimpleRoutingDSL},
		{"server trailing slash routing", testdata.ServerTrailingSlashRoutingDSL},
		{"server simple routing with a redirect", testdata.ServerSimpleRoutingWithRedirectDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ServerFiles()
			sections := codegentest.Sections(fs, "server.go", "server-handler")
			require.Greater(t, len(sections), 0)
			code := codegen.SectionCode(t, sections[0])
			testutil.AssertGo(t, "testdata/golden/server_handler_"+c.Name+".go.golden", code)
		})
	}
}
