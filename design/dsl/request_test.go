package dsl_test

import (
	"testing"

	"reflect"

	"goa.design/goa.v2/design"
	. "goa.design/goa.v2/design/dsl"
	"goa.design/goa.v2/eval"
)

func TestRequest(t *testing.T) {
	var RequestType = Type("Request", func() {
		Description("Optional description")
		Attribute("required", design.String)
		Attribute("name", design.String)
		Required("required")
	})
	cases := map[string]struct {
		DSL         func()
		Assert      func(t *testing.T, req design.DataType)
		ExpectError bool
	}{
		"DefinedUserType": {
			DSL: func() {
				Request(RequestType, func() {
					Required("name")
				})
			},
			Assert: func(t *testing.T, req design.DataType) {
				v, ok := req.(*design.UserTypeExpr)
				if !ok || v == nil {
					t.Errorf("expected request to be a design.UserTypeExpr got %v ", reflect.TypeOf(v))
					return
				}
				t.Log(v.AllRequired())
				if v.Name() != "TestServiceRequest" {
					t.Errorf("expected the Request UserType to have the name TestServiceRequest but got %s", v.Name())
				}
			},
			ExpectError: false,
		},
		"InlineAttributes": {
			DSL: func() {
				Request(func() {
					Attribute("testOne", design.String, "a test attribute")
					Attribute("testTwo")
					Attribute("testNumber", design.Int64, "a number")
					Required("testOne")
				})
			},
			Assert: func(t *testing.T, req design.DataType) {
				v, ok := req.(*design.UserTypeExpr)
				if !ok || v == nil {
					t.Errorf("expected request to be a design.UserTypeExpr got %v ", reflect.TypeOf(v))
					return
				}
				o, ok := v.Type.(design.Object)
				if !ok || o == nil {
					t.Errorf("expected request type to be a design.Object got %v ", reflect.TypeOf(o))
					return
				}
				keys := []string{}
				for k, _ := range o {
					keys = append(keys, k)
				}

				assertHasAll(t, []string{"testOne"}, v.AllRequired())
				assertHasAll(t, []string{"testOne", "testTwo", "testNumber"}, keys)
				assertDescription(t, "a test attribute", o["testOne"].Description)
				assertDescription(t, "", o["testTwo"].Description)
				assertDescription(t, "a number", o["testNumber"].Description)
				assertAttributeType(t, o["testOne"], design.String)
				assertAttributeType(t, o["testTwo"], design.String)
				assertAttributeType(t, o["testNumber"], design.Int64)
			},
			ExpectError: false,
		},
		"IncorrectRequiredFields": {
			DSL: func() {
				Request(func() {
					Attribute("testOne", design.String, "a test attribute")
					Required("testTwo")
				})
			},
			Assert:      nil,
			ExpectError: true,
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			endpointExpr := &design.EndpointExpr{Service: &design.ServiceExpr{Name: "test service"}}
			eval.Execute(tc.DSL, endpointExpr)
			if eval.Context.Errors != nil && !tc.ExpectError {
				t.Fatalf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			if eval.Context.Errors == nil && tc.ExpectError {
				t.Fatalf("%s: Expected context error but got none", k)
			}
			if tc.Assert != nil {
				tc.Assert(t, endpointExpr.Request)
			}
		})
	}
}

func assertHasAll(t *testing.T, expected []string, has []string) {
	if len(expected) != len(has) {
		t.Errorf("expected the expected and has required fields to match in length expected %d got %d ", len(expected), len(has))
	}
	for _, r := range expected {
		found := false
		for _, h := range has {
			if h == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("failed to find required value %s in %v ", r, has)
		}
	}
}

func assertAttributeType(t *testing.T, actual *design.AttributeExpr, expected design.DataType) {
	if actual.Type != expected {
		t.Errorf("expected attribute type to match %v but got %v", actual.Type.Name(), expected.Name())
	}
}
