package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestServiceClientInterceptorPlan checks the retained interface and each method's
// applicability before and after the source expression lists change.
func TestServiceClientInterceptorPlan(t *testing.T) {
	cases := []struct {
		name    string
		declare bool
		server  bool
		api     bool
		service bool
		method  bool
	}{
		{name: "none"},
		{name: "unused declaration", declare: true},
		{name: "server only", server: true},
		{name: "method only", method: true},
		{name: "service only", service: true},
		{name: "API only", api: true},
		{name: "repeated attachment", api: true, service: true, method: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var trace *expr.InterceptorExpr
			root := codegen.RunDSL(t, func() {
				if tc.declare || tc.server || tc.api || tc.service || tc.method {
					trace = dsl.Interceptor("Trace")
				}
				dsl.API("interceptor plan", func() {
					if tc.api {
						dsl.ClientInterceptor(trace)
					}
				})
				dsl.Service("Events", func() {
					if tc.service {
						dsl.ClientInterceptor(trace)
					}
					if tc.server {
						dsl.ServerInterceptor(trace)
					}
					dsl.Method("Watch", func() {
						if tc.method {
							dsl.ClientInterceptor(trace)
						}
					})
					dsl.Method("Read", func() {})
				})
			})
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
			plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
			require.NoError(t, err)
			svc := root.Service("Events")
			want := tc.api || tc.service || tc.method
			present, err := plan.ServiceHasClientInterceptors(svc)
			require.NoError(t, err)
			require.Equal(t, want, present)

			// Later expression edits cannot give the import planner a different
			// answer from the service interface retained by the same plan.
			root.API.ClientInterceptors = nil
			svc.ClientInterceptors = nil
			for _, method := range svc.Methods {
				method.ClientInterceptors = nil
			}
			if !want {
				svc.ClientInterceptors = []*expr.InterceptorExpr{{Name: "AddedAfterPlanning"}}
			}
			present, err = plan.ServiceHasClientInterceptors(svc)
			require.NoError(t, err)
			require.Equal(t, want, present)
			require.NoError(t, generation.Freeze())
			require.NoError(t, plan.Link())
			data := plan.Services().Get("Events")
			require.Equal(t, want, len(data.ClientInterceptors) > 0)
			if want {
				require.Len(t, data.ClientInterceptors, 1)
			}
			for _, method := range data.Methods {
				applies := tc.api || tc.service || (tc.method && method.Name == "Watch")
				if applies {
					require.Equal(t, []string{"Trace"}, method.ClientInterceptors)
				} else {
					require.Empty(t, method.ClientInterceptors)
				}
			}
		})
	}
}

// TestServiceClientInterceptorPlanRejectsForeignService checks that a service
// with the same name from another root does not borrow this plan's answer.
func TestServiceClientInterceptorPlanRejectsForeignService(t *testing.T) {
	foreign := codegen.RunDSL(t, func() {
		dsl.API("foreign", func() {})
		dsl.Service("Events", func() {
			dsl.Method("Read", func() {})
		})
	})
	root := codegen.RunDSL(t, func() {
		dsl.API("owned", func() {})
		dsl.Service("Events", func() {
			dsl.Method("Read", func() {})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	for _, svc := range []*expr.ServiceExpr{nil, foreign.Service("Events")} {
		present, err := plan.ServiceHasClientInterceptors(svc)
		require.ErrorContains(t, err, "not part of this plan")
		require.False(t, present)
	}
}
