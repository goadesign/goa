package rest

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"goa.design/goa.v2/codegen"
	. "goa.design/goa.v2/codegen/files/rest/testing"
	. "goa.design/goa.v2/codegen/testing"
	restdesign "goa.design/goa.v2/design/rest"
	. "goa.design/goa.v2/design/rest/testing"
)

func TestPayloadConstructor(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"query-bool", PayloadQueryBoolDSL, PayloadQueryBoolConstructorCode},
		{"query-bool-validate", PayloadQueryBoolValidateDSL, PayloadQueryBoolValidateConstructorCode},
		{"query-int", PayloadQueryIntDSL, PayloadQueryIntConstructorCode},
		{"query-int-validate", PayloadQueryIntValidateDSL, PayloadQueryIntValidateConstructorCode},
		{"query-int32", PayloadQueryInt32DSL, PayloadQueryInt32ConstructorCode},
		{"query-int32-validate", PayloadQueryInt32ValidateDSL, PayloadQueryInt32ValidateConstructorCode},
		{"query-int64", PayloadQueryInt64DSL, PayloadQueryInt64ConstructorCode},
		{"query-int64-validate", PayloadQueryInt64ValidateDSL, PayloadQueryInt64ValidateConstructorCode},
		{"query-uint", PayloadQueryUIntDSL, PayloadQueryUIntConstructorCode},
		{"query-uint-validate", PayloadQueryUIntValidateDSL, PayloadQueryUIntValidateConstructorCode},
		{"query-uint32", PayloadQueryUInt32DSL, PayloadQueryUInt32ConstructorCode},
		{"query-uint32-validate", PayloadQueryUInt32ValidateDSL, PayloadQueryUInt32ValidateConstructorCode},
		{"query-uint64", PayloadQueryUInt64DSL, PayloadQueryUInt64ConstructorCode},
		{"query-uint64-validate", PayloadQueryUInt64ValidateDSL, PayloadQueryUInt64ValidateConstructorCode},
		{"query-float32", PayloadQueryFloat32DSL, PayloadQueryFloat32ConstructorCode},
		{"query-float32-validate", PayloadQueryFloat32ValidateDSL, PayloadQueryFloat32ValidateConstructorCode},
		{"query-float64", PayloadQueryFloat64DSL, PayloadQueryFloat64ConstructorCode},
		{"query-float64-validate", PayloadQueryFloat64ValidateDSL, PayloadQueryFloat64ValidateConstructorCode},
		{"query-string", PayloadQueryStringDSL, PayloadQueryStringConstructorCode},
		{"query-string-validate", PayloadQueryStringValidateDSL, PayloadQueryStringValidateConstructorCode},
		{"query-bytes", PayloadQueryBytesDSL, PayloadQueryBytesConstructorCode},
		{"query-bytes-validate", PayloadQueryBytesValidateDSL, PayloadQueryBytesValidateConstructorCode},
		{"query-any", PayloadQueryAnyDSL, PayloadQueryAnyConstructorCode},
		{"query-any-validate", PayloadQueryAnyValidateDSL, PayloadQueryAnyValidateConstructorCode},
		{"query-array-bool", PayloadQueryArrayBoolDSL, PayloadQueryArrayBoolConstructorCode},
		{"query-array-bool-validate", PayloadQueryArrayBoolValidateDSL, PayloadQueryArrayBoolValidateConstructorCode},
		{"query-array-int", PayloadQueryArrayIntDSL, PayloadQueryArrayIntConstructorCode},
		{"query-array-int-validate", PayloadQueryArrayIntValidateDSL, PayloadQueryArrayIntValidateConstructorCode},
		{"query-array-int32", PayloadQueryArrayInt32DSL, PayloadQueryArrayInt32ConstructorCode},
		{"query-array-int32-validate", PayloadQueryArrayInt32ValidateDSL, PayloadQueryArrayInt32ValidateConstructorCode},
		{"query-array-int64", PayloadQueryArrayInt64DSL, PayloadQueryArrayInt64ConstructorCode},
		{"query-array-int64-validate", PayloadQueryArrayInt64ValidateDSL, PayloadQueryArrayInt64ValidateConstructorCode},
		{"query-array-uint", PayloadQueryArrayUIntDSL, PayloadQueryArrayUIntConstructorCode},
		{"query-array-uint-validate", PayloadQueryArrayUIntValidateDSL, PayloadQueryArrayUIntValidateConstructorCode},
		{"query-array-uint32", PayloadQueryArrayUInt32DSL, PayloadQueryArrayUInt32ConstructorCode},
		{"query-array-uint32-validate", PayloadQueryArrayUInt32ValidateDSL, PayloadQueryArrayUInt32ValidateConstructorCode},
		{"query-array-uint64", PayloadQueryArrayUInt64DSL, PayloadQueryArrayUInt64ConstructorCode},
		{"query-array-uint64-validate", PayloadQueryArrayUInt64ValidateDSL, PayloadQueryArrayUInt64ValidateConstructorCode},
		{"query-array-float32", PayloadQueryArrayFloat32DSL, PayloadQueryArrayFloat32ConstructorCode},
		{"query-array-float32-validate", PayloadQueryArrayFloat32ValidateDSL, PayloadQueryArrayFloat32ValidateConstructorCode},
		{"query-array-float64", PayloadQueryArrayFloat64DSL, PayloadQueryArrayFloat64ConstructorCode},
		{"query-array-float64-validate", PayloadQueryArrayFloat64ValidateDSL, PayloadQueryArrayFloat64ValidateConstructorCode},
		{"query-array-string", PayloadQueryArrayStringDSL, PayloadQueryArrayStringConstructorCode},
		{"query-array-string-validate", PayloadQueryArrayStringValidateDSL, PayloadQueryArrayStringValidateConstructorCode},
		{"query-array-bytes", PayloadQueryArrayBytesDSL, PayloadQueryArrayBytesConstructorCode},
		{"query-array-bytes-validate", PayloadQueryArrayBytesValidateDSL, PayloadQueryArrayBytesValidateConstructorCode},
		{"query-array-any", PayloadQueryArrayAnyDSL, PayloadQueryArrayAnyConstructorCode},
		{"query-array-any-validate", PayloadQueryArrayAnyValidateDSL, PayloadQueryArrayAnyValidateConstructorCode},
		{"query-map-string-string", PayloadQueryMapStringStringDSL, PayloadQueryMapStringStringConstructorCode},
		{"query-map-string-string-validate", PayloadQueryMapStringStringValidateDSL, PayloadQueryMapStringStringValidateConstructorCode},
		{"query-map-string-bool", PayloadQueryMapStringBoolDSL, PayloadQueryMapStringBoolConstructorCode},
		{"query-map-string-bool-validate", PayloadQueryMapStringBoolValidateDSL, PayloadQueryMapStringBoolValidateConstructorCode},
		{"query-map-bool-string", PayloadQueryMapBoolStringDSL, PayloadQueryMapBoolStringConstructorCode},
		{"query-map-bool-string-validate", PayloadQueryMapBoolStringValidateDSL, PayloadQueryMapBoolStringValidateConstructorCode},
		{"query-map-bool-bool", PayloadQueryMapBoolBoolDSL, PayloadQueryMapBoolBoolConstructorCode},
		{"query-map-bool-bool-validate", PayloadQueryMapBoolBoolValidateDSL, PayloadQueryMapBoolBoolValidateConstructorCode},
		{"query-map-string-array-string", PayloadQueryMapStringArrayStringDSL, PayloadQueryMapStringArrayStringConstructorCode},
		{"query-map-string-array-string-validate", PayloadQueryMapStringArrayStringValidateDSL, PayloadQueryMapStringArrayStringValidateConstructorCode},
		{"query-map-string-array-bool", PayloadQueryMapStringArrayBoolDSL, PayloadQueryMapStringArrayBoolConstructorCode},
		{"query-map-string-array-bool-validate", PayloadQueryMapStringArrayBoolValidateDSL, PayloadQueryMapStringArrayBoolValidateConstructorCode},
		{"query-map-bool-array-string", PayloadQueryMapBoolArrayStringDSL, PayloadQueryMapBoolArrayStringConstructorCode},
		{"query-map-bool-array-string-validate", PayloadQueryMapBoolArrayStringValidateDSL, PayloadQueryMapBoolArrayStringValidateConstructorCode},
		{"query-map-bool-array-bool", PayloadQueryMapBoolArrayBoolDSL, PayloadQueryMapBoolArrayBoolConstructorCode},
		{"query-map-bool-array-bool-validate", PayloadQueryMapBoolArrayBoolValidateDSL, PayloadQueryMapBoolArrayBoolValidateConstructorCode},

		{"path-string", PayloadPathStringDSL, PayloadPathStringConstructorCode},
		{"path-string-validate", PayloadPathStringValidateDSL, PayloadPathStringValidateConstructorCode},
		{"path-array-string", PayloadPathArrayStringDSL, PayloadPathArrayStringConstructorCode},
		{"path-array-string-validate", PayloadPathArrayStringValidateDSL, PayloadPathArrayStringValidateConstructorCode},

		{"header-string", PayloadHeaderStringDSL, PayloadHeaderStringConstructorCode},
		{"header-string-validate", PayloadHeaderStringValidateDSL, PayloadHeaderStringValidateConstructorCode},
		{"header-array-string", PayloadHeaderArrayStringDSL, PayloadHeaderArrayStringConstructorCode},
		{"header-array-string-validate", PayloadHeaderArrayStringValidateDSL, PayloadHeaderArrayStringValidateConstructorCode},

		{"body-query-object", PayloadBodyQueryObjectDSL, PayloadBodyQueryObjectConstructorCode},
		{"body-query-object-validate", PayloadBodyQueryObjectValidateDSL, PayloadBodyQueryObjectValidateConstructorCode},
		{"body-query-user", PayloadBodyQueryUserDSL, PayloadBodyQueryUserConstructorCode},
		{"body-query-user-validate", PayloadBodyQueryUserValidateDSL, PayloadBodyQueryUserValidateConstructorCode},

		{"body-path-object", PayloadBodyPathObjectDSL, PayloadBodyPathObjectConstructorCode},
		{"body-path-object-validate", PayloadBodyPathObjectValidateDSL, PayloadBodyPathObjectValidateConstructorCode},
		{"body-path-user", PayloadBodyPathUserDSL, PayloadBodyPathUserConstructorCode},
		{"body-path-user-validate", PayloadBodyPathUserValidateDSL, PayloadBodyPathUserValidateConstructorCode},

		{"body-query-path-object", PayloadBodyQueryPathObjectDSL, PayloadBodyQueryPathObjectConstructorCode},
		{"body-query-path-object-validate", PayloadBodyQueryPathObjectValidateDSL, PayloadBodyQueryPathObjectValidateConstructorCode},
		{"body-query-path-user", PayloadBodyQueryPathUserDSL, PayloadBodyQueryPathUserConstructorCode},
		{"body-query-path-user-validate", PayloadBodyQueryPathUserValidateDSL, PayloadBodyQueryPathUserValidateConstructorCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			if len(restdesign.Root.Resources) != 1 {
				t.Fatalf("got %d file(s), expected 1", len(restdesign.Root.Resources))
			}
			fs := MarshalTypes(restdesign.Root.Resources[0])
			sections := fs.Sections("")
			if len(sections) < 2 {
				t.Fatalf("got %d sections, expected at least 2", len(sections))
			}
			var code bytes.Buffer
			if err := sections[1].Write(&code); err != nil {
				t.Fatal(err)
			}

			// format code...
			tmp, err := ioutil.TempFile("", "types_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmp.Name())
			_, err = tmp.WriteString("package foo\n" + code.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := tmp.Close(); err != nil {
				t.Fatal(err)
			}
			if err := codegen.Format(tmp.Name()); err != nil {
				t.Fatal(err)
			}
			content, err := ioutil.ReadFile(tmp.Name())
			if err != nil {
				t.Fatal(err)
			}
			formatted := strings.Join(strings.Split(string(content), "\n")[2:], "\n")

			// ... and compare
			if formatted != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", formatted, Diff(t, formatted, c.Code))
			}
		})
	}
}
