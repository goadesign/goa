package expr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
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

func TestEquivalentInlineMethodErrorsShareOrigin(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("secured", func() {
			for _, method := range []string{"read", "write"} {
				dsl.Method(method, func() {
					dsl.Error("invalid_scopes", dsl.String)
				})
			}
		})
	})
	service := root.Service("secured")
	first := service.Method("read").Error("invalid_scopes").Type.(expr.UserType)
	second := service.Method("write").Error("invalid_scopes").Type.(expr.UserType)

	require.NotSame(t, first, second)
	require.Same(t, first.Origin(), second.Origin())
	firstIdentity, ok := expr.GeneratedUserTypeExampleIdentity(first)
	require.True(t, ok)
	secondIdentity, ok := expr.GeneratedUserTypeExampleIdentity(second)
	require.True(t, ok)
	require.NotEqual(t, firstIdentity, secondIdentity)
}

func TestIncompatibleInlineMethodErrorsAreRejected(t *testing.T) {
	expr.ResetDSL(t)
	design := func() {
		dsl.Service("secured", func() {
			dsl.Method("read", func() {
				dsl.Error("invalid_scopes", dsl.String)
			})
			dsl.Method("write", func() {
				dsl.Error("invalid_scopes", dsl.Int)
			})
		})
	}
	require.True(t, eval.Execute(design, nil))
	err := eval.RunDSL()
	require.ErrorContains(t, err, `inline error "invalid_scopes" must define the same value contract in every method of service "secured"`)
}

func TestRepeatedStandardErrorsMustUseSameQualifiers(t *testing.T) {
	qualifiers := []struct {
		name  string
		apply func()
	}{
		{name: "temporary", apply: dsl.Temporary},
		{name: "timeout", apply: func() { dsl.Timeout() }},
		{name: "fault", apply: dsl.Fault},
	}
	for _, qualifier := range qualifiers {
		t.Run("service reference inherits "+qualifier.name, func(t *testing.T) {
			root := expr.RunDSL(t, func() {
				dsl.Service("jobs", func() {
					dsl.Error("busy", qualifier.apply)
					dsl.Method("run", func() {
						dsl.Error("busy")
					})
				})
			})
			service := root.Service("jobs")
			require.Same(t, service.Error("busy"), service.Method("run").Error("busy"))
		})

		t.Run("two methods "+qualifier.name, func(t *testing.T) {
			expr.ResetDSL(t)
			design := func() {
				dsl.Service("jobs", func() {
					dsl.Method("start", func() {
						dsl.Error("busy", qualifier.apply)
					})
					dsl.Method("resume", func() {
						dsl.Error("busy")
					})
				})
			}
			require.True(t, eval.Execute(design, nil))
			err := eval.RunDSL()
			require.ErrorContains(t, err, qualifier.name+" setting differs")
		})

		t.Run("matching "+qualifier.name, func(t *testing.T) {
			expr.RunDSL(t, func() {
				dsl.Service("jobs", func() {
					dsl.Error("busy", qualifier.apply)
					dsl.Method("run", func() {
						dsl.Error("busy", qualifier.apply)
					})
				})
			})
		})
	}
}

func TestAPIErrorDoesNotAddErrorsToServicesOrMethods(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.API("jobs", func() {
			dsl.Error("busy", dsl.Temporary)
		})
		dsl.Service("jobs", func() {
			dsl.Method("run", func() {})
			dsl.Method("retry", func() {
				dsl.Error("busy")
			})
		})
	})
	service := root.Service("jobs")

	require.Empty(t, service.Errors)
	require.Empty(t, service.Method("run").Errors)
	require.Equal(t, []*expr.ErrorExpr{root.Error("busy")}, service.Method("retry").Errors)
}

func TestRepeatedAuthoredErrorTypesDoNotShareGeneratedConstructors(t *testing.T) {
	custom := dsl.Type("CustomError", func() {
		dsl.Attribute("message", dsl.String)
	})
	expr.RunDSL(t, func() {
		dsl.Service("jobs", func() {
			dsl.Error("busy", custom, dsl.Temporary)
			dsl.Method("run", func() {
				dsl.Error("busy", custom)
			})
		})
	})
}

func TestServiceExprError(t *testing.T) {
	var (
		errorFoo = &expr.ErrorExpr{
			Name: "foo",
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
		"not exist": {
			name:     "qux",
			expected: nil,
		},
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
		{"service errors", testdata.ServiceErrorDSL, `attribute: error name "a" must be required in type "ServiceError"`},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, tc.DSL)
			assert.EqualError(t, err, tc.Error)
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
			`attribute: error name "a" must be required in type "ServiceError"
attribute: duplicate error names in type "Error"
attribute: error name "a" must be a string in type "Error"
attribute: error name "a" must be required in type "Error"
attribute: type "ErrorType" is used to define multiple errors and must identify the attribute containing the error name with ErrorName
attribute: type "ErrorType" is used to define multiple errors and must identify the attribute containing the error name with ErrorName`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Error == "" {
				expr.RunDSL(t, tc.DSL)
			} else {
				err := expr.RunInvalidDSL(t, tc.DSL)
				assert.EqualError(t, err, tc.Error)
			}
		})
	}
}
