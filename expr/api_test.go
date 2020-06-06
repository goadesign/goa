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
				Servers: []*ServerExpr{{}},
			},
			expected: nil,
		},
		"single scheme": {
			expr: APIExpr{
				Servers: []*ServerExpr{{
					Hosts: []*HostExpr{
						{URIs: []URIExpr{"http://example.com"}},
					}},
				},
			},
			expected: []string{"http"},
		},
		"multiple schemes": {
			expr: APIExpr{
				Servers: []*ServerExpr{{
					Hosts: []*HostExpr{{URIs: []URIExpr{"http://example.com"}}},
				}, {
					Hosts: []*HostExpr{{URIs: []URIExpr{"https://example.net"}}},
				}},
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

func TestDocsExprEvalName(t *testing.T) {
	cases := map[string]struct {
		url         string
		expected    string
		description string
	}{
		"test, only with url":       {url: "http://parrot.com", expected: "Documentation http://parrot.com"},
		"test, description and url": {url: "http://parrot.com", description: "A website for a parrot API", expected: "Documentation http://parrot.com"},
		"test, only description":    {description: "A website for a parrot API", expected: "Documentation "},
		"test, empty url":           {url: "", expected: "Documentation "},
	}
	for k, tc := range cases {
		docExpr := DocsExpr{URL: tc.url, Description: tc.description}
		if actual := docExpr.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
func TestContactExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		email    string
		url      string
		expected string
	}{
		"test, only with name":   {name: "parrot", expected: "Contact parrot"},
		"test, name and email":   {name: "parrot", email: "parrot@gopher.com", expected: "Contact parrot"},
		"test, name and url":     {name: "parrot", url: "https://parrot.com", expected: "Contact parrot"},
		"test, name, url, email": {name: "parrot", email: "parrot@gopher.com", url: "https://parrot.com", expected: "Contact parrot"},
		"test, empty name":       {name: "", expected: "Contact "},
	}
	for k, tc := range cases {
		contactExpr := ContactExpr{Name: tc.name, Email: tc.email, URL: tc.url}
		if actual := contactExpr.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
func TestLicenseExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo": {name: "foo", expected: "License foo"},
	}

	for k, tc := range cases {
		license := LicenseExpr{Name: tc.name}
		if actual := license.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
