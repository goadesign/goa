package design

import (
	"fmt"
	"testing"

	"goa.design/goa/eval"
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
	var (
		goadesign = &ServerExpr{
			URL: "https://goa.design",
		}
		localhost = &ServerExpr{
			URL: "http://localhost:8080",
		}
	)
	cases := map[string]struct {
		servers  []*ServerExpr
		expected []*ServerExpr
	}{
		"no server": {servers: nil, expected: []*ServerExpr{localhost}},
		"1 server":  {servers: []*ServerExpr{goadesign}, expected: []*ServerExpr{goadesign}},
		"2 servers": {servers: []*ServerExpr{goadesign, localhost}, expected: []*ServerExpr{goadesign, localhost}},
	}

	for k, tc := range cases {
		api := APIExpr{Servers: tc.servers}
		api.Finalize()
		if actual := api.Servers; len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of scheme values to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v.URL != tc.expected[i].URL {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}

func TestServerExprValidate(t *testing.T) {
	cases := map[string]struct {
		url      string
		params   *AttributeExpr
		expected *eval.ValidationErrors
	}{
		"no error": {
			url:      "http://example.com/cellar/accounts/{accountID}",
			params:   &AttributeExpr{Type: &Object{&NamedAttributeExpr{Name: "accountID", Attribute: &AttributeExpr{DefaultValue: "foo"}}}},
			expected: &eval.ValidationErrors{},
		},
		"missing param expression": {
			url:      "http://example.com/cellar/accounts/{accountID}",
			params:   nil,
			expected: &eval.ValidationErrors{Errors: []error{fmt.Errorf("missing Param expressions")}},
		},
		"invalid parameter count": {
			url:      "http://example.com/cellar/accounts/{accountID}",
			params:   &AttributeExpr{Type: &Object{}},
			expected: &eval.ValidationErrors{Errors: []error{fmt.Errorf("invalid parameter count, expected %d, got %d", 1, 0)}},
		},
		"parameter not defined": {
			url:      "http://example.com/cellar/accounts/{accountID}",
			params:   &AttributeExpr{Type: &Object{&NamedAttributeExpr{Name: "bottleID", Attribute: &AttributeExpr{DefaultValue: "foo"}}}},
			expected: &eval.ValidationErrors{Errors: []error{fmt.Errorf("parameter %s is not defined", "accountID")}},
		},
		"parameter has no default value": {
			url:      "http://example.com/cellar/accounts/{accountID}",
			params:   &AttributeExpr{Type: &Object{&NamedAttributeExpr{Name: "accountID", Attribute: &AttributeExpr{DefaultValue: nil}}}},
			expected: &eval.ValidationErrors{Errors: []error{fmt.Errorf("parameter %s has no default value", "accountID")}},
		},
	}

	for k, tc := range cases {
		server := ServerExpr{
			URL:    tc.url,
			Params: tc.params,
		}
		if actual := server.Validate().(*eval.ValidationErrors); len(tc.expected.Errors) != len(actual.Errors) {
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

func TestServerParamExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo": {name: "foo", expected: "URL parameter foo"},
	}

	for k, tc := range cases {
		param := ServerParamExpr{Name: tc.name}
		if actual := param.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestURLParams(t *testing.T) {
	cases := map[string]struct {
		url      string
		expected []string
	}{
		"empty url":        {url: "", expected: []string{}},
		"no match":         {url: "http://example.com", expected: []string{}},
		"single match":     {url: "http://example.com/cellar/accounts/{accountID}", expected: []string{"accountID"}},
		"multiple matches": {url: "http://example.com/cellar/accounts/{accountID}/bottles/{bottleID}", expected: []string{"accountID", "bottleID"}},
	}

	for k, tc := range cases {
		if actual := URLParams(tc.url); len(tc.expected) != len(actual) {
			t.Errorf("%s: expected the number of param values to match %d got %d ", k, len(tc.expected), len(actual))
		} else {
			for i, v := range actual {
				if v != tc.expected[i] {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, v, tc.expected[i], i)
				}
			}
		}
	}
}
