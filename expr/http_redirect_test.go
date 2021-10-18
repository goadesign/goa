package expr_test

import (
	"net/http"
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestHTTPRedirectExprEvalName(t *testing.T) {
	cases := map[string]struct {
		url        string
		statusCode int
		parent     eval.Expression
		expected   string
	}{
		"without parent": {
			url:        "/redirect/dest",
			statusCode: http.StatusMovedPermanently,
			expected:   "redirect to /redirect/dest with status code 301",
		},
		"parent is HTTPEndpointExpr": {
			url:        "/redirect/dest",
			statusCode: http.StatusMovedPermanently,
			parent:     &expr.HTTPEndpointExpr{MethodExpr: &expr.MethodExpr{Name: "method"}},
			expected:   `HTTP endpoint "method" redirect to /redirect/dest with status code 301`,
		},
		"parent is HTTPFileServerExpr": {
			url:        "/redirect/dest",
			statusCode: http.StatusMovedPermanently,
			parent:     &expr.HTTPFileServerExpr{FilePath: "/file.json"},
			expected:   `file server /file.json redirect to /redirect/dest with status code 301`,
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			r := expr.HTTPRedirectExpr{URL: tc.url, StatusCode: tc.statusCode, Parent: tc.parent}
			if actual := r.EvalName(); actual != tc.expected {
				t.Errorf("got %#v, expected %#v", actual, tc.expected)
			}
		})
	}
}
