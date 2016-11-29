package dsl_test

import (
	"testing"

	"fmt"

	"goa.design/goa.v2/design"
	. "goa.design/goa.v2/design/dsl"
	"goa.design/goa.v2/eval"
)

func TestRequest(t *testing.T) {
	var RequestType = Type("Request", func() {
		Description("Optional description")
		Attribute("required", design.String)
		Attribute("name", design.String, "a name")
		Required("required")
	})
	//useful for this test but not others so defined here
	commonRequestTypeAsserts := func(t *testing.T, o design.Object, ut *design.UserTypeExpr) {
		keys := []string{}
		for k, _ := range o {
			keys = append(keys, k)
		}
		assertHasAll(t, []string{"required"}, ut.AllRequired())
		assertHasAll(t, []string{"required", "name"}, keys)
		assertDescription(t, "", o["required"].Description)
		assertDescription(t, "a name", o["name"].Description)
		assertAttributeType(t, o["name"], design.String)
		assertAttributeType(t, o["required"], design.String)
	}
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, req design.DataType)
	}{
		"ArrayOfRequest": {
			DSL: func() {
				Request(ArrayOf(RequestType))
			},
			Assert: func(t *testing.T, req design.DataType) {
				v, ok := req.(*design.Array)
				if !ok || v == nil {
					t.Errorf("expected request to be a design.Array got %s ", req.Name())
					return
				}
				ut, err := populateAttr(v.ElemType)
				if err != nil {
					t.Fatalf("error populating Attr %s", err.Error())
				}
				o, ok := ut.Type.(design.Object)
				if !ok || o == nil {
					t.Errorf("expected request type to be a design.Object got %v ", v.ElemType.Type.Name())
					return
				}
				commonRequestTypeAsserts(t, o, ut)
			},
		},
		"MapOfRequest": {
			DSL: func() {
				Request(MapOf(design.String, RequestType))
			},
			Assert: func(t *testing.T, req design.DataType) {
				v, ok := req.(*design.Map)
				if !ok || v == nil {
					t.Errorf("expected request to be a design.Array got %s ", req.Name())
					return
				}
				ut, err := populateAttr(v.ElemType)
				if err != nil {
					t.Fatalf("error populating Attr %s", err.Error())
				}
				o, ok := ut.Type.(design.Object)
				if !ok || o == nil {
					t.Errorf("expected request type to be a design.Object got %v ", v.ElemType.Type.Name())
					return
				}
				commonRequestTypeAsserts(t, o, ut)
			},
		},
		"DefinedUserType": {
			DSL: func() {
				Request(RequestType)
			},
			Assert: func(t *testing.T, req design.DataType) {
				v, ok := req.(*design.UserTypeExpr)
				if !ok || v == nil {
					t.Errorf("expected request to be a design.UserTypeExpr got %s ", v.Name())
					return
				}
				if !eval.Execute(v.DSL(), v.AttributeExpr) {
					t.Fatalf("failed to execute UserType DSL func. err %s ", eval.Context.Error())
				}
				o, ok := v.Type.(design.Object)
				if !ok || o == nil {
					t.Errorf("expected request type to be a design.Object got %s ", v.Type.Name())
					return
				}
				commonRequestTypeAsserts(t, o, v)
			},
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
					t.Errorf("expected request to be a design.UserTypeExpr got %s ", req.Name())
					return
				}
				o, ok := v.Type.(design.Object)
				if !ok || o == nil {
					t.Errorf("expected request type to be a design.Object got %s ", v.Type.Name())
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
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			endpointExpr := &design.EndpointExpr{Service: &design.ServiceExpr{Name: "test service"}}
			eval.Execute(tc.DSL, endpointExpr)
			if eval.Context.Errors != nil {
				t.Fatalf("%s: failed unexpectedly with %s", k, eval.Context.Errors)
			}
			if tc.Assert != nil {
				tc.Assert(t, endpointExpr.Request)
			}
		})
	}
}

func assertHasAll(t *testing.T, expected []string, has []string) {
	if len(expected) != len(has) {
		t.Errorf("expected fields to match in length expected %d got %d ", len(expected), len(has))
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

// poputlates embeded AttributeExpr
func populateAttr(v *design.AttributeExpr) (*design.UserTypeExpr, error) {
	ut, ok := v.Type.(*design.UserTypeExpr)
	if !ok || ut == nil {
		return nil, fmt.Errorf("expected request to be a design.UserTypeExpr got %v ", v.Type.Name())
	}
	if !eval.Execute(ut.DSL(), ut.AttributeExpr) {
		return nil, fmt.Errorf("failed to execute ArrayOf DSL func. err %s ", eval.Context.Error())
	}
	if eval.Context.Errors != nil {
		return nil, fmt.Errorf("unexpected error executing attribute dsl %s ", eval.Context.Error())
	}
	return ut, nil
}

func assertAttributeType(t *testing.T, actual *design.AttributeExpr, expected design.DataType) {
	if actual.Type != expected {
		t.Errorf("expected attribute type to match %v but got %v", actual.Type.Name(), expected.Name())
	}
}
