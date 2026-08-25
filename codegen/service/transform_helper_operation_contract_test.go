// This file verifies recursive transform calls retain the field-presence
// operation selected by each result view from planning through rendered code.
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

// TestRecursiveTransformCallsRetainRequiredness checks that optional values
// are tested for nil before the same strict conversion used by required values.
func TestRecursiveTransformCallsRetainRequiredness(t *testing.T) {
	forwardPlan := recursiveTransformPlan(t, false)
	operations := retainedRecursiveTransformOperations(t, forwardPlan)
	required := operations[expr.DefaultView]
	optional := operations["optional"]
	require.NotNil(t, required)
	require.NotNil(t, optional)

	require.NotSame(t, required.definition.Declaration, optional.definition.Declaration)
	require.NotContains(t, required.definition.Code, "if v == nil")
	require.NotContains(t, optional.definition.Code, "if v == nil")
	require.Contains(t, required.init.Code, required.definition.Declaration.Name()+"(")
	require.Contains(t, optional.init.Code, "if res.OptionalNode != nil {")
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
	compileGeneratedServiceFiles(t, compileFiles)
	reverseCompileFiles := append([]*codegen.File(nil), reverseFiles...)
	reverseCompileFiles = append(reverseCompileFiles, ExampleServiceFiles(reversePlan)...)
	compileGeneratedServiceFiles(t, reverseCompileFiles)
}

// TestRecursiveTransformCallsShareSiblingHelper checks that required and
// optional fields call one strict conversion while the optional call stays
// inside its nil check.
func TestRecursiveTransformCallsShareSiblingHelper(t *testing.T) {
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
			dsl.Required("left")
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
	plan := retainedServicePlanForPackage(t, root)
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
	require.True(t, helpers[0].Required)
	require.False(t, helpers[1].Required)
	require.Same(t, helpers[0].Declaration, helpers[1].Declaration)

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
	require.Len(t, init.Helpers, 1)
	require.Same(t, helpers[0].Declaration, init.Helpers[0].Declaration)
	require.NotContains(t, init.Helpers[0].Code, "if v == nil")
	require.Contains(t, init.Code, "vres.Left = "+helpers[0].Declaration.Name()+"(res.Left)")
	require.Contains(t, init.Code, "if res.Right != nil {")
	require.Contains(t, init.Code, "vres.Right = "+helpers[0].Declaration.Name()+"(res.Right)")

	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	compileGeneratedServiceFiles(t, files)
}

// TestTransformHelperNamesIgnoreSiblingFieldOrder verifies that the authored
// field path, not discovery order, decides which colliding helper name keeps
// its preferred spelling.
func TestTransformHelperNamesIgnoreSiblingFieldOrder(t *testing.T) {
	declarations := func(reverse bool) map[string]string {
		root := codegen.RunDSL(t, func() {
			node := dsl.Type("Node", func() {
				dsl.Attribute("label", dsl.String)
				dsl.Attribute("next", "Node")
				dsl.Required("label")
			})
			tree := dsl.ResultType("application/vnd.stable-tree", func() {
				dsl.TypeName("StableTree")
				field := func(name string) {
					dsl.Attribute(name, node, func() {
						dsl.Meta("test:helper", name)
					})
				}
				if reverse {
					field("right")
					field("left")
				} else {
					field("left")
					field("right")
				}
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
		plan := retainedServicePlanForPackage(t, root)
		facts := plan.facts.serviceByID["Trees"]
		projected := facts.projections[facts.methods[0]].types[0]
		var conversion *viewConversionFacts
		for _, candidate := range projected.conversions {
			if !candidate.toResult && candidate.viewName == expr.DefaultView {
				conversion = candidate
				break
			}
		}
		require.NotNil(t, conversion)
		definitions := conversion.plan.HelperDefinitions()
		require.Len(t, definitions, 2)
		names := make(map[string]string, len(definitions))
		for _, definition := range definitions {
			marker := definition.Source.Meta["test:helper"]
			require.Len(t, marker, 1)
			names[marker[0]] = definition.Declaration.Name()
		}
		return names
	}

	require.Equal(t, declarations(false), declarations(true))
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
	return retainedServicePlanForPackage(t, root)
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
