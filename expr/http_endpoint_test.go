package expr_test

import (
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/expr/testdata"
)

func TestHTTPRouteValidation(t *testing.T) {
	cases := []struct {
		Name  string
		DSL   func()
		Error string
	}{
		{"valid", testdata.ValidRouteDSL, ""},
		{"invalid", testdata.DuplicateWCRouteDSL, `route POST "/{id}" of service "InvalidRoute" HTTP endpoint "Method": Wildcard "id" appears multiple times in full path "/{id}/{id}"`},
		{"disallow-response-body", testdata.DisallowResponseBodyHeadDSL, `route HEAD "/" of service "DisallowResponseBody" HTTP endpoint "Method": HTTP status 200: Response body defined for HEAD method which does not allow response body.
route HEAD "/" of service "DisallowResponseBody" HTTP endpoint "Method": HTTP status 404: Response body defined for HEAD method which does not allow response body.`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q\nexpected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

func TestHTTPEndpointPrepare(t *testing.T) {
	cases := map[string]struct {
		DSL     func()
		Headers []string
		Params  []string
		Cookies []string
		Error   string
	}{
		"valid": {
			DSL:    testdata.ValidRouteDSL,
			Params: []string{"base_id", "id"},
		},
		"with parent": {
			DSL:     testdata.EndpointWithParentDSL,
			Headers: []string{"pheader", "header"},
			Params:  []string{"pparam", "param"},
			Cookies: []string{"pcookie", "cookie"},
		},
		"with parent revert": {
			DSL:     testdata.EndpointWithParentRevertDSL,
			Headers: []string{"pheader", "header"},
			Params:  []string{"pparam", "param"},
			Cookies: []string{"pcookie", "cookie"},
		},
		"error": {
			DSL:   testdata.EndpointRecursiveParentDSL,
			Error: "service \"Parent\": Parent service Child is also child\nservice \"Child\": Parent service Parent is also child",
		},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			if c.Error == "" {
				root := expr.RunDSL(t, c.DSL)
				e := root.API.HTTP.Services[len(root.API.HTTP.Services)-1].HTTPEndpoints[0]

				ht := expr.AsObject(e.Headers.AttributeExpr.Type)
				if len(*ht) != len(c.Headers) {
					t.Errorf("got %d headers, expected %d", len(*ht), len(c.Headers))
				} else {
					for _, n := range c.Headers {
						if ht.Attribute(n) == nil {
							t.Errorf("header %q is missing", n)
						}
					}
				}

				ct := expr.AsObject(e.Cookies.AttributeExpr.Type)
				if len(*ct) != len(c.Cookies) {
					t.Errorf("got %d cookies, expected %d", len(*ct), len(c.Cookies))
				} else {
					for _, n := range c.Cookies {
						if ct.Attribute(n) == nil {
							t.Errorf("cookie %q is missing", n)
						}
					}
				}

				pt := expr.AsObject(e.Params.AttributeExpr.Type)
				if len(*pt) != len(c.Params) {
					t.Errorf("got %d params, expected %d", len(*pt), len(c.Params))
				} else {
					for _, n := range c.Params {
						if pt.Attribute(n) == nil {
							t.Errorf("param %q is missing", n)
						}
					}
				}
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

func TestHTTPEndpointValidation(t *testing.T) {
	cases := map[string]struct {
		DSL   func()
		Error string
	}{
		"endpoint-body-as-payload-prop": {
			DSL: testdata.EndpointBodyAsPayloadProp,
		},
		"endpoint-body-as-missed-payload-prop": {
			DSL:   testdata.EndpointBodyAsMissedPayloadProp,
			Error: `Request type does not have an attribute named "name" in service "Service" HTTP endpoint "Method"`,
		},
		"endpoint-body-extend-payload": {
			DSL: testdata.EndpointBodyExtendPayload,
		},
		"endpoint-body-as-user-type": {
			DSL: testdata.EndpointBodyAsUserType,
		},
		"endpoint-missing-token": {
			DSL:   testdata.EndpointMissingToken,
			Error: `service "Service" method "Method": payload of method "Method" of service "Service" does not define a JWT attribute, use Token to define one`,
		},
		"endpoint-missing-token-payload": {
			DSL:   testdata.EndpointMissingTokenPayload,
			Error: `service "Service" method "Method": payload of method "Method" of service "Service" does not define a JWT attribute, use Token to define one`,
		},
		"endpoint-missing-token-extend": {
			DSL: testdata.EndpointExtendToken,
		},
		"endpoint-has-parent": {
			DSL: testdata.EndpointHasParent,
		},
		"endpoint-has-parent-and-other": {
			DSL: testdata.EndpointHasParentAndOther,
		},
		"endpoint-has-skip-request-encode-and-payload-streaming": {
			DSL:   testdata.EndpointHasSkipRequestEncodeAndPayloadStreaming,
			Error: `service "Service" HTTP endpoint "Method": Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingPayload.`,
		},
		"endpoint-has-skip-request-encode-and-result-streaming": {
			DSL:   testdata.EndpointHasSkipRequestEncodeAndResultStreaming,
			Error: `service "Service" HTTP endpoint "Method": Endpoint cannot use SkipRequestBodyEncodeDecode when method defines a StreamingResult. Use SkipResponseBodyEncodeDecode instead.`,
		},
		"endpoint-has-skip-response-encode-and-payload-streaming": {
			DSL:   testdata.EndpointHasSkipResponseEncodeAndPayloadStreaming,
			Error: `service "Service" HTTP endpoint "Method": Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingPayload. Use SkipRequestBodyEncodeDecode instead.`,
		},
		"endpoint-has-skip-response-encode-and-result-streaming": {
			DSL: testdata.EndpointHasSkipResponseEncodeAndResultStreaming,
			Error: `service "Service" HTTP endpoint "Method": Endpoint cannot use SkipResponseBodyEncodeDecode when method defines a StreamingResult.
service "Service" HTTP endpoint "Method": HTTP endpoint response body must be empty when using SkipResponseBodyEncodeDecode. Make sure to define headers and cookies as needed.`,
		},
		"endpoint-has-skip-encode-and-grpc": {
			DSL:   testdata.EndpointHasSkipEncodeAndGRPC,
			Error: `service "Service" HTTP endpoint "Method": Endpoint cannot use SkipRequestBodyEncodeDecode and define a gRPC transport.`,
		},
		"endpoint-payload-missing-required": {
			DSL:   testdata.EndpointPayloadMissingRequired,
			Error: `service "Service" HTTP endpoint "Method": The following HTTP request body attribute is required but the corresponding method payload attribute is not: nonreq. Use 'Required' to make the attribute required in the method payload as well.`,
		},
		"streaming-endpoint-has-request-body": {
			DSL: testdata.StreamingEndpointRequestBody,
			Error: `service "Service" HTTP endpoint "MethodA": HTTP endpoint request body must be empty when the endpoint uses streaming. Payload attributes must be mapped to headers and/or params.
service "Service" HTTP endpoint "MethodB": HTTP endpoint request body must be empty when the endpoint uses streaming. Payload attributes must be mapped to headers and/or params.
service "Service" HTTP endpoint "MethodC": HTTP endpoint request body must be empty when the endpoint uses streaming. Payload attributes must be mapped to headers and/or params.`,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				var errors []error
				err := expr.RunInvalidDSL(t, c.DSL)
				if err != nil {
					if merr, ok := err.(eval.MultiError); ok {
						for _, e := range merr {
							errors = append(errors, e.GoError)
						}
					} else {
						errors = append(errors, err)
					}
				}
				if len(errors) > 1 || len(errors) == 0 {
					t.Errorf("got %d errors, expected 1", len(errors))
				} else {
					if errors[0].Error() != c.Error {
						t.Errorf("got `%s`, expected `%s`", err.Error(), c.Error)
					}
				}
			}
		})
	}
}

func TestHTTPEndpointParentRequired(t *testing.T) {
	root := expr.RunDSL(t, testdata.EndpointHasParent)
	svc := root.Service("Child")
	if svc == nil {
		t.Fatal(`unexpected error, service "Child" not found`)
	}
	m := svc.Method("Method")
	if m == nil {
		t.Fatal(`unexpected error, method "Method" not found`)
	}
	if !m.Payload.IsRequired("ancestor_id") {
		t.Errorf(`expected "ancestor_id" is required, but not so`)
	}
	if !m.Payload.IsRequired("parent_id") {
		t.Errorf(`expected "parent_id" is required, but not so`)
	}
}

func TestHTTPEndpointFinalization(t *testing.T) {
	cases := map[string]struct {
		DSL          func()
		ExpectedBody expr.DataType
	}{
		"body-as-extend-type": {
			DSL:          testdata.FinalizeEndpointBodyAsExtendedTypeDSL,
			ExpectedBody: testdata.FinalizeEndpointBodyAsExtendedType,
		},
		"body-as-prop-with-extend-type": {
			DSL:          testdata.FinalizeEndpointBodyAsPropWithExtendedTypeDSL,
			ExpectedBody: testdata.FinalizeEndpointBodyAsPropWithExtendedType,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			root := expr.RunDSL(t, tc.DSL)
			e := root.API.HTTP.Services[0].HTTPEndpoints[0]

			if tc.ExpectedBody != nil {
				if e.Body == nil {
					t.Errorf("got endpoint without body, expected endpoint with body")
					return
				}
				bodyObj := *expr.AsObject(e.Body.Type)
				expectedBodyObj := *expr.AsObject(tc.ExpectedBody)
				if len(bodyObj) != len(expectedBodyObj) {
					t.Errorf("got %d, expected %d attribute(s) in endpoint body", len(bodyObj), len(expectedBodyObj))
				} else {
					for i := 0; i < len(expectedBodyObj); i++ {
						if bodyObj[i].Name != expectedBodyObj[i].Name {
							t.Errorf("got %q, expected %q attribute in endpoint body", bodyObj[i].Name, expectedBodyObj[i].Name)
						}
					}
				}
			}
		})
	}
}

func TestHTTPAuthorizationMapping(t *testing.T) {
	cases := []struct {
		Name           string
		DSL            func()
		ExpectedHeader string
	}{{
		Name:           "explicit",
		DSL:            testdata.ExplicitAuthHeaderDSL,
		ExpectedHeader: "token",
	}, {
		Name:           "implicit",
		DSL:            testdata.ImplicitAuthHeaderDSL,
		ExpectedHeader: "Authorization",
	},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			root := expr.RunDSL(t, tc.DSL)
			e := root.API.HTTP.Services[0].HTTPEndpoints[0]
			if e.Headers == nil {
				t.Errorf("got endpoint without header, expected endpoint with HTTP header")
				return
			}
			if len(*expr.AsObject(e.Headers.Type)) != 1 {
				t.Errorf("got %d, expected 1 attribute in endpoint headers", len(*expr.AsObject(e.Headers.Type)))
				return
			}
			n := e.Headers.ElemName("token")
			if n != tc.ExpectedHeader {
				t.Errorf("got %q, expected %q attribute in endpoint headers", n, tc.ExpectedHeader)
			}
		})
	}
}
