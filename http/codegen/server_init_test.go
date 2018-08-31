package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestServerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, testdata.ServerMultiEndpointsConstructorCode, 3},
		{"multiple bases", testdata.ServerMultiBasesDSL, testdata.ServerMultiBasesConstructorCode, 3},
		{"file server", testdata.ServerFileServerDSL, testdata.ServerFileServerConstructorCode, 3},
		{"mixed", testdata.ServerMixedDSL, testdata.ServerMixedConstructorCode, 3},
		{"multipart", testdata.ServerMultipartDSL, testdata.ServerMultipartConstructorCode, 4},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 6 {
				t.Fatalf("got %d sections, expected at least 6", len(sections))
			}
			code := codegen.SectionCode(t, sections[c.SectionNum])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
