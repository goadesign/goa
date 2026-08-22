// This file verifies retained service plans aggregate shared packages
// deterministically and render without changing their declaration catalogs.
package service

import (
	"bytes"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestFilesRejectPlansFromDifferentGenerations verifies that aggregation cannot
// render declarations through a package catalog that did not plan them.
func TestFilesRejectPlansFromDifferentGenerations(t *testing.T) {
	first := mustTestGeneration(t, "example.com/first/gen", nil)
	second := mustTestGeneration(t, "example.com/second/gen", nil)

	_, err := Files(&Plan{generation: first}, &Plan{generation: second})
	require.ErrorContains(t, err, "different generations")
}

type retainedServiceNameID struct {
	root    int
	service string
	symbol  serviceSymbolID
}

// TestServicePlansRenderByteIdenticallyAcrossRootAndServiceOrder catches
// shared-package names or section order that depend on discovery order.
func TestServicePlansRenderByteIdenticallyAcrossRootAndServiceOrder(t *testing.T) {
	forwardPlans := orderedServicePlans(t, false)
	forwardFiles, err := Files(forwardPlans...)
	require.NoError(t, err)
	forward := renderedServiceFiles(t, forwardFiles)

	reversePlans := orderedServicePlans(t, true)
	reverseFiles, err := Files(reversePlans...)
	require.NoError(t, err)
	reverse := renderedServiceFiles(t, reverseFiles)

	requireRenderedServiceFilesEqual(t, forward, reverse)
	requireRenderedServiceFile(t, forward, filepath.Join(codegen.Gendir, "types", "alpha_envelope.go"))
	requireRenderedServiceFile(t, forward, filepath.Join(codegen.Gendir, "types", "beta_envelope.go"))
	requireRenderedServiceFile(t, forward, filepath.Join(codegen.Gendir, "types", "omega_envelope.go"))
	requireRenderedServiceFile(t, forward, filepath.Join(codegen.Gendir, "types", "unions.go"))

	compileFiles := append([]*codegen.File(nil), forwardFiles...)
	for _, plan := range forwardPlans {
		compileFiles = append(compileFiles, ExampleServiceFiles(plan)...)
	}
	compileGeneratedServiceFiles(t, "generated.local", compileFiles)
}

// TestServicePlanRenderingIsPure catches renderers that rebuild analysis,
// replace retained declaration records, or append sections on a second read.
func TestServicePlanRenderingIsPure(t *testing.T) {
	plans := orderedServicePlans(t, false)
	before := retainedServiceNamePointers(plans)
	firstFiles, err := Files(plans...)
	require.NoError(t, err)
	first := renderedServiceFiles(t, firstFiles)

	secondFiles, err := Files(plans...)
	require.NoError(t, err)
	second := renderedServiceFiles(t, secondFiles)
	after := retainedServiceNamePointers(plans)

	require.Equal(t, first, second)
	require.Len(t, after, len(before))
	for id, declaration := range before {
		require.Same(t, declaration, after[id], "service symbol changed: %+v", id)
	}
}

// requireRenderedServiceFile reports a missing shared-package contribution
// without printing the complete generated output map.
func requireRenderedServiceFile(t *testing.T, files map[string][]byte, path string) {
	t.Helper()
	_, exists := files[path]
	require.True(t, exists, "missing generated file %s", path)
}

// requireRenderedServiceFilesEqual compares the same sorted output paths one
// at a time so an order-dependent package reports the precise changed file.
func requireRenderedServiceFilesEqual(t *testing.T, expected, actual map[string][]byte) {
	t.Helper()
	expectedPaths := make([]string, 0, len(expected))
	actualPaths := make([]string, 0, len(actual))
	for path := range expected {
		expectedPaths = append(expectedPaths, path)
	}
	for path := range actual {
		actualPaths = append(actualPaths, path)
	}
	sort.Strings(expectedPaths)
	sort.Strings(actualPaths)
	require.Equal(t, expectedPaths, actualPaths)
	for _, path := range expectedPaths {
		require.Equal(t, string(expected[path]), string(actual[path]), path)
	}
}

// orderedServicePlans builds equivalent fresh designs with both root and
// service discovery reversed, then runs collection, freeze, and link once.
func orderedServicePlans(t *testing.T, reverse bool) []*Plan {
	t.Helper()
	first := orderedServiceRoot(t, reverse)
	second := singleServiceRoot(t)
	roots := []*expr.RootExpr{first, second}
	if reverse {
		roots[0], roots[1] = roots[1], roots[0]
	}
	evaluated := make([]eval.Root, len(roots))
	for index, root := range roots {
		evaluated[index] = root
	}
	generation, err := codegen.NewGeneration("generated.local/gen", evaluated)
	require.NoError(t, err)
	inputs := make([]PlanInput, len(roots))
	for index, root := range roots {
		inputs[index] = PlanInput{Root: root, Examples: expr.NewExampleGenerator(root.API.RandomizerFactory)}
	}
	plans, err := NewPlans(generation, inputs...)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	for _, plan := range plans {
		require.NoError(t, plan.Link())
	}
	return plans
}

// orderedServiceRoot defines two services that emit distinct same-base unions
// into one relocated package.
func orderedServiceRoot(t *testing.T, reverse bool) *expr.RootExpr {
	t.Helper()
	return codegen.RunDSL(t, func() {
		alpha := relocatedOrderType("AlphaEnvelope", "text", dsl.String)
		omega := relocatedOrderType("OmegaEnvelope", "count", dsl.Int)
		alphaService := func() {
			dsl.Service("Alpha", func() {
				dsl.Method("Read", func() { dsl.Payload(alpha) })
			})
		}
		omegaService := func() {
			dsl.Service("Omega", func() {
				dsl.Method("Read", func() { dsl.Payload(omega) })
			})
		}
		if reverse {
			omegaService()
			alphaService()
		} else {
			alphaService()
			omegaService()
		}
	})
}

// singleServiceRoot defines the second root contributing to the same
// relocated package used by orderedServiceRoot.
func singleServiceRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return codegen.RunDSL(t, func() {
		beta := relocatedOrderType("BetaEnvelope", "enabled", dsl.Boolean)
		dsl.Service("Beta", func() {
			dsl.Method("Read", func() { dsl.Payload(beta) })
		})
	})
}

// relocatedOrderType creates one force-generated type with a Value union in
// the shared generated types package.
func relocatedOrderType(name, branch string, dataType expr.DataType) expr.UserType {
	return dsl.Type(name, func() {
		dsl.Meta("struct:pkg:path", "types")
		dsl.Meta("type:generate:force")
		dsl.OneOf("Value", func() {
			dsl.Attribute(branch, dataType)
		})
	})
}

// renderedServiceFiles executes every retained section and indexes the exact
// bytes by output path, rejecting duplicate contributions.
func renderedServiceFiles(t *testing.T, files []*codegen.File) map[string][]byte {
	t.Helper()
	rendered := make(map[string][]byte, len(files))
	for _, file := range files {
		_, duplicate := rendered[file.Path]
		require.False(t, duplicate, "duplicate generated file %s", file.Path)
		var buffer bytes.Buffer
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&buffer))
		}
		rendered[file.Path] = bytes.Clone(buffer.Bytes())
	}
	return rendered
}

// retainedServiceNamePointers snapshots the exact declaration records held by
// each retained plan so rendering cannot replace them unnoticed.
func retainedServiceNamePointers(plans []*Plan) map[retainedServiceNameID]*codegen.NameDeclaration {
	pointers := make(map[retainedServiceNameID]*codegen.NameDeclaration)
	for rootIndex, plan := range plans {
		for _, facts := range plan.facts.services {
			for id, name := range facts.names {
				pointers[retainedServiceNameID{
					root:    rootIndex,
					service: facts.service.Name,
					symbol:  id,
				}] = name.declaration
			}
		}
	}
	return pointers
}
