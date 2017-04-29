package dsl_test

import (
	"testing"

	"goa.design/goa.v2/design"
	. "goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
)

func TestEndpoint(t *testing.T) {
	const (
		desc = "test description"
		url  = "test URL"
	)
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, s []*design.EndpointExpr)
	}{
		"a": {
			func() {
				Endpoint("a", func() {})
			},
			func(t *testing.T, endpoints []*design.EndpointExpr) {
				if len(endpoints) != 1 {
					t.Fatalf("a: expected 1 endpoint, got %d", len(endpoints))
				}
				endpoint := endpoints[0]
				if endpoint.Name != "a" {
					t.Fatalf("a: expected endpoint name to be %s, got %s", "a", endpoint.Name)
				}
			},
		},
		"b": {
			func() {
				Endpoint("b", func() {
					Docs(func() {
						Description(desc)
						URL(url)
					})
				})
			},
			func(t *testing.T, endpoints []*design.EndpointExpr) {
				if len(endpoints) != 1 {
					t.Fatalf("b: expected 1 endpoint, got %d", len(endpoints))
				}
				endpoint := endpoints[0]
				doc := endpoint.Docs
				if doc == nil {
					t.Fatalf("b: endpoint docs is nil")
				}
				if doc.Description != desc {
					t.Errorf("b: expected docs description '%s' to match '%s' ", desc, doc.Description)
				}
				if doc.URL != url {
					t.Errorf("b: expected docs url '%s' to match '%s' ", url, doc.URL)
				}
			},
		},
		"c": {
			func() {
				Endpoint("c", func() {
					Payload(func() {
						Description(desc)
						Attribute("required", design.String)
						Required("required")
					})
				})
			},
			func(t *testing.T, endpoints []*design.EndpointExpr) {
				if len(endpoints) != 1 {
					t.Fatalf("b: expected 1 endpoint, got %d", len(endpoints))
				}
				endpoint := endpoints[0]
				if endpoint == nil {
					t.Fatalf("c: endpoint is nil")
				}
				payload := endpoint.Payload
				if payload == nil {
					t.Fatalf("c: endpoint payload is nil")
				}
				if payload.Description != desc {
					t.Errorf("c: expected payload description '%s' to match '%s' ", desc, payload.Description)
				}
				attrs := design.AsObject(payload.Type)
				if _, ok := attrs["required"]; !ok {
					t.Errorf("c: expected a payload field with key required")
				}
				if !payload.IsRequired("required") {
					t.Errorf("c: expected the required field to be required")
				}
			},
		},
	}
	//Run our tests
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			serviceExpr := &design.ServiceExpr{}
			eval.Execute(tc.DSL, serviceExpr)
			if eval.Context.Errors != nil {
				t.Errorf("%s: Service DSL failed unexpectedly with %s", k, eval.Context.Errors)
			}
			for _, endpointExpr := range serviceExpr.Endpoints {
				eval.Execute(endpointExpr.DSLFunc, endpointExpr)
				if eval.Context.Errors != nil {
					t.Errorf("%s: Endpoint DSL failed unexpectedly with %s", k, eval.Context.Errors)
				}
			}
			tc.Assert(t, serviceExpr.Endpoints)
		})
	}
}
