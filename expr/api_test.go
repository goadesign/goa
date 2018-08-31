package expr

import (
	"testing"
)

func TestAPIExprSchemes(t *testing.T) {
	cases := map[string]struct {
		expr     APIExpr
		expected []string
	}{
		"default scheme": {
			expr: APIExpr{
				Servers: []*ServerExpr{&ServerExpr{}},
			},
			expected: nil,
		},
		"single scheme": {
			expr: APIExpr{
				Servers: []*ServerExpr{
					&ServerExpr{
						Hosts: []*HostExpr{
							{URIs: []URIExpr{"http://example.com"}},
						},
					},
				},
			},
			expected: []string{"http"},
		},
		"multiple schemes": {
			expr: APIExpr{
				Servers: []*ServerExpr{
					&ServerExpr{
						Hosts: []*HostExpr{
							{URIs: []URIExpr{"http://example.com"}},
						},
					},
					&ServerExpr{
						Hosts: []*HostExpr{
							{URIs: []URIExpr{"https://example.net"}},
						},
					},
				},
			},
			expected: []string{"http", "https"},
		},
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

func TestAPIExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo": {name: "foo", expected: "API foo"},
	}

	for k, tc := range cases {
		api := APIExpr{Name: tc.name}
		if actual := api.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestAPIExprFinalize(t *testing.T) {
	cases := map[string]struct {
		api      *APIExpr
		expected []string
	}{
		"empty name, empty server": {
			api:      &APIExpr{},
			expected: []string{"Default server for api"},
		},
		"empty server": {
			api: &APIExpr{
				Name: "my api",
			},
			expected: []string{"Default server for my api"},
		},
		"with server": {
			api: &APIExpr{
				Name:    "my api",
				Servers: []*ServerExpr{{Description: "my server"}},
			},
			expected: []string{"my server"},
		},
	}

	for k, tc := range cases {
		tc.api.Finalize()

		if actual := tc.api.Servers; len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of servers to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v.Description != tc.expected[i] {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}

func TestApiExpr_Hash(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo":   {name: "foo", expected: "_api_+foo"},
		"empty": {name: "", expected: "_api_+"},
	}

	for k, tc := range cases {
		api := APIExpr{Name: tc.name}
		if actual := api.Hash(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
