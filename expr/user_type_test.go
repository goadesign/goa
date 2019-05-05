package expr

import "testing"

func TestUserTypeExprName(t *testing.T) {
	var (
		userTypeExprWithoutAttribute = UserTypeExpr{
			TypeName: "foo",
		}
		userTypeExprHasMeta = UserTypeExpr{
			TypeName: "foo",
			AttributeExpr: &AttributeExpr{
				Meta: MetaExpr{
					"struct:type:name": []string{"bar"},
				},
			},
		}
		userTypeExprHasAnotherMeta = UserTypeExpr{
			TypeName: "foo",
			AttributeExpr: &AttributeExpr{
				Meta: MetaExpr{
					"struct:field:name": []string{"baz"},
				},
			},
		}
	)
	cases := map[string]struct {
		userType UserTypeExpr
		expected string
	}{
		"attribute in user type is nill": {
			userType: userTypeExprWithoutAttribute,
			expected: "foo",
		},
		"user type has meta": {
			userType: userTypeExprHasMeta,
			expected: "bar",
		},
		"user type has another meta": {
			userType: userTypeExprHasAnotherMeta,
			expected: "foo",
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := tc.userType.Name(); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}

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
