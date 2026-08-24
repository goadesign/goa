package expr_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestMethodExprValidate(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"valid-security-schemes-extend", testdata.ValidSecuritySchemesExtendDSL, ""},
		{"invalid-security-schemes", testdata.InvalidSecuritySchemesDSL,
			`service "InvalidSecuritySchemesService" method "SecureMethod": payload of method "SecureMethod" of service "InvalidSecuritySchemesService" does not define a username attribute, use Username to define one
service "InvalidSecuritySchemesService" method "SecureMethod": payload of method "SecureMethod" of service "InvalidSecuritySchemesService" does not define a password attribute, use Password to define one
service "InvalidSecuritySchemesService" method "SecureMethod": payload of method "SecureMethod" of service "InvalidSecuritySchemesService" does not define a Bearer token attribute, use BearerToken to define one
service "InvalidSecuritySchemesService" method "SecureMethod": payload of method "SecureMethod" of service "InvalidSecuritySchemesService" does not define a JWT attribute, use Token to define one
service "InvalidSecuritySchemesService" method "SecureMethod": security scope "not:found" not found in any of the security schemes.
flow authorization_code: invalid token URL "^example:/token<>": parse "^example:/token<>": first path segment in URL cannot contain colon
flow authorization_code: invalid authorization URL "http://^authorization": parse "http://^authorization": invalid character "^" in host name
flow authorization_code: invalid refresh URL "http://refresh^": parse "http://refresh^": invalid character "^" in host name
service "InvalidSecuritySchemesService" method "InheritedSecureMethod": payload of method "InheritedSecureMethod" of service "InvalidSecuritySchemesService" does not define a OAuth2 access token attribute, use AccessToken to define one
service "InvalidSecuritySchemesService" method "InheritedSecureMethod": payload of method "InheritedSecureMethod" of service "InvalidSecuritySchemesService" does not define an API key attribute, use APIKey to define one
service "InvalidSecuritySchemesService" method "InheritedSecureMethod": security scope "not:found" not found in any of the security schemes.
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines a username attribute, but no basic auth security scheme exist
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines a password attribute, but no basic auth security scheme exist
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines an API key attribute, but no APIKey security scheme exist
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines a Bearer token attribute, but no Bearer security scheme exist
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines a JWT token attribute, but no JWT auth security scheme exist
service "AnotherInvalidSecuritySchemesService" method "Method": payload of method "Method" of service "AnotherInvalidSecuritySchemesService" defines a OAuth2 access token attribute, but no OAuth2 security scheme exist`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Error == "" {
				expr.RunDSL(t, tc.DSL)
			} else {
				err := expr.RunInvalidDSL(t, tc.DSL)
				assert.Equal(t, tc.Error, stripValidationLocations(err.Error()))
			}
		})
	}
}

func TestMethodExprFinalizeInheritsBearerFormat(t *testing.T) {
	cases := []struct {
		name       string
		kind       expr.SchemeKind
		setup      func(*expr.SecurityExpr) (*expr.APIExpr, *expr.ServiceExpr)
		schemeName string
	}{
		{
			name: "service security",
			kind: expr.BearerKind,
			setup: func(req *expr.SecurityExpr) (*expr.APIExpr, *expr.ServiceExpr) {
				return &expr.APIExpr{}, &expr.ServiceExpr{
					Name:         "Service",
					Requirements: []*expr.SecurityExpr{req},
				}
			},
			schemeName: "bearer",
		},
		{
			name: "API security",
			kind: expr.JWTKind,
			setup: func(req *expr.SecurityExpr) (*expr.APIExpr, *expr.ServiceExpr) {
				return &expr.APIExpr{
					Requirements: []*expr.SecurityExpr{req},
				}, &expr.ServiceExpr{Name: "Service"}
			},
			schemeName: "jwt",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := &expr.SchemeExpr{
				Kind:         tc.kind,
				SchemeName:   tc.schemeName,
				BearerFormat: "JWT",
				In:           "header",
				Name:         "Authorization",
			}
			req := &expr.SecurityExpr{Schemes: []*expr.SchemeExpr{scheme}}
			api, service := tc.setup(req)
			designAPI := expr.NewAPIExpr("test", func() {})
			designAPI.Requirements = api.Requirements
			design := &expr.RootExpr{API: designAPI, Services: []*expr.ServiceExpr{service}}
			design.WalkSets(func(eval.ExpressionSet) {})

			method := &expr.MethodExpr{
				Name:    "Method",
				Service: service,
			}
			method.Finalize()

			require.Len(t, method.Requirements, 1)
			require.Len(t, method.Requirements[0].Schemes, 1)
			inherited := method.Requirements[0].Schemes[0]
			require.NotSame(t, scheme, inherited)
			require.Equal(t, "JWT", inherited.BearerFormat)
		})
	}
}

func TestMethodExprFinalizePreservesNoSecurityMarker(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Service"}
	api := expr.NewAPIExpr("test", func() {})
	api.Requirements = []*expr.SecurityExpr{{
		Schemes: []*expr.SchemeExpr{{
			Kind:       expr.JWTKind,
			SchemeName: "jwt",
		}},
	}}
	design := &expr.RootExpr{API: api, Services: []*expr.ServiceExpr{service}}
	design.WalkSets(func(eval.ExpressionSet) {})
	method := &expr.MethodExpr{
		Name: "Health",
		Requirements: []*expr.SecurityExpr{{
			Schemes: []*expr.SchemeExpr{{Kind: expr.NoKind}},
		}},
		Service: service,
	}

	method.Finalize()

	require.True(t, expr.HasNoSecurity(method.Requirements))
	require.Len(t, method.Requirements, 1)
	require.Len(t, method.Requirements[0].Schemes, 1)
	require.Equal(t, expr.NoKind, method.Requirements[0].Schemes[0].Kind)
	require.Nil(t, expr.EffectiveSecurityRequirements(method.Requirements))
}

func TestMethodExprError(t *testing.T) {
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
		"exist in method": {
			name:     "foo",
			expected: errorFoo,
		},
		"exist in service": {
			name:     "bar",
			expected: errorBar,
		},
		"not exist": {
			name:     "qux",
			expected: nil,
		},
	}

	s := expr.ServiceExpr{
		Errors: []*expr.ErrorExpr{
			errorBar,
		},
	}
	m := expr.MethodExpr{
		Errors: []*expr.ErrorExpr{
			errorFoo,
		},
		Service: &s,
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := m.Error(tc.name); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}

func TestMethodExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		service  *expr.ServiceExpr
		expected string
	}{
		"unnamed": {name: "", service: nil, expected: "unnamed method"},
		"foo":     {name: "foo", service: nil, expected: fmt.Sprintf("method %#v", "foo")},
		"bar":     {name: "bar", service: &expr.ServiceExpr{Name: ""}, expected: fmt.Sprintf("unnamed service method %#v", "bar")},
		"baz":     {name: "baz", service: &expr.ServiceExpr{Name: "baz service"}, expected: fmt.Sprintf("service %#v method %#v", "baz service", "baz")},
	}
	for k, tc := range cases {
		m := expr.MethodExpr{Name: tc.name, Service: tc.service}
		if actual := m.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestMethodExprIsPayloadStreaming(t *testing.T) {
	cases := map[string]struct {
		stream   expr.StreamKind
		expected bool
	}{
		"no stream": {
			stream:   expr.NoStreamKind,
			expected: false,
		},
		"client stream": {
			stream:   expr.ClientStreamKind,
			expected: true,
		},
		"server stream": {
			stream:   expr.ServerStreamKind,
			expected: false,
		},
		"BidirectionalStreamKind": {
			stream:   expr.BidirectionalStreamKind,
			expected: true,
		},
	}
	for k, tc := range cases {
		m := expr.MethodExpr{
			Stream: tc.stream,
		}
		if actual := m.IsPayloadStreaming(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestMethodExprValidateInterceptors(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"no-interceptors", testdata.NoInterceptorsDSL, ""},
		{"valid-interceptors", testdata.ValidInterceptorsDSL, ""},
		{"duplicate-interceptors", testdata.DuplicateInterceptorsDSL, ""}, // Duplicates are handled by merging
		{"mixed-interceptors", testdata.MixedInterceptorsDSL, ""},
		{"undefined-interceptor", testdata.UndefinedInterceptorDSL,
			`ServerInterceptor: interceptor "undefined" not found in service "Service" method "Method"`},
		{"empty-interceptor-name", testdata.EmptyInterceptorNameDSL,
			`ServerInterceptor: interceptor name cannot be empty`},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
				return
			}
			err := expr.RunInvalidDSL(t, c.DSL)
			assert.Contains(t, err.Error(), c.Error)
		})
	}
}
