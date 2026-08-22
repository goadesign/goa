// This file verifies service package declarations are collected only for the
// conditions that emit them and never depend on declarations in other packages.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestRelocatedResultViewConstructorsCompile catches constructor declarations
// that incorrectly depend on a result type declaration owned by another Go
// package.
func TestRelocatedResultViewConstructorsCompile(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		reading := dsl.ResultType("application/vnd.reading", func() {
			dsl.TypeName("Reading")
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("value", dsl.Int)
			dsl.Required("name", "value")
			dsl.View("default", func() {
				dsl.Attribute("name")
				dsl.Attribute("value")
			})
			dsl.View("summary", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(reading)
			})
		})
	})

	plan := retainedServicePlanForPackage(t, root, "generated.local/gen")
	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	compileGeneratedServiceFiles(t, "generated.local", files)
}

// TestJSONRPCSSEEventNameIsDeclaredOnlyWhenEmitted catches an unused Event
// declaration that changes collision suffixes for methods with no result.
func TestJSONRPCSSEEventNameIsDeclaredOnlyWhenEmitted(t *testing.T) {
	cases := []struct {
		name      string
		result    expr.DataType
		wantEvent bool
		wantName  string
	}{
		{name: "no result", result: expr.Empty},
		{name: "emits event", result: expr.String, wantEvent: true, wantName: "Event2"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			servicePackage := mustClaimTestPackage(t, generation, "generated.local/gen/events")
			mustClaimTestPackage(t, generation, "generated.local/gen/events/views")
			authored, err := servicePackage.DeclareUserType(&expr.UserTypeExpr{
				AttributeExpr: &expr.AttributeExpr{Type: expr.String},
				TypeName:      "Event",
				UID:           "authored-event",
			})
			require.NoError(t, err)

			result := &expr.AttributeExpr{Type: test.result}
			method := &expr.MethodExpr{
				Name:             "Watch",
				Payload:          &expr.AttributeExpr{Type: expr.Empty},
				Result:           result,
				Meta:             expr.MetaExpr{"jsonrpc": []string{}},
				Stream:           expr.ServerStreamKind,
				StreamingPayload: &expr.AttributeExpr{Type: expr.Empty},
				StreamingResult:  result,
			}
			service := &expr.ServiceExpr{Name: "Events", Methods: []*expr.MethodExpr{method}}
			method.Service = service
			facts := &serviceFacts{
				service: service,
				methods: []*expr.MethodExpr{method},
				methodByExpr: map[*expr.MethodExpr]*methodFacts{
					method: {
						method:       method,
						varName:      "Watch",
						isJSONRPCSSE: true,
					},
				},
				projections: make(map[*expr.MethodExpr]*projectionFacts),
			}

			require.NoError(t, collectServiceNames(facts, &rootTypeSet{byOrigin: make(map[expr.UserType]expr.UserType)}, generation))
			require.NoError(t, generation.Freeze())
			require.Equal(t, "Event", authored.Name())
			event, exists := facts.names[serviceSymbolID{
				role:    serviceEventNameRole,
				service: service.Name,
			}]
			require.Equal(t, test.wantEvent, exists)
			if test.wantEvent {
				require.Equal(t, test.wantName, event.declaration.Name())
			}
		})
	}
}
