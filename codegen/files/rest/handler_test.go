package rest

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/codegen/files/rest/testing"
	"goa.design/goa.v2/design/rest"
)

func TestHandlerInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no payload no result", ServerNoPayloadNoResult, ServerNoPayloadNoResultHandlerConstructorCode},
		{"payload no result", ServerPayloadNoResult, ServerPayloadNoResultHandlerConstructorCode},
		{"no payload result", ServerNoPayloadResult, ServerNoPayloadResultHandlerConstructorCode},
		{"payload result", ServerPayloadResult, ServerPayloadResultHandlerConstructorCode},
		{"payload result error", ServerPayloadResultError, ServerPayloadResultErrorHandlerConstructorCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			fs := Servers(rest.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].Sections("")
			if len(sections) < 8 {
				t.Fatalf("got %d sections, expected a least 8", len(sections))
			}
			code := SectionCode(t, sections[5])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
