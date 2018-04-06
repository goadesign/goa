package design

import (
	"testing"
)

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
