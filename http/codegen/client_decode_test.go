package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"empty-body", testdata.EmptyServerResponseDSL},
		{"body-result-multiple-views", testdata.ResultBodyMultipleViewsDSL},
		{"empty-body-result-multiple-views", testdata.EmptyBodyResultMultipleViewsDSL},
		{"explicit-body-primitive-result", testdata.ExplicitBodyPrimitiveResultMultipleViewsDSL},
		{"explicit-body-result-multiple-views", testdata.ExplicitBodyUserResultMultipleViewsDSL},
		{"explicit-body-result-collection", testdata.ExplicitBodyResultCollectionDSL},
		{"tag-result-multiple-views", testdata.ResultMultipleViewsTagDSL},
		{"status-tag-required", testdata.ResultStatusTagRequiredDSL},
		{"empty-server-response-with-tags", testdata.EmptyServerResponseWithTagsDSL},
		{"header-string-implicit", testdata.ResultHeaderStringImplicitDSL},
		{"header-string-array", testdata.ResultHeaderStringArrayDSL},
		{"header-string-array-validate", testdata.ResultHeaderStringArrayValidateDSL},
		{"header-array", testdata.ResultHeaderArrayDSL},
		{"header-array-validate", testdata.ResultHeaderArrayValidateDSL},
		{"with-headers-dsl", testdata.WithHeadersBlockDSL},
		{"with-headers-dsl-viewed-result", testdata.WithHeadersBlockViewedResultDSL},
		{"validate-error-response-type", testdata.ValidateErrorResponseTypeDSL},
		{"empty-error-response-body", testdata.EmptyErrorResponseBodyDSL},
		{"required-primitive-arrays", testdata.RequiredPrimitiveArrayDSL},
		{"skip-response-body-encode-decode", testdata.ServerSkipResponseBodyEncodeDecodeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			var section *codegen.SectionTemplate
			for _, s := range sections {
				if s.Name == "response-decoder" {
					section = s
				}
			}
			require.NotNil(t, section)
			code := codegen.SectionCode(t, section)
			testutil.AssertGo(t, "testdata/golden/client_decode_"+c.Name+".go.golden", code)
		})
	}
}
