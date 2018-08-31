package expr

import "testing"

func TestUserTypeExprIsCompatible(t *testing.T) {
	var (
		b = true
		i = 1
	)
	cases := map[string]struct {
		typ      DataType
		values   []interface{}
		expected bool
	}{
		"compatible": {
			typ:      Int,
			values:   []interface{}{i},
			expected: true,
		},
		"not compatible": {
			typ:      Int,
			values:   []interface{}{b},
			expected: false,
		},
		"type is nil": {
			typ:      nil,
			values:   []interface{}{b, i},
			expected: true,
		},
	}

	for k, tc := range cases {
		u := UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type: tc.typ,
			},
		}
		for _, value := range tc.values {
			if actual := u.IsCompatible(value); tc.expected != actual {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}
	}
}
