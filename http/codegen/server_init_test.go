package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, testdata.ServerMultiEndpointsConstructorCode, 2, 3},
		{"multiple bases", testdata.ServerMultiBasesDSL, testdata.ServerMultiBasesConstructorCode, 2, 3},
		{"file server", testdata.ServerFileServerDSL, testdata.ServerFileServerConstructorCode, 1, 3},
		{"file server with a redirect", testdata.ServerFileServerWithRedirectDSL, testdata.ServerFileServerConstructorCode, 1, 3},
		{"mixed", testdata.ServerMixedDSL, testdata.ServerMixedConstructorCode, 2, 3},
		{"multipart", testdata.ServerMultipartDSL, testdata.ServerMultipartConstructorCode, 2, 4},
		{"streaming", testdata.StreamingResultDSL, testdata.ServerStreamingConstructorCode, 3, 3},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			if len(fs) != c.FileCount {
				t.Fatalf("got %d files, expected %v", len(fs), c.FileCount)
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
