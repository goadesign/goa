package codegen

import (
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestHandlerInit(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no payload no result", testdata.ServerNoPayloadNoResultDSL, testdata.ServerNoPayloadNoResultHandlerConstructorCode},
		{"no payload no result with a redirect", testdata.ServerNoPayloadNoResultWithRedirectDSL, testdata.ServerNoPayloadNoResultWithRedirectHandlerConstructorCode},
		{"payload no result", testdata.ServerPayloadNoResultDSL, testdata.ServerPayloadNoResultHandlerConstructorCode},
		{"payload no result with a redirect", testdata.ServerPayloadNoResultWithRedirectDSL, testdata.ServerPayloadNoResultWithRedirectHandlerConstructorCode},
		{"no payload result", testdata.ServerNoPayloadResultDSL, testdata.ServerNoPayloadResultHandlerConstructorCode},
		{"payload result", testdata.ServerPayloadResultDSL, testdata.ServerPayloadResultHandlerConstructorCode},
		{"payload result error", testdata.ServerPayloadResultErrorDSL, testdata.ServerPayloadResultErrorHandlerConstructorCode},
		{"skip response body encode decode", testdata.ServerSkipResponseBodyEncodeDecodeDSL, testdata.ServerSkipResponseBodyEncodeDecodeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			sections := codegentest.Sections(fs, filepath.Join("", "server.go"), "server-handler-init")
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
