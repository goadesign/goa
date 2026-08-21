package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	// "goa.design/goa/v3/http/codegen/testdata"
)

func TestTransformHelperServer(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Offset int
	}{
		// {"body-user-inner-default-1", testdata.PayloadBodyUserInnerDefaultDSL1, 1},
		// {"body-user-recursive-default-1", testdata.PayloadBodyInlineRecursiveUserDSL1, 1},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			f := ServerEncodeDecodeFile(root.API.HTTP.Services[0], services)
			sections := f.SectionTemplates
			require.Greater(t, len(sections), c.Offset)
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			testutil.AssertGo(t, "testdata/golden/transform_helper_"+c.Name+".go.golden", code)
		})
	}
}

func TestTransformHelperCLI(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Offset int
	}{
		// {"cli-body-user-inner-default-1", testdata.PayloadBodyUserInnerDefaultDSLCLI1, 1},
		// {"cli-body-user-inner-default-2", testdata.PayloadBodyUserInnerDefaultDSLCLI2, 2},
		// {"cli-body-user-recursive-default-1", testdata.PayloadBodyInlineRecursiveUserDSLCLI1, 1},
		// {"cli-body-user-recursive-default-2", testdata.PayloadBodyInlineRecursiveUserDSLCLI2, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			f := ClientEncodeDecodeFile(root.API.HTTP.Services[0], services)
			sections := f.SectionTemplates
			require.Greater(t, len(sections), c.Offset)
			code := codegen.SectionCode(t, sections[len(sections)-c.Offset])
			testutil.AssertGo(t, "testdata/golden/transform_helper_"+c.Name+".go.golden", code)
		})
	}
}
