package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/http/codegen/testdata"
	httpdesign "goa.design/goa/http/design"
)

func TestClientEncode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"query-bool", testdata.PayloadQueryBoolDSL, testdata.PayloadQueryBoolEncodeCode},
		{"query-bool-validate", testdata.PayloadQueryBoolValidateDSL, testdata.PayloadQueryBoolValidateEncodeCode},
		{"query-int", testdata.PayloadQueryIntDSL, testdata.PayloadQueryIntEncodeCode},
		{"query-int-validate", testdata.PayloadQueryIntValidateDSL, testdata.PayloadQueryIntValidateEncodeCode},
		{"query-int32", testdata.PayloadQueryInt32DSL, testdata.PayloadQueryInt32EncodeCode},
		{"query-int32-validate", testdata.PayloadQueryInt32ValidateDSL, testdata.PayloadQueryInt32ValidateEncodeCode},
		{"query-int64", testdata.PayloadQueryInt64DSL, testdata.PayloadQueryInt64EncodeCode},
		{"query-int64-validate", testdata.PayloadQueryInt64ValidateDSL, testdata.PayloadQueryInt64ValidateEncodeCode},
		{"query-uint", testdata.PayloadQueryUIntDSL, testdata.PayloadQueryUIntEncodeCode},
		{"query-uint-validate", testdata.PayloadQueryUIntValidateDSL, testdata.PayloadQueryUIntValidateEncodeCode},
		{"query-uint32", testdata.PayloadQueryUInt32DSL, testdata.PayloadQueryUInt32EncodeCode},
		{"query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL, testdata.PayloadQueryUInt32ValidateEncodeCode},
		{"query-uint64", testdata.PayloadQueryUInt64DSL, testdata.PayloadQueryUInt64EncodeCode},
		{"query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL, testdata.PayloadQueryUInt64ValidateEncodeCode},
		{"query-float32", testdata.PayloadQueryFloat32DSL, testdata.PayloadQueryFloat32EncodeCode},
		{"query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL, testdata.PayloadQueryFloat32ValidateEncodeCode},
		{"query-float64", testdata.PayloadQueryFloat64DSL, testdata.PayloadQueryFloat64EncodeCode},
		{"query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL, testdata.PayloadQueryFloat64ValidateEncodeCode},
		{"query-string", testdata.PayloadQueryStringDSL, testdata.PayloadQueryStringEncodeCode},
		{"query-string-validate", testdata.PayloadQueryStringValidateDSL, testdata.PayloadQueryStringValidateEncodeCode},
		{"query-bytes", testdata.PayloadQueryBytesDSL, testdata.PayloadQueryBytesEncodeCode},
		{"query-bytes-validate", testdata.PayloadQueryBytesValidateDSL, testdata.PayloadQueryBytesValidateEncodeCode},
		{"query-any", testdata.PayloadQueryAnyDSL, testdata.PayloadQueryAnyEncodeCode},
		{"query-any-validate", testdata.PayloadQueryAnyValidateDSL, testdata.PayloadQueryAnyValidateEncodeCode},
		{"query-array-bool", testdata.PayloadQueryArrayBoolDSL, testdata.PayloadQueryArrayBoolEncodeCode},
		{"query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL, testdata.PayloadQueryArrayBoolValidateEncodeCode},
		{"query-array-int", testdata.PayloadQueryArrayIntDSL, testdata.PayloadQueryArrayIntEncodeCode},
		{"query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL, testdata.PayloadQueryArrayIntValidateEncodeCode},
		{"query-array-int32", testdata.PayloadQueryArrayInt32DSL, testdata.PayloadQueryArrayInt32EncodeCode},
		{"query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL, testdata.PayloadQueryArrayInt32ValidateEncodeCode},
		{"query-array-int64", testdata.PayloadQueryArrayInt64DSL, testdata.PayloadQueryArrayInt64EncodeCode},
		{"query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL, testdata.PayloadQueryArrayInt64ValidateEncodeCode},
		{"query-array-uint", testdata.PayloadQueryArrayUIntDSL, testdata.PayloadQueryArrayUIntEncodeCode},
		{"query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL, testdata.PayloadQueryArrayUIntValidateEncodeCode},
		{"query-array-uint32", testdata.PayloadQueryArrayUInt32DSL, testdata.PayloadQueryArrayUInt32EncodeCode},
		{"query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL, testdata.PayloadQueryArrayUInt32ValidateEncodeCode},
		{"query-array-uint64", testdata.PayloadQueryArrayUInt64DSL, testdata.PayloadQueryArrayUInt64EncodeCode},
		{"query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL, testdata.PayloadQueryArrayUInt64ValidateEncodeCode},
		{"query-array-float32", testdata.PayloadQueryArrayFloat32DSL, testdata.PayloadQueryArrayFloat32EncodeCode},
		{"query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL, testdata.PayloadQueryArrayFloat32ValidateEncodeCode},
		{"query-array-float64", testdata.PayloadQueryArrayFloat64DSL, testdata.PayloadQueryArrayFloat64EncodeCode},
		{"query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL, testdata.PayloadQueryArrayFloat64ValidateEncodeCode},
		{"query-array-string", testdata.PayloadQueryArrayStringDSL, testdata.PayloadQueryArrayStringEncodeCode},
		{"query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL, testdata.PayloadQueryArrayStringValidateEncodeCode},
		{"query-array-bytes", testdata.PayloadQueryArrayBytesDSL, testdata.PayloadQueryArrayBytesEncodeCode},
		{"query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL, testdata.PayloadQueryArrayBytesValidateEncodeCode},
		{"query-array-any", testdata.PayloadQueryArrayAnyDSL, testdata.PayloadQueryArrayAnyEncodeCode},
		{"query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL, testdata.PayloadQueryArrayAnyValidateEncodeCode},

		{"query-primitive-string-validate", testdata.PayloadQueryPrimitiveStringValidateDSL, testdata.PayloadQueryPrimitiveStringValidateEncodeCode},
		{"query-primitive-bool-validate", testdata.PayloadQueryPrimitiveBoolValidateDSL, testdata.PayloadQueryPrimitiveBoolValidateEncodeCode},
		{"query-primitive-array-string-validate", testdata.PayloadQueryPrimitiveArrayStringValidateDSL, testdata.PayloadQueryPrimitiveArrayStringValidateEncodeCode},
		{"query-primitive-array-bool-validate", testdata.PayloadQueryPrimitiveArrayBoolValidateDSL, testdata.PayloadQueryPrimitiveArrayBoolValidateEncodeCode},

		{"query-string-mapped", testdata.PayloadQueryStringMappedDSL, testdata.PayloadQueryStringMappedEncodeCode},

		{"query-string-default", testdata.PayloadQueryStringDefaultDSL, testdata.PayloadQueryStringDefaultEncodeCode},
		{"query-primitive-string-default", testdata.PayloadQueryPrimitiveStringDefaultDSL, testdata.PayloadQueryPrimitiveStringDefaultEncodeCode},

		{"path-string", testdata.PayloadPathStringDSL, testdata.PayloadPathStringEncodeCode},
		{"path-string-validate", testdata.PayloadPathStringValidateDSL, testdata.PayloadPathStringValidateEncodeCode},
		{"path-array-string", testdata.PayloadPathArrayStringDSL, testdata.PayloadPathArrayStringEncodeCode},
		{"path-array-string-validate", testdata.PayloadPathArrayStringValidateDSL, testdata.PayloadPathArrayStringValidateEncodeCode},

		{"path-primitive-string-validate", testdata.PayloadPathPrimitiveStringValidateDSL, testdata.PayloadPathPrimitiveStringValidateEncodeCode},
		{"path-primitive-bool-validate", testdata.PayloadPathPrimitiveBoolValidateDSL, testdata.PayloadPathPrimitiveBoolValidateEncodeCode},
		{"path-primitive-array-string-validate", testdata.PayloadPathPrimitiveArrayStringValidateDSL, testdata.PayloadPathPrimitiveArrayStringValidateEncodeCode},
		{"path-primitive-array-bool-validate", testdata.PayloadPathPrimitiveArrayBoolValidateDSL, testdata.PayloadPathPrimitiveArrayBoolValidateEncodeCode},

		{"header-string", testdata.PayloadHeaderStringDSL, testdata.PayloadHeaderStringEncodeCode},
		{"header-string-validate", testdata.PayloadHeaderStringValidateDSL, testdata.PayloadHeaderStringValidateEncodeCode},
		{"header-array-string", testdata.PayloadHeaderArrayStringDSL, testdata.PayloadHeaderArrayStringEncodeCode},
		{"header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL, testdata.PayloadHeaderArrayStringValidateEncodeCode},

		{"header-primitive-string-validate", testdata.PayloadHeaderPrimitiveStringValidateDSL, testdata.PayloadHeaderPrimitiveStringValidateEncodeCode},
		{"header-primitive-bool-validate", testdata.PayloadHeaderPrimitiveBoolValidateDSL, testdata.PayloadHeaderPrimitiveBoolValidateEncodeCode},
		{"header-primitive-array-string-validate", testdata.PayloadHeaderPrimitiveArrayStringValidateDSL, testdata.PayloadHeaderPrimitiveArrayStringValidateEncodeCode},
		{"header-primitive-array-bool-validate", testdata.PayloadHeaderPrimitiveArrayBoolValidateDSL, testdata.PayloadHeaderPrimitiveArrayBoolValidateEncodeCode},

		{"header-string-default", testdata.PayloadHeaderStringDefaultDSL, testdata.PayloadHeaderStringDefaultEncodeCode},
		{"header-primitive-string-default", testdata.PayloadHeaderPrimitiveStringDefaultDSL, testdata.PayloadHeaderPrimitiveStringDefaultEncodeCode},

		{"body-string", testdata.PayloadBodyStringDSL, testdata.PayloadBodyStringEncodeCode},
		{"body-string-validate", testdata.PayloadBodyStringValidateDSL, testdata.PayloadBodyStringValidateEncodeCode},
		{"body-user", testdata.PayloadBodyUserDSL, testdata.PayloadBodyUserEncodeCode},
		{"body-user-validate", testdata.PayloadBodyUserValidateDSL, testdata.PayloadBodyUserValidateEncodeCode},
		{"body-array-string", testdata.PayloadBodyArrayStringDSL, testdata.PayloadBodyArrayStringEncodeCode},
		{"body-array-string-validate", testdata.PayloadBodyArrayStringValidateDSL, testdata.PayloadBodyArrayStringValidateEncodeCode},
		{"body-array-user", testdata.PayloadBodyArrayUserDSL, testdata.PayloadBodyArrayUserEncodeCode},
		{"body-array-user-validate", testdata.PayloadBodyArrayUserValidateDSL, testdata.PayloadBodyArrayUserValidateEncodeCode},
		{"body-map-string", testdata.PayloadBodyMapStringDSL, testdata.PayloadBodyMapStringEncodeCode},
		{"body-map-string-validate", testdata.PayloadBodyMapStringValidateDSL, testdata.PayloadBodyMapStringValidateEncodeCode},
		{"body-map-user", testdata.PayloadBodyMapUserDSL, testdata.PayloadBodyMapUserEncodeCode},
		{"body-map-user-validate", testdata.PayloadBodyMapUserValidateDSL, testdata.PayloadBodyMapUserValidateEncodeCode},

		{"body-primitive-string-validate", testdata.PayloadBodyPrimitiveStringValidateDSL, testdata.PayloadBodyPrimitiveStringValidateEncodeCode},
		{"body-primitive-bool-validate", testdata.PayloadBodyPrimitiveBoolValidateDSL, testdata.PayloadBodyPrimitiveBoolValidateEncodeCode},
		{"body-primitive-array-string-validate", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, testdata.PayloadBodyPrimitiveArrayStringValidateEncodeCode},
		{"body-primitive-array-bool-validate", testdata.PayloadBodyPrimitiveArrayBoolValidateDSL, testdata.PayloadBodyPrimitiveArrayBoolValidateEncodeCode},

		{"body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL, testdata.PayloadBodyPrimitiveArrayUserValidateEncodeCode},
		{"body-primitive-field-array-user", testdata.PayloadBodyPrimitiveFieldArrayUserDSL, testdata.PayloadBodyPrimitiveFieldArrayUserEncodeCode},
		{"body-primitive-field-array-user-validate", testdata.PayloadBodyPrimitiveFieldArrayUserValidateDSL, testdata.PayloadBodyPrimitiveFieldArrayUserValidateEncodeCode},

		{"body-query-object", testdata.PayloadBodyQueryObjectDSL, testdata.PayloadBodyQueryObjectEncodeCode},
		{"body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL, testdata.PayloadBodyQueryObjectValidateEncodeCode},
		{"body-query-user", testdata.PayloadBodyQueryUserDSL, testdata.PayloadBodyQueryUserEncodeCode},
		{"body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL, testdata.PayloadBodyQueryUserValidateEncodeCode},

		{"body-path-object", testdata.PayloadBodyPathObjectDSL, testdata.PayloadBodyPathObjectEncodeCode},
		{"body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL, testdata.PayloadBodyPathObjectValidateEncodeCode},
		{"body-path-user", testdata.PayloadBodyPathUserDSL, testdata.PayloadBodyPathUserEncodeCode},
		{"body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, testdata.PayloadBodyPathUserValidateEncodeCode},

		{"body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL, testdata.PayloadBodyQueryPathObjectEncodeCode},
		{"body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL, testdata.PayloadBodyQueryPathObjectValidateEncodeCode},
		{"body-query-path-user", testdata.PayloadBodyQueryPathUserDSL, testdata.PayloadBodyQueryPathUserEncodeCode},
		{"body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL, testdata.PayloadBodyQueryPathUserValidateEncodeCode},
	}
	golden := makeGolden(t, "testdata/payload_encode_functions.go")
	if golden != nil {
		golden.WriteString("package testdata\n")
		defer golden.Close()
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ClientFiles("", httpdesign.Root)
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
