package expr

import (
	"reflect"
	"testing"
)

func TestServerExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo": {name: "foo", expected: "Server foo"},
	}

	for k, tc := range cases {
		server := ServerExpr{Name: tc.name}
		if actual := server.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestHostExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name       string
		serverName string
		expected   string
	}{
		"foo": {name: "foo", serverName: "bar", expected: `host "foo" of server "bar"`},
	}

	for k, tc := range cases {
		host := HostExpr{Name: tc.name, ServerName: tc.serverName}
		if actual := host.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestHostExprAttribute(t *testing.T) {
	cases := map[string]struct {
		attributeExpr *AttributeExpr
		expected      *AttributeExpr
	}{
		"nil": {
			attributeExpr: nil,
			expected:      &AttributeExpr{Type: &Object{}},
		},
		"non-nil": {
			attributeExpr: &AttributeExpr{Description: "foo"},
			expected:      &AttributeExpr{Description: "foo"},
		},
	}

	for k, tc := range cases {
		host := HostExpr{Variables: tc.attributeExpr}
		actual := host.Attribute()

		actualType := reflect.TypeOf(actual)
		expectedValue := reflect.ValueOf(tc.expected)

		if !reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual) {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
