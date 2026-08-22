// This file verifies recursive transform helpers retain the field-presence
// operation selected by each result view from planning through rendered calls.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type retainedTransformOperation struct {
	conversion *viewConversionFacts
	init       *InitData
	helper     codegen.TransformHelper
	definition *codegen.TransformFunctionData
}

// TestRecursiveTransformHelpersRetainRequiredness catches required and
// optional recursive field operations that are collapsed because their named
// source and target types have the same origins.
func TestRecursiveTransformHelpersRetainRequiredness(t *testing.T) {
	forwardPlan := recursiveTransformPlan(t, false)
	operations := retainedRecursiveTransformOperations(t, forwardPlan)
	required := operations[expr.DefaultView]
	optional := operations["optional"]
	require.NotNil(t, required)
	require.NotNil(t, optional)

	require.NotSame(t, required.definition.Declaration, optional.definition.Declaration)
	require.NotEqual(t, required.definition.Code, optional.definition.Code)
	require.NotContains(t, required.definition.Code, "if v == nil")
	require.Contains(t, optional.definition.Code, "if v == nil")
	require.Contains(t, required.init.Code, required.definition.Declaration.Name()+"(")
	require.Contains(t, optional.init.Code, optional.definition.Declaration.Name()+"(")
	require.Same(
		t,
		required.helper.Declaration,
		required.definition.Declaration,
	)
	require.Same(
		t,
		optional.helper.Declaration,
		optional.definition.Declaration,
	)

	reversePlan := recursiveTransformPlan(t, true)
	reverseOperations := retainedRecursiveTransformOperations(t, reversePlan)
	for _, view := range []string{expr.DefaultView, "optional"} {
		require.Equal(t, operations[view].definition.Declaration.Name(), reverseOperations[view].definition.Declaration.Name())
		require.Equal(t, operations[view].definition.Code, reverseOperations[view].definition.Code)
	}

	forwardFiles, err := Files(forwardPlan)
	require.NoError(t, err)
	reverseFiles, err := Files(reversePlan)
	require.NoError(t, err)

	compileFiles := append([]*codegen.File(nil), forwardFiles...)
	compileFiles = append(compileFiles, ExampleServiceFiles(forwardPlan)...)
	compileGeneratedServiceFiles(t, "generated.local", compileFiles)
	reverseCompileFiles := append([]*codegen.File(nil), reverseFiles...)
	reverseCompileFiles = append(reverseCompileFiles, ExampleServiceFiles(reversePlan)...)
	compileGeneratedServiceFiles(t, "generated.local", reverseCompileFiles)
}

// TestRecursiveTransformHelpersRetainSiblingOccurrences catches package-name
// planning that collapses two optional fields because their recursive types
// share an authored origin and requiredness.
func TestRecursiveTransformHelpersRetainSiblingOccurrences(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		node := dsl.Type("Node", func() {
			dsl.Attribute("label", dsl.String)
			dsl.Attribute("next", "Node")
			dsl.Required("label")
		})
		tree := dsl.ResultType("application/vnd.sibling-tree", func() {
			dsl.TypeName("SiblingTree")
			dsl.Attribute("left", node)
			dsl.Attribute("right", node)
			dsl.View(expr.DefaultView, func() {
				dsl.Attribute("left")
				dsl.Attribute("right")
			})
		})
		dsl.Service("Trees", func() {
			dsl.Method("Read", func() {
				dsl.Result(tree)
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root, "generated.local/gen")
	facts := plan.facts.serviceByID["Trees"]
	require.NotNil(t, facts)
	projected := facts.projections[facts.methods[0]].types[0]
	var conversion *viewConversionFacts
	for _, candidate := range projected.conversions {
		if !candidate.toResult && candidate.viewName == expr.DefaultView {
			conversion = candidate
			break
		}
	}
	require.NotNil(t, conversion)
	helpers := conversion.plan.Helpers()
	require.Len(t, helpers, 2)
	require.False(t, helpers[0].Required)
	require.False(t, helpers[1].Required)
	require.NotSame(t, helpers[0].Declaration, helpers[1].Declaration)

	serviceData := plan.Services().Get("Trees")
	require.NotNil(t, serviceData)
	var init *InitData
	for _, projectedData := range serviceData.projectedTypes {
		for _, candidate := range projectedData.Projections {
			if candidate.Declaration == conversion.constructor {
				init = candidate
				break
			}
		}
	}
	require.NotNil(t, init)
	require.Len(t, init.Helpers, 2)
	for _, helper := range helpers {
		require.Contains(t, init.Code, helper.Declaration.Name()+"(")
		var definition *codegen.TransformFunctionData
		for _, candidate := range init.Helpers {
			if candidate.ID == helper.ID {
				definition = candidate
				break
			}
		}
		require.NotNil(t, definition)
		require.Same(t, helper.Declaration, definition.Declaration)
	}

	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	compileGeneratedServiceFiles(t, "generated.local", files)
}

// recursiveTransformPlan builds equivalent result designs in either field and
// view order, then completes their retained service planning lifecycle.
func recursiveTransformPlan(t *testing.T, reverse bool) *Plan {
	t.Helper()
	root := codegen.RunDSL(t, func() {
		node := dsl.Type("Node", func() {
			dsl.Attribute("label", dsl.String)
			dsl.Attribute("next", "Node")
			dsl.Required("label")
		})
		tree := dsl.ResultType("application/vnd.tree", func() {
			dsl.TypeName("Tree")
			requiredField := func() {
				dsl.Attribute("required_node", node)
			}
			optionalField := func() {
				dsl.Attribute("optional_node", node)
			}
			if reverse {
				optionalField()
				requiredField()
			} else {
				requiredField()
				optionalField()
			}
			dsl.Required("required_node")

			requiredView := func() {
				dsl.View(expr.DefaultView, func() {
					dsl.Attribute("required_node")
				})
			}
			optionalView := func() {
				dsl.View("optional", func() {
					dsl.Attribute("optional_node")
				})
			}
			if reverse {
				optionalView()
				requiredView()
			} else {
				requiredView()
				optionalView()
			}
		})
		dsl.Service("Trees", func() {
			dsl.Method("Read", func() {
				dsl.Result(tree)
			})
		})
	})
	return retainedServicePlanForPackage(t, root, "generated.local/gen")
}

// retainedRecursiveTransformOperations returns the service-to-view helper
// operation, its call binding, and its rendered definition for each view.
func retainedRecursiveTransformOperations(t *testing.T, plan *Plan) map[string]*retainedTransformOperation {
	t.Helper()
	facts := plan.facts.serviceByID["Trees"]
	require.NotNil(t, facts)
	data := plan.Services().Get("Trees")
	require.NotNil(t, data)

	var projectedFacts *projectedTypeFacts
	for _, candidate := range facts.projections[facts.methods[0]].types {
		if candidate.pair.source.Name() == "Tree" {
			projectedFacts = candidate
			break
		}
	}
	require.NotNil(t, projectedFacts)
	var projectedData *ProjectedTypeData
	for _, candidate := range data.projectedTypes {
		if candidate.Type.Origin() == projectedFacts.pair.projected.Origin() {
			projectedData = candidate
			break
		}
	}
	require.NotNil(t, projectedData)

	operations := make(map[string]*retainedTransformOperation)
	for _, conversion := range projectedFacts.conversions {
		if conversion.toResult {
			continue
		}
		helpers := conversion.plan.Helpers()
		require.NotEmpty(t, helpers, conversion.viewName)
		var selected codegen.TransformHelper
		required := conversion.viewName == expr.DefaultView
		for _, helper := range helpers {
			if helper.Required == required {
				selected = helper
				break
			}
		}
		require.NotNil(t, selected.Declaration, conversion.viewName)
		var init *InitData
		for _, candidate := range projectedData.Projections {
			if candidate.Declaration == conversion.constructor {
				init = candidate
				break
			}
		}
		require.NotNil(t, init, conversion.viewName)
		var definition *codegen.TransformFunctionData
		for _, helper := range init.Helpers {
			if helper.ID == selected.ID {
				definition = helper
				break
			}
		}
		require.NotNil(t, definition, conversion.viewName)
		operations[conversion.viewName] = &retainedTransformOperation{
			conversion: conversion,
			init:       init,
			helper:     selected,
			definition: definition,
		}
	}
	return operations
}
