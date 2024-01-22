package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			require.Greater(t, len(sections), c.Offset)
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			assert.Equal(t, c.Code, code)
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
			require.Greater(t, len(sections), c.Offset)
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			assert.Equal(t, c.Code, code)
		})
	}
}
