package design_test

import (
	"testing"

	. "goa.design/goa/design"
	"goa.design/goa/eval"
	"goa.design/goa/http/design"
	"goa.design/goa/http/design/testdata"
)

func TestRouteValidation(t *testing.T) {
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
				design.RunHTTPDSL(t, c.DSL)
			} else {
				err := design.RunInvalidHTTPDSL(t, c.DSL)
				if err.Error() != c.Error {
					t.Errorf("got error %q, expected %q", err.Error(), c.Error)
				}
			}
		})
	}
}

func TestEndpointValidation(t *testing.T) {
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
				"\"name\" is not found in result type in service \"Service\" HTTP endpoint \"Method\"",
			},
		},
		"endpoint-body-extend-payload": {
			DSL: testdata.EndpointBodyExtendPayload,
		},
		"endpoint-body-as-user-type": {
			DSL: testdata.EndpointBodyAsUserType,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if c.Errors == nil || len(c.Errors) == 0 {
				design.RunHTTPDSL(t, c.DSL)
			} else {
				var errors []error

				err := design.RunInvalidHTTPDSL(t, c.DSL)
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
					t.Errorf("%s: got %d, expected the number of error values to match %d", name, len(errors), len(c.Errors))
				} else {
					for i, err := range errors {
						if err.Error() != c.Errors[i] {
							t.Errorf("%s:\ngot \t\t%q,\nexpected\t%q at index %d", name, err.Error(), c.Errors[i], i)
						}
					}
				}
			}
		})
	}
}

func TestEndpointFinalization(t *testing.T) {
	cases := map[string]struct {
		DSL          func()
		ExpectedBody DataType
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
			root := design.RunHTTPDSL(t, tc.DSL)
			e := root.HTTPServices[0].HTTPEndpoints[0]

			if tc.ExpectedBody != nil {
				if e.Body == nil {
					t.Errorf("%s: got endpoint without body, expected endpoint with body", name)
					return
				}
				bodyObj := *AsObject(e.Body.Type)
				expectedBodyObj := *AsObject(tc.ExpectedBody)
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
