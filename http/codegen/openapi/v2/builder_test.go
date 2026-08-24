// This file verifies OpenAPI v2 construction from evaluated HTTP endpoint
// designs, including request and response example ownership.
package openapiv2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"gopkg.in/yaml.v3"
)

func TestNewV2WithValues(t *testing.T) {
	root := codegen.RunDSL(t, localizedValuesDSL)
	service := root.Service("messages")
	method := service.Method("show")
	values := (openapi.Values{}).
		WithTitle(root.API, "Localized API").
		WithDescription(root.API, "Localized API description").
		WithDescription(service, "Localized service description").
		WithDescription(method, "Localized method description")

	spec, err := NewV2WithValues(
		root,
		root.API.Servers[0].Hosts[0],
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		values,
	)
	require.NoError(t, err)
	require.Equal(t, "Localized API", spec.Info.Title)
	require.Equal(t, "Localized API description", spec.Info.Description)
	operation := spec.Paths["/messages"].(*Path).Get
	require.Equal(t, "Localized method description", operation.Description)
	require.Contains(t, operation.Tags, "messages")
	require.Equal(t, "Original API", root.API.Title)
	require.Equal(t, "Original method description", method.Description)
}

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
				Paths: make(map[string]any),
			}
			root := &expr.RootExpr{
				API: &expr.APIExpr{
					RandomizerFactory: expr.NewFakerRandomizerFactory("test"),
				},
			}
			fs := &expr.HTTPFileServerExpr{
				Service: &expr.HTTPServiceExpr{
					ServiceExpr: &expr.ServiceExpr{
						Name: "service",
					},
				},
				RequestPaths: []string{tc.path},
			}
			buildPathFromFileServer(s, root, fs, newSchemaBuilder(openapi.Values{}), expr.NewExampleGenerator(root.API.RandomizerFactory), openapi.Values{})
			for actual := range s.Paths {
				if actual != tc.expected {
					t.Errorf("got %#v, expected %#v", actual, tc.expected)
				}
			}
		})
	}
}

func TestNoSecurityOverridesAPISecurity(t *testing.T) {
	root := codegen.RunDSL(t, noSecurityOverridesAPISecurityDSL)
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)

	cases := map[string]struct {
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		"json": {
			marshal:   json.Marshal,
			unmarshal: json.Unmarshal,
		},
		"yaml": {
			marshal:   yaml.Marshal,
			unmarshal: yaml.Unmarshal,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			data, err := tc.marshal(spec)
			require.NoError(t, err)

			var actual struct {
				Paths map[string]map[string]struct {
					Security []map[string][]string `json:"security" yaml:"security"`
				} `json:"paths" yaml:"paths"`
			}
			require.NoError(t, tc.unmarshal(data, &actual))

			require.NotEmpty(t, actual.Paths["/secure"]["get"].Security, "secure operation has no security requirements")
			security := actual.Paths["/public"]["get"].Security
			require.NotNil(t, security, "NoSecurity operation omitted the operation security override")
			require.Empty(t, security, "NoSecurity operation security expected empty override")
		})
	}
}

func TestNoSecurityOverridesServiceSecurity(t *testing.T) {
	root := codegen.RunDSL(t, noSecurityOverridesServiceSecurityDSL)
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)

	cases := map[string]struct {
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		"json": {
			marshal:   json.Marshal,
			unmarshal: json.Unmarshal,
		},
		"yaml": {
			marshal:   yaml.Marshal,
			unmarshal: yaml.Unmarshal,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			data, err := tc.marshal(spec)
			require.NoError(t, err)

			var actual struct {
				Paths map[string]map[string]struct {
					Security []map[string][]string `json:"security" yaml:"security"`
				} `json:"paths" yaml:"paths"`
			}
			require.NoError(t, tc.unmarshal(data, &actual))

			require.NotEmpty(t, actual.Paths["/service-secure"]["get"].Security, "secure operation has no security requirements")
			security := actual.Paths["/service-public"]["get"].Security
			require.NotNil(t, security, "NoSecurity operation omitted the operation security override")
			require.Empty(t, security, "NoSecurity operation security expected empty override")
		})
	}
}

func TestStreamingResponseStatusCodes(t *testing.T) {
	root := codegen.RunDSL(t, streamingResponseStatusDSL)
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)

	sseResponses := spec.Paths["/sse"].(*Path).Get.Responses
	require.Contains(t, sseResponses, "200")
	require.NotContains(t, sseResponses, "101")
	require.NotContains(t, spec.Paths["/sse"].(*Path).Get.Schemes, "ws")
	require.NotContains(t, spec.Paths["/sse"].(*Path).Get.Schemes, "wss")

	websocketResponses := spec.Paths["/websocket"].(*Path).Get.Responses
	require.Contains(t, websocketResponses, "101")
	require.NotContains(t, websocketResponses, "200")
	require.Contains(t, spec.Paths["/websocket"].(*Path).Get.Schemes, "ws")
}

func TestOperationSecurityMarshal(t *testing.T) {
	securityCases := map[string]struct {
		operation Operation
		expected  map[string]any
	}{
		"nil security is omitted": {
			operation: Operation{},
			expected:  map[string]any{},
		},
		"empty security is emitted": {
			operation: Operation{Security: SecurityRequirements{}},
			expected:  map[string]any{"security": []any{}},
		},
	}
	cases := map[string]struct {
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		"json": {
			marshal:   json.Marshal,
			unmarshal: json.Unmarshal,
		},
		"yaml": {
			marshal:   yaml.Marshal,
			unmarshal: yaml.Unmarshal,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for securityName, securityTC := range securityCases {
				t.Run(securityName, func(t *testing.T) {
					data, err := tc.marshal(securityTC.operation)
					require.NoError(t, err)

					var actual map[string]any
					require.NoError(t, tc.unmarshal(data, &actual))
					require.Equal(t, securityTC.expected, actual)
				})
			}
		})
	}
}

func TestSecurityDefinitionsIncludeVisibleOperationsOnly(t *testing.T) {
	root := codegen.RunDSL(t, visibleSecuritySchemesDSL)
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)

	visible := root.Service("visible").Method("read").Requirements[0].Schemes[0].Hash()
	hiddenMethod := root.Service("mixed").Method("hidden").Requirements[0].Schemes[0].Hash()
	hiddenService := root.Service("hidden").Method("read").Requirements[0].Schemes[0].Hash()
	require.Contains(t, spec.SecurityDefinitions, visible)
	require.NotContains(t, spec.SecurityDefinitions, hiddenMethod)
	require.NotContains(t, spec.SecurityDefinitions, hiddenService)
}

var noSecurityOverridesAPISecurityDSL = func() {
	var JWTAuth = dsl.JWTSecurity("jwt")

	dsl.API("test", func() {
		dsl.Security(JWTAuth)
	})

	dsl.Service("test", func() {
		dsl.Method("secure", func() {
			dsl.Payload(func() {
				dsl.Token("token", dsl.String)
				dsl.Required("token")
			})
			dsl.HTTP(func() {
				dsl.GET("/secure")
			})
		})
		dsl.Method("public", func() {
			dsl.NoSecurity()
			dsl.HTTP(func() {
				dsl.GET("/public")
			})
		})
	})
}

var localizedValuesDSL = func() {
	dsl.API("messages", func() {
		dsl.Title("Original API")
		dsl.Description("Original API description")
	})
	dsl.Service("messages", func() {
		dsl.Description("Original service description")
		dsl.Method("show", func() {
			dsl.Description("Original method description")
			dsl.HTTP(func() {
				dsl.GET("/messages")
			})
		})
	})
}

var visibleSecuritySchemesDSL = func() {
	var (
		VisibleAuth       = dsl.JWTSecurity("visible_auth")
		HiddenMethodAuth  = dsl.JWTSecurity("hidden_method_auth")
		HiddenServiceAuth = dsl.JWTSecurity("hidden_service_auth")
	)

	dsl.Service("visible", func() {
		dsl.Method("read", func() {
			dsl.Security(VisibleAuth)
			dsl.Payload(func() {
				dsl.Token("token", dsl.String)
			})
			dsl.HTTP(func() {
				dsl.GET("/visible")
			})
		})
	})
	dsl.Service("mixed", func() {
		dsl.Method("hidden", func() {
			dsl.Meta("openapi:generate", "false")
			dsl.Security(HiddenMethodAuth)
			dsl.Payload(func() {
				dsl.Token("token", dsl.String)
			})
			dsl.HTTP(func() {
				dsl.GET("/hidden-method")
			})
		})
	})
	dsl.Service("hidden", func() {
		dsl.Meta("openapi:generate", "false")
		dsl.Method("read", func() {
			dsl.Security(HiddenServiceAuth)
			dsl.Payload(func() {
				dsl.Token("token", dsl.String)
			})
			dsl.HTTP(func() {
				dsl.GET("/hidden-service")
			})
		})
	})
}

var noSecurityOverridesServiceSecurityDSL = func() {
	var JWTAuth = dsl.JWTSecurity("jwt")

	dsl.API("test", func() {})

	dsl.Service("test", func() {
		dsl.Security(JWTAuth)
		dsl.Method("service-secure", func() {
			dsl.Payload(func() {
				dsl.Token("token", dsl.String)
				dsl.Required("token")
			})
			dsl.HTTP(func() {
				dsl.GET("/service-secure")
			})
		})
		dsl.Method("service-public", func() {
			dsl.NoSecurity()
			dsl.HTTP(func() {
				dsl.GET("/service-public")
			})
		})
	})
}

var streamingResponseStatusDSL = func() {
	dsl.Service("streaming", func() {
		dsl.Method("sse", func() {
			dsl.StreamingResult(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/sse")
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("websocket", func() {
			dsl.StreamingResult(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/websocket")
			})
		})
	})
}

func TestBuildPathFromExpr(t *testing.T) {
	cases := map[string]struct {
		multipartRequest bool
		deprecated       bool
		expected         Operation
	}{
		"multipart request": {
			multipartRequest: true,
			deprecated:       false,
			expected: Operation{
				Deprecated: false,
				Consumes:   []string{"multipart/form-data"},
				Parameters: []*Parameter{{In: "formData"}},
			},
		},
		"non multipart request": {
			multipartRequest: false,
			deprecated:       true,
			expected: Operation{
				Deprecated: true,
				Consumes:   nil,
				Parameters: []*Parameter{{In: "body"}},
			},
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			s := &V2{
				Consumes: []string{"application/json"},
				Paths:    make(map[string]any),
			}
			root := &expr.RootExpr{
				API: &expr.APIExpr{
					HTTP: &expr.HTTPExpr{
						Path: "/",
					},
				},
			}
			expr.Root = root
			h := &expr.HostExpr{}
			route := &expr.RouteExpr{
				Method: "POST",
				Endpoint: &expr.HTTPEndpointExpr{
					MethodExpr: &expr.MethodExpr{
						Payload: &expr.AttributeExpr{},
					},
					Service: &expr.HTTPServiceExpr{
						Root:        root.API.HTTP,
						ServiceExpr: &expr.ServiceExpr{},
						Paths:       []string{"/foo"},
						Params:      expr.NewEmptyMappedAttributeExpr(),
					},
					Headers: expr.NewEmptyMappedAttributeExpr(),
					Body: &expr.AttributeExpr{
						Type: expr.String,
					},
					MultipartRequest: tc.multipartRequest,
					Meta:             expr.MetaExpr{},
				},
			}
			route.Endpoint.MethodExpr.Name = "method"
			route.Endpoint.Service.ServiceExpr.Name = "service"
			route.Endpoint.MethodExpr.Service = route.Endpoint.Service.ServiceExpr

			if tc.deprecated {
				route.Endpoint.Meta["openapi:deprecated"] = []string{"true"}
			}

			basePath := "/"
			generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test"))
			buildPathFromExpr(s, root, h, route, basePath, newSchemaBuilder(openapi.Values{}), generator, openapi.Values{})
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
