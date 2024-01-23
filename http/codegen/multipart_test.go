package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestServerMultipartFuncType(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveDecoderFuncTypeCode},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeDecoderFuncTypeCode},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeDecoderFuncTypeCode},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeDecoderFuncTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			require.Len(t, fs, 2)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), 5)
			code := codegen.SectionCode(t, sections[3])
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestClientMultipartFuncType(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveEncoderFuncTypeCode},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeEncoderFuncTypeCode},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeEncoderFuncTypeCode},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeEncoderFuncTypeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles(genpkg, expr.Root)
			require.Len(t, fs, 2)
			sections := fs[0].SectionTemplates
			require.Greater(t, len(sections), 4)
			code := codegen.SectionCode(t, sections[2])
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestServerMultipartNewFunc(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"server-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveDecoderFuncCode},
		{"server-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeDecoderFuncCode},
		{"server-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeDecoderFuncCode},
		{"server-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeDecoderFuncCode},
		{"server-multipart-with-param", testdata.PayloadMultipartWithParamDSL, testdata.MultipartWithParamDecoderFuncCode},
		{"server-multipart-with-params-and-headers", testdata.PayloadMultipartWithParamsAndHeadersDSL, testdata.MultipartWithParamsAndHeadersDecoderFuncCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 3)
			code := codegen.SectionCode(t, sections[3])
			assert.Equal(t, c.Code, code)
		})
	}
}

func TestClientMultipartNewFunc(t *testing.T) {
	const genpkg = "gen"
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"client-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveEncoderFuncCode},
		{"client-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeEncoderFuncCode},
		{"client-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeEncoderFuncCode},
		{"client-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeEncoderFuncCode},
		{"client-multipart-with-param", testdata.PayloadMultipartWithParamDSL, testdata.MultipartWithParamEncoderFuncCode},
		{"client-multipart-with-params-and-headers", testdata.PayloadMultipartWithParamsAndHeadersDSL, testdata.MultipartWithParamsAndHeadersEncoderFuncCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles(genpkg, expr.Root)
			require.Len(t, fs, 2)
			sections := fs[1].SectionTemplates
			require.Greater(t, len(sections), 3)
			code := codegen.SectionCode(t, sections[3])
			assert.Equal(t, c.Code, code)
		})
	}
}
