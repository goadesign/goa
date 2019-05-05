package dsl_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestMethod(t *testing.T) {
	const (
		desc = "test description"
		url  = "test URL"
	)
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, s []*expr.MethodExpr)
	}{
		"a": {
			func() {
				Method("a", func() {})
			},
			func(t *testing.T, methods []*expr.MethodExpr) {
				if len(methods) != 1 {
					t.Fatalf("a: expected 1 method, got %d", len(methods))
				}
				method := methods[0]
				if method.Name != "a" {
					t.Fatalf("a: expected method name to be %s, got %s", "a", method.Name)
				}
			},
		},
		"b": {
			func() {
				Method("b", func() {
					Docs(func() {
						Description(desc)
						URL(url)
					})
				})
			},
			func(t *testing.T, methods []*expr.MethodExpr) {
				if len(methods) != 1 {
					t.Fatalf("b: expected 1 method, got %d", len(methods))
				}
				method := methods[0]
				doc := method.Docs
				if doc == nil {
					t.Fatalf("b: method docs is nil")
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
				Method("c", func() {
					Payload(func() {
						Description(desc)
						Attribute("required", expr.String)
						Required("required")
					})
				})
			},
			func(t *testing.T, methods []*expr.MethodExpr) {
				if len(methods) != 1 {
					t.Fatalf("b: expected 1 method, got %d", len(methods))
				}
				method := methods[0]
				if method == nil {
					t.Fatalf("c: method is nil")
				}
				payload := method.Payload
				if payload == nil {
					t.Fatalf("c: method payload is nil")
				}
				if payload.Description != desc {
					t.Errorf("c: expected payload description '%s' to match '%s' ", desc, payload.Description)
				}
				obj := expr.AsObject(payload.Type)
				if att := obj.Attribute("required"); att == nil {
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
			serviceExpr := &expr.ServiceExpr{}
			eval.Execute(tc.DSL, serviceExpr)
			if eval.Context.Errors != nil {
				t.Errorf("%s: Service DSL failed unexpectedly with %s", k, eval.Context.Errors)
			}
			for _, methodExpr := range serviceExpr.Methods {
				eval.Execute(methodExpr.DSLFunc, methodExpr)
				if eval.Context.Errors != nil {
					t.Errorf("%s: Method DSL failed unexpectedly with %s", k, eval.Context.Errors)
				}
			}
			tc.Assert(t, serviceExpr.Methods)
		})
	}
}
