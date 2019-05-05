package expr_test

import (
	"testing"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestServiceExprMethod(t *testing.T) {
	var (
		methodFoo = &expr.MethodExpr{
			Name: "foo",
		}
		methodBar = &expr.MethodExpr{
			Name: "bar",
		}
	)
	cases := map[string]struct {
		name     string
		expected *expr.MethodExpr
	}{
		"exist": {
			name:     "foo",
			expected: methodFoo,
		},
		"not exist": {
			name:     "baz",
			expected: nil,
		},
	}

	for k, tc := range cases {
		s := expr.ServiceExpr{
			Methods: []*expr.MethodExpr{
				methodFoo,
				methodBar,
			},
		}
		if actual := s.Method(tc.name); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestServiceExprError(t *testing.T) {
	var (
		errorFoo = &expr.ErrorExpr{
			Name: "foo",
		}
		errorBar = &expr.ErrorExpr{
			Name: "bar",
		}
	)
	cases := map[string]struct {
		name     string
		expected *expr.ErrorExpr
	}{
		"exist in service": {
			name:     "foo",
			expected: errorFoo,
		},
		"exist in root": {
			name:     "bar",
			expected: errorBar,
		},
		"not exist": {
			name:     "qux",
			expected: nil,
		},
	}

	expr.Root.Errors = []*expr.ErrorExpr{
		errorBar,
	}
	s := expr.ServiceExpr{
		Errors: []*expr.ErrorExpr{
			errorFoo,
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := s.Error(tc.name); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}

func TestServiceExprValidate(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"service errors", testdata.ServiceErrorDSL, `attribute: attribute "a" with 'struct:error:name' in the meta must be required in "ServiceError" type`},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, tc.DSL)
			if tc.Error != err.Error() {
				t.Errorf("invalid error:\ngot:\n%s\n\ngot vs expected:\n%s", err.Error(), expr.Diff(t, err.Error(), tc.Error))
			}
		})
	}
}

func TestErrorExprValidate(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"no error", testdata.ValidErrorsDSL, ""},
		{"invalid-struct-error-name-meta", testdata.InvalidStructErrorNameDSL,
			`attribute: attribute "a" with 'struct:error:name' in the meta must be required in "ServiceError" type
attribute: attribute "a" has 'struct:error:name' meta which is already set for attribute "b" in "Error" type
attribute: attribute "a" with 'struct:error:name' in the meta must be a string in "Error" type
attribute: attribute "a" with 'struct:error:name' in the meta must be required in "Error" type
attribute: type "ErrorType" is used to define multiple errors and must identify the attribute containing error name. Use Meta with the key 'struct:error:name' on the error name attribute
attribute: type "ErrorType" is used to define multiple errors and must identify the attribute containing error name. Use Meta with the key 'struct:error:name' on the error name attribute`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Error == "" {
				expr.RunDSL(t, tc.DSL)
			} else {
				err := expr.RunInvalidDSL(t, tc.DSL)
				if tc.Error != err.Error() {
					t.Errorf("invalid error:\ngot:\n%s\n\ngot vs expected:\n%s", err.Error(), expr.Diff(t, err.Error(), tc.Error))
				}
			}
		})
	}
}
