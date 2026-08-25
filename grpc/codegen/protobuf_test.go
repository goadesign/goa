package codegen

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestProtobufify(t *testing.T) {
	cases := []struct {
		Name       string
		String     string
		FirstUpper bool
		Acronym    bool
		Expected   string
	}{{
		"AllLower", "lower", false, false, "lower",
	}, {
		"AllLowerFirstUpper", "lower", true, false, "Lower",
	}, {
		"AllUpper", "UPPER", false, false, "uPPER",
	}, {
		"AllUpperFirstUpper", "UPPER", true, false, "UPPER",
	}, {
		"StartUpperThenLower", "Upper", false, false, "upper",
	}, {
		"StartUpperThenLowerFirstUpper", "Upper", true, false, "Upper",
	}, {
		"StartsWithUnderscore", "_foo", false, false, "foo",
	}, {
		"EndsWithUnderscore", "foo_", false, false, "foo",
	}, {
		"ContainsUnderscore", "foo_bar", false, false, "fooBar",
	}, {
		"StartsWithDigits", "123foo", false, false, "123Foo",
	}, {
		"EndsWithDigits", "foo123", false, false, "foo123",
	}, {
		"ContainsDigits", "foo123bar", false, false, "foo123Bar",
	}, {
		"ContainsIgnoredAcronym", "foo_jwt", false, false, "fooJwt",
	}, {
		"ContainsAcronym", "foo_jwt", false, true, "fooJWT",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoBufify(c.String, c.FirstUpper, c.Acronym)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestProtoNativeType(t *testing.T) {
	cases := []struct {
		Name     string
		DataType expr.DataType
		Expected string
	}{{
		"Boolean", expr.Boolean, "bool",
	}, {
		"Int", expr.Int, "sint32",
	}, {
		"Int32", expr.Int32, "sint32",
	}, {
		"Int64", expr.Int64, "sint64",
	}, {
		"UInt", expr.UInt, "uint32",
	}, {
		"UInt32", expr.UInt32, "uint32",
	}, {
		"UInt64", expr.UInt64, "uint64",
	}, {
		"Float32", expr.Float32, "float",
	}, {
		"Float64", expr.Float64, "double",
	}, {
		"String", expr.String, "string",
	}, {
		"Bytes", expr.Bytes, "bytes",
	}, {
		"Any", expr.Any, "google.protobuf.Value",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoNativeType(c.DataType)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestProtoBufNativeGoTypeName(t *testing.T) {
	cases := []struct {
		Name     string
		DataType expr.DataType
		Expected string
	}{{
		"Boolean", expr.Boolean, "bool",
	}, {
		"Int", expr.Int, "int32",
	}, {
		"Int32", expr.Int32, "int32",
	}, {
		"Int64", expr.Int64, "int64",
	}, {
		"UInt", expr.UInt, "uint32",
	}, {
		"UInt32", expr.UInt32, "uint32",
	}, {
		"UInt64", expr.UInt64, "uint64",
	}, {
		"Float32", expr.Float32, "float32",
	}, {
		"Float64", expr.Float64, "float64",
	}, {
		"String", expr.String, "string",
	}, {
		"Bytes", expr.Bytes, "[]byte",
	}, {
		"Any", expr.Any, "*structpb.Value",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoBufNativeGoTypeName(c.DataType)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestHasAnyType(t *testing.T) {
	cases := []struct {
		Name     string
		Attr     *expr.AttributeExpr
		Expected bool
	}{{
		"NoAnyType", &expr.AttributeExpr{Type: expr.String}, false,
	}, {
		"DirectAnyType", &expr.AttributeExpr{Type: expr.Any}, true,
	}, {
		"ArrayWithAnyType", &expr.AttributeExpr{
			Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Any}},
		}, true,
	}, {
		"MapWithAnyKeyType", &expr.AttributeExpr{
			Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.Any},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			},
		}, true,
	}, {
		"MapWithAnyElemType", &expr.AttributeExpr{
			Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.Any},
			},
		}, true,
	}, {
		"ObjectWithAnyField", &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{
					Name:      "field",
					Attribute: &expr.AttributeExpr{Type: expr.Any},
				},
			},
		}, true,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := hasAnyType(c.Attr)
			if got != c.Expected {
				t.Errorf("got %t, expected %t", got, c.Expected)
			}
		})
	}
}

func TestProtoBufMessageDefJSONNameOption(t *testing.T) {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{
				Name: "value",
				Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{
						"rpc:tag":        []string{"1"},
						"proto:tag:json": []string{"customValue"},
					},
				},
			},
		},
	}
	sd := &ServiceData{Scope: codegen.NewNameScope()}
	def := protoBufMessageDef(attr, sd)
	if !strings.Contains(def, `json_name = "customValue"`) {
		t.Fatalf("expected json_name option, got %q", def)
	}
}

func TestProtoBufMessageDefJSONNameOptionOneOf(t *testing.T) {
	attr := &expr.AttributeExpr{
		Type: &expr.Union{
			TypeName: "Animal",
			Values: []*expr.NamedAttributeExpr{
				{
					Name: "Cat",
					Attribute: &expr.AttributeExpr{
						Type: expr.String,
						Meta: expr.MetaExpr{
							"rpc:tag":        []string{"1"},
							"proto:tag:json": []string{"cat"},
						},
					},
				},
			},
		},
	}
	sd := &ServiceData{Scope: codegen.NewNameScope()}
	def := protoBufMessageDef(attr, sd)
	if !strings.Contains(def, `json_name = "cat"`) {
		t.Fatalf("expected json_name option in oneof, got %q", def)
	}
}

func TestProtoBufMessageDefOneOfUsesAttributeName(t *testing.T) {
	tests := []struct {
		name          string
		unionName     string
		members       []string
		wantOneof     string
		wantGoField   string
		unwantedOneof string
	}{
		{
			name:          "custom service name",
			unionName:     "subject",
			members:       []string{"equipment", "facility"},
			wantOneof:     "subject",
			wantGoField:   "Subject",
			unwantedOneof: "alarm_info_subject",
		},
		{
			name:        "member name collision",
			unionName:   "subject",
			members:     []string{"subject", "facility"},
			wantOneof:   "subject_oneof",
			wantGoField: "SubjectOneof",
		},
		{
			name:        "repeated member name collision",
			unionName:   "subject",
			members:     []string{"subject", "subject_oneof", "facility"},
			wantOneof:   "subject_oneof_oneof",
			wantGoField: "SubjectOneofOneof",
		},
		{
			name:        "generated method collision",
			unionName:   "reset",
			members:     []string{"value"},
			wantOneof:   "reset",
			wantGoField: "Reset_",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				Type("Alarm", func() {
					OneOf(test.unionName, func() {
						TypeName("AlarmInfoSubject")
						for i, member := range test.members {
							Field(i+1, member, String)
						}
					})
				})
			})
			alarm := root.UserType("Alarm")
			subject := alarm.Attribute().Find(test.unionName)
			require.Equal(t, "AlarmInfoSubject", subject.Type.Name())

			sd := &ServiceData{Scope: codegen.NewNameScope()}
			def := protoBufMessageDef(alarm.Attribute(), sd)

			require.Contains(t, def, "oneof "+test.wantOneof+" {")
			if test.unwantedOneof != "" {
				require.NotContains(t, def, "oneof "+test.unwantedOneof+" {")
			}

			proto := "syntax = \"proto3\";\npackage test;\noption go_package = \"example.com/test;test\";\nmessage Alarm" + def
			fpath := codegen.CreateTempFile(t, proto)
			require.NoError(t, protoc(defaultProtocCmd, fpath, nil))
			generated, err := os.ReadFile(fpath + ".pb.go")
			require.NoError(t, err)
			require.Regexp(
				t,
				regexp.MustCompile(`\n\t`+regexp.QuoteMeta(test.wantGoField)+`\s+isAlarm_`),
				string(generated),
			)

			source := makeProtoBufMessage(expr.DupAtt(alarm.Attribute()), alarm.Name(), sd)
			transform, _, err := protoBufTransform(
				source,
				alarm.Attribute(),
				"source",
				"target",
				protoBufTypeContext("test", sd.Scope, true),
				serviceTypeContext("test", sd.Scope),
				false,
				true,
			)
			require.NoError(t, err)
			code := codegen.FormatTestCode(t, "package test\nfunc transform() {\n"+transform+"}")
			require.Contains(t, code, "source."+test.wantGoField)
			require.NotContains(t, code, "source.AlarmInfoSubject")

			transform, _, err = protoBufTransform(
				alarm.Attribute(),
				source,
				"source",
				"target",
				serviceTypeContext("test", sd.Scope),
				protoBufTypeContext("test", sd.Scope, true),
				true,
				true,
			)
			require.NoError(t, err)
			code = codegen.FormatTestCode(t, "package test\nfunc transform() {\n"+transform+"}")
			require.Contains(t, code, "target."+test.wantGoField)
			require.NotContains(t, code, "target.AlarmInfoSubject")
		})
	}
}

func TestProtoBufMessageDefReservations(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		Type("Activation", func() {
			Meta("rpc:reserved:number", "20", "3", "15")
			Meta("rpc:reserved:name", "linked_control_point_id", "deployment_id")
			Field(1, "id", String)
		})
	})
	activation := root.UserType("Activation")
	sd := &ServiceData{Scope: codegen.NewNameScope()}
	def := protoBufMessageDef(activation.Attribute(), sd)
	require.Contains(t, def, "reserved 3, 15, 20;")
	require.Contains(t, def, `reserved "deployment_id", "linked_control_point_id";`)

	proto := "syntax = \"proto3\";\npackage test;\noption go_package = \"example.com/test;test\";\nmessage Activation" + def
	fpath := codegen.CreateTempFile(t, proto)
	require.NoError(t, protoc(defaultProtocCmd, fpath, nil))
}

func TestMakeProtoBufMessageMarksWrappers(t *testing.T) {
	cases := []struct {
		Name string
		Type func() expr.DataType
		// ElemWrapped indicates that the wrapped field is an array whose
		// element is itself a synthesized wrapper message.
		ElemWrapped bool
		// FieldKind is the expected kind of the attribute wrapped by the
		// top-level message.
		FieldKind expr.Kind
	}{{
		Name:      "primitive",
		Type:      func() expr.DataType { return expr.Int },
		FieldKind: expr.IntKind,
	}, {
		Name: "array",
		Type: func() expr.DataType {
			return &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}
		},
		FieldKind: expr.ArrayKind,
	}, {
		Name: "map",
		Type: func() expr.DataType {
			return &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.Int},
			}
		},
		FieldKind: expr.MapKind,
	}, {
		Name: "array-of-array",
		Type: func() expr.DataType {
			return &expr.Array{ElemType: &expr.AttributeExpr{
				Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}},
			}}
		},
		ElemWrapped: true,
		FieldKind:   expr.ArrayKind,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			sd := &ServiceData{Name: "Service", Scope: codegen.NewNameScope()}
			att := makeProtoBufMessage(&expr.AttributeExpr{Type: c.Type()}, "Message", sd)
			require.True(t, isWrappedAttr(att), "expected message to be marked as a wrapper")
			field := unwrapAttr(att)
			assert.Equal(t, c.FieldKind, field.Type.Kind(), "unexpected wrapped field kind")
			if c.ElemWrapped {
				elem := expr.AsArray(field.Type).ElemType
				require.True(t, isWrappedAttr(elem), "expected nested array element to be marked as a wrapper")
				inner := unwrapAttr(elem)
				assert.Equal(t, expr.ArrayKind, inner.Type.Kind(), "unexpected nested wrapped field kind")
			}
		})
	}
}

func TestUnwrapAttrPanicsOnNonWrapper(t *testing.T) {
	cases := []struct {
		Name string
		Att  *expr.AttributeExpr
	}{{
		Name: "primitive",
		Att:  &expr.AttributeExpr{Type: expr.Int},
	}, {
		Name: "object",
		Att: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "id", Attribute: &expr.AttributeExpr{Type: expr.Int}},
		}},
	}, {
		// A user type with an attribute literally named "field" but no
		// marker: only the marker set at wrapping time identifies wrappers.
		Name: "unmarked user type with field attribute",
		Att: &expr.AttributeExpr{Type: &expr.UserTypeExpr{
			TypeName: "NotAWrapper",
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "field", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			}},
		}},
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			assert.False(t, isWrappedAttr(c.Att))
			require.Panics(t, func() { unwrapAttr(c.Att) })
		})
	}
}
