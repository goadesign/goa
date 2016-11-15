package dsl_test

import (
	"testing"

	"fmt"

	"github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
)

func TestEndpoint(t *testing.T) {
	cases := map[string]struct {
		Expr   eval.Expression
		DSL    func()
		Assert func(testName string, t *testing.T, s *design.ServiceExpr)
	}{
		"basic endpoint": {
			&design.ServiceExpr{},
			func() {
				Endpoint("basic", func() {
					Description("basic endpoint")
				})
			},
			func(testName string, t *testing.T, s *design.ServiceExpr) {
				if len(s.Endpoints) != 1 {
					t.Errorf("%s: expected %d endpoints but got %d", testName, 1, len(s.Endpoints))
				}
				for _, e := range s.Endpoints {
					if err := assertEndpointDescription("basic endpoint", e.Description); err != nil {
						t.Errorf("%s assert failed %s ", testName, err.Error())
					}
				}
			},
		},
		"basic endpoint with docs": {
			&design.ServiceExpr{},
			func() {
				Endpoint("basic", func() {
					Description("basic endpoint")
					Docs(func() {
						URL("http://example.com")
						Description("some docs")
					})
				})
			},
			func(testName string, t *testing.T, s *design.ServiceExpr) {
				if len(s.Endpoints) != 1 {
					t.Errorf("%s: expected %d endpoints but got %d", testName, 1, len(s.Endpoints))
				}
				for _, e := range s.Endpoints {
					if err := assertEndpointDescription("basic endpoint", e.Description); err != nil {
						t.Errorf("%s assert failed %s ", testName, err.Error())
					}
					if err := assertEndpointDocs(e.Docs, "http://example.com", "some docs"); err != nil {
						t.Errorf("%s assert failed %s ", testName, err.Error())
					}
				}
			},
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			eval.Execute(tc.DSL, tc.Expr)
			//After evaling the service our endpoints are present but need to also need to have thier DSL func exectuted.
			evalService := tc.Expr.(*design.ServiceExpr)
			for _, endpointExp := range evalService.Endpoints {
				eval.Execute(endpointExp.DSLFunc, endpointExp)
			}
			if eval.Context.Errors != nil {
				t.Errorf("%s: Endpoint failed unexpectedly with %s", k, eval.Context.Errors)
			}
			if tc.Assert != nil {
				tc.Assert(k, t, evalService)
			}
		})
	}
}

//helper funcs
func assertEndpointDocs(doc *design.DocsExpr, url, desc string) error {
	if doc.Description != desc {
		return fmt.Errorf("expected docs description '%s' to match '%s' ", desc, doc.Description)
	}
	if doc.URL != url {
		return fmt.Errorf("expected docs url '%s' to match '%s' ", url, doc.URL)
	}
	return nil
}

func assertEndpointDescription(expectedDesc, actualDesc string) error {
	if expectedDesc != actualDesc {
		return fmt.Errorf("expected description '%s' to match '%s' ", actualDesc, expectedDesc)
	}
	return nil
}
