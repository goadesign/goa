package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"decode-path-custom-float32", testdata.PayloadPathCustomFloat32DSL, testdata.PayloadPathCustomFloat32DecodeCode},
		{"decode-path-custom-float64", testdata.PayloadPathCustomFloat64DSL, testdata.PayloadPathCustomFloat64DecodeCode},
		{"decode-path-custom-int", testdata.PayloadPathCustomIntDSL, testdata.PayloadPathCustomIntDecodeCode},
		{"decode-path-custom-int32", testdata.PayloadPathCustomInt32DSL, testdata.PayloadPathCustomInt32DecodeCode},
		{"decode-path-custom-int64", testdata.PayloadPathCustomInt64DSL, testdata.PayloadPathCustomInt64DecodeCode},
		{"decode-path-custom-uint", testdata.PayloadPathCustomUIntDSL, testdata.PayloadPathCustomUIntDecodeCode},
		{"decode-path-custom-uint32", testdata.PayloadPathCustomUInt32DSL, testdata.PayloadPathCustomUInt32DecodeCode},
		{"decode-path-custom-uint64", testdata.PayloadPathCustomUInt64DSL, testdata.PayloadPathCustomUInt64DecodeCode},
		{"decode-query-bool", testdata.PayloadQueryBoolDSL, testdata.PayloadQueryBoolDecodeCode},
		{"decode-query-bool-validate", testdata.PayloadQueryBoolValidateDSL, testdata.PayloadQueryBoolValidateDecodeCode},
		{"decode-query-int", testdata.PayloadQueryIntDSL, testdata.PayloadQueryIntDecodeCode},
		{"decode-query-int-validate", testdata.PayloadQueryIntValidateDSL, testdata.PayloadQueryIntValidateDecodeCode},
		{"decode-query-int32", testdata.PayloadQueryInt32DSL, testdata.PayloadQueryInt32DecodeCode},
		{"decode-query-int32-validate", testdata.PayloadQueryInt32ValidateDSL, testdata.PayloadQueryInt32ValidateDecodeCode},
		{"decode-query-int64", testdata.PayloadQueryInt64DSL, testdata.PayloadQueryInt64DecodeCode},
		{"decode-query-int64-validate", testdata.PayloadQueryInt64ValidateDSL, testdata.PayloadQueryInt64ValidateDecodeCode},
		{"decode-query-uint", testdata.PayloadQueryUIntDSL, testdata.PayloadQueryUIntDecodeCode},
		{"decode-query-uint-validate", testdata.PayloadQueryUIntValidateDSL, testdata.PayloadQueryUIntValidateDecodeCode},
		{"decode-query-uint32", testdata.PayloadQueryUInt32DSL, testdata.PayloadQueryUInt32DecodeCode},
		{"decode-query-uint32-validate", testdata.PayloadQueryUInt32ValidateDSL, testdata.PayloadQueryUInt32ValidateDecodeCode},
		{"decode-query-uint64", testdata.PayloadQueryUInt64DSL, testdata.PayloadQueryUInt64DecodeCode},
		{"decode-query-uint64-validate", testdata.PayloadQueryUInt64ValidateDSL, testdata.PayloadQueryUInt64ValidateDecodeCode},
		{"decode-query-float32", testdata.PayloadQueryFloat32DSL, testdata.PayloadQueryFloat32DecodeCode},
		{"decode-query-float32-validate", testdata.PayloadQueryFloat32ValidateDSL, testdata.PayloadQueryFloat32ValidateDecodeCode},
		{"decode-query-float64", testdata.PayloadQueryFloat64DSL, testdata.PayloadQueryFloat64DecodeCode},
		{"decode-query-float64-validate", testdata.PayloadQueryFloat64ValidateDSL, testdata.PayloadQueryFloat64ValidateDecodeCode},
		{"decode-query-string", testdata.PayloadQueryStringDSL, testdata.PayloadQueryStringDecodeCode},
		{"decode-query-string-validate", testdata.PayloadQueryStringValidateDSL, testdata.PayloadQueryStringValidateDecodeCode},
		{"decode-query-string-not-required-validate", testdata.PayloadQueryStringNotRequiredValidateDSL, testdata.PayloadQueryStringNotRequiredValidateDecodeCode},
		{"decode-query-bytes", testdata.PayloadQueryBytesDSL, testdata.PayloadQueryBytesDecodeCode},
		{"decode-query-bytes-validate", testdata.PayloadQueryBytesValidateDSL, testdata.PayloadQueryBytesValidateDecodeCode},
		{"decode-query-any", testdata.PayloadQueryAnyDSL, testdata.PayloadQueryAnyDecodeCode},
		{"decode-query-any-validate", testdata.PayloadQueryAnyValidateDSL, testdata.PayloadQueryAnyValidateDecodeCode},
		{"decode-query-array-bool", testdata.PayloadQueryArrayBoolDSL, testdata.PayloadQueryArrayBoolDecodeCode},
		{"decode-query-array-bool-validate", testdata.PayloadQueryArrayBoolValidateDSL, testdata.PayloadQueryArrayBoolValidateDecodeCode},
		{"decode-query-array-int", testdata.PayloadQueryArrayIntDSL, testdata.PayloadQueryArrayIntDecodeCode},
		{"decode-query-array-int-validate", testdata.PayloadQueryArrayIntValidateDSL, testdata.PayloadQueryArrayIntValidateDecodeCode},
		{"decode-query-array-int32", testdata.PayloadQueryArrayInt32DSL, testdata.PayloadQueryArrayInt32DecodeCode},
		{"decode-query-array-int32-validate", testdata.PayloadQueryArrayInt32ValidateDSL, testdata.PayloadQueryArrayInt32ValidateDecodeCode},
		{"decode-query-array-int64", testdata.PayloadQueryArrayInt64DSL, testdata.PayloadQueryArrayInt64DecodeCode},
		{"decode-query-array-int64-validate", testdata.PayloadQueryArrayInt64ValidateDSL, testdata.PayloadQueryArrayInt64ValidateDecodeCode},
		{"decode-query-array-uint", testdata.PayloadQueryArrayUIntDSL, testdata.PayloadQueryArrayUIntDecodeCode},
		{"decode-query-array-uint-validate", testdata.PayloadQueryArrayUIntValidateDSL, testdata.PayloadQueryArrayUIntValidateDecodeCode},
		{"decode-query-array-uint32", testdata.PayloadQueryArrayUInt32DSL, testdata.PayloadQueryArrayUInt32DecodeCode},
		{"decode-query-array-uint32-validate", testdata.PayloadQueryArrayUInt32ValidateDSL, testdata.PayloadQueryArrayUInt32ValidateDecodeCode},
		{"decode-query-array-uint64", testdata.PayloadQueryArrayUInt64DSL, testdata.PayloadQueryArrayUInt64DecodeCode},
		{"decode-query-array-uint64-validate", testdata.PayloadQueryArrayUInt64ValidateDSL, testdata.PayloadQueryArrayUInt64ValidateDecodeCode},
		{"decode-query-array-float32", testdata.PayloadQueryArrayFloat32DSL, testdata.PayloadQueryArrayFloat32DecodeCode},
		{"decode-query-array-float32-validate", testdata.PayloadQueryArrayFloat32ValidateDSL, testdata.PayloadQueryArrayFloat32ValidateDecodeCode},
		{"decode-query-array-float64", testdata.PayloadQueryArrayFloat64DSL, testdata.PayloadQueryArrayFloat64DecodeCode},
		{"decode-query-array-float64-validate", testdata.PayloadQueryArrayFloat64ValidateDSL, testdata.PayloadQueryArrayFloat64ValidateDecodeCode},
		{"decode-query-array-string", testdata.PayloadQueryArrayStringDSL, testdata.PayloadQueryArrayStringDecodeCode},
		{"decode-query-array-string-validate", testdata.PayloadQueryArrayStringValidateDSL, testdata.PayloadQueryArrayStringValidateDecodeCode},
		{"decode-query-array-bytes", testdata.PayloadQueryArrayBytesDSL, testdata.PayloadQueryArrayBytesDecodeCode},
		{"decode-query-array-bytes-validate", testdata.PayloadQueryArrayBytesValidateDSL, testdata.PayloadQueryArrayBytesValidateDecodeCode},
		{"decode-query-array-any", testdata.PayloadQueryArrayAnyDSL, testdata.PayloadQueryArrayAnyDecodeCode},
		{"decode-query-array-any-validate", testdata.PayloadQueryArrayAnyValidateDSL, testdata.PayloadQueryArrayAnyValidateDecodeCode},
		{"decode-query-map-string-string", testdata.PayloadQueryMapStringStringDSL, testdata.PayloadQueryMapStringStringDecodeCode},
		{"decode-query-map-string-string-validate", testdata.PayloadQueryMapStringStringValidateDSL, testdata.PayloadQueryMapStringStringValidateDecodeCode},
		{"decode-query-map-string-bool", testdata.PayloadQueryMapStringBoolDSL, testdata.PayloadQueryMapStringBoolDecodeCode},
		{"decode-query-map-string-bool-validate", testdata.PayloadQueryMapStringBoolValidateDSL, testdata.PayloadQueryMapStringBoolValidateDecodeCode},
		{"decode-query-map-bool-string", testdata.PayloadQueryMapBoolStringDSL, testdata.PayloadQueryMapBoolStringDecodeCode},
		{"decode-query-map-bool-string-validate", testdata.PayloadQueryMapBoolStringValidateDSL, testdata.PayloadQueryMapBoolStringValidateDecodeCode},
		{"decode-query-map-bool-bool", testdata.PayloadQueryMapBoolBoolDSL, testdata.PayloadQueryMapBoolBoolDecodeCode},
		{"decode-query-map-bool-bool-validate", testdata.PayloadQueryMapBoolBoolValidateDSL, testdata.PayloadQueryMapBoolBoolValidateDecodeCode},
		{"decode-query-map-string-array-string", testdata.PayloadQueryMapStringArrayStringDSL, testdata.PayloadQueryMapStringArrayStringDecodeCode},
		{"decode-query-map-string-array-string-validate", testdata.PayloadQueryMapStringArrayStringValidateDSL, testdata.PayloadQueryMapStringArrayStringValidateDecodeCode},
		{"decode-query-map-string-array-bool", testdata.PayloadQueryMapStringArrayBoolDSL, testdata.PayloadQueryMapStringArrayBoolDecodeCode},
		{"decode-query-map-string-array-bool-validate", testdata.PayloadQueryMapStringArrayBoolValidateDSL, testdata.PayloadQueryMapStringArrayBoolValidateDecodeCode},
		{"decode-query-map-bool-array-string", testdata.PayloadQueryMapBoolArrayStringDSL, testdata.PayloadQueryMapBoolArrayStringDecodeCode},
		{"decode-query-map-bool-array-string-validate", testdata.PayloadQueryMapBoolArrayStringValidateDSL, testdata.PayloadQueryMapBoolArrayStringValidateDecodeCode},
		{"decode-query-map-bool-array-bool", testdata.PayloadQueryMapBoolArrayBoolDSL, testdata.PayloadQueryMapBoolArrayBoolDecodeCode},
		{"decode-query-map-bool-array-bool-validate", testdata.PayloadQueryMapBoolArrayBoolValidateDSL, testdata.PayloadQueryMapBoolArrayBoolValidateDecodeCode},

		{"decode-query-primitive-string-validate", testdata.PayloadQueryPrimitiveStringValidateDSL, testdata.PayloadQueryPrimitiveStringValidateDecodeCode},
		{"decode-query-primitive-bool-validate", testdata.PayloadQueryPrimitiveBoolValidateDSL, testdata.PayloadQueryPrimitiveBoolValidateDecodeCode},
		{"decode-query-primitive-array-string-validate", testdata.PayloadQueryPrimitiveArrayStringValidateDSL, testdata.PayloadQueryPrimitiveArrayStringValidateDecodeCode},
		{"decode-query-primitive-array-bool-validate", testdata.PayloadQueryPrimitiveArrayBoolValidateDSL, testdata.PayloadQueryPrimitiveArrayBoolValidateDecodeCode},
		{"decode-query-primitive-map-string-array-string-validate", testdata.PayloadQueryPrimitiveMapStringArrayStringValidateDSL, testdata.PayloadQueryPrimitiveMapStringArrayStringValidateDecodeCode},
		{"decode-query-primitive-map-string-bool-validate", testdata.PayloadQueryPrimitiveMapStringBoolValidateDSL, testdata.PayloadQueryPrimitiveMapStringBoolValidateDecodeCode},
		{"decode-query-primitive-map-bool-array-bool-validate", testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateDSL, testdata.PayloadQueryPrimitiveMapBoolArrayBoolValidateDecodeCode},
		{"decode-query-map-string-map-int-string-validate", testdata.PayloadQueryMapStringMapIntStringValidateDSL, testdata.PayloadQueryMapStringMapIntStringValidateDecodeCode},
		{"decode-query-map-int-map-string-array-int-validate", testdata.PayloadQueryMapIntMapStringArrayIntValidateDSL, testdata.PayloadQueryMapIntMapStringArrayIntValidateDecodeCode},

		{"decode-query-string-mapped", testdata.PayloadQueryStringMappedDSL, testdata.PayloadQueryStringMappedDecodeCode},

		{"decode-query-string-default", testdata.PayloadQueryStringDefaultDSL, testdata.PayloadQueryStringDefaultDecodeCode},
		{"decode-query-string-slice-default", testdata.PayloadQueryStringSliceDefaultDSL, testdata.PayloadQueryStringSliceDefaultDecodeCode},
		{"decode-query-string-default-validate", testdata.PayloadQueryStringDefaultValidateDSL, testdata.PayloadQueryStringDefaultValidateDecodeCode},
		{"decode-query-primitive-string-default", testdata.PayloadQueryPrimitiveStringDefaultDSL, testdata.PayloadQueryPrimitiveStringDefaultDecodeCode},
		{"decode-query-string-extended-payload", testdata.PayloadExtendedQueryStringDSL, testdata.PayloadExtendedQueryStringDecodeCode},

		{"decode-path-string", testdata.PayloadPathStringDSL, testdata.PayloadPathStringDecodeCode},
		{"decode-path-string-validate", testdata.PayloadPathStringValidateDSL, testdata.PayloadPathStringValidateDecodeCode},
		{"decode-path-array-string", testdata.PayloadPathArrayStringDSL, testdata.PayloadPathArrayStringDecodeCode},
		{"decode-path-array-string-validate", testdata.PayloadPathArrayStringValidateDSL, testdata.PayloadPathArrayStringValidateDecodeCode},

		{"decode-path-primitive-string-validate", testdata.PayloadPathPrimitiveStringValidateDSL, testdata.PayloadPathPrimitiveStringValidateDecodeCode},
		{"decode-path-primitive-bool-validate", testdata.PayloadPathPrimitiveBoolValidateDSL, testdata.PayloadPathPrimitiveBoolValidateDecodeCode},
		{"decode-path-primitive-array-string-validate", testdata.PayloadPathPrimitiveArrayStringValidateDSL, testdata.PayloadPathPrimitiveArrayStringValidateDecodeCode},
		{"decode-path-primitive-array-bool-validate", testdata.PayloadPathPrimitiveArrayBoolValidateDSL, testdata.PayloadPathPrimitiveArrayBoolValidateDecodeCode},

		{"decode-header-string", testdata.PayloadHeaderStringDSL, testdata.PayloadHeaderStringDecodeCode},
		{"decode-header-string-validate", testdata.PayloadHeaderStringValidateDSL, testdata.PayloadHeaderStringValidateDecodeCode},
		{"decode-header-array-string", testdata.PayloadHeaderArrayStringDSL, testdata.PayloadHeaderArrayStringDecodeCode},
		{"decode-header-array-string-validate", testdata.PayloadHeaderArrayStringValidateDSL, testdata.PayloadHeaderArrayStringValidateDecodeCode},

		{"decode-header-primitive-string-validate", testdata.PayloadHeaderPrimitiveStringValidateDSL, testdata.PayloadHeaderPrimitiveStringValidateDecodeCode},
		{"decode-header-primitive-bool-validate", testdata.PayloadHeaderPrimitiveBoolValidateDSL, testdata.PayloadHeaderPrimitiveBoolValidateDecodeCode},
		{"decode-header-primitive-array-string-validate", testdata.PayloadHeaderPrimitiveArrayStringValidateDSL, testdata.PayloadHeaderPrimitiveArrayStringValidateDecodeCode},
		{"decode-header-primitive-array-bool-validate", testdata.PayloadHeaderPrimitiveArrayBoolValidateDSL, testdata.PayloadHeaderPrimitiveArrayBoolValidateDecodeCode},

		{"decode-header-string-default", testdata.PayloadHeaderStringDefaultDSL, testdata.PayloadHeaderStringDefaultDecodeCode},
		{"decode-header-string-default-validate", testdata.PayloadHeaderStringDefaultValidateDSL, testdata.PayloadHeaderStringDefaultValidateDecodeCode},
		{"decode-header-primitive-string-default", testdata.PayloadHeaderPrimitiveStringDefaultDSL, testdata.PayloadHeaderPrimitiveStringDefaultDecodeCode},

		{"decode-cookie-string", testdata.PayloadCookieStringDSL, testdata.PayloadCookieStringDecodeCode},
		{"decode-cookie-string-validate", testdata.PayloadCookieStringValidateDSL, testdata.PayloadCookieStringValidateDecodeCode},

		{"decode-cookie-primitive-string-validate", testdata.PayloadCookiePrimitiveStringValidateDSL, testdata.PayloadCookiePrimitiveStringValidateDecodeCode},
		{"decode-cookie-primitive-bool-validate", testdata.PayloadCookiePrimitiveBoolValidateDSL, testdata.PayloadCookiePrimitiveBoolValidateDecodeCode},

		{"decode-cookie-string-default", testdata.PayloadCookieStringDefaultDSL, testdata.PayloadCookieStringDefaultDecodeCode},
		{"decode-cookie-string-default-validate", testdata.PayloadCookieStringDefaultValidateDSL, testdata.PayloadCookieStringDefaultValidateDecodeCode},
		{"decode-cookie-primitive-string-default", testdata.PayloadCookiePrimitiveStringDefaultDSL, testdata.PayloadCookiePrimitiveStringDefaultDecodeCode},

		{"decode-body-string", testdata.PayloadBodyStringDSL, testdata.PayloadBodyStringDecodeCode},
		{"decode-body-string-validate", testdata.PayloadBodyStringValidateDSL, testdata.PayloadBodyStringValidateDecodeCode},
		{"decode-body-user", testdata.PayloadBodyUserDSL, testdata.PayloadBodyUserDecodeCode},
		{"decode-body-user-required", testdata.PayloadBodyUserRequiredDSL, testdata.PayloadBodyUserRequiredDecodeCode},
		{"decode-body-user-nested", testdata.PayloadBodyNestedUserDSL, testdata.PayloadBodyNestedUserDecodeCode},
		{"decode-body-user-validate", testdata.PayloadBodyUserValidateDSL, testdata.PayloadBodyUserValidateDecodeCode},
		{"decode-body-object", testdata.PayloadBodyObjectDSL, testdata.PayloadBodyObjectDecodeCode},
		{"decode-body-object-validate", testdata.PayloadBodyObjectValidateDSL, testdata.PayloadBodyObjectValidateDecodeCode},
		{"decode-body-union", testdata.PayloadBodyUnionDSL, testdata.PayloadBodyUnionDecodeCode},
		{"decode-body-union-validate", testdata.PayloadBodyUnionValidateDSL, testdata.PayloadBodyUnionValidateDecodeCode},
		{"decode-body-union-user", testdata.PayloadBodyUnionUserDSL, testdata.PayloadBodyUnionUserDecodeCode},
		{"decode-body-union-user-validate", testdata.PayloadBodyUnionUserValidateDSL, testdata.PayloadBodyUnionUserValidateDecodeCode},
		{"decode-body-array-string", testdata.PayloadBodyArrayStringDSL, testdata.PayloadBodyArrayStringDecodeCode},
		{"decode-body-array-string-validate", testdata.PayloadBodyArrayStringValidateDSL, testdata.PayloadBodyArrayStringValidateDecodeCode},
		{"decode-body-array-user", testdata.PayloadBodyArrayUserDSL, testdata.PayloadBodyArrayUserDecodeCode},
		{"decode-body-array-user-validate", testdata.PayloadBodyArrayUserValidateDSL, testdata.PayloadBodyArrayUserValidateDecodeCode},
		{"decode-body-map-string", testdata.PayloadBodyMapStringDSL, testdata.PayloadBodyMapStringDecodeCode},
		{"decode-body-map-string-validate", testdata.PayloadBodyMapStringValidateDSL, testdata.PayloadBodyMapStringValidateDecodeCode},
		{"decode-body-map-user", testdata.PayloadBodyMapUserDSL, testdata.PayloadBodyMapUserDecodeCode},
		{"decode-body-map-user-validate", testdata.PayloadBodyMapUserValidateDSL, testdata.PayloadBodyMapUserValidateDecodeCode},

		{"decode-body-primitive-string-validate", testdata.PayloadBodyPrimitiveStringValidateDSL, testdata.PayloadBodyPrimitiveStringValidateDecodeCode},
		{"decode-body-primitive-bool-validate", testdata.PayloadBodyPrimitiveBoolValidateDSL, testdata.PayloadBodyPrimitiveBoolValidateDecodeCode},
		{"decode-body-primitive-array-string-validate", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, testdata.PayloadBodyPrimitiveArrayStringValidateDecodeCode},
		{"decode-body-primitive-array-bool-validate", testdata.PayloadBodyPrimitiveArrayBoolValidateDSL, testdata.PayloadBodyPrimitiveArrayBoolValidateDecodeCode},

		{"decode-body-primitive-array-user-validate", testdata.PayloadBodyPrimitiveArrayUserValidateDSL, testdata.PayloadBodyPrimitiveArrayUserValidateDecodeCode},
		{"decode-body-primitive-field-array-user", testdata.PayloadBodyPrimitiveFieldArrayUserDSL, testdata.PayloadBodyPrimitiveFieldArrayUserDecodeCode},
		{"decode-body-extend-primitive-field-array-user", testdata.PayloadExtendBodyPrimitiveFieldArrayUserDSL, testdata.PayloadBodyPrimitiveFieldArrayUserDecodeCode},
		{"decode-body-extend-primitive-field-string", testdata.PayloadExtendBodyPrimitiveFieldStringDSL, testdata.PayloadBodyPrimitiveFieldStringDecodeCode},
		{"decode-body-primitive-field-array-user-validate", testdata.PayloadBodyPrimitiveFieldArrayUserValidateDSL, testdata.PayloadBodyPrimitiveFieldArrayUserValidateDecodeCode},

		{"decode-body-query-object", testdata.PayloadBodyQueryObjectDSL, testdata.PayloadBodyQueryObjectDecodeCode},
		{"decode-body-query-object-validate", testdata.PayloadBodyQueryObjectValidateDSL, testdata.PayloadBodyQueryObjectValidateDecodeCode},
		{"decode-body-query-user", testdata.PayloadBodyQueryUserDSL, testdata.PayloadBodyQueryUserDecodeCode},
		{"decode-body-query-user-validate", testdata.PayloadBodyQueryUserValidateDSL, testdata.PayloadBodyQueryUserValidateDecodeCode},

		{"decode-body-path-object", testdata.PayloadBodyPathObjectDSL, testdata.PayloadBodyPathObjectDecodeCode},
		{"decode-body-path-object-validate", testdata.PayloadBodyPathObjectValidateDSL, testdata.PayloadBodyPathObjectValidateDecodeCode},
		{"decode-body-path-user", testdata.PayloadBodyPathUserDSL, testdata.PayloadBodyPathUserDecodeCode},
		{"decode-body-path-user-validate", testdata.PayloadBodyPathUserValidateDSL, testdata.PayloadBodyPathUserValidateDecodeCode},

		{"decode-body-query-path-object", testdata.PayloadBodyQueryPathObjectDSL, testdata.PayloadBodyQueryPathObjectDecodeCode},
		{"decode-body-query-path-object-validate", testdata.PayloadBodyQueryPathObjectValidateDSL, testdata.PayloadBodyQueryPathObjectValidateDecodeCode},
		{"decode-body-query-path-user", testdata.PayloadBodyQueryPathUserDSL, testdata.PayloadBodyQueryPathUserDecodeCode},
		{"decode-body-query-path-user-validate", testdata.PayloadBodyQueryPathUserValidateDSL, testdata.PayloadBodyQueryPathUserValidateDecodeCode},

		{"decode-map-query-primitive-primitive", testdata.PayloadMapQueryPrimitivePrimitiveDSL, testdata.PayloadMapQueryPrimitivePrimitiveDecodeCode},
		{"decode-map-query-primitive-array", testdata.PayloadMapQueryPrimitiveArrayDSL, testdata.PayloadMapQueryPrimitiveArrayDecodeCode},
		{"decode-map-query-object", testdata.PayloadMapQueryObjectDSL, testdata.PayloadMapQueryObjectDecodeCode},
		{"decode-multipart-body-primitive", testdata.PayloadMultipartPrimitiveDSL, testdata.PayloadMultipartPrimitiveDecodeCode},
		{"decode-multipart-body-user-type", testdata.PayloadMultipartUserTypeDSL, testdata.PayloadMultipartUserTypeDecodeCode},
		{"decode-multipart-body-array-type", testdata.PayloadMultipartArrayTypeDSL, testdata.PayloadMultipartArrayTypeDecodeCode},
		{"decode-multipart-body-map-type", testdata.PayloadMultipartMapTypeDSL, testdata.PayloadMultipartMapTypeDecodeCode},
		{"decode-with-params-and-headers-dsl", testdata.WithParamsAndHeadersBlockDSL, testdata.WithParamsAndHeadersBlockDecodeCode},

		{"decode-query-int-alias", testdata.QueryIntAliasDSL, testdata.QueryIntAliasDecodeCode},
		{"decode-query-int-alias-validate", testdata.QueryIntAliasValidateDSL, testdata.QueryIntAliasValidateDecodeCode},
		{"decode-query-array-alias", testdata.QueryArrayAliasDSL, testdata.QueryArrayAliasDecodeCode},
		{"decode-query-array-alias-validate", testdata.QueryArrayAliasValidateDSL, testdata.QueryArrayAliasValidateDecodeCode},
		{"decode-query-map-alias", testdata.QueryMapAliasDSL, testdata.QueryMapAliasDecodeCode},
		{"decode-query-map-alias-validate", testdata.QueryMapAliasValidateDSL, testdata.QueryMapAliasValidateDecodeCode},
		{"decode-query-array-nested-alias-validate", testdata.QueryArrayNestedAliasValidateDSL, testdata.QueryArrayNestedAliasValidateDecodeCode},
		{"decode-header-int-alias", testdata.HeaderIntAliasDSL, testdata.HeaderIntAliasDecodeCode},
		{"decode-path-int-alias", testdata.PathIntAliasDSL, testdata.PathIntAliasDecodeCode},
	}
	golden := makeGolden(t, "testdata/payload_decode_functions.go")
	if golden != nil {
		if _, err := golden.WriteString("package testdata\n"); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := ServerFiles("", expr.Root)
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
				name = strings.ReplaceAll(name, "Uint", "UInt")
				code = "\nvar Payload" + name + "DecodeCode = `" + code + "`"
				if _, err := golden.WriteString(code + "\n"); err != nil {
					t.Fatal(err)
				}
			} else if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
