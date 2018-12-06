package expr

import (
	"testing"
)

func TestResultTypeExprIsError(t *testing.T) {
	cases := map[string]struct {
		identifier string
		expected   bool
	}{
		"error": {
			identifier: "application/vnd.goa.error",
			expected:   true,
		},
		"not error": {
			identifier: "application/vnd.goa.foo",
			expected:   false,
		},
	}

	for k, tc := range cases {
		r := ResultTypeExpr{
			Identifier: tc.identifier,
		}
		if actual := r.IsError(); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestCanonicalIdentifier(t *testing.T) {
	cases := map[string]struct {
		identifier string
		expected   string
	}{
		"standards": {
			identifier: "application/json",
			expected:   "application/json",
		},
		"standards with parameter": {
			identifier: "application/json; charset=utf-8",
			expected:   "application/json; charset=utf-8",
		},
		"vendor": {
			identifier: "application/vnd.goa.error+json",
			expected:   "application/vnd.goa.error",
		},
		"vendor with parameter": {
			identifier: "application/vnd.goa.error+json; type=collection",
			expected:   "application/vnd.goa.error; type=collection",
		},
	}

	for k, tc := range cases {
		if actual := CanonicalIdentifier(tc.identifier); tc.expected != actual {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
