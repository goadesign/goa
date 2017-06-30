package rest

import (
	"testing"

	. "goa.design/goa.v2/codegen/files/rest/testing"
	. "goa.design/goa.v2/codegen/testing"
	restdesign "goa.design/goa.v2/design/rest"
)

func TestEncode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"header-bool", ResultHeaderBoolDSL, ResultHeaderBoolEncodeCode},
		{"header-int", ResultHeaderIntDSL, ResultHeaderIntEncodeCode},
		{"header-int32", ResultHeaderInt32DSL, ResultHeaderInt32EncodeCode},
		{"header-int64", ResultHeaderInt64DSL, ResultHeaderInt64EncodeCode},
		{"header-uint", ResultHeaderUIntDSL, ResultHeaderUIntEncodeCode},
		{"header-uint32", ResultHeaderUInt32DSL, ResultHeaderUInt32EncodeCode},
		{"header-uint64", ResultHeaderUInt64DSL, ResultHeaderUInt64EncodeCode},
		{"header-float32", ResultHeaderFloat32DSL, ResultHeaderFloat32EncodeCode},
		{"header-float64", ResultHeaderFloat64DSL, ResultHeaderFloat64EncodeCode},
		{"header-string", ResultHeaderStringDSL, ResultHeaderStringEncodeCode},
		{"header-bytes", ResultHeaderBytesDSL, ResultHeaderBytesEncodeCode},
		{"header-any", ResultHeaderAnyDSL, ResultHeaderAnyEncodeCode},
		{"header-array-bool", ResultHeaderArrayBoolDSL, ResultHeaderArrayBoolEncodeCode},
		{"header-array-int", ResultHeaderArrayIntDSL, ResultHeaderArrayIntEncodeCode},
		{"header-array-int32", ResultHeaderArrayInt32DSL, ResultHeaderArrayInt32EncodeCode},
		{"header-array-int64", ResultHeaderArrayInt64DSL, ResultHeaderArrayInt64EncodeCode},
		{"header-array-uint", ResultHeaderArrayUIntDSL, ResultHeaderArrayUIntEncodeCode},
		{"header-array-uint32", ResultHeaderArrayUInt32DSL, ResultHeaderArrayUInt32EncodeCode},
		{"header-array-uint64", ResultHeaderArrayUInt64DSL, ResultHeaderArrayUInt64EncodeCode},
		{"header-array-float32", ResultHeaderArrayFloat32DSL, ResultHeaderArrayFloat32EncodeCode},
		{"header-array-float64", ResultHeaderArrayFloat64DSL, ResultHeaderArrayFloat64EncodeCode},
		{"header-array-string", ResultHeaderArrayStringDSL, ResultHeaderArrayStringEncodeCode},
		{"header-array-bytes", ResultHeaderArrayBytesDSL, ResultHeaderArrayBytesEncodeCode},
		{"header-array-any", ResultHeaderArrayAnyDSL, ResultHeaderArrayAnyEncodeCode},

		{"header-bool-default", ResultHeaderBoolDefaultDSL, ResultHeaderBoolDefaultEncodeCode},
		{"header-bool-required-default", ResultHeaderBoolRequiredDefaultDSL, ResultHeaderBoolRequiredDefaultEncodeCode},
		{"header-string-default", ResultHeaderStringDefaultDSL, ResultHeaderStringDefaultEncodeCode},
		{"header-string-required-default", ResultHeaderStringRequiredDefaultDSL, ResultHeaderStringRequiredDefaultEncodeCode},
		{"header-array-bool-default", ResultHeaderArrayBoolDefaultDSL, ResultHeaderArrayBoolDefaultEncodeCode},
		{"header-array-bool-required-default", ResultHeaderArrayBoolRequiredDefaultDSL, ResultHeaderArrayBoolRequiredDefaultEncodeCode},
		{"header-array-string-default", ResultHeaderArrayStringDefaultDSL, ResultHeaderArrayStringDefaultEncodeCode},
		{"header-array-string-required-default", ResultHeaderArrayStringRequiredDefaultDSL, ResultHeaderArrayStringRequiredDefaultEncodeCode},

		{"body-string", ResultBodyStringDSL, ResultBodyStringEncodeCode},
		{"body-object", ResultBodyObjectDSL, ResultBodyObjectEncodeCode},
		{"body-user", ResultBodyUserDSL, ResultBodyUserEncodeCode},
		{"body-array-string", ResultBodyArrayStringDSL, ResultBodyArrayStringEncodeCode},
		{"body-array-user", ResultBodyArrayUserDSL, ResultBodyArrayUserEncodeCode},

		{"body-primitive-string", ResultBodyPrimitiveStringDSL, ResultBodyPrimitiveStringEncodeCode},
		{"body-primitive-bool", ResultBodyPrimitiveBoolDSL, ResultBodyPrimitiveBoolEncodeCode},
		{"body-primitive-array-string", ResultBodyPrimitiveArrayStringDSL, ResultBodyPrimitiveArrayStringEncodeCode},
		{"body-primitive-array-bool", ResultBodyPrimitiveArrayBoolDSL, ResultBodyPrimitiveArrayBoolEncodeCode},

		{"body-header-object", ResultBodyHeaderObjectDSL, ResultBodyHeaderObjectEncodeCode},
		{"body-header-user", ResultBodyHeaderUserDSL, ResultBodyHeaderUserEncodeCode},

		{"tag-string", ResultTagStringDSL, ResultTagStringEncodeCode},
		{"tag-string-required", ResultTagStringRequiredDSL, ResultTagStringRequiredEncodeCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			fs := Servers(restdesign.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].Sections("")
			if len(sections) != 7 {
				t.Fatalf("got %d sections, expected 7", len(sections))
			}
			code := SectionCode(t, sections[6])
			if code != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
			}
		})
	}
}
