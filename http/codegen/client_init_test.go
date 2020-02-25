package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientInit(t *testing.T) {
	cases := []struct {
		Name       string
		DSL        func()
		Code       string
		FileCount  int
		SectionNum int
	}{
		{"multiple endpoints", testdata.ServerMultiEndpointsDSL, testdata.MultipleEndpointsClientInitCode, 2, 2},
		{"streaming", testdata.StreamingResultDSL, testdata.StreamingClientInitCode, 3, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			if len(fs) != c.FileCount {
				t.Fatalf("got %d files, expected %v", len(fs), c.FileCount)
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 3 {
				t.Fatalf("got %d sections, expected at least 3", len(sections))
			}
			code := codegen.SectionCode(t, sections[c.SectionNum])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
