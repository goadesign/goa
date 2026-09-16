// This file proves service rendering uses only facts collected before the
// generation freezes, even if callers later mutate the evaluated expressions.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type retainedExpressionFixture struct {
	root        *expr.RootExpr
	service     *expr.ServiceExpr
	method      *expr.MethodExpr
	result      *expr.ResultTypeExpr
	interceptor *expr.InterceptorExpr
}

// TestServicePlanIgnoresRetainedExpressionMutation catches linking and
// rendering that reread mutable service, method, view, error, interceptor,
// security, example, or stream expressions after NewPlan returns.
func TestServicePlanIgnoresRetainedExpressionMutation(t *testing.T) {
	baselineFixture := retainedExpressionMutationFixture(t)
	baselinePlan := retainedServicePlanForPackage(t, baselineFixture.root)
	baseline := renderedPlanAndExamples(t, baselinePlan)
	baselineMethod := baselinePlan.Services().Get("RetainedMutable").Methods[0]

	tests := []struct {
		name   string
		mutate func(*retainedExpressionFixture)
	}{
		{"service and method", func(f *retainedExpressionFixture) {
			f.service.Name = "MutatedService"
			f.service.Description = "mutated service"
			f.method.Name = "MutatedMethod"
			f.method.Description = "mutated method"
			f.method.Idempotent = !f.method.Idempotent
		}},
		{"errors", func(f *retainedExpressionFixture) {
			f.service.Errors[0].Description = "mutated service error"
			f.method.Errors[0].Description = "mutated method error"
			f.method.Errors[0].Meta = expr.MetaExpr{"goa:error:fault": nil}
		}},
		{"interceptor", func(f *retainedExpressionFixture) {
			f.interceptor.Description = "mutated interceptor"
			f.interceptor.ReadPayload = nil
			f.interceptor.ReadStreamingPayload = nil
		}},
		{"security", func(f *retainedExpressionFixture) {
			f.method.Requirements[0].Scopes[0] = "mutated"
			f.method.Requirements[0].Schemes[0].Scopes[0].Name = "mutated"
		}},
		{"examples", func(f *retainedExpressionFixture) {
			f.method.Payload.UserExamples[0].Value = map[string]any{"key": "mutated"}
			f.method.StreamingPayload.UserExamples[0].Value = map[string]any{"chunk": "mutated"}
		}},
		{"stream", func(f *retainedExpressionFixture) {
			f.method.Stream = expr.NoStreamKind
			f.method.StreamingPayload.Description = "mutated streaming payload"
			f.method.StreamingResult.Description = "mutated streaming result"
		}},
		{"type layout", func(f *retainedExpressionFixture) {
			field := expr.AsObject(f.method.Payload.Type).Attribute("key")
			field.Description = "mutated field"
			field.Meta = expr.MetaExpr{"struct:field:name": []string{"MutatedKey"}}
		}},
		{"result and view", func(f *retainedExpressionFixture) {
			f.method.Result.Description = "mutated result"
			f.result.Views[0].Description = "mutated view"
			viewObject := expr.AsObject(f.result.Views[0].Type)
			viewObject.Set("extra", &expr.AttributeExpr{Type: expr.String})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := retainedExpressionMutationFixture(t)
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{fixture.root})
			plan, err := NewPlan(fixture.root, generation, expr.NewExampleGenerator(fixture.root.API.RandomizerFactory))
			require.NoError(t, err)
			test.mutate(fixture)
			require.NoError(t, generation.Freeze())
			require.NoError(t, plan.Link())
			requireRenderedServiceFilesEqual(t, baseline, renderedPlanAndExamples(t, plan))

			method := plan.Services().Get("RetainedMutable").Methods[0]
			require.Equal(t, baselineMethod.PayloadEx, method.PayloadEx)
			require.Equal(t, baselineMethod.StreamingPayloadEx, method.StreamingPayloadEx)
		})
	}
}

// retainedExpressionMutationFixture builds one service that exercises every
// expression family the retained core service plan must finish collecting.
func retainedExpressionMutationFixture(t *testing.T) *retainedExpressionFixture {
	t.Helper()
	fixture := new(retainedExpressionFixture)
	fixture.root = codegen.RunDSL(t, func() {
		auth := dsl.APIKeySecurity("key", func() {
			dsl.Scope("read", "Read values")
		})
		fixture.interceptor = dsl.Interceptor("Audit", func() {
			dsl.Description("Audits request values.")
			dsl.ReadPayload(func() { dsl.Attribute("key") })
			dsl.ReadStreamingPayload(func() { dsl.Attribute("chunk") })
		})
		result := dsl.ResultType("application/vnd.retained", func() {
			dsl.TypeName("RetainedResult")
			dsl.Description("The retained result.")
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
			dsl.View("default", func() { dsl.Attribute("value") })
			dsl.View("summary", func() { dsl.Attribute("value") })
		})
		fixture.result = result
		streamResult := dsl.Type("RetainedStreamResult", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		serviceError := dsl.Type("RetainedServiceError", func() {
			dsl.Attribute("message", dsl.String)
			dsl.Required("message")
		})
		methodError := dsl.Type("RetainedMethodError", func() {
			dsl.Attribute("message", dsl.String)
			dsl.Required("message")
		})
		fixture.service = dsl.Service("RetainedMutable", func() {
			dsl.Description("The retained mutable service.")
			dsl.Security(auth, func() { dsl.Scope("read") })
			dsl.ServerInterceptor(fixture.interceptor)
			dsl.ClientInterceptor(fixture.interceptor)
			dsl.Error("service_failed", serviceError, "The service failed.")
			fixture.method = dsl.Method("Watch", func() {
				dsl.Description("Watches retained values.")
				dsl.Payload(func() {
					dsl.APIKey("key", "key", dsl.String)
					dsl.Required("key")
					dsl.Example(map[string]any{"key": "original"})
				})
				dsl.StreamingPayload(func() {
					dsl.Attribute("chunk", dsl.String)
					dsl.Required("chunk")
					dsl.Example(map[string]any{"chunk": "original"})
				})
				dsl.Result(result)
				dsl.StreamingResult(streamResult)
				dsl.Error("method_failed", methodError, "The method failed.")
			})
		})
	})
	return fixture
}

// renderedPlanAndExamples renders both generated service packages and their
// starter implementation so post-link expression reads cannot hide in either.
func renderedPlanAndExamples(t *testing.T, plan *Plan) map[string][]byte {
	t.Helper()
	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	return renderedServiceFiles(t, files)
}
