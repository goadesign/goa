package codegen

import (
	"testing"

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
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].SectionTemplates
			if len(sections) < 3 {
				t.Fatalf("got %d sections, expected at least 3", len(sections))
			}
			code := codegen.SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
