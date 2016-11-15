package dsl_test

import (
	"testing"

	"github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
)

func TestEndpoint(t *testing.T) {
	cases := map[string]struct {
		Expr        eval.Expression
		DSL         func()
		Assert      func(testName string, t *testing.T, e *design.EndpointExpr)
		Invocations int
	}{
		"basic endpoint": {
			&design.ServiceExpr{},
			func() {
				Endpoint("basic", func() {
					Description("basic endpoint")
				})
			},
			func(testName string, t *testing.T, e *design.EndpointExpr) {
				expected := "basic endpoint"
				if e.Description != expected {
					t.Errorf("%s: expected the description to be '%s' but got '%s' ", testName, expected, e.Description)
				}
			},
			1,
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			for i := tc.Invocations; i > 0; i-- {
				eval.Execute(tc.DSL, tc.Expr)
				//After evaling the service our endpoints are present but need to also need to have thier DSL func exectuted.
				evalService := tc.Expr.(*design.ServiceExpr)
				for _, endpointExp := range evalService.Endpoints {
					eval.Execute(endpointExp.DSLFunc, endpointExp)
				}
			}
			if eval.Context.Errors != nil {
				t.Errorf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			endpoints := tc.Expr.(*design.ServiceExpr).Endpoints
			if len(endpoints) != tc.Invocations {
				t.Errorf("%s: expected %d endpoints but got %d", k, tc.Invocations, len(endpoints))
			}
			for _, endpoint := range endpoints {
				if tc.Assert != nil {
					tc.Assert(k, t, endpoint)
				}
			}
		})
	}
}
