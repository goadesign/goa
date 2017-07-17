package rest

import (
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/codegen/rest/testing"
	"goa.design/goa.v2/design/rest"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"query-bool", PayloadQueryBoolDSL, PayloadQueryBoolDecodeCode},
		{"query-bool-validate", PayloadQueryBoolValidateDSL, PayloadQueryBoolValidateDecodeCode},
		{"query-int", PayloadQueryIntDSL, PayloadQueryIntDecodeCode},
		{"query-int-validate", PayloadQueryIntValidateDSL, PayloadQueryIntValidateDecodeCode},
		{"query-int32", PayloadQueryInt32DSL, PayloadQueryInt32DecodeCode},
		{"query-int32-validate", PayloadQueryInt32ValidateDSL, PayloadQueryInt32ValidateDecodeCode},
		{"query-int64", PayloadQueryInt64DSL, PayloadQueryInt64DecodeCode},
		{"query-int64-validate", PayloadQueryInt64ValidateDSL, PayloadQueryInt64ValidateDecodeCode},
		{"query-uint", PayloadQueryUIntDSL, PayloadQueryUIntDecodeCode},
		{"query-uint-validate", PayloadQueryUIntValidateDSL, PayloadQueryUIntValidateDecodeCode},
		{"query-uint32", PayloadQueryUInt32DSL, PayloadQueryUInt32DecodeCode},
		{"query-uint32-validate", PayloadQueryUInt32ValidateDSL, PayloadQueryUInt32ValidateDecodeCode},
		{"query-uint64", PayloadQueryUInt64DSL, PayloadQueryUInt64DecodeCode},
		{"query-uint64-validate", PayloadQueryUInt64ValidateDSL, PayloadQueryUInt64ValidateDecodeCode},
		{"query-float32", PayloadQueryFloat32DSL, PayloadQueryFloat32DecodeCode},
		{"query-float32-validate", PayloadQueryFloat32ValidateDSL, PayloadQueryFloat32ValidateDecodeCode},
		{"query-float64", PayloadQueryFloat64DSL, PayloadQueryFloat64DecodeCode},
		{"query-float64-validate", PayloadQueryFloat64ValidateDSL, PayloadQueryFloat64ValidateDecodeCode},
		{"query-string", PayloadQueryStringDSL, PayloadQueryStringDecodeCode},
		{"query-string-validate", PayloadQueryStringValidateDSL, PayloadQueryStringValidateDecodeCode},
		{"query-bytes", PayloadQueryBytesDSL, PayloadQueryBytesDecodeCode},
		{"query-bytes-validate", PayloadQueryBytesValidateDSL, PayloadQueryBytesValidateDecodeCode},
		{"query-any", PayloadQueryAnyDSL, PayloadQueryAnyDecodeCode},
		{"query-any-validate", PayloadQueryAnyValidateDSL, PayloadQueryAnyValidateDecodeCode},
		{"query-array-bool", PayloadQueryArrayBoolDSL, PayloadQueryArrayBoolDecodeCode},
		{"query-array-bool-validate", PayloadQueryArrayBoolValidateDSL, PayloadQueryArrayBoolValidateDecodeCode},
		{"query-array-int", PayloadQueryArrayIntDSL, PayloadQueryArrayIntDecodeCode},
		{"query-array-int-validate", PayloadQueryArrayIntValidateDSL, PayloadQueryArrayIntValidateDecodeCode},
		{"query-array-int32", PayloadQueryArrayInt32DSL, PayloadQueryArrayInt32DecodeCode},
		{"query-array-int32-validate", PayloadQueryArrayInt32ValidateDSL, PayloadQueryArrayInt32ValidateDecodeCode},
		{"query-array-int64", PayloadQueryArrayInt64DSL, PayloadQueryArrayInt64DecodeCode},
		{"query-array-int64-validate", PayloadQueryArrayInt64ValidateDSL, PayloadQueryArrayInt64ValidateDecodeCode},
		{"query-array-uint", PayloadQueryArrayUIntDSL, PayloadQueryArrayUIntDecodeCode},
		{"query-array-uint-validate", PayloadQueryArrayUIntValidateDSL, PayloadQueryArrayUIntValidateDecodeCode},
		{"query-array-uint32", PayloadQueryArrayUInt32DSL, PayloadQueryArrayUInt32DecodeCode},
		{"query-array-uint32-validate", PayloadQueryArrayUInt32ValidateDSL, PayloadQueryArrayUInt32ValidateDecodeCode},
		{"query-array-uint64", PayloadQueryArrayUInt64DSL, PayloadQueryArrayUInt64DecodeCode},
		{"query-array-uint64-validate", PayloadQueryArrayUInt64ValidateDSL, PayloadQueryArrayUInt64ValidateDecodeCode},
		{"query-array-float32", PayloadQueryArrayFloat32DSL, PayloadQueryArrayFloat32DecodeCode},
		{"query-array-float32-validate", PayloadQueryArrayFloat32ValidateDSL, PayloadQueryArrayFloat32ValidateDecodeCode},
		{"query-array-float64", PayloadQueryArrayFloat64DSL, PayloadQueryArrayFloat64DecodeCode},
		{"query-array-float64-validate", PayloadQueryArrayFloat64ValidateDSL, PayloadQueryArrayFloat64ValidateDecodeCode},
		{"query-array-string", PayloadQueryArrayStringDSL, PayloadQueryArrayStringDecodeCode},
		{"query-array-string-validate", PayloadQueryArrayStringValidateDSL, PayloadQueryArrayStringValidateDecodeCode},
		{"query-array-bytes", PayloadQueryArrayBytesDSL, PayloadQueryArrayBytesDecodeCode},
		{"query-array-bytes-validate", PayloadQueryArrayBytesValidateDSL, PayloadQueryArrayBytesValidateDecodeCode},
		{"query-array-any", PayloadQueryArrayAnyDSL, PayloadQueryArrayAnyDecodeCode},
		{"query-array-any-validate", PayloadQueryArrayAnyValidateDSL, PayloadQueryArrayAnyValidateDecodeCode},
		{"query-map-string-string", PayloadQueryMapStringStringDSL, PayloadQueryMapStringStringDecodeCode},
		{"query-map-string-string-validate", PayloadQueryMapStringStringValidateDSL, PayloadQueryMapStringStringValidateDecodeCode},
		{"query-map-string-bool", PayloadQueryMapStringBoolDSL, PayloadQueryMapStringBoolDecodeCode},
		{"query-map-string-bool-validate", PayloadQueryMapStringBoolValidateDSL, PayloadQueryMapStringBoolValidateDecodeCode},
		{"query-map-bool-string", PayloadQueryMapBoolStringDSL, PayloadQueryMapBoolStringDecodeCode},
		{"query-map-bool-string-validate", PayloadQueryMapBoolStringValidateDSL, PayloadQueryMapBoolStringValidateDecodeCode},
		{"query-map-bool-bool", PayloadQueryMapBoolBoolDSL, PayloadQueryMapBoolBoolDecodeCode},
		{"query-map-bool-bool-validate", PayloadQueryMapBoolBoolValidateDSL, PayloadQueryMapBoolBoolValidateDecodeCode},
		{"query-map-string-array-string", PayloadQueryMapStringArrayStringDSL, PayloadQueryMapStringArrayStringDecodeCode},
		{"query-map-string-array-string-validate", PayloadQueryMapStringArrayStringValidateDSL, PayloadQueryMapStringArrayStringValidateDecodeCode},
		{"query-map-string-array-bool", PayloadQueryMapStringArrayBoolDSL, PayloadQueryMapStringArrayBoolDecodeCode},
		{"query-map-string-array-bool-validate", PayloadQueryMapStringArrayBoolValidateDSL, PayloadQueryMapStringArrayBoolValidateDecodeCode},
		{"query-map-bool-array-string", PayloadQueryMapBoolArrayStringDSL, PayloadQueryMapBoolArrayStringDecodeCode},
		{"query-map-bool-array-string-validate", PayloadQueryMapBoolArrayStringValidateDSL, PayloadQueryMapBoolArrayStringValidateDecodeCode},
		{"query-map-bool-array-bool", PayloadQueryMapBoolArrayBoolDSL, PayloadQueryMapBoolArrayBoolDecodeCode},
		{"query-map-bool-array-bool-validate", PayloadQueryMapBoolArrayBoolValidateDSL, PayloadQueryMapBoolArrayBoolValidateDecodeCode},

		{"query-primitive-string-validate", PayloadQueryPrimitiveStringValidateDSL, PayloadQueryPrimitiveStringValidateDecodeCode},
		{"query-primitive-bool-validate", PayloadQueryPrimitiveBoolValidateDSL, PayloadQueryPrimitiveBoolValidateDecodeCode},
		{"query-primitive-array-string-validate", PayloadQueryPrimitiveArrayStringValidateDSL, PayloadQueryPrimitiveArrayStringValidateDecodeCode},
		{"query-primitive-array-bool-validate", PayloadQueryPrimitiveArrayBoolValidateDSL, PayloadQueryPrimitiveArrayBoolValidateDecodeCode},
		{"query-primitive-map-string-array-string-validate", PayloadQueryPrimitiveMapStringArrayStringValidateDSL, PayloadQueryPrimitiveMapStringArrayStringValidateDecodeCode},
		{"query-primitive-map-string-bool-validate", PayloadQueryPrimitiveMapStringBoolValidateDSL, PayloadQueryPrimitiveMapStringBoolValidateDecodeCode},
		{"query-primitive-map-bool-array-bool-validate", PayloadQueryPrimitiveMapBoolArrayBoolValidateDSL, PayloadQueryPrimitiveMapBoolArrayBoolValidateDecodeCode},

		{"query-string-default", PayloadQueryStringDefaultDSL, PayloadQueryStringDefaultDecodeCode},
		{"query-primitive-string-default", PayloadQueryPrimitiveStringDefaultDSL, PayloadQueryPrimitiveStringDefaultDecodeCode},

		{"path-string", PayloadPathStringDSL, PayloadPathStringDecodeCode},
		{"path-string-validate", PayloadPathStringValidateDSL, PayloadPathStringValidateDecodeCode},
		{"path-array-string", PayloadPathArrayStringDSL, PayloadPathArrayStringDecodeCode},
		{"path-array-string-validate", PayloadPathArrayStringValidateDSL, PayloadPathArrayStringValidateDecodeCode},

		{"path-primitive-string-validate", PayloadPathPrimitiveStringValidateDSL, PayloadPathPrimitiveStringValidateDecodeCode},
		{"path-primitive-bool-validate", PayloadPathPrimitiveBoolValidateDSL, PayloadPathPrimitiveBoolValidateDecodeCode},
		{"path-primitive-array-string-validate", PayloadPathPrimitiveArrayStringValidateDSL, PayloadPathPrimitiveArrayStringValidateDecodeCode},
		{"path-primitive-array-bool-validate", PayloadPathPrimitiveArrayBoolValidateDSL, PayloadPathPrimitiveArrayBoolValidateDecodeCode},

		{"header-string", PayloadHeaderStringDSL, PayloadHeaderStringDecodeCode},
		{"header-string-validate", PayloadHeaderStringValidateDSL, PayloadHeaderStringValidateDecodeCode},
		{"header-array-string", PayloadHeaderArrayStringDSL, PayloadHeaderArrayStringDecodeCode},
		{"header-array-string-validate", PayloadHeaderArrayStringValidateDSL, PayloadHeaderArrayStringValidateDecodeCode},

		{"header-primitive-string-validate", PayloadHeaderPrimitiveStringValidateDSL, PayloadHeaderPrimitiveStringValidateDecodeCode},
		{"header-primitive-bool-validate", PayloadHeaderPrimitiveBoolValidateDSL, PayloadHeaderPrimitiveBoolValidateDecodeCode},
		{"header-primitive-array-string-validate", PayloadHeaderPrimitiveArrayStringValidateDSL, PayloadHeaderPrimitiveArrayStringValidateDecodeCode},
		{"header-primitive-array-bool-validate", PayloadHeaderPrimitiveArrayBoolValidateDSL, PayloadHeaderPrimitiveArrayBoolValidateDecodeCode},

		{"header-string-default", PayloadHeaderStringDefaultDSL, PayloadHeaderStringDefaultDecodeCode},
		{"header-primitive-string-default", PayloadHeaderPrimitiveStringDefaultDSL, PayloadHeaderPrimitiveStringDefaultDecodeCode},

		{"body-string", PayloadBodyStringDSL, PayloadBodyStringDecodeCode},
		{"body-string-validate", PayloadBodyStringValidateDSL, PayloadBodyStringValidateDecodeCode},
		{"body-user", PayloadBodyUserDSL, PayloadBodyUserDecodeCode},
		{"body-user-validate", PayloadBodyUserValidateDSL, PayloadUserBodyValidateDecodeCode},
		{"body-array-string", PayloadBodyArrayStringDSL, PayloadBodyArrayStringDecodeCode},
		{"body-array-string-validate", PayloadBodyArrayStringValidateDSL, PayloadBodyArrayStringValidateDecodeCode},
		{"body-array-user", PayloadBodyArrayUserDSL, PayloadBodyArrayUserDecodeCode},
		{"body-array-user-validate", PayloadBodyArrayUserValidateDSL, PayloadBodyArrayUserValidateDecodeCode},
		{"body-map-string", PayloadBodyMapStringDSL, PayloadBodyMapStringDecodeCode},
		{"body-map-string-validate", PayloadBodyMapStringValidateDSL, PayloadBodyMapStringValidateDecodeCode},
		{"body-map-user", PayloadBodyMapUserDSL, PayloadBodyMapUserDecodeCode},
		{"body-map-user-validate", PayloadBodyMapUserValidateDSL, PayloadBodyMapUserValidateDecodeCode},

		{"body-primitive-string-validate", PayloadBodyPrimitiveStringValidateDSL, PayloadBodyPrimitiveStringValidateDecodeCode},
		{"body-primitive-bool-validate", PayloadBodyPrimitiveBoolValidateDSL, PayloadBodyPrimitiveBoolValidateDecodeCode},
		{"body-primitive-array-string-validate", PayloadBodyPrimitiveArrayStringValidateDSL, PayloadBodyPrimitiveArrayStringValidateDecodeCode},
		{"body-primitive-array-bool-validate", PayloadBodyPrimitiveArrayBoolValidateDSL, PayloadBodyPrimitiveArrayBoolValidateDecodeCode},

		{"body-primitive-array-user-validate", PayloadBodyPrimitiveArrayUserValidateDSL, PayloadBodyPrimitiveArrayUserValidateDecodeCode},
		{"body-primitive-field-array-user", PayloadBodyPrimitiveFieldArrayUserDSL, PayloadBodyPrimitiveFieldArrayUserDecodeCode},
		{"body-primitive-field-array-user-validate", PayloadBodyPrimitiveFieldArrayUserValidateDSL, PayloadBodyPrimitiveFieldArrayUserValidateDecodeCode},

		{"body-query-object", PayloadBodyQueryObjectDSL, PayloadBodyQueryObjectDecodeCode},
		{"body-query-object-validate", PayloadBodyQueryObjectValidateDSL, PayloadBodyQueryObjectValidateDecodeCode},
		{"body-query-user", PayloadBodyQueryUserDSL, PayloadBodyQueryUserDecodeCode},
		{"body-query-user-validate", PayloadBodyQueryUserValidateDSL, PayloadBodyQueryUserValidateDecodeCode},

		{"body-path-object", PayloadBodyPathObjectDSL, PayloadBodyPathObjectDecodeCode},
		{"body-path-object-validate", PayloadBodyPathObjectValidateDSL, PayloadBodyPathObjectValidateDecodeCode},
		{"body-path-user", PayloadBodyPathUserDSL, PayloadBodyPathUserDecodeCode},
		{"body-path-user-validate", PayloadBodyPathUserValidateDSL, PayloadBodyPathUserValidateDecodeCode},

		{"body-query-path-object", PayloadBodyQueryPathObjectDSL, PayloadBodyQueryPathObjectDecodeCode},
		{"body-query-path-object-validate", PayloadBodyQueryPathObjectValidateDSL, PayloadBodyQueryPathObjectValidateDecodeCode},
		{"body-query-path-user", PayloadBodyQueryPathUserDSL, PayloadBodyQueryPathUserDecodeCode},
		{"body-query-path-user-validate", PayloadBodyQueryPathUserValidateDSL, PayloadBodyQueryPathUserValidateDecodeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			fs := Servers(rest.Root)
			if len(fs) != 2 {
				t.Fatalf("got %d files, expected two", len(fs))
			}
			sections := fs[1].Sections("")
			if len(sections) != 3 {
				t.Fatalf("got %d sections, expected 3", len(sections))
			}
			code := SectionCode(t, sections[2])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
			}
		})
	}
}
