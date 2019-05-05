package expr

import (
	"fmt"
	"net/url"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestFlowExprEvalName(t *testing.T) {
	cases := []struct {
		name     string
		kind     FlowKind
		expected string
	}{
		{"authorization-code", AuthorizationCodeFlowKind, "flow authorization_code"},
		{"implicit", ImplicitFlowKind, "flow implicit"},
		{"client-credentials", ClientCredentialsFlowKind, "flow client_credentials"},
		{"password", PasswordFlowKind, "flow password"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fe := &FlowExpr{Kind: tc.kind}
			if actual := fe.EvalName(); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}

func TestFlowExprType(t *testing.T) {
	var noKind FlowKind
	cases := map[string]struct {
		kind     FlowKind
		expected string
	}{
		"authorization code": {
			kind:     AuthorizationCodeFlowKind,
			expected: "authorization_code",
		},
		"implicit": {
			kind:     ImplicitFlowKind,
			expected: "implicit",
		},
		"password": {
			kind:     PasswordFlowKind,
			expected: "password",
		},
		"client credentials": {
			kind:     ClientCredentialsFlowKind,
			expected: "client_credentials",
		},
		"no kind": {
			kind:     noKind,
			expected: "", // should have panicked!
		},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "no kind" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			f := &FlowExpr{Kind: tc.kind}
			if actual := f.Type(); actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}

func TestFlowExprValidate(t *testing.T) {
	var (
		tokenURL                   = "http://example.com/token"
		authorizationURL           = "http://example.com/auth"
		refreshURL                 = "http://example.com/refresh"
		invalidURL                 = "http://%"
		escapeError                = url.EscapeError("%")
		parseError                 = url.Error{Op: "parse", URL: invalidURL, Err: escapeError}
		errInvalidTokenURL         = fmt.Errorf("invalid token URL %q: %s", invalidURL, parseError.Error())
		errInvalidAuthorizationURL = fmt.Errorf("invalid authorization URL %q: %s", invalidURL, parseError.Error())
		errInvalidRefreshURL       = fmt.Errorf("invalid refresh URL %q: %s", invalidURL, parseError.Error())
	)
	cases := map[string]struct {
		tokenURL         string
		authorizationURL string
		refreshURL       string
		expected         *eval.ValidationErrors
	}{
		"no error": {
			tokenURL:         tokenURL,
			authorizationURL: authorizationURL,
			refreshURL:       refreshURL,
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"invalid token url": {
			tokenURL:         invalidURL,
			authorizationURL: authorizationURL,
			refreshURL:       refreshURL,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidTokenURL,
				},
			},
		},
		"invalid authorization url": {
			tokenURL:         tokenURL,
			authorizationURL: invalidURL,
			refreshURL:       refreshURL,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidAuthorizationURL,
				},
			},
		},
		"invalid refresh url": {
			tokenURL:         tokenURL,
			authorizationURL: authorizationURL,
			refreshURL:       invalidURL,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidRefreshURL,
				},
			},
		},
		"invalid token url, authorization url and refresh url": {
			tokenURL:         invalidURL,
			authorizationURL: invalidURL,
			refreshURL:       invalidURL,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidTokenURL,
					errInvalidAuthorizationURL,
					errInvalidRefreshURL,
				},
			},
		},
	}

	for k, tc := range cases {
		f := FlowExpr{
			TokenURL:         tc.tokenURL,
			AuthorizationURL: tc.authorizationURL,
			RefreshURL:       tc.refreshURL,
		}
		if actual := f.Validate(); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestSchemeExprEvalName(t *testing.T) {
	cases := map[string]struct {
		kind     SchemeKind
		expected string
	}{
		"OAuth2Kind":    {kind: OAuth2Kind, expected: "OAuth2Security"},
		"BasicAuthKind": {kind: BasicAuthKind, expected: "BasicAuthSecurity"},
		"APIKeyKind":    {kind: APIKeyKind, expected: "APIKeySecurity"},
		"JWTKind":       {kind: JWTKind, expected: "JWTSecurity"},
		"NoKind":        {kind: NoKind, expected: "This case is panic"},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "NoKind" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			se := &SchemeExpr{Kind: tc.kind}
			if actual := se.EvalName(); actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}

func TestSchemeExprType(t *testing.T) {
	cases := map[string]struct {
		kind     SchemeKind
		expected string
	}{
		"oauth2": {
			kind:     OAuth2Kind,
			expected: "OAuth2",
		},
		"basic auth": {
			kind:     BasicAuthKind,
			expected: "BasicAuth",
		},
		"api key": {
			kind:     APIKeyKind,
			expected: "APIKey",
		},
		"jwt": {
			kind:     JWTKind,
			expected: "JWT",
		},
		"NoKind": {
			kind:     NoKind,
			expected: "", // should have panicked!
		},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "NoKind" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			f := &SchemeExpr{Kind: tc.kind}
			if actual := f.Type(); actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}

func TestSecurityExprEvalName(t *testing.T) {
	scheme1 := &SchemeExpr{SchemeName: "A"}
	scheme2 := &SchemeExpr{SchemeName: ""}

	cases := map[string]struct {
		schemes  []*SchemeExpr
		expected string
	}{
		"security with suffix":           {schemes: []*SchemeExpr{scheme1}, expected: "Securityscheme A"},
		"empty string are security only": {schemes: []*SchemeExpr{scheme2}, expected: "Security"},
		"in case of security only":       {schemes: nil, expected: "Security"},
		"also in case of security only":  {schemes: []*SchemeExpr{}, expected: "Security"},
	}

	for k, tc := range cases {
		se := &SecurityExpr{Schemes: tc.schemes}
		if actual := se.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestSchemeExprValidate(t *testing.T) {
	var (
		tokenURL         = "http://example.com/token"
		authorizationURL = "http://example.com/auth"
		refreshURL       = "http://example.com/refresh"
		invalidURL       = "http://%"
		escapeError      = url.EscapeError("%")
		parseError       = url.Error{Op: "parse", URL: invalidURL, Err: escapeError}
		validFlow        = &FlowExpr{
			TokenURL:         tokenURL,
			AuthorizationURL: authorizationURL,
			RefreshURL:       refreshURL,
		}
		invalidTokenURLFlow = &FlowExpr{
			TokenURL:         invalidURL,
			AuthorizationURL: authorizationURL,
			RefreshURL:       refreshURL,
		}
		invalidAuthorizationURLFlow = &FlowExpr{
			TokenURL:         tokenURL,
			AuthorizationURL: invalidURL,
			RefreshURL:       refreshURL,
		}
		errInvalidTokenURL         = fmt.Errorf("invalid token URL %q: %s", invalidURL, parseError.Error())
		errInvalidAuthorizationURL = fmt.Errorf("invalid authorization URL %q: %s", invalidURL, parseError.Error())
	)
	cases := map[string]struct {
		flows    []*FlowExpr
		expected *eval.ValidationErrors
	}{
		"no error": {
			flows: []*FlowExpr{
				validFlow,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"single error": {
			flows: []*FlowExpr{
				invalidTokenURLFlow,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidTokenURL,
				},
			},
		},
		"multiple errors": {
			flows: []*FlowExpr{
				invalidTokenURLFlow,
				invalidAuthorizationURLFlow,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidTokenURL,
					errInvalidAuthorizationURL,
				},
			},
		},
	}

	for k, tc := range cases {
		s := SchemeExpr{
			Flows: tc.flows,
		}
		if actual := s.Validate(); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestSchemeKindString(t *testing.T) {
	var unknownKind SchemeKind
	cases := map[string]struct {
		kind     SchemeKind
		expected string
	}{
		"basic auth": {
			kind:     BasicAuthKind,
			expected: "Basic",
		},
		"api key": {
			kind:     APIKeyKind,
			expected: "APIKey",
		},
		"jwt": {
			kind:     JWTKind,
			expected: "JWT",
		},
		"oauth2": {
			kind:     OAuth2Kind,
			expected: "OAuth2",
		},
		"no kind": {
			kind:     NoKind,
			expected: "None",
		},
		"unknown kind": {
			kind:     unknownKind,
			expected: "", // should have panicked!
		},
	}

	for k, tc := range cases {
		func() {
			// panic recover
			defer func() {
				if k != "unknown kind" {
					return
				}

				if recover() == nil {
					t.Errorf("should have panicked!")
				}
			}()

			if actual := tc.kind.String(); actual != tc.expected {
				t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
			}
		}()
	}
}
