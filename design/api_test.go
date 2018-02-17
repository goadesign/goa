package design

import (
	"testing"
)

func TestAPIExprSchemes(t *testing.T) {
	cases := map[string]struct {
		expr     APIExpr
		expected []string
	}{
		"default scheme":   {expr: APIExpr{Servers: []*ServerExpr{&ServerExpr{}}}, expected: []string{"http"}},
		"single scheme":    {expr: APIExpr{Servers: []*ServerExpr{&ServerExpr{URL: "http://example.com"}}}, expected: []string{"http"}},
		"multiple schemes": {expr: APIExpr{Servers: []*ServerExpr{&ServerExpr{URL: "http://example.com"}, &ServerExpr{URL: "https://example.net"}}}, expected: []string{"http", "https"}},
	}

	for k, tc := range cases {
		if actual := tc.expr.Schemes(); len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of scheme values to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v != tc.expected[i] {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}
