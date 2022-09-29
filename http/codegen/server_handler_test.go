package codegen

import (
	"path/filepath"
	"testing"

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
			if len(sections) == 0 {
				t.Fatal("section not found")
			}
			code := codegen.SectionCode(t, sections[0])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
