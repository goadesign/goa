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
		Invocations int
		Description string
		Doc         string
		DocURL      string
		Name        string
		RequestAttr map[string]design.Primitive
	}{
		"basic endpoint": {
			&design.EndpointExpr{Service: &design.ServiceExpr{}},
			1,
			"A description",
			"some docs",
			"http://docs.com",
			"A basic endpoint",
			map[string]design.Primitive{
				"test": design.String,
			},
		},
		"basic endpoint no doc": {
			&design.EndpointExpr{Service: &design.ServiceExpr{}},
			1,
			"A description",
			"",
			"",
			"A basic endpoint",
			map[string]design.Primitive{
				"test": design.String,
			},
		},
		"basic endpoint no description": {
			&design.EndpointExpr{Service: &design.ServiceExpr{}},
			1,
			"",
			"some docs",
			"http://docs.com",
			"A basic endpoint",
			map[string]design.Primitive{
				"test": design.String,
			},
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			for i := tc.Invocations; i > 0; i-- {
				eval.Execute(func() {
					if tc.Description != "" {
						Description(tc.Description)
					}
					if tc.Doc != "" {
						Docs(func() {
							Description(tc.Doc)
							URL(tc.DocURL)
						})
					}
					Request(func() {
						for k, v := range tc.RequestAttr {
							Attribute(k, v)
						}

					})
					//NOTE no Response func is defined yet
				}, tc.Expr)
			}
			if eval.Context.Errors != nil {
				t.Errorf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			endpoint := tc.Expr.(*design.EndpointExpr)

			if tc.Description != "" && endpoint.Description != tc.Description {
				t.Errorf("%s: expected endpoint.Description to match: '%s' but got: '%s' ", k, tc.Description, endpoint.Description)
			}
			if tc.Doc != "" && endpoint.Docs == nil {
				t.Errorf("%s did not expect endpoint.Docs to be nil", k)
			}
			if endpoint.Docs != nil && endpoint.Docs.Description != tc.Doc {
				t.Errorf("%s: expected the endpoint.Docs.Description '%s' to match '%s' ", k, endpoint.Docs.Description, tc.Doc)
			}
			if endpoint.Docs != nil && endpoint.Docs.URL != tc.DocURL {
				t.Errorf("%s: expected the endpoint.Docs.URL '%s' to match '%s' ", k, endpoint.Docs.URL, tc.DocURL)
			}
			if endpoint.Request == nil {
				t.Errorf("%s the endpoint.Request definition should not be nil", k)
			}
			if endpoint.Request.Name() != "Request" {
				t.Errorf("%s endpoint.Request should have the name Request ", k)
			}
		})
	}
}
