package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestEncodeError(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"primitive-error-response", testdata.PrimitiveErrorResponseDSL, testdata.PrimitiveErrorResponseEncoderCode},
		{"primitive-error-in-response-header", testdata.PrimitiveErrorInResponseHeaderDSL, testdata.PrimitiveErrorInResponseHeaderEncoderCode},
		{"api-primitive-error-response", testdata.APIPrimitiveErrorResponseDSL, testdata.APIPrimitiveErrorResponseEncoderCode},
		{"default-error-response", testdata.DefaultErrorResponseDSL, testdata.DefaultErrorResponseEncoderCode},
		{"default-error-response-with-content-type", testdata.DefaultErrorResponseWithContentTypeDSL, testdata.DefaultErrorResponseWithContentTypeEncoderCode},
		{"service-error-response", testdata.ServiceErrorResponseDSL, testdata.ServiceErrorResponseEncoderCode},
		{"api-error-response", testdata.APIErrorResponseDSL, testdata.ServiceErrorResponseEncoderCode},
		{"api-error-response-with-content-type", testdata.APIErrorResponseWithContentTypeDSL, testdata.ServiceErrorResponseWithContentTypeEncoderCode},
		{"no-body-error-response", testdata.NoBodyErrorResponseDSL, testdata.NoBodyErrorResponseEncoderCode},
		{"no-body-error-response-with-content-type", testdata.NoBodyErrorResponseWithContentTypeDSL, testdata.NoBodyErrorResponseWithContentTypeEncoderCode},
		{"api-no-body-error-response", testdata.APINoBodyErrorResponseDSL, testdata.NoBodyErrorResponseEncoderCode},
		{"api-no-body-error-response-with-content-type", testdata.APINoBodyErrorResponseWithContentTypeDSL, testdata.NoBodyErrorResponseWithContentTypeEncoderCode},
		{"empty-error-response-body", testdata.EmptyErrorResponseBodyDSL, testdata.EmptyErrorResponseBodyEncoderCode},
		{"empty-custom-error-response-body", testdata.EmptyCustomErrorResponseBodyDSL, testdata.EmptyCustomErrorResponseBodyEncoderCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[2])
			assert.Equal(t, c.Code, code)
		})
	}
}
