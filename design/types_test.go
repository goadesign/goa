package design

import "testing"

func TestIsPrimitive(t *testing.T) {
	var (
		primitiveUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: Boolean,
			},
		}
		notPrimitiveUserType = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: &Object{},
			},
		}
		primitiveResultType = &ResultTypeExpr{
			UserTypeExpr: primitiveUserType,
		}
		notPrimitiveResultType = &ResultTypeExpr{
			UserTypeExpr: notPrimitiveUserType,
		}
	)
	cases := map[string]struct {
		dt       DataType
		expected bool
	}{
		"boolean": {
			dt:       Boolean,
			expected: true,
		},
		"int": {
			dt:       Int,
			expected: true,
		},
		"int32": {
			dt:       Int32,
			expected: true,
		},
		"int64": {
			dt:       Int64,
			expected: true,
		},
		"uint": {
			dt:       UInt,
			expected: true,
		},
		"uint32": {
			dt:       UInt32,
			expected: true,
		},
		"uint64": {
			dt:       UInt64,
			expected: true,
		},
		"float32": {
			dt:       Float32,
			expected: true,
		},
		"float64": {
			dt:       Float64,
			expected: true,
		},
		"string": {
			dt:       String,
			expected: true,
		},
		"bytes": {
			dt:       Bytes,
			expected: true,
		},
		"any": {
			dt:       Any,
			expected: true,
		},
		"primitive user type": {
			dt:       primitiveUserType,
			expected: true,
		},
		"not primitive user type": {
			dt:       notPrimitiveUserType,
			expected: false,
		},
		"primitive result type": {
			dt:       primitiveResultType,
			expected: true,
		},
		"not primitive result type": {
			dt:       notPrimitiveResultType,
			expected: false,
		},
		"not primitive": {
			dt:       &Object{},
			expected: false,
		},
	}

	for k, tc := range cases {
		if actual := IsPrimitive(tc.dt); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
