package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestTransformHelperServer(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Code   string
		Offset int
	}{
		{"body-user-inner-default-1", testdata.PayloadBodyUserInnerDefaultDSL, testdata.PayloadBodyUserInnerDefaultTransformCode1, 1},
		{"body-user-recursive-default-1", testdata.PayloadBodyInlineRecursiveUserDSL, testdata.PayloadBodyInlineRecursiveUserTransformCode1, 1},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			f := serverEncodeDecodeFile("", expr.Root.API.HTTP.Services[0])
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
		{"cli-body-user-inner-default-1", testdata.PayloadBodyUserInnerDefaultDSL, testdata.PayloadBodyUserInnerDefaultTransformCodeCLI1, 1},
		{"cli-body-user-inner-default-2", testdata.PayloadBodyUserInnerDefaultDSL, testdata.PayloadBodyUserInnerDefaultTransformCodeCLI2, 2},
		{"cli-body-user-recursive-default-1", testdata.PayloadBodyInlineRecursiveUserDSL, testdata.PayloadBodyInlineRecursiveUserTransformCodeCLI1, 1},
		{"cli-body-user-recursive-default-2", testdata.PayloadBodyInlineRecursiveUserDSL, testdata.PayloadBodyInlineRecursiveUserTransformCodeCLI2, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			f := clientEncodeDecodeFile("", expr.Root.API.HTTP.Services[0])
			sections := f.SectionTemplates
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
