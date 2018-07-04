package design

import (
	"fmt"
	"net/url"
	"testing"

	"goa.design/goa/eval"
)

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
