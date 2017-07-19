package rest

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/codegen/rest/testing"
	"goa.design/goa.v2/design/rest"
)

func TestHandlerInit(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"no payload no result", ServerNoPayloadNoResultDSL, ServerNoPayloadNoResultHandlerConstructorCode},
		{"payload no result", ServerPayloadNoResultDSL, ServerPayloadNoResultHandlerConstructorCode},
		{"no payload result", ServerNoPayloadResultDSL, ServerNoPayloadResultHandlerConstructorCode},
		{"payload result", ServerPayloadResultDSL, ServerPayloadResultHandlerConstructorCode},
		{"payload result error", ServerPayloadResultErrorDSL, ServerPayloadResultErrorHandlerConstructorCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			fs := Servers(rest.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].Sections("")
			if len(sections) < 6 {
				t.Fatalf("got %d sections, expected a least 6", len(sections))
			}
			code := SectionCode(t, sections[5])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
