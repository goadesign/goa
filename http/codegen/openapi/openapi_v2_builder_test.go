package openapi

import (
	"testing"

	"goa.design/goa/v3/expr"
)

func TestBuildPathFromFileServer(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{
			path:     "/foo",
			expected: "/foo",
		},
		{
			path:     "/foo/{bar}",
			expected: "/foo/{bar}",
		},
		{
			path:     "/foo/{*bar}",
			expected: "/foo/{bar}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			s := &V2{
				Paths: make(map[string]interface{}),
			}
			root := &expr.RootExpr{
				API: &expr.APIExpr{},
			}
			fs := &expr.HTTPFileServerExpr{
				Service: &expr.HTTPServiceExpr{
					ServiceExpr: &expr.ServiceExpr{
						Name: "service",
					},
				},
				RequestPaths: []string{tc.path},
			}
			buildPathFromFileServer(s, root, fs)
			for actual := range s.Paths {
				if actual != tc.expected {
					t.Errorf("got %#v, expected %#v", actual, tc.expected)
				}
			}
		})
	}
}
