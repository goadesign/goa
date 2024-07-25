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

func TestClientTypeFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"client-payload-with-nested-types", testdata.PayloadWithNestedTypesDSL, testdata.PayloadWithNestedTypesClientTypeCode},
		{"client-payload-with-duplicate-use", testdata.PayloadWithMultipleUseTypesDSL, testdata.PayloadWithMultipleUseTypesClientTypeCode},
		{"client-payload-with-alias-type", testdata.PayloadWithAliasTypeDSL, testdata.PayloadWithAliasTypeClientTypeCode},
		{"client-result-collection", testdata.ResultWithCollectionDSL, testdata.ResultWithCollectionClientTypeCode},
		{"client-alias-validation", testdata.ResultWithAliasValidation, testdata.ResultWithAliasValidationClientTypeCode},
		{"client-with-errors", testdata.UnaryRPCWithErrorsDSL, testdata.WithErrorsClientTypeCode},
		{"client-bidirectional-streaming-same-type", testdata.BidirectionalStreamingRPCSameTypeDSL, testdata.BidirectionalStreamingRPCSameTypeClientTypeCode},
		{"client-struct-meta-type", testdata.StructMetaTypeDSL, testdata.StructMetaTypeTypeCode},
		{"client-struct-field-name-meta-type", testdata.StructFieldNameMetaTypeDSL, testdata.StructFieldNameMetaTypeClientTypesCode},
		{"client-default-fields", testdata.DefaultFieldsDSL, testdata.DefaultFieldsTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunGRPCDSL(t, c.DSL)
			fs := ClientTypeFiles("", expr.Root)
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
