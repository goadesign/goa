package openapiv2

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

func TestBuildPathFromExpr(t *testing.T) {
	cases := map[string]struct {
		multipartRequest bool
		expected         Operation
	}{
		"multipart request": {
			multipartRequest: true,
			expected: Operation{
				Consumes:   []string{"multipart/form-data"},
				Parameters: []*Parameter{{In: "formData"}},
			},
		},
		"non multipart request": {
			multipartRequest: false,
			expected: Operation{
				Consumes:   nil,
				Parameters: []*Parameter{{In: "body"}},
			},
		},
	}
	expr.Root.API = &expr.APIExpr{
		HTTP: &expr.HTTPExpr{
			Path: "/",
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			s := &V2{
				Consumes: []string{"application/json"},
				Paths:    make(map[string]interface{}),
			}
			root := &expr.RootExpr{
				API: &expr.APIExpr{},
			}
			h := &expr.HostExpr{}
			route := &expr.RouteExpr{
				Method: "POST",
				Endpoint: &expr.HTTPEndpointExpr{
					MethodExpr: &expr.MethodExpr{
						Payload: &expr.AttributeExpr{},
					},
					Service: &expr.HTTPServiceExpr{
						ServiceExpr: &expr.ServiceExpr{},
						Paths:       []string{"/foo"},
						Params:      expr.NewEmptyMappedAttributeExpr(),
					},
					Headers: expr.NewEmptyMappedAttributeExpr(),
					Body: &expr.AttributeExpr{
						Type: expr.String,
					},
					MultipartRequest: tc.multipartRequest,
				},
			}
			basePath := "/"
			buildPathFromExpr(s, root, h, route, basePath)
			for _, path := range s.Paths {
				actual := path.(*Path).Post
				if len(actual.Consumes) != len(tc.expected.Consumes) {
					t.Errorf("expected the number of consumes to match %d got %d", len(actual.Consumes), len(tc.expected.Consumes))
				} else {
					for i, v := range actual.Consumes {
						if v != tc.expected.Consumes[i] {
							t.Errorf("got %#v, expected %#v at index %d", v, tc.expected.Consumes[i], i)
						}
					}
				}
				if len(actual.Parameters) != len(tc.expected.Parameters) {
					t.Errorf("expected the number of parameters to match %d got %d", len(actual.Parameters), len(tc.expected.Parameters))
				} else {
					for i, v := range actual.Parameters {
						if v.In != tc.expected.Parameters[i].In {
							t.Errorf("got %#v, expected %#v at index %d", v.In, tc.expected.Parameters[i].In, i)
						}
					}
				}
			}
		})
	}
}
