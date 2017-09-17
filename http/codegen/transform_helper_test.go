package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	. "goa.design/goa/http/codegen/testing"
	httpdesign "goa.design/goa/http/design"
)

func TestTransformHelperServer(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Code   string
		Offset int
	}{
		{"body-user-inner-default-1", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCode1, 1},
		{"body-user-inner-default-2", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCode2, 1},
		{"body-user-recursive-default-1", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCode1, 1},
		{"body-user-recursive-default-2", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCode2, 1},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			f := serverEncodeDecode("", httpdesign.Root.HTTPServices[0])
			sections := f.SectionTemplates
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}

func TestTransformHelperCLI(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Code   string
		Offset int
	}{
		{"cli-body-user-inner-default-1", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCodeCLI1, 1},
		{"cli-body-user-inner-default-2", PayloadBodyUserInnerDefaultDSL, PayloadBodyUserInnerDefaultTransformCodeCLI2, 2},
		{"cli-body-user-recursive-default-1", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCodeCLI1, 1},
		{"cli-body-user-recursive-default-2", PayloadBodyInlineRecursiveUserDSL, PayloadBodyInlineRecursiveUserTransformCodeCLI2, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			f := clientEncodeDecode("", httpdesign.Root.HTTPServices[0])
			sections := f.SectionTemplates
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
