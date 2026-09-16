package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestEncodeError(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"primitive-error-response", testdata.PrimitiveErrorResponseDSL},
		{"primitive-error-in-response-header", testdata.PrimitiveErrorInResponseHeaderDSL},
		{"api-primitive-error-response", testdata.APIPrimitiveErrorResponseDSL},
		{"default-error-response", testdata.DefaultErrorResponseDSL},
		{"default-error-response-with-content-type", testdata.DefaultErrorResponseWithContentTypeDSL},
		{"service-error-response", testdata.ServiceErrorResponseDSL},
		{"api-error-response", testdata.APIErrorResponseDSL},
		{"api-error-response-with-content-type", testdata.APIErrorResponseWithContentTypeDSL},
		{"no-body-error-response", testdata.NoBodyErrorResponseDSL},
		{"no-body-error-response-with-content-type", testdata.NoBodyErrorResponseWithContentTypeDSL},
		{"api-no-body-error-response", testdata.APINoBodyErrorResponseDSL},
		{"api-no-body-error-response-with-content-type", testdata.APINoBodyErrorResponseWithContentTypeDSL},
		{"empty-error-response-body", testdata.EmptyErrorResponseBodyDSL},
		{"empty-custom-error-response-body", testdata.EmptyCustomErrorResponseBodyDSL},
		{"error-body-attribute-response", testdata.ErrorBodyAttributeResponseDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ServerFiles()
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[2])
			testutil.AssertGo(t, "testdata/golden/server_error_encoder_"+c.Name+".go.golden", code)
		})
	}
}
