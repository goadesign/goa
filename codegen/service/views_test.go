package service

import (
	"bytes"
	"go/format"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
)

func TestViews(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"result-with-multiple-views", testdata.ResultWithMultipleViewsDSL, testdata.ResultWithMultipleViewsCode},
		{"result-collection-multiple-views", testdata.ResultCollectionMultipleViewsDSL, testdata.ResultCollectionMultipleViewsCode},
		{"result-with-user-type", testdata.ResultWithUserTypeDSL, testdata.ResultWithUserTypeCode},
		{"result-with-result-type", testdata.ResultWithResultTypeDSL, testdata.ResultWithResultTypeCode},
		{"result-with-recursive-result-type", testdata.ResultWithRecursiveResultTypeDSL, testdata.ResultWithRecursiveResultTypeCode},
		{"result-type-with-custom-fields", testdata.ResultWithCustomFieldsDSL, testdata.ResultWithCustomFieldsCode},
		{"result-with-recursive-collection-of-result-type", testdata.ResultWithRecursiveCollectionOfResultTypeDSL, testdata.ResultWithRecursiveCollectionOfResultTypeCode},
		{"result-with-multiple-methods", testdata.ResultWithMultipleMethodsDSL, testdata.ResultWithMultipleMethodsCode},
		{"result-with-enum-type", testdata.ResultWithEnumTypeDSL, testdata.ResultWithEnumType},
		{"result-with-pkg-path", testdata.ResultWithPkgPathDSL, testdata.ResultWithPkgPathCode},
		{"result-with-oneof-in-result-type", testdata.ResultWithOneOfInResultTypeDSL, testdata.ResultWithOneOfInResultTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			services := NewServicesData(root)
			require.Len(t, root.Services, 1)
			fs := ViewsFile("goa.design/goa/example", root.Services[0], services)
			require.NotNil(t, fs)
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := string(bs)
			code = strings.ReplaceAll(code, "\r\n", "\n")
			assert.Equal(t, c.Code, code)
		})
	}
}
