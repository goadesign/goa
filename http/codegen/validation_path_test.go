// This file verifies that generated HTTP validators keep complete error paths
// while reusing named validators for nested and recursive values.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	gencodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestHTTPValidationPathsUseGeneratedCalls(t *testing.T) {
	root := expr.RunDSL(t, recursiveValidationPathDSL)
	code := renderedFiles(t, linkedHTTPPlanForRoot(t, root).ServerTypeFiles())

	require.Contains(t, code, `validateNodeRequestBody(body.First, "body.first")`)
	require.Contains(t, code, `validateNodeRequestBody(body.Second, "body.second")`)
	require.Contains(t, code, `validateNodeRequestBody(body.Next, "body.next")`)
	require.Contains(t, code, `validateNodeRequestBody(body.Next, path+".next")`)
	require.Contains(t, code, `validateNodeRequestBody(e, path+".children[*]")`)
	require.Contains(t, code, `validateNodeRequestBody(v, path+".children_by_name[key]")`)
	require.Contains(t, code, `goa.InvalidLengthError("body.value"`)
	require.Contains(t, code, `goa.InvalidLengthError(path+".value"`)
	require.Equal(t, 1, strings.Count(code, "func validateNodeRequestBody("))
	require.NotContains(t, code, "fmt.Sprintf")
}

func TestHTTPValidationPathsKeepMutualRecursion(t *testing.T) {
	root := expr.RunDSL(t, mutualValidationPathDSL)
	code := renderedFiles(t, linkedHTTPPlanForRoot(t, root).ServerTypeFiles())

	require.Contains(t, code, `validateLeftRequestBody(body.Left, "body.left")`)
	require.Contains(t, code, `validateRightRequestBody(body.Right, path+".right")`)
	require.Contains(t, code, `validateLeftRequestBody(body.Left, path+".left")`)
	require.Equal(t, 1, strings.Count(code, "func validateLeftRequestBody("))
	require.Equal(t, 1, strings.Count(code, "func validateRightRequestBody("))
}

func TestHTTPValidationPathsOmitUnusedNestedHelper(t *testing.T) {
	root := expr.RunDSL(t, unusedClientNestedValidationDSL)
	code := renderedFiles(t, linkedHTTPPlanForRoot(t, root).ClientTypeFiles())

	require.Contains(t, code, "func ValidateChildRequestBody(")
	require.NotContains(t, code, "func validateChildRequestBody(")
}

func TestHTTPValidationPathsOmitUnusedViewedSSENestedHelper(t *testing.T) {
	root := expr.RunDSL(t, viewedSSENestedValidationDSL)
	var plan *Plan
	require.NotPanics(t, func() {
		plan = linkedHTTPPlanForRoot(t, root)
	})
	types := renderedFiles(t, plan.ClientTypeFiles())
	serviceFiles, err := service.Files(plan.servicePlan)
	require.NoError(t, err)
	views := renderedFiles(t, serviceFiles)

	require.NotContains(t, types, "func validateProfile(")
	require.Contains(t, views, `goa.ValidatePattern("result.code"`)
}

func TestHTTPValidationPathsOmitUnusedClientArrayElementValidator(t *testing.T) {
	root := expr.RunDSL(t, testdata.PayloadBodyPrimitiveArrayUserRequiredDSL)
	generation, err := gencodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)

	planned := plans[0].wireTypes[root.API.HTTP.Services[0]]
	var clientElement, serverElement *wireTypeRecord
	for _, record := range planned.client.records {
		if record.identity.preferred == "PayloadType" {
			clientElement = record
			break
		}
	}
	for _, record := range planned.server.records {
		if record.identity.preferred == "PayloadType" {
			serverElement = record
			break
		}
	}
	require.NotNil(t, clientElement)
	require.False(t, clientElement.needsNestedCall)
	require.Nil(t, clientElement.nestedValidator)
	require.NotNil(t, serverElement)
	require.True(t, serverElement.needsNestedCall)
	require.NotNil(t, serverElement.nestedValidator)
}

// recursiveValidationPathDSL defines one named type used by two fields and by
// its own object, array, and map fields.
func recursiveValidationPathDSL() {
	node := Type("Node", func() {
		Attribute("value", String, func() {
			MinLength(1)
		})
		Attribute("next", "Node")
		Attribute("children", ArrayOf("Node"))
		Attribute("children_by_name", MapOf(String, "Node"))
	})
	payload := Type("Payload", func() {
		Attribute("first", node)
		Attribute("second", node)
	})
	Service("RecursiveValidation", func() {
		Method("Check", func() {
			Payload(payload)
			HTTP(func() {
				POST("/check")
			})
		})
	})
}

// mutualValidationPathDSL defines two named types that refer to each other.
func mutualValidationPathDSL() {
	left := Type("Left", func() {
		Attribute("right", "Right")
	})
	Type("Right", func() {
		Attribute("code", String, func() {
			Pattern("^[a-z]+$")
		})
		Attribute("left", "Left")
	})
	payload := Type("MutualPayload", func() {
		Attribute("left", left)
	})
	Service("MutualValidation", func() {
		Method("Check", func() {
			Payload(payload)
			HTTP(func() {
				POST("/check")
			})
		})
	})
}

// unusedClientNestedValidationDSL defines a child validator whose client
// request body has no generated validation call to that child.
func unusedClientNestedValidationDSL() {
	child := Type("Child", func() {
		Attribute("code", String, func() {
			Pattern("^[a-z]+$")
		})
	})
	payload := Type("UnusedNestedPayload", func() {
		Attribute("child", child)
	})
	Service("UnusedNestedValidation", func() {
		Method("Check", func() {
			Payload(payload)
			HTTP(func() {
				POST("/check")
			})
		})
	})
}

// viewedSSENestedValidationDSL defines a viewed event whose views package
// validates one nested named value.
func viewedSSENestedValidationDSL() {
	profile := Type("Profile", func() {
		Attribute("code", String, func() {
			Pattern("^[a-z]+$")
		})
	})
	event := ResultType("application/vnd.viewed-sse-nested-validation", func() {
		TypeName("ViewedSSENestedValidation")
		Attribute("profile", profile)
		Required("profile")
		View("summary", func() {
			Attribute("profile")
		})
		View("detailed", func() {
			Attribute("profile")
		})
	})
	Service("Viewed SSE Nested Validation", func() {
		Method("Watch", func() {
			StreamingResult(event)
			HTTP(func() {
				GET("/watch")
				ServerSentEvents()
			})
		})
	})
}
