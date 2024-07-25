package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestServerTypeFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"server-payload-with-nested-types", testdata.PayloadWithNestedTypesDSL, testdata.PayloadWithNestedTypesServerTypeCode},
		{"server-payload-with-duplicate-use", testdata.PayloadWithMultipleUseTypesDSL, testdata.PayloadWithMultipleUseTypesServerTypeCode},
		{"server-payload-with-alias-type", testdata.PayloadWithAliasTypeDSL, testdata.PayloadWithAliasTypeServerTypeCode},
		{"server-payload-with-mixed-attributes", testdata.PayloadWithMixedAttributesDSL, testdata.PayloadWithMixedAttributesServerTypeCode},
		{"server-payload-with-custom-type-package", testdata.PayloadWithCustomTypePackageDSL, testdata.PayloadWithCustomTypePackageServerTypeCode},
		{"server-result-collection", testdata.ResultWithCollectionDSL, testdata.ResultWithCollectionServerTypeCode},
		{"server-with-errors", testdata.UnaryRPCWithErrorsDSL, testdata.WithErrorsServerTypeCode},
		{"server-elem-validation", testdata.ElemValidationDSL, testdata.ElemValidationServerTypesFile},
		{"server-alias-validation", testdata.AliasValidationDSL, testdata.AliasValidationServerTypesFile},
		{"server-struct-meta-type", testdata.StructMetaTypeDSL, testdata.StructMetaTypeServerTypeCode},
		{"server-struct-field-name-meta-type", testdata.StructFieldNameMetaTypeDSL, testdata.StructFieldNameMetaTypeServerTypesCode},
		{"server-default-fields", testdata.DefaultFieldsDSL, testdata.DefaultFieldsServerTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ServerTypeFiles("", expr.Root)
			require.Len(t, fs, 1)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			assert.Equal(t, c.Code, code)
		})
	}
}
