package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerHandler(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"server simple routing", testdata.ServerSimpleRoutingDSL, testdata.ServerSimpleRoutingCode, 2, 7},
		{"server trailing slash routing", testdata.ServerTrailingSlashRoutingDSL, testdata.ServerTrailingSlashRoutingCode, 2, 7},
		{"server simple routing with a redirect", testdata.ServerSimpleRoutingWithRedirectDSL, testdata.ServerSimpleRoutingCode, 1, 7},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			if len(fs) != c.FileCount {
				t.Fatalf("got %d files, expected %d", len(fs), c.FileCount)
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 8 {
				t.Fatalf("got %d sections, expected at least 8", len(sections))
			}
			code := codegen.SectionCode(t, sections[c.SectionNum])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
