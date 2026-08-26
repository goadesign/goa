// This file verifies method layouts retain fields below consecutive named
// types for transports and plugins that plan complete service conversions.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestMethodLayoutsRetainNestedNamedValues(t *testing.T) {
	var payloadChild, payloadGrandchild, resultGrandchild, streamingPayloadGrandchild, streamingResultGrandchild expr.UserType
	root := codegen.RunDSL(t, func() {
		payloadGrandchild = dsl.Type("PayloadGrandchild", func() {
			dsl.Attribute("value", dsl.String)
		})
		payloadChild = dsl.Type("PayloadChild", func() {
			dsl.Attribute("grandchild", payloadGrandchild)
		})
		payload := dsl.Type("ReadPayload", func() {
			dsl.Attribute("child", payloadChild)
		})
		resultGrandchild = dsl.Type("ResultGrandchild", func() {
			dsl.Attribute("value", dsl.String)
		})
		resultChild := dsl.Type("ResultChild", func() {
			dsl.Attribute("grandchild", resultGrandchild)
		})
		result := dsl.ResultType("ReadResult", func() {
			dsl.Attribute("child", resultChild)
		})
		streamingPayloadGrandchild = dsl.Type("StreamingPayloadGrandchild", func() {
			dsl.Attribute("value", dsl.String)
		})
		streamingPayloadChild := dsl.Type("StreamingPayloadChild", func() {
			dsl.Attribute("grandchild", streamingPayloadGrandchild)
		})
		streamingPayload := dsl.Type("ReadStreamingPayload", func() {
			dsl.Attribute("child", streamingPayloadChild)
		})
		streamingResultGrandchild = dsl.Type("StreamingResultGrandchild", func() {
			dsl.Attribute("value", dsl.String)
		})
		streamingResultChild := dsl.Type("StreamingResultChild", func() {
			dsl.Attribute("grandchild", streamingResultGrandchild)
		})
		streamingResult := dsl.Type("ReadStreamingResult", func() {
			dsl.Attribute("child", streamingResultChild)
		})
		dsl.Service("Documents", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
				dsl.StreamingPayload(streamingPayload)
				dsl.StreamingResult(streamingResult)
				dsl.Result(result)
			})
		})
	})
	plan := mustServicePlan(t, root)
	method := root.Service("Documents").Method("Read")

	for _, test := range []struct {
		name       string
		grandchild expr.UserType
		layout     func(*expr.MethodExpr) (*codegen.GoTypePlan, error)
	}{
		{name: "payload", grandchild: payloadGrandchild, layout: plan.MethodPayloadLayout},
		{name: "result", grandchild: resultGrandchild, layout: plan.MethodResultLayout},
		{name: "streaming result", grandchild: streamingResultGrandchild, layout: plan.StreamingResultLayout},
	} {
		t.Run(test.name, func(t *testing.T) {
			layout, err := test.layout(method)
			require.NoError(t, err)
			matches := layout.PlansForOccurrence(test.grandchild.Attribute().Find("value"))
			require.Len(t, matches, 1)
			require.Equal(t, codegen.GoPrimitive, matches[0].Kind())
		})
	}

	methodFacts := plan.facts.services[0].methodByExpr[method]
	streamingPayloadMatches := methodFacts.streamingPayload.layout.PlansForOccurrence(
		streamingPayloadGrandchild.Attribute().Find("value"),
	)
	require.Len(t, streamingPayloadMatches, 1)

	var payloadFacts *userTypeFacts
	for _, candidate := range plan.facts.services[0].userTypes {
		if candidate.userType == payloadChild {
			payloadFacts = candidate
			break
		}
	}
	require.NotNil(t, payloadFacts)
	require.Empty(t, payloadFacts.layout.PlansForOccurrence(payloadGrandchild.Attribute().Find("value")))
}

// TestMethodTypeLayoutKeepsSourceAndProjectedDeclarations checks that creating
// a result view does not move the original service type into the views package.
func TestMethodTypeLayoutKeepsSourceAndProjectedDeclarations(t *testing.T) {
	var viewed, plain *expr.MethodExpr
	root := codegen.RunDSL(t, func() {
		child := dsl.Type("SharedDeclarationChild", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		result := dsl.ResultType("application/vnd.view-bindings", func() {
			dsl.TypeName("ViewedContainer")
			dsl.Attribute("child", child)
			dsl.Required("child")
			dsl.View("default", func() {
				dsl.Attribute("child")
			})
		})
		dsl.Service("ViewBindings", func() {
			viewed = dsl.Method("Viewed", func() {
				dsl.Result(result)
			})
			plain = dsl.Method("Plain", func() {
				dsl.Payload(func() {
					dsl.Attribute("child", child)
				})
			})
		})
	})
	plan := mustServicePlan(t, root)

	plainLayout, err := plan.MethodTypeLayout(plain, plain.Payload)
	require.NoError(t, err)
	plainChild := plainLayout.PlansForOccurrence(plain.Payload.Find("child"))
	require.Len(t, plainChild, 1)
	require.Equal(t, "goa.design/goa/example/view_bindings", plainChild[0].Owner())
	require.Equal(t, "SharedDeclarationChild", plainChild[0].TypeDeclaration().Name())

	projected, err := plan.ProjectedResult(viewed)
	require.NoError(t, err)
	projectedLayout, err := plan.MethodTypeLayout(viewed, projected)
	require.NoError(t, err)
	projectedChild := projectedLayout.PlansForOccurrence(projected.Find("child"))
	require.Len(t, projectedChild, 1)
	require.Equal(t, "goa.design/goa/example/view_bindings/views", projectedChild[0].Owner())
	require.Equal(t, "SharedDeclarationChildView", projectedChild[0].TypeDeclaration().Name())
}
