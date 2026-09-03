// This file verifies external conversions are owned once by their generated
// receiver package and retain every reflected package reference before freeze.
package service

import (
	"bytes"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	nestedalpha "goa.design/goa/v3/codegen/service/testdata/nested-alpha"
	nestedbeta "goa.design/goa/v3/codegen/service/testdata/nested-beta"
	nestedouter "goa.design/goa/v3/codegen/service/testdata/nested-outer"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestExternalConversionsBelongToGeneratedReceiverPackage catches conversion
// files duplicated by two services that reference one relocated receiver. It
// also compiles two same-named reflected children from distinct Go packages.
func TestExternalConversionsBelongToGeneratedReceiverPackage(t *testing.T) {
	root := externalConversionContractRoot(t)
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())

	files, err := Files(plan)
	require.NoError(t, err)
	conversionPath := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
	require.Len(t, filesAtPath(files, conversionPath), 1)
	conversion := renderSingleFileAtPath(t, files, conversionPath)
	require.NotContains(t, conversion, "goa.design/goa/v3/codegen/service/testdata/a-nested-alpha")
	require.Contains(t, conversion, "nestedalpha.Child")
	require.NotContains(t, conversion, "nestedalpha2.Child")
	compileGeneratedServiceFiles(t, files)
}

// TestExternalConversionPlanIgnoresLaterTypeMapMutation proves linked output
// is byte-for-byte determined by facts retained in NewPlan.
func TestExternalConversionPlanIgnoresLaterTypeMapMutation(t *testing.T) {
	baseline := retainedServicePlanForPackage(t, externalConversionContractRoot(t))
	baselineFiles, err := Files(baseline)
	require.NoError(t, err)
	conversionPath := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
	before := renderSingleFileAtPath(t, baselineFiles, conversionPath)

	root := externalConversionContractRoot(t)
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	originalServiceName := root.Services[0].Name
	root.Services[0].Name = "MutatedService"
	for _, mapping := range append(root.Conversions, root.Creations...) {
		mapping.User.Attribute().AddMeta("struct:pkg:path", "mutated/types")
		if object := expr.AsObject(mapping.User); object != nil && len(*object) > 0 {
			(*object)[0].Attribute.AddMeta(
				"struct:field:type",
				"mutated.Value",
				"mutated.local/value",
				"mutated",
			)
		}
		mapping.User.Rename("Mutated" + mapping.User.Name())
		mapping.User = nil
		mapping.External = struct{}{}
	}
	root.Conversions = nil
	root.Creations = nil
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	require.Equal(t, "generated.local/gen/alpha", plan.Services().ServiceImport("generated.local", originalServiceName).Path)
	afterFiles, err := Files(plan)
	require.NoError(t, err)
	after := renderSingleFileAtPath(t, afterFiles, conversionPath)
	require.Equal(t, before, after)
	compileGeneratedServiceFiles(t, afterFiles)
}

// TestExternalConversionOperationsHaveCanonicalOrder catches convert.go output
// that follows TypeMap traversal rather than stable receiver identities.
func TestExternalConversionOperationsHaveCanonicalOrder(t *testing.T) {
	forward := externalConversionContractRoot(t)
	reverse := externalConversionContractRoot(t)
	slices.Reverse(reverse.Conversions)
	slices.Reverse(reverse.Creations)
	forwardPlan := retainedServicePlanForPackage(t, forward)
	reversePlan := retainedServicePlanForPackage(t, reverse)
	forwardFiles, err := Files(forwardPlan)
	require.NoError(t, err)
	reverseFiles, err := Files(reversePlan)
	require.NoError(t, err)
	conversionPath := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
	require.Equal(
		t,
		renderSingleFileAtPath(t, forwardFiles, conversionPath),
		renderSingleFileAtPath(t, reverseFiles, conversionPath),
	)
}

// TestExternalConversionReachabilityCoversEveryServiceValue catches mappings
// omitted when a type is reachable only through a stream or error contract.
func TestExternalConversionReachabilityCoversEveryServiceValue(t *testing.T) {
	tests := map[string]func(expr.UserType){
		"streaming payload": func(mapped expr.UserType) {
			dsl.Service("Reach", func() {
				dsl.Method("Use", func() {
					dsl.StreamingPayload(mapped)
					dsl.Result(dsl.String)
				})
			})
		},
		"mixed streaming result": func(mapped expr.UserType) {
			dsl.Service("Reach", func() {
				dsl.Method("Use", func() {
					dsl.Result(dsl.String)
					dsl.StreamingResult(mapped)
				})
			})
		},
		"service error": func(mapped expr.UserType) {
			dsl.Service("Reach", func() {
				dsl.Error("failed", mapped)
				dsl.Method("Use", func() {})
			})
		},
		"method error": func(mapped expr.UserType) {
			dsl.Service("Reach", func() {
				dsl.Method("Use", func() {
					dsl.Error("failed", mapped)
				})
			})
		},
	}
	for name, use := range tests {
		t.Run(name, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				mapped := dsl.Type("Mapped", func() {
					dsl.ConvertTo(nestedalpha.Child{})
					dsl.Attribute("value", dsl.String)
					dsl.Required("value")
				})
				use(mapped)
			})
			plan := retainedServicePlanForPackage(t, root)
			files, err := Files(plan)
			require.NoError(t, err)
			conversionPath := filepath.Join(codegen.Gendir, "reach", "convert.go")
			require.Len(t, filesAtPath(files, conversionPath), 1)
			compileGeneratedServiceFiles(t, files)
		})
	}
}

// TestExternalConversionsAggregateAcrossRoots catches root-local conversion
// files that target one generated package and change with root order.
func TestExternalConversionsAggregateAcrossRoots(t *testing.T) {
	forwardPlans := convertedRootPlans(t, false)
	forwardFiles, err := Files(forwardPlans...)
	require.NoError(t, err)
	conversionPath := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
	require.Len(t, filesAtPath(forwardFiles, conversionPath), 1)
	forward := renderSingleFileAtPath(t, forwardFiles, conversionPath)

	reversePlans := convertedRootPlans(t, true)
	reverseFiles, err := Files(reversePlans...)
	require.NoError(t, err)
	require.Len(t, filesAtPath(reverseFiles, conversionPath), 1)
	require.Equal(t, forward, renderSingleFileAtPath(t, reverseFiles, conversionPath))
	require.Contains(t, forward, "func (t *AlphaMapped) ConvertToChild()")
	require.Contains(t, forward, "func (t *BetaMapped) ConvertToChild()")
	compileGeneratedServiceFiles(t, forwardFiles)
}

// TestExternalConversionsShareReceiverMethodNamesAcrossRoots catches method
// names assigned independently by roots that contribute operations for the
// same canonical receiver declaration.
func TestExternalConversionsShareReceiverMethodNamesAcrossRoots(t *testing.T) {
	forwardPlans := sharedConvertedReceiverPlans(t, false)
	forwardFiles, err := Files(forwardPlans...)
	require.NoError(t, err)
	conversionPath := filepath.Join(codegen.Gendir, "shared", "types", "convert.go")
	require.Len(t, filesAtPath(forwardFiles, conversionPath), 1)
	forward := renderSingleFileAtPath(t, forwardFiles, conversionPath)

	reversePlans := sharedConvertedReceiverPlans(t, true)
	reverseFiles, err := Files(reversePlans...)
	require.NoError(t, err)
	require.Len(t, filesAtPath(reverseFiles, conversionPath), 1)
	require.Equal(t, forward, renderSingleFileAtPath(t, reverseFiles, conversionPath))
	require.Contains(t, forward, "func (t *SharedMapped) ConvertToChild()")
	require.Contains(t, forward, "func (t *SharedMapped) ConvertToChild2()")
	compileGeneratedServiceFiles(t, forwardFiles)
}

// TestNewPlansRejectDuplicateExternalConversionsAcrossRoots proves the batch
// boundary rejects one exact receiver operation instead of inventing X2.
func TestNewPlansRejectDuplicateExternalConversionsAcrossRoots(t *testing.T) {
	var shared expr.UserType
	first := codegen.RunDSL(t, func() {
		shared = dsl.Type("SharedMapped", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.ConvertTo(nestedalpha.Child{})
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("Alpha", func() {
			dsl.Method("Use", func() { dsl.Payload(shared) })
		})
	})
	second := codegen.RunDSL(t, func() {
		dsl.Service("Beta", func() {
			dsl.Method("Use", func() { dsl.Payload(shared) })
		})
	})
	second.Conversions = append(second.Conversions, &expr.TypeMap{
		User:     shared,
		External: nestedalpha.Child{},
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{first, second})
	_, err := NewPlans(
		generation,
		PlanInput{Root: first, Examples: expr.NewExampleGenerator(first.API.RandomizerFactory)},
		PlanInput{Root: second, Examples: expr.NewExampleGenerator(second.API.RandomizerFactory)},
	)
	require.ErrorContains(t, err, "duplicate external conversion")
}

// externalConversionContractRoot builds one relocated receiver referenced by
// two services and mapped in both conversion directions.
func externalConversionContractRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return codegen.RunDSL(t, func() {
		alpha := dsl.Type("AlphaChild", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		beta := dsl.Type("BetaChild", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		envelope := dsl.Type("Envelope", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.ConvertTo(nestedouter.Envelope{})
			dsl.CreateFrom(nestedouter.Envelope{})
			dsl.Attribute("alpha", alpha)
			dsl.Attribute("beta", beta)
			dsl.Required("alpha", "beta")
		})
		for _, service := range []string{"Alpha", "Beta"} {
			dsl.Service(service, func() {
				dsl.Method("Read", func() {
					dsl.Payload(envelope)
				})
			})
		}
	})
}

// convertedRootPlans creates two roots whose distinct converted receivers are
// relocated into one generated package, in either discovery order.
func convertedRootPlans(t *testing.T, reverse bool) []*Plan {
	t.Helper()
	roots := []*expr.RootExpr{
		convertedReceiverRoot(t, "Alpha", nestedalpha.Child{}),
		convertedReceiverRoot(t, "Beta", nestedbeta.Child{}),
	}
	if reverse {
		slices.Reverse(roots)
	}
	evaluated := make([]eval.Root, len(roots))
	for index, root := range roots {
		evaluated[index] = root
	}
	generation := mustTestGeneration(t, "generated.local/gen", evaluated)
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

// sharedConvertedReceiverPlans creates two roots that contribute distinct
// same-named external mappings for one exact relocated receiver declaration.
func sharedConvertedReceiverPlans(t *testing.T, reverse bool) []*Plan {
	t.Helper()
	var shared expr.UserType
	first := codegen.RunDSL(t, func() {
		shared = dsl.Type("SharedMapped", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.ConvertTo(nestedalpha.Child{})
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("Alpha", func() {
			dsl.Method("Use", func() {
				dsl.Payload(shared)
			})
		})
	})
	second := codegen.RunDSL(t, func() {
		dsl.Service("Beta", func() {
			dsl.Method("Use", func() {
				dsl.Payload(shared)
			})
		})
	})
	second.Conversions = append(second.Conversions, &expr.TypeMap{
		User:     shared,
		External: nestedbeta.Child{},
	})
	roots := []*expr.RootExpr{first, second}
	if reverse {
		slices.Reverse(roots)
	}
	evaluated := make([]eval.Root, len(roots))
	for index, root := range roots {
		evaluated[index] = root
	}
	generation := mustTestGeneration(t, "generated.local/gen", evaluated)
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

// convertedReceiverRoot builds one relocated receiver mapping for the
// multi-root conversion aggregation contract.
func convertedReceiverRoot(t *testing.T, service string, external any) *expr.RootExpr {
	t.Helper()
	return codegen.RunDSL(t, func() {
		mapped := dsl.Type(service+"Mapped", func() {
			dsl.Meta("struct:pkg:path", "shared/types")
			dsl.ConvertTo(external)
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service(service, func() {
			dsl.Method("Use", func() {
				dsl.Payload(mapped)
			})
		})
	})
}

// filesAtPath returns every generated file that targets path.
func filesAtPath(files []*codegen.File, path string) []*codegen.File {
	var matches []*codegen.File
	for _, file := range files {
		if file.Path == path {
			matches = append(matches, file)
		}
	}
	return matches
}

// renderSingleFileAtPath renders the unique file targeting path.
func renderSingleFileAtPath(t *testing.T, files []*codegen.File, path string) string {
	t.Helper()
	matches := filesAtPath(files, path)
	require.Len(t, matches, 1)
	var rendered bytes.Buffer
	for _, section := range matches[0].SectionTemplates {
		require.NoError(t, section.Write(&rendered))
	}
	return rendered.String()
}
