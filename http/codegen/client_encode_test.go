package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
)

func TestClientEncode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"query-bool-validate", testdata.PayloadQueryBoolValidateDSL, testdata.PayloadQueryBoolValidateEncodeCode},
		{"query-int-validate", testdata.PayloadQueryIntValidateDSL, testdata.PayloadQueryIntValidateEncodeCode},
		{"query-int32-validate", testdata.PayloadQueryInt32ValidateDSL, testdata.PayloadQueryInt32ValidateEncodeCode},
		{"query-int64-validate", testdata.PayloadQueryInt64ValidateDSL, testdata.PayloadQueryInt64ValidateEncodeCode},
		{"query-uint-validate", testdata.PayloadQueryUIntValidateDSL, testdata.PayloadQueryUIntValidateEncodeCode},
		{"query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL, testdata.PayloadQueryUInt32ValidateEncodeCode},
		{"query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL, testdata.PayloadQueryUInt64ValidateEncodeCode},
		{"query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL, testdata.PayloadQueryFloat32ValidateEncodeCode},
		{"query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL, testdata.PayloadQueryFloat64ValidateEncodeCode},
		{"query-string-validate", testdata.PayloadQueryStringValidateDSL, testdata.PayloadQueryStringValidateEncodeCode},
		{"query-bytes-validate", testdata.PayloadQueryBytesValidateDSL, testdata.PayloadQueryBytesValidateEncodeCode},
		{"query-any-validate", testdata.PayloadQueryAnyValidateDSL, testdata.PayloadQueryAnyValidateEncodeCode},
		{"query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL, testdata.PayloadQueryArrayBoolValidateEncodeCode},
		{"query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL, testdata.PayloadQueryArrayIntValidateEncodeCode},
		{"query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL, testdata.PayloadQueryArrayInt32ValidateEncodeCode},
		{"query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL, testdata.PayloadQueryArrayInt64ValidateEncodeCode},
		{"query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL, testdata.PayloadQueryArrayUIntValidateEncodeCode},
		{"query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL, testdata.PayloadQueryArrayUInt32ValidateEncodeCode},
		{"query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL, testdata.PayloadQueryArrayUInt64ValidateEncodeCode},
		{"query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL, testdata.PayloadQueryArrayFloat32ValidateEncodeCode},
		{"query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL, testdata.PayloadQueryArrayFloat64ValidateEncodeCode},
		{"query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL, testdata.PayloadQueryArrayStringValidateEncodeCode},
		{"query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL, testdata.PayloadQueryArrayBytesValidateEncodeCode},
		{"query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL, testdata.PayloadQueryArrayAnyValidateEncodeCode},
		{"query-map-string-string-validate", testdata.PayloadQueryMapStringStringValidateDSL, testdata.PayloadQueryMapStringStringValidateEncodeCode},
		{"query-map-string-bool-validate", testdata.PayloadQueryMapStringBoolValidateDSL, testdata.PayloadQueryMapStringBoolValidateEncodeCode},
		{"query-map-bool-string-validate", testdata.PayloadQueryMapBoolStringValidateDSL, testdata.PayloadQueryMapBoolStringValidateEncodeCode},
		{"query-map-bool-bool-validate", testdata.PayloadQueryMapBoolBoolValidateDSL, testdata.PayloadQueryMapBoolBoolValidateEncodeCode},
		{"query-map-string-array-string-validate", testdata.PayloadQueryMapStringArrayStringValidateDSL, testdata.PayloadQueryMapStringArrayStringValidateEncodeCode},
		{"query-map-string-array-bool-validate", testdata.PayloadQueryMapStringArrayBoolValidateDSL, testdata.PayloadQueryMapStringArrayBoolValidateEncodeCode},
		{"query-map-bool-array-string-validate", testdata.PayloadQueryMapBoolArrayStringValidateDSL, testdata.PayloadQueryMapBoolArrayStringValidateEncodeCode},
		{"query-map-bool-array-bool-validate", testdata.PayloadQueryMapBoolArrayBoolValidateDSL, testdata.PayloadQueryMapBoolArrayBoolValidateEncodeCode},

		{"query-primitive-string-validate", testdata.PayloadQueryPrimitiveStringValidateDSL, testdata.PayloadQueryPrimitiveStringValidateEncodeCode},
		{"query-primitive-bool-validate", testdata.PayloadQueryPrimitiveBoolValidateDSL, testdata.PayloadQueryPrimitiveBoolValidateEncodeCode},
		{"query-primitive-array-string-validate", testdata.PayloadQueryPrimitiveArrayStringValidateDSL, testdata.PayloadQueryPrimitiveArrayStringValidateEncodeCode},
		{"query-primitive-array-bool-validate", testdata.PayloadQueryPrimitiveArrayBoolValidateDSL, testdata.PayloadQueryPrimitiveArrayBoolValidateEncodeCode},
		{"query-primitive-map-string-array-string-validate", testdata.PayloadQueryPrimitiveMapStringArrayStringValidateDSL, testdata.PayloadQueryPrimitiveMapStringArrayStringValidateEncodeCode},
		{"query-primitive-map-string-bool-validate", testdata.PayloadQueryPrimitiveMapStringBoolValidateDSL, testdata.PayloadQueryPrimitiveMapStringBoolValidateEncodeCode},
		{"query-primitive-map-bool-array-bool-validate", testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateDSL, testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateEncodeCode},
		{"query-map-string-map-int-string-validate", testdata.PayloadQueryMapStringMapIntStringValidateDSL, testdata.PayloadQueryMapStringMapIntStringValidateEncodeCode},
		{"query-map-int-map-string-array-int-validate", testdata.PayloadQueryMapIntMapStringArrayIntValidateDSL, testdata.PayloadQueryMapIntMapStringArrayIntValidateEncodeCode},

		{"query-string-mapped", testdata.PayloadQueryStringMappedDSL, testdata.PayloadQueryStringMappedEncodeCode},

		{"query-string-default", testdata.PayloadQueryStringDefaultDSL, testdata.PayloadQueryStringDefaultEncodeCode},
		{"query-primitive-string-default", testdata.PayloadQueryPrimitiveStringDefaultDSL, testdata.PayloadQueryPrimitiveStringDefaultEncodeCode},

		{"header-string-validate", testdata.PayloadHeaderStringValidateDSL, testdata.PayloadHeaderStringValidateEncodeCode},
		{"header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL, testdata.PayloadHeaderArrayStringValidateEncodeCode},

		{"header-primitive-string-validate", testdata.PayloadHeaderPrimitiveStringValidateDSL, testdata.PayloadHeaderPrimitiveStringValidateEncodeCode},
		{"header-primitive-bool-validate", testdata.PayloadHeaderPrimitiveBoolValidateDSL, testdata.PayloadHeaderPrimitiveBoolValidateEncodeCode},
		{"header-primitive-array-string-validate", testdata.PayloadHeaderPrimitiveArrayStringValidateDSL, testdata.PayloadHeaderPrimitiveArrayStringValidateEncodeCode},
		{"header-primitive-array-bool-validate", testdata.PayloadHeaderPrimitiveArrayBoolValidateDSL, testdata.PayloadHeaderPrimitiveArrayBoolValidateEncodeCode},

		{"header-string-default", testdata.PayloadHeaderStringDefaultDSL, testdata.PayloadHeaderStringDefaultEncodeCode},
		{"header-primitive-string-default", testdata.PayloadHeaderPrimitiveStringDefaultDSL, testdata.PayloadHeaderPrimitiveStringDefaultEncodeCode},

		{"body-string-validate", testdata.PayloadBodyStringValidateDSL, testdata.PayloadBodyStringValidateEncodeCode},
		{"body-user-validate", testdata.PayloadBodyUserValidateDSL, testdata.PayloadBodyUserValidateEncodeCode},
		{"body-array-string-validate", testdata.PayloadBodyArrayStringValidateDSL, testdata.PayloadBodyArrayStringValidateEncodeCode},
		{"body-array-user-validate", testdata.PayloadBodyArrayUserValidateDSL, testdata.PayloadBodyArrayUserValidateEncodeCode},
		{"body-map-string-validate", testdata.PayloadBodyMapStringValidateDSL, testdata.PayloadBodyMapStringValidateEncodeCode},
		{"body-map-user-validate", testdata.PayloadBodyMapUserValidateDSL, testdata.PayloadBodyMapUserValidateEncodeCode},

		{"body-primitive-string-validate", testdata.PayloadBodyPrimitiveStringValidateDSL, testdata.PayloadBodyPrimitiveStringValidateEncodeCode},
		{"body-primitive-bool-validate", testdata.PayloadBodyPrimitiveBoolValidateDSL, testdata.PayloadBodyPrimitiveBoolValidateEncodeCode},
		{"body-primitive-array-string-validate", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, testdata.PayloadBodyPrimitiveArrayStringValidateEncodeCode},
		{"body-primitive-array-bool-validate", testdata.PayloadBodyPrimitiveArrayBoolValidateDSL, testdata.PayloadBodyPrimitiveArrayBoolValidateEncodeCode},

		{"body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL, testdata.PayloadBodyPrimitiveArrayUserValidateEncodeCode},
		{"body-primitive-field-array-user-validate", testdata.PayloadBodyPrimitiveFieldArrayUserValidateDSL, testdata.PayloadBodyPrimitiveFieldArrayUserValidateEncodeCode},

		{"body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL, testdata.PayloadBodyQueryObjectValidateEncodeCode},
		{"body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL, testdata.PayloadBodyQueryUserValidateEncodeCode},

		{"body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL, testdata.PayloadBodyPathObjectValidateEncodeCode},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, testdata.PayloadBodyPathUserValidateEncodeCode},

		{"body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL, testdata.PayloadBodyQueryPathObjectValidateEncodeCode},
		{"body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL, testdata.PayloadBodyQueryPathUserValidateEncodeCode},

		{"map-query-primitive-primitive", testdata.PayloadMapQueryPrimitivePrimitiveDSL, testdata.PayloadMapQueryPrimitivePrimitiveEncodeCode},
		{"map-query-primitive-array", testdata.PayloadMapQueryPrimitiveArrayDSL, testdata.PayloadMapQueryPrimitiveArrayEncodeCode},
		{"map-query-object", testdata.PayloadMapQueryObjectDSL, testdata.PayloadMapQueryObjectEncodeCode},
		{"multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.PayloadMultipartBodyPrimitiveEncodeCode},
		{"multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.PayloadMultipartBodyUserTypeEncodeCode},
		{"multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.PayloadMultipartBodyArrayTypeEncodeCode},
		{"multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.PayloadMultipartBodyMapTypeEncodeCode},
	}
	golden := makeGolden(t, "testdata/payload_encode_functions.go")
	if golden != nil {
		golden.WriteString("package testdata\n")
		defer golden.Close()
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

			if golden != nil {
				name := codegen.Goify(c.Name, true)
				name = strings.Replace(name, "Uint", "UInt", -1)
				code = "\nvar Payload" + name + "EncodeCode = `" + code + "`"
				golden.WriteString(code + "\n")
			} else if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
