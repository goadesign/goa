package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerHandler(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"server simple routing", testdata.ServerSimpleRoutingDSL, testdata.ServerSimpleRoutingCode},
		{"server trailing slash routing", testdata.ServerTrailingSlashRoutingDSL, testdata.ServerTrailingSlashRoutingCode},
		{"server simple routing with a redirect", testdata.ServerSimpleRoutingWithRedirectDSL, testdata.ServerSimpleRoutingCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			sections := codegentest.Sections(fs, filepath.Join("", "server.go"), "server-handler")
			require.Greater(t, len(sections), 0)
			code := codegen.SectionCode(t, sections[0])
			assert.Equal(t, c.Code, code)
		})
	}
}
