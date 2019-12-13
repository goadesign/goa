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
		SectionNum int
	}{
		{
			Name:       "multiple files constructor",
			DSL:        testdata.ServerMultipleFilesDSL,
			Code:       testdata.ServerMultipleFilesConstructorCode,
			SectionNum: 6,
		},
		{
			Name:       "multiple files mounter",
			DSL:        testdata.ServerMultipleFilesDSL,
			Code:       testdata.ServerMultipleFilesMounterCode,
			SectionNum: 9,
		},
		{
			Name:       "multiple files constructor /w prefix path",
			DSL:        testdata.ServerMultipleFilesWithPrefixPathDSL,
			Code:       testdata.ServerMultipleFilesWithPrefixPathConstructorCode,
			SectionNum: 6,
		},
		{
			Name:       "multiple files mounter /w prefix path",
			DSL:        testdata.ServerMultipleFilesWithPrefixPathDSL,
			Code:       testdata.ServerMultipleFilesWithPrefixPathMounterCode,
			SectionNum: 9,
		},
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
