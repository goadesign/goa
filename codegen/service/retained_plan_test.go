// This file verifies that service planning retains one immutable render model
// per design root. Definitions and references must consume the exact package-
// owned declaration record collected by that plan.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestServicePlanSharesDefinitionAndReferenceDeclaration catches service
// analysis that reconstructs a payload name independently from its definition.
func TestServicePlanSharesDefinitionAndReferenceDeclaration(t *testing.T) {
	var payload expr.UserType
	root := codegen.RunDSL(t, func() {
		payload = dsl.Type("Payload", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})

	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	root.Service("Values").Methods = nil
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())

	owner := generation.Package("generated.local/gen/values")
	declaration, err := owner.UserType(payload)
	require.NoError(t, err)
	services := plan.Services()
	require.Len(t, services.Get("Values").Methods, 1)
	require.Same(t, declaration, services.Get("Values").Methods[0].PayloadDeclaration)
}

// TestServicePlanSharesNestedValidatorDeclaration verifies that a projected
// parent call and the child function definition retain one package declaration
// even when another projected type collides with the child's preferred
// validator name.
func TestServicePlanSharesNestedValidatorDeclaration(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		child := dsl.ResultType("application/vnd.child", func() {
			dsl.TypeName("Child")
			dsl.Attribute("name", dsl.String)
			dsl.Required("name")
		})
		collision := dsl.Type("ValidateChild", func() {
			dsl.Attribute("value", dsl.String)
		})
		parent := dsl.ResultType("application/vnd.parent", func() {
			dsl.TypeName("Parent")
			dsl.Attribute("child", child)
			dsl.Attribute("collision", collision)
			dsl.Required("child")
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(parent)
			})
		})
	})

	plan := mustServicePlan(t, root)
	data := plan.Services().Get("Values")
	var child, parent *ValidateData
	for _, projected := range data.projectedTypes {
		for _, validation := range projected.Validations {
			switch projected.Name {
			case "ChildView":
				child = validation
			case "ParentView":
				parent = validation
			}
		}
	}
	require.NotNil(t, child)
	require.NotNil(t, parent)
	require.Len(t, parent.Calls, 1)
	require.Same(t, child.Declaration, parent.Calls[0].Declaration)
	require.Equal(t, child.Declaration.Name(), parent.Calls[0].Declaration.Name())
}
