package codegen

import (
	"testing"

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
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].SectionTemplates
			if len(sections) < 2 {
				t.Fatalf("got %d sections, expected at least 2", len(sections))
			}
			code := codegen.SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
