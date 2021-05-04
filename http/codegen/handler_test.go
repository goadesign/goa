package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestHandlerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name      string
		DSL       func()
		Code      string
		FileCount int
	}{
		{"no payload no result", testdata.ServerNoPayloadNoResultDSL, testdata.ServerNoPayloadNoResultHandlerConstructorCode, 2},
		{"no payload no result with a redirect", testdata.ServerNoPayloadNoResultWithRedirectDSL, testdata.ServerNoPayloadNoResultWithRedirectHandlerConstructorCode, 1},
		{"payload no result", testdata.ServerPayloadNoResultDSL, testdata.ServerPayloadNoResultHandlerConstructorCode, 2},
		{"payload no result with a redirect", testdata.ServerPayloadNoResultWithRedirectDSL, testdata.ServerPayloadNoResultWithRedirectHandlerConstructorCode, 2},
		{"no payload result", testdata.ServerNoPayloadResultDSL, testdata.ServerNoPayloadResultHandlerConstructorCode, 2},
		{"payload result", testdata.ServerPayloadResultDSL, testdata.ServerPayloadResultHandlerConstructorCode, 2},
		{"payload result error", testdata.ServerPayloadResultErrorDSL, testdata.ServerPayloadResultErrorHandlerConstructorCode, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			if len(fs) != c.FileCount {
				t.Fatalf("got %d files, expected %d", len(fs), c.FileCount)
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 7 {
				t.Fatalf("got %d sections, expected at least 6", len(sections))
			}
			code := codegen.SectionCode(t, sections[8])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
