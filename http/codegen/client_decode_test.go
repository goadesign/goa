package codegen

import (
	"goa.design/goa/v3/codegen/testutil"
	"testing"

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
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ClientFiles("", services)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 2)
			code := codegen.SectionCode(t, sections[2])
			testutil.AssertGo(t, "testdata/golden/client_decode_"+c.Name+".go.golden", code)
		})
	}
}
