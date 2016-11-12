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
			&design.ServiceExpr{},
			1,
			"A description",
			"some docs",
			"http://docs.com",
			"BasicEndpoint",
			map[string]design.Primitive{
				"test": design.String,
			},
		},
		"basic endpoint no doc": {
			&design.ServiceExpr{},
			1,
			"A description",
			"",
			"",
			"BasicEndpoint",
			map[string]design.Primitive{
				"test": design.String,
			},
		},
		"basic endpoint no description": {
			&design.ServiceExpr{},
			1,
			"",
			"some docs",
			"http://docs.com",
			"BasicEndpoint",
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
					Endpoint(tc.Name, func() {
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
					})

				}, tc.Expr)
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
				if endpoint.Request.Name() != tc.Name+"Request" {
					t.Errorf("%s endpoint.Request should have the name Request but had %s ", k, endpoint.Request.Name())
				}
			}
		})
	}
}
