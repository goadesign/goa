package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserTypeOrigin(t *testing.T) {
	original := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "Value",
	}
	copy := original.Dup(DupAtt(original.Attribute())).(*UserTypeExpr)
	copyOfCopy := copy.Dup(DupAtt(copy.Attribute())).(*UserTypeExpr)
	require.Same(t, original, original.Origin())
	require.Same(t, original, copy.Origin())
	require.Same(t, original, copyOfCopy.Origin())

	copy.Rename("RenamedValue")
	renamedCopy := copy.Dup(DupAtt(copy.Attribute())).(*UserTypeExpr)
	require.Same(t, copy, copy.Origin())
	require.Same(t, copy, renamedCopy.Origin())
	require.Same(t, original, copyOfCopy.Origin())
}

func TestIndependentUserTypesHaveDistinctOrigins(t *testing.T) {
	first := &UserTypeExpr{AttributeExpr: &AttributeExpr{Type: String}, TypeName: "Value"}
	second := &UserTypeExpr{AttributeExpr: &AttributeExpr{Type: String}, TypeName: "Value"}
	require.NotSame(t, first.Origin(), second.Origin())
}

func TestResultTypeOriginPreservesDynamicType(t *testing.T) {
	original := NewResultTypeExpr("Value", "application/vnd.value", nil)
	copy := original.Dup(DupAtt(original.Attribute())).(*ResultTypeExpr)
	require.Same(t, original, copy.Origin())

	copy.Rename("RenamedValue")
	renamedCopy := copy.Dup(DupAtt(copy.Attribute())).(*ResultTypeExpr)
	require.Same(t, copy, copy.Origin())
	require.Same(t, copy, renamedCopy.Origin())
}

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
		values   []any
		expected bool
	}{
		"compatible": {
			typ:      Int,
			values:   []any{i},
			expected: true,
		},
		"not compatible": {
			typ:      Int,
			values:   []any{b},
			expected: false,
		},
		"type is nil": {
			typ:      nil,
			values:   []any{b, i},
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
