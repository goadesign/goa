package codegen

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/http/codegen/testing"
	httpdesign "goa.design/goa.v2/http/design"
)

func TestTransformHelper(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Code   string
		Offset int
	}{
		{"body-user-inner-default-1", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCode1, 1},
		{"body-user-inner-default-2", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCode2, 2},
		{"body-user-recursive-default-1", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCode1, 1},
		{"body-user-recursive-default-2", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCode2, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			f := clientEncodeDecode(httpdesign.Root.HTTPServices[0])
			sections := f.Sections("")
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
