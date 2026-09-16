// This file verifies that service analysis retains the exact mutable example
// generator owned by its generation run.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestServicesDataRetainsRunExampleGenerator(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(dsl.String)
			})
		})
	})
	root.API.RandomizerFactory = expr.NewDeterministicRandomizerFactory()
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	examples := expr.NewExampleGenerator(root.API.RandomizerFactory)
	plan, err := NewPlan(root, generation, examples)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	services := plan.Services()
	require.NoError(t, err)
	attribute := &expr.AttributeExpr{Type: expr.String}
	method := root.Services[0].Methods[0]
	owner := expr.MethodResultExampleIdentity(method)
	require.Equal(t, "abc123", services.Example(attribute, owner))
	require.Equal(t, "abc123", services.FieldExample(attribute, attribute, "value", owner))
}

func TestRepeatedServiceReadsKeepAnonymousExamplesStable(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Primitive", func() {
				dsl.Payload(dsl.String)
				dsl.Result(dsl.String)
			})
			dsl.Method("Array", func() {
				dsl.Payload(dsl.ArrayOf(dsl.String))
				dsl.Result(dsl.ArrayOf(dsl.Int))
			})
			dsl.Method("Map", func() {
				dsl.Payload(dsl.MapOf(dsl.String, dsl.Int))
				dsl.Result(dsl.MapOf(dsl.Int, dsl.String))
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	examples := expr.NewExampleGenerator(root.API.RandomizerFactory)
	plan, err := NewPlan(root, generation, examples)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	first := plan.Services()
	second := plan.Services()
	require.Len(t, first.Get("Values").Methods, 3)
	require.Len(t, second.Get("Values").Methods, 3)
	for index, firstMethod := range first.Get("Values").Methods {
		secondMethod := second.Get("Values").Methods[index]
		require.Equal(t, firstMethod.PayloadEx, secondMethod.PayloadEx, firstMethod.Name+" payload")
		require.Equal(t, firstMethod.ResultEx, secondMethod.ResultEx, firstMethod.Name+" result")
	}
}
