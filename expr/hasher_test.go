package expr

import "testing"

func TestPrimitiveHash(t *testing.T) {
	cases := map[string]Primitive{
		"boolean": Boolean,
		"int":     Int,
		"int32":   Int32,
		"int64":   Int64,
		"uint":    UInt,
		"uint32":  UInt32,
		"uint64":  UInt64,
		"float32": Float32,
		"float64": Float64,
		"string":  String,
		"bytes":   Bytes,
		"any":     Any,
	}
	for k, p := range cases {
		if actual := Hash(p, true, false, false); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, true, false, true); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, true, true, false); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, true, true, true); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, false, false, false); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, false, false, true); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, false, true, false); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
		if actual := Hash(p, false, true, true); k != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, k)
		}
	}
}

func TestObjectHash(t *testing.T) {
	var (
		attributeInt            = &AttributeExpr{Type: Int}
		attributeString         = &AttributeExpr{Type: String}
		attributeArray          = &AttributeExpr{Type: &Array{ElemType: attributeString}}
		attributeMap            = &AttributeExpr{Type: &Map{KeyType: attributeInt, ElemType: attributeString}}
		userType                = &UserTypeExpr{AttributeExpr: attributeString, TypeName: "quux"}
		namedAttributePrimitive = &NamedAttributeExpr{Name: "foo", Attribute: attributeInt}
		namedAttributeArray     = &NamedAttributeExpr{Name: "bar", Attribute: attributeArray}
		namedAttributeMap       = &NamedAttributeExpr{Name: "baz", Attribute: attributeMap}
		namedAttributeUserType  = &NamedAttributeExpr{Name: "qux", Attribute: &AttributeExpr{Type: userType}}
	)
	cases := map[string]struct {
		object   Object
		expected string
	}{
		"nil": {
			object:   nil,
			expected: "_o_",
		},
		"single attribute": {
			object:   Object{namedAttributePrimitive},
			expected: "_o_-foo/int",
		},
		"multiple attributes": {
			object: Object{
				namedAttributePrimitive,
				namedAttributeArray,
				namedAttributeMap,
				namedAttributeUserType,
			},
			expected: "_o_-bar/_a_string-baz/_m_int:string-foo/int-qux/_u_quux",
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := Hash(&tc.object, true, false, true); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}
