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
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Error == "" {
				expr.RunDSL(t, c.DSL)
			} else {
				err := expr.RunInvalidDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
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
		},
		"with parent revert": {
			DSL:     testdata.EndpointWithParentRevertDSL,
			Headers: []string{"pheader", "header"},
			Params:  []string{"pparam", "param"},
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
		DSL    func()
		Errors []string
	}{
		"endpoint-body-as-payload-prop": {
			DSL: testdata.EndpointBodyAsPayloadProp,
		},
		"endpoint-body-as-missed-payload-prop": {
			DSL: testdata.EndpointBodyAsMissedPayloadProp,
			Errors: []string{
				"Request type does not have an attribute named \"name\" in service \"Service\" HTTP endpoint \"Method\"",
			},
		},
		"endpoint-body-extend-payload": {
			DSL: testdata.EndpointBodyExtendPayload,
		},
		"endpoint-body-as-user-type": {
			DSL: testdata.EndpointBodyAsUserType,
		},
		"endpoint-missing-token": {
			DSL: testdata.EndpointMissingToken,
			Errors: []string{
				"service \"Service\" method \"Method\": payload of method \"Method\" of service \"Service\" does not define a JWT attribute, use Token to define one",
			},
		},
		"endpoint-missing-token-payload": {
			DSL: testdata.EndpointMissingTokenPayload,
			Errors: []string{
				"service \"Service\" method \"Method\": payload of method \"Method\" of service \"Service\" does not define a JWT attribute, use Token to define one",
			},
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
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.Errors == nil || len(c.Errors) == 0 {
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

				if len(c.Errors) != len(errors) {
					t.Errorf("got %d, expected the number of error values to match %d\nerrors:\n%s", len(errors), len(c.Errors), err.Error())
				} else {
					for i, err := range errors {
						if err.Error() != c.Errors[i] {
							t.Errorf("got \t\t%q,\nexpected\t%q at index %d", err.Error(), c.Errors[i], i)
						}
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
					t.Errorf("%s: got endpoint without body, expected endpoint with body", name)
					return
				}
				bodyObj := *expr.AsObject(e.Body.Type)
				expectedBodyObj := *expr.AsObject(tc.ExpectedBody)
				if len(bodyObj) != len(expectedBodyObj) {
					t.Errorf("%s: got %d, expected %d attribute(s) in endpoint body", name, len(bodyObj), len(expectedBodyObj))
				} else {
					for i := 0; i < len(expectedBodyObj); i++ {
						if bodyObj[i].Name != expectedBodyObj[i].Name {
							t.Errorf("%s: got %q, expected %q attribute in endpoint body", name, bodyObj[i].Name, expectedBodyObj[i].Name)
						}
					}
				}
			}
		})
	}
}
