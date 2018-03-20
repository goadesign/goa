package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/http/codegen/testdata"
	httpdesign "goa.design/goa/http/design"
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
			fs := ServerFiles(genpkg, httpdesign.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 6 {
				t.Fatalf("got %d sections, expected at least 6", len(sections))
			}
			code := codegen.SectionCode(t, sections[3])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
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
			fs := ClientFiles(genpkg, httpdesign.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[0].SectionTemplates
			if len(sections) < 5 {
				t.Fatalf("got %d sections, expected at least 5", len(sections))
			}
			code := codegen.SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
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
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveDecoderFuncCode},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeDecoderFuncCode},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeDecoderFuncCode},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeDecoderFuncCode},
		{"multipart-with-params", testdata.PayloadMultipartWithParams, testdata.MultipartWithParamsDecoderFuncCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles(genpkg, httpdesign.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].SectionTemplates
			if len(sections) < 4 {
				t.Fatalf("got %d sections, expected at least 4", len(sections))
			}
			code := codegen.SectionCode(t, sections[3])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
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
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.MultipartPrimitiveEncoderFuncCode},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.MultipartUserTypeEncoderFuncCode},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.MultipartArrayTypeEncoderFuncCode},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.MultipartMapTypeEncoderFuncCode},
		{"multipart-with-params", testdata.PayloadMultipartWithParams, testdata.MultipartWithParamsEncoderFuncCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles(genpkg, httpdesign.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].SectionTemplates
			if len(sections) < 4 {
				t.Fatalf("got %d sections, expected at least 4", len(sections))
			}
			code := codegen.SectionCode(t, sections[3])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
