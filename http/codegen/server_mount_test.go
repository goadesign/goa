package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerMount(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"simple routing constructor", testdata.ServerSimpleRoutingDSL, testdata.ServerSimpleRoutingConstructorCode, 2, 6},
		{"simple routing with a redirect constructor", testdata.ServerSimpleRoutingWithRedirectDSL, testdata.ServerSimpleRoutingConstructorCode, 1, 6},
		{"multiple files constructor", testdata.ServerMultipleFilesDSL, testdata.ServerMultipleFilesConstructorCode, 1, 6},
		{"multiple files mounter", testdata.ServerMultipleFilesDSL, testdata.ServerMultipleFilesMounterCode, 1, 10},
		{"multiple files constructor /w prefix path", testdata.ServerMultipleFilesWithPrefixPathDSL, testdata.ServerMultipleFilesWithPrefixPathConstructorCode, 1, 6},
		{"multiple files mounter /w prefix path", testdata.ServerMultipleFilesWithPrefixPathDSL, testdata.ServerMultipleFilesWithPrefixPathMounterCode, 1, 10},
		{"multiple files with a redirect constructor", testdata.ServerMultipleFilesWithRedirectDSL, testdata.ServerMultipleFilesWithRedirectConstructorCode, 1, 6},
		{"multiple files with a redirect mounter", testdata.ServerMultipleFilesWithRedirectDSL, testdata.ServerMultipleFilesMounterCode, 1, 10},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			if len(fs) != c.FileCount {
				t.Fatalf("got %d files, expected %d", len(fs), c.FileCount)
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
