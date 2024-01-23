package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"empty-body", testdata.EmptyServerResponseDSL, testdata.EmptyServerResponseDecodeCode},
		{"body-result-multiple-views", testdata.ResultBodyMultipleViewsDSL, testdata.ResultBodyMultipleViewsDecodeCode},
		{"empty-body-result-multiple-views", testdata.EmptyBodyResultMultipleViewsDSL, testdata.EmptyBodyResultMultipleViewsDecodeCode},
		{"explicit-body-primitive-result", testdata.ExplicitBodyPrimitiveResultMultipleViewsDSL, testdata.ExplicitBodyPrimitiveResultDecodeCode},
		{"explicit-body-result-multiple-views", testdata.ExplicitBodyUserResultMultipleViewsDSL, testdata.ExplicitBodyUserResultMultipleViewsDecodeCode},
		{"explicit-body-result-collection", testdata.ExplicitBodyResultCollectionDSL, testdata.ExplicitBodyResultCollectionDecodeCode},
		{"tag-result-multiple-views", testdata.ResultMultipleViewsTagDSL, testdata.ResultMultipleViewsTagDecodeCode},
		{"empty-server-response-with-tags", testdata.EmptyServerResponseWithTagsDSL, testdata.EmptyServerResponseWithTagsDecodeCode},
		{"header-string-implicit", testdata.ResultHeaderStringImplicitDSL, testdata.ResultHeaderStringImplicitResponseDecodeCode},
		{"header-string-array", testdata.ResultHeaderStringArrayDSL, testdata.ResultHeaderStringArrayResponseDecodeCode},
		{"header-string-array-validate", testdata.ResultHeaderStringArrayValidateDSL, testdata.ResultHeaderStringArrayValidateResponseDecodeCode},
		{"header-array", testdata.ResultHeaderArrayDSL, testdata.ResultHeaderArrayResponseDecodeCode},
		{"header-array-validate", testdata.ResultHeaderArrayValidateDSL, testdata.ResultHeaderArrayValidateResponseDecodeCode},
		{"with-headers-dsl", testdata.WithHeadersBlockDSL, testdata.WithHeadersBlockResponseDecodeCode},
		{"with-headers-dsl-viewed-result", testdata.WithHeadersBlockViewedResultDSL, testdata.WithHeadersBlockViewedResultResponseDecodeCode},
		{"validate-error-response-type", testdata.ValidateErrorResponseTypeDSL, testdata.ValidateErrorResponseTypeDecodeCode},
		{"empty-error-response-body", testdata.EmptyErrorResponseBodyDSL, testdata.EmptyErrorResponseBodyDecodeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles("", expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 2)
			code := codegen.SectionCode(t, sections[2])
			assert.Equal(t, c.Code, code)
		})
	}
}
