// This file verifies service render analysis and the generated service files
// built from its immutable data.
package service

import (
	"bytes"
	"go/format"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestServicesDataUsesFrozenPackageDeclarations(t *testing.T) {
	var shared expr.UserType
	root := codegen.RunDSL(t, func() {
		shared = dsl.Type("Shared", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		first := dsl.Type("FirstPayload", func() {
			dsl.Attribute("shared", shared)
		})
		second := dsl.Type("SecondPayload", func() {
			dsl.Attribute("shared", shared)
		})
		dsl.Service("First", func() {
			dsl.Method("Read", func() {
				dsl.Payload(first)
			})
		})
		dsl.Service("Second", func() {
			dsl.Method("Read", func() {
				dsl.Payload(second)
			})
		})
	})

	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	require.Panics(t, func() { plan.Services() })
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	services := plan.Services()

	first := services.Get("First")
	second := services.Get("Second")
	firstShared := findUserTypeData(first.userTypes, shared)
	secondShared := findUserTypeData(second.userTypes, shared)
	require.NotNil(t, firstShared)
	require.NotNil(t, secondShared)
	require.Same(t, firstShared.Declaration, secondShared.Declaration)
	require.Len(t, first.unions, 1)
	require.Len(t, second.unions, 1)
	require.Same(t, first.unions[0].Declaration, second.unions[0].Declaration)
	require.Equal(t, "Value", first.unions[0].Name)
	require.Equal(t, "ValueKind", first.unions[0].KindName)

	_, err = generation.Package("goa.design/goa/example/types").DeclareUserType(shared)
	require.ErrorContains(t, err, "frozen")
}

// TestPlanOwnsNormalizedMethodNames verifies that semantic wrappers receive
// names from the service package catalog and collide only with local exact
// declarations.
func TestPlanOwnsNormalizedMethodNames(t *testing.T) {
	var local expr.UserType
	root := codegen.RunDSL(t, func() {
		local = dsl.Type("UsePayload", func() {
			dsl.Attribute("existing", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Existing", func() {
				dsl.Payload(local)
			})
			dsl.Method("Use", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	require.NoError(t, planTestServices(root, generation))
	require.NoError(t, generation.Freeze())

	service := root.Service("Values")
	wrapper := service.Method("Use").Payload.Type.(expr.UserType)
	declaration, err := generation.Package("generated.local/gen/values").Type(wrapper)
	require.NoError(t, err)
	require.Equal(t, "UsePayload2", declaration.Name())
}

// TestPlanPreservesGeneratedPackageClaims verifies that service planning
// rejects distinct metadata spellings before path normalization can merge
// their declarations into one output package.
func TestPlanPreservesGeneratedPackageClaims(t *testing.T) {
	tests := []struct {
		name       string
		firstPath  string
		secondPath string
		contains   string
	}{
		{"normalized collision", "types", "domain/../types", "normalize to import path"},
		{"portable collision", "Types", "types", "case-insensitive filesystem"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				first := dsl.Type("First", func() {
					dsl.Meta("struct:pkg:path", test.firstPath)
					dsl.Attribute("value", dsl.String)
				})
				second := dsl.Type("Second", func() {
					dsl.Meta("struct:pkg:path", test.secondPath)
					dsl.Attribute("value", dsl.String)
				})
				dsl.Service("Values", func() {
					dsl.Method("First", func() { dsl.Payload(first) })
					dsl.Method("Second", func() { dsl.Payload(second) })
				})
			})
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})

			err := planTestServices(root, generation)
			require.ErrorContains(t, err, test.contains)
		})
	}
}

// TestPlanRejectsInvalidGeneratedPackageLocations verifies that relative Goa
// metadata cannot escape its generated module or use filesystem separators in
// a Go import path.
func TestPlanRejectsInvalidGeneratedPackageLocations(t *testing.T) {
	tests := []struct {
		name     string
		location string
	}{
		{"absolute", "/outside"},
		{"escape", "../outside"},
		{"backslash", `domain\types`},
		{"colon", "domain:types"},
		{"space", "domain types"},
		{"control", "domain\x00types"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				value := dsl.Type("Value", func() {
					dsl.Meta("struct:pkg:path", test.location)
					dsl.Attribute("value", dsl.String)
				})
				dsl.Service("Values", func() {
					dsl.Method("Read", func() { dsl.Payload(value) })
				})
			})
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})

			require.Error(t, planTestServices(root, generation))
		})
	}
}

// TestPlanIgnoresUnusedRelocatedTypes verifies that a type excluded from
// service output does not needlessly claim a package or contribute imports.
func TestPlanIgnoresUnusedRelocatedTypes(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Type("Unused", func() {
			dsl.Meta("struct:pkg:path", "unused")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})

	require.NoError(t, planTestServices(root, generation))
	require.NoError(t, generation.Freeze())
}

// TestFilesUseCanonicalOwnedOutputDirectory verifies that a lone noncanonical
// metadata spelling emits the declaration beneath its owned canonical package.
func TestFilesUseCanonicalOwnedOutputDirectory(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		value := dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "domain/../types")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() { dsl.Payload(value) })
		})
	})
	plan := mustServicePlan(t, root)

	require.NotNil(t, findFile(mustServiceFiles(t, plan),
		filepath.Join("gen", "types", "value.go")))
}

// TestServicesDataUsesRebuiltViewDeclarations verifies that planning and
// rendering can rebuild view expressions while sharing frozen declarations.
func TestServicesDataUsesRebuiltViewDeclarations(t *testing.T) {
	var result *expr.ResultTypeExpr
	root := codegen.RunDSL(t, func() {
		result = dsl.ResultType("application/vnd.value", func() {
			dsl.TypeName("Value")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(result)
			})
		})
	})

	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	views := mustClaimTestPackage(t, generation, "goa.design/goa/example/values/views")
	plannedProjected, err := views.DerivedType(codegen.NewProjectedTypeID(result))
	require.NoError(t, err)
	plannedViewed, err := views.DerivedType(codegen.NewViewedResultTypeID(result))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	services := plan.Services()
	service := services.Get("Values")
	require.Len(t, service.projectedTypes, 1)
	require.Len(t, service.viewedResultTypes, 1)
	require.Same(t, plannedProjected, service.projectedTypes[0].Declaration)
	require.Same(t, plannedViewed, service.viewedResultTypes[0].Declaration)
	require.Equal(t, "ValueView", plannedProjected.Name())
	require.Equal(t, "Value", plannedViewed.Name())
}

func TestFilesEmitsPackageDeclarationsOnce(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionNameScopeDSL)
	files := mustServiceFiles(t, mustServicePlan(t, root))

	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "first_value.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "second_value.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "third_value.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "unions.go")))

	unionFile := findFile(files, filepath.Join("gen", "types", "unions.go"))
	require.NotNil(t, unionFile)
	code := renderSections(t, unionFile.SectionTemplates)
	require.Equal(t, 1, strings.Count(code, "type Value struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type ValueKind string"), code)
}

func TestFilesEmitsDifferentSameBaseUnionsWithFrozenNames(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		first := dsl.Type("First", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		second := dsl.Type("Second", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("number", dsl.Int)
			})
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(first)
			})
		})
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(second)
			})
		})
	})

	files := mustServiceFiles(t, mustServicePlan(t, root))
	unionFile := findFile(files, filepath.Join("gen", "types", "unions.go"))
	require.NotNil(t, unionFile)
	code := renderSections(t, unionFile.SectionTemplates)
	require.Equal(t, 1, strings.Count(code, "type Value struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type Value2 struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type ValueKind string"), code)
	require.Equal(t, 1, strings.Count(code, "type Value2Kind string"), code)
}

func TestFilesEmitsSharedPackagesOnceAcrossRoots(t *testing.T) {
	var firstType expr.UserType
	firstRoot := codegen.RunDSL(t, func() {
		firstType = dsl.Type("First", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(firstType)
			})
		})
	})
	var secondType expr.UserType
	secondRoot := codegen.RunDSL(t, func() {
		secondType = dsl.Type("Second", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(secondType)
			})
		})
	})

	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	plans, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	firstPlan, secondPlan := plans[0], plans[1]
	firstUnion := expr.AsObject(firstType).Attribute("Value").Type.(*expr.Union)
	secondUnion := expr.AsObject(secondType).Attribute("Value").Type.(*expr.Union)
	generatedPackage := mustClaimTestPackage(t, generation, "goa.design/goa/example/types")
	firstBranch, err := generatedPackage.UnionBranchType(firstUnion, "text")
	require.NoError(t, err)
	secondBranch, err := generatedPackage.UnionBranchType(secondUnion, "text")
	require.NoError(t, err)
	require.Same(t, firstBranch, secondBranch)

	require.NoError(t, generation.Freeze())
	require.NoError(t, firstPlan.Link())
	require.NoError(t, secondPlan.Link())
	files := mustServiceFiles(t, firstPlan, secondPlan)

	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "first.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "second.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "value_text.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "types", "unions.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "first_service", "service.go")))
	require.Equal(t, 1, countFiles(files, filepath.Join("gen", "second_service", "service.go")))
}

func TestFilesEmitCanonicalSharedDeclarationAcrossRoots(t *testing.T) {
	forwardPlans := sharedDeclarationPlans(t, false)
	forwardFiles := mustServiceFiles(t, forwardPlans...)
	sharedPath := filepath.Join("gen", "types", "shared.go")
	require.Equal(t, 1, countFiles(forwardFiles, sharedPath))
	forward := renderSingleFileAtPath(t, forwardFiles, sharedPath)

	reversePlans := sharedDeclarationPlans(t, true)
	reverseFiles := mustServiceFiles(t, reversePlans...)
	require.Equal(t, 1, countFiles(reverseFiles, sharedPath))
	require.Equal(t, forward, renderSingleFileAtPath(t, reverseFiles, sharedPath))
}

func TestNewPlansRejectConflictingSharedDeclarationEmissionCandidates(t *testing.T) {
	firstRoot, secondRoot := conflictingSharedDeclarationRoots(t)
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	_, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.ErrorContains(t, err, "conflicting generated type emission")
}

// TestNewPlansAcceptEquivalentSharedDeclarationCopies proves compiler-created
// copies coalesce when every retained type fact is structurally identical.
func TestNewPlansAcceptEquivalentSharedDeclarationCopies(t *testing.T) {
	firstRoot, secondRoot := copiedSharedDeclarationRoots(t, nil)
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	plans, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	for _, plan := range plans {
		require.NoError(t, plan.Link())
	}
	require.Equal(t, 1, countFiles(mustServiceFiles(t, plans...), filepath.Join("gen", "types", "shared.go")))
}

// TestNewPlansRejectSharedDeclarationLayoutConflicts proves a shared package
// declaration cannot silently select one compiler copy's field spelling.
func TestNewPlansRejectSharedDeclarationLayoutConflicts(t *testing.T) {
	firstRoot, secondRoot := copiedSharedDeclarationRoots(t, func(copy expr.UserType) {
		field := expr.AsObject(copy).Attribute("value")
		field.AddMeta("struct:field:name", "OtherValue")
	})
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	_, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.ErrorContains(t, err, "conflicting generated type emission")
}

// TestNewPlansAcceptDistinctTransportValidationForSharedDeclaration proves
// validation does not become false service-file ownership. HTTP and gRPC own
// their validation programs; the shared service file owns only the Go layout.
func TestNewPlansAcceptDistinctTransportValidationForSharedDeclaration(t *testing.T) {
	firstRoot, secondRoot := copiedSharedDeclarationRoots(t, func(copy expr.UserType) {
		expr.AsObject(copy).Attribute("value").Validation = &expr.ValidationExpr{Pattern: "^[a-z]+$"}
	})
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	plans, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	for _, plan := range plans {
		require.NoError(t, plan.Link())
	}
	require.Equal(t, 1, countFiles(mustServiceFiles(t, plans...), filepath.Join("gen", "types", "shared.go")))
}

// TestNewPlansRejectSharedUnionBranchLayoutConflicts proves one canonical
// union declaration cannot select between differing retained branch layouts.
func TestNewPlansRejectSharedUnionBranchLayoutConflicts(t *testing.T) {
	firstRoot, secondRoot := copiedSharedUnionRoots(t, func(union *expr.Union) {
		union.Values[0].Attribute.Description = "a conflicting branch description"
	})
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	_, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.ErrorContains(t, err, "conflicting generated union emission")
}

func TestNewPlanRejectsPartialMultiRootPlanning(t *testing.T) {
	_, firstRoot, secondRoot := sharedDeclarationRoots(t)
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{firstRoot, secondRoot})
	_, err := NewPlan(firstRoot, generation, expr.NewExampleGenerator(firstRoot.API.RandomizerFactory))
	require.ErrorContains(t, err, "requires all 2 generation roots")
}

func sharedDeclarationPlans(t *testing.T, reverse bool) []*Plan {
	t.Helper()
	_, firstRoot, secondRoot := sharedDeclarationRoots(t)
	roots := []*expr.RootExpr{firstRoot, secondRoot}
	if reverse {
		slices.Reverse(roots)
	}
	evaluated := make([]eval.Root, len(roots))
	for index, root := range roots {
		evaluated[index] = root
	}
	generation := mustTestGeneration(t, "goa.design/goa/example", evaluated)
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

func sharedDeclarationRoots(t *testing.T) (expr.UserType, *expr.RootExpr, *expr.RootExpr) {
	t.Helper()
	var shared expr.UserType
	firstRoot := codegen.RunDSL(t, func() {
		shared = dsl.Type("Shared", func() {
			dsl.Description("The canonical shared declaration.")
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(shared)
			})
		})
	})
	secondRoot := codegen.RunDSL(t, func() {
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(shared)
			})
		})
	})
	return shared, firstRoot, secondRoot
}

func conflictingSharedDeclarationRoots(t *testing.T) (*expr.RootExpr, *expr.RootExpr) {
	t.Helper()
	var shared expr.UserType
	firstRoot := codegen.RunDSL(t, func() {
		shared = dsl.Type("Shared", func() {
			dsl.Description("The first retained declaration.")
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(shared)
			})
		})
	})
	conflicting := shared.Dup(expr.DupAtt(shared.Attribute())).(expr.UserType)
	conflicting.Attribute().Description = "The conflicting retained declaration."
	secondRoot := codegen.RunDSL(t, func() {
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(conflicting)
			})
		})
	})
	return firstRoot, secondRoot
}

// copiedSharedDeclarationRoots returns two roots whose compiler copies share
// one authored origin and therefore one generated declaration.
func copiedSharedDeclarationRoots(t *testing.T, mutate func(expr.UserType)) (*expr.RootExpr, *expr.RootExpr) {
	t.Helper()
	var shared expr.UserType
	firstRoot := codegen.RunDSL(t, func() {
		shared = dsl.Type("Shared", func() {
			dsl.Description("The canonical shared declaration.")
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() { dsl.Payload(shared) })
		})
	})
	copy := shared.Dup(expr.DupAtt(shared.Attribute())).(expr.UserType)
	if mutate != nil {
		mutate(copy)
	}
	secondRoot := codegen.RunDSL(t, func() {
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() { dsl.Payload(copy) })
		})
	})
	return firstRoot, secondRoot
}

// copiedSharedUnionRoots returns two roots whose equal union identities bind
// the same generated declaration while retaining independent branch facts.
func copiedSharedUnionRoots(t *testing.T, mutate func(*expr.Union)) (*expr.RootExpr, *expr.RootExpr) {
	t.Helper()
	var container expr.UserType
	firstRoot := codegen.RunDSL(t, func() {
		container = dsl.Type("Container", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() { dsl.Payload(container) })
		})
	})
	copy := container.Dup(expr.DupAtt(container.Attribute())).(expr.UserType)
	union := expr.AsObject(copy).Attribute("value").Type.(*expr.Union)
	if mutate != nil {
		mutate(union)
	}
	secondRoot := codegen.RunDSL(t, func() {
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() { dsl.Payload(copy) })
		})
	})
	return firstRoot, secondRoot
}

func TestGeneratedUnionBranchCollisionDoesNotCanonicalizeToRootType(t *testing.T) {
	var (
		exact     expr.UserType
		container expr.UserType
	)
	root := codegen.RunDSL(t, func() {
		exact = dsl.Type("Value-Text", dsl.String, func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
		})
		container = dsl.Type("Container", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		dsl.Service("Collision", func() {
			dsl.Method("Read", func() {
				dsl.Payload(container)
			})
		})
	})

	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	union := expr.AsObject(container).Attribute("Value").Type.(*expr.Union)
	generatedPackage := mustClaimTestPackage(t, generation, "goa.design/goa/example/types")
	exactDeclaration, err := generatedPackage.UserType(exact)
	require.NoError(t, err)
	branchDeclaration, err := generatedPackage.UnionBranchType(union, "text")
	require.NoError(t, err)
	require.NotSame(t, exactDeclaration, branchDeclaration)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "ValueText", exactDeclaration.Name())
	require.Equal(t, "ValueText2", branchDeclaration.Name())
	require.NoError(t, plan.Link())
	typeFile := findFile(
		mustServiceFiles(t, plan),
		filepath.Join("gen", "types", "value_text.go"),
	)
	require.NotNil(t, typeFile)
	code := renderSections(t, typeFile.SectionTemplates)
	require.Contains(t, code, "type ValueText string")
	require.Contains(t, code, "type ValueText2 string")
}

func TestService(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL},
		{"service-single", testdata.SingleMethodDSL},
		{"service-multiple", testdata.MultipleMethodsDSL},
		{"service-union", testdata.UnionMethodDSL},
		{"service-multi-union", testdata.MultiUnionMethodDSL},
		{"service-union-alias-cross-pkg", testdata.UnionWithAliasCrossPkgDSL},
		{"service-no-payload-no-result", testdata.EmptyMethodDSL},
		{"service-payload-no-result", testdata.EmptyResultMethodDSL},
		{"service-no-payload-result", testdata.EmptyPayloadMethodDSL},
		{"service-payload-result-with-default", testdata.WithDefaultDSL},
		{"service-result-with-multiple-views", testdata.MultipleMethodsResultMultipleViewsDSL},
		{"service-result-with-explicit-and-default-views", testdata.WithExplicitAndDefaultViewsDSL},
		{"service-result-collection-multiple-views", testdata.ResultCollectionMultipleViewsMethodDSL},
		{"service-result-with-other-result", testdata.ResultWithOtherResultMethodDSL},
		{"service-result-with-result-collection", testdata.ResultWithResultCollectionMethodDSL},
		{"service-result-with-dashed-mime-type", testdata.ResultWithDashedMimeTypeMethodDSL},
		{"service-result-with-one-of-type", testdata.ResultWithOneOfTypeMethodDSL},
		{"service-result-with-inline-validation", testdata.ResultWithInlineValidationDSL},
		{"service-service-level-error", testdata.ServiceErrorDSL},
		{"service-custom-errors", testdata.CustomErrorsDSL},
		{"service-custom-errors-custom-field", testdata.CustomErrorsCustomFieldDSL},
		{"service-force-generate-type", testdata.ForceGenerateTypeDSL},
		{"service-force-generate-type-explicit", testdata.ForceGenerateTypeExplicitDSL},
		{"service-streaming-result", testdata.StreamingResultMethodDSL},
		{"service-mixed-results", testdata.MixedResultsEndpointDSL},
		{"service-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL},
		{"service-streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewMethodDSL},
		{"service-streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL},
		{"service-streaming-payload", testdata.StreamingPayloadMethodDSL},
		{"service-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL},
		{"service-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL},
		{"service-streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsMethodDSL},
		{"service-streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewMethodDSL},
		{"service-bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL},
		{"service-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL},
		{"service-bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsMethodDSL},
		{"service-bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewMethodDSL},
		{"service-multiple-api-key-security", testdata.MultipleAPIKeySecurityDSL},
		{"service-mixed-and-multiple-api-key-security", testdata.MixedAndMultipleAPIKeySecurityDSL},
		{"service-bearer-security", testdata.BearerSecurityDSL},
		{"service-raw-object-payload-type-name-collision", testdata.RawObjectPayloadTypeNameCollisionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			files := mustServiceFiles(t, plan)
			require.Greater(t, len(files), 0)

			code := renderServiceGolden(t, files, files[0])

			// Compare with golden file
			testutil.AssertGo(t, "testdata/golden/service_"+c.Name+".go.golden", code)
		})
	}
}

func TestStructPkgPath(t *testing.T) {
	fooPath := filepath.Join("gen", "foo", "foo.go")
	recursiveFooPath := filepath.Join("gen", "foo", "recursive_foo.go")
	barPath := filepath.Join("gen", "bar", "bar.go")
	bazPath := filepath.Join("gen", "baz", "baz.go")
	cases := []struct {
		Name      string
		DSL       func()
		TypeFiles []string
	}{
		{"none", testdata.SingleMethodDSL, nil},
		{"single", testdata.PkgPathDSL, []string{fooPath}},
		{"array", testdata.PkgPathArrayDSL, []string{fooPath}},
		{"recursive", testdata.PkgPathRecursiveDSL, []string{fooPath, recursiveFooPath}},
		{"multiple", testdata.PkgPathMultipleDSL, []string{barPath, bazPath}},
		{"nopkg", testdata.PkgPathNoDirDSL, nil},
		{"dupes", testdata.PkgPathDupeDSL, []string{fooPath}},
		{"payload_attribute", testdata.PkgPathPayloadAttributeDSL, []string{fooPath}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			services := plan.Services()
			files := mustServiceFiles(t, plan)

			serviceFile := findFile(files, filepath.Join(codegen.Gendir, services.Get(root.Services[0].Name).PathName, "service.go"))
			require.NotNil(t, serviceFile)
			testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_service.go.golden", renderServiceGolden(t, files, serviceFile))

			// Type files
			for _, typeFile := range c.TypeFiles {
				file := findFile(files, typeFile)
				require.NotNil(t, file)
				buf := new(bytes.Buffer)
				for _, s := range file.SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err)
				goldenName := filepath.Base(typeFile)
				testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_"+goldenName+".golden", string(bs))
			}

			// For dupes case, test the second service
			if c.Name == "dupes" && len(root.Services) > 1 {
				files = serviceFiles(plan, plan.facts.services[1])
				require.Len(t, files, 1)
				buf := new(bytes.Buffer)
				for _, s := range files[0].SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err)
				testutil.AssertGo(t, "testdata/golden/pkg_path_"+c.Name+"_service2.go.golden", string(bs))
			}
		})
	}
}

func TestStructPkgPath_UnionImportsJSON(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionDSL)
	plan := mustServicePlan(t, root)
	require.Len(t, root.Services, 1)

	files := mustServiceFiles(t, plan)
	require.GreaterOrEqual(t, len(files), 2, "expected at least service.go + one struct:pkg:path file")

	unionFile := findFile(files, filepath.Join("gen", "types", "unions.go"))
	require.NotNil(t, unionFile, "expected generated union file in struct:pkg:path package")

	buf := new(bytes.Buffer)
	for _, s := range unionFile.SectionTemplates {
		require.NoError(t, s.Write(buf))
	}
	code := buf.String()
	require.Contains(t, code, "\"bytes\"", "expected bytes import in generated file:\n%s", code)
	require.Contains(t, code, "\"encoding/json\"", "expected encoding/json import in generated file:\n%s", code)
}

func TestStructPkgPath_UnionNamesSharePackageScopeAcrossServices(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionNameScopeDSL)
	var generated strings.Builder
	files := mustServiceFiles(t, mustServicePlan(t, root))
	for _, file := range files {
		if !strings.Contains(file.Path, filepath.Join("gen", "types")) {
			continue
		}
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&generated))
		}
	}

	code := generated.String()
	require.Equal(t, 1, strings.Count(code, "type Value struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type ValueKind string"), code)
	firstUsesValue := unionFieldType(code, "FirstValue")
	secondUsesValue := unionFieldType(code, "SecondValue")
	thirdUsesValue := unionFieldType(code, "ThirdValue")
	require.Equal(t, []string{"Value", "Value", "Value"}, []string{firstUsesValue, secondUsesValue, thirdUsesValue})
}

func unionFieldType(code, owner string) string {
	prefix := "type " + owner + " struct {\n\tValue "
	start := strings.Index(code, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.IndexByte(code[start:], '\n')
	if end == -1 {
		return ""
	}
	return code[start : start+end]
}

// mustServicePlan runs the complete retained-plan lifecycle used by service
// renderer tests.
func mustServicePlan(t *testing.T, root *expr.RootExpr) *Plan {
	t.Helper()
	generation := mustTestGeneration(t, "goa.design/goa/example", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	return plan
}

// mustServicesData returns the linked render model for focused analysis tests.
func mustServicesData(t *testing.T, root *expr.RootExpr) *ServicesData {
	t.Helper()
	return mustServicePlan(t, root).Services()
}

// mustServiceFiles renders linked plans or fails the calling test.
func mustServiceFiles(t *testing.T, plans ...*Plan) []*codegen.File {
	t.Helper()
	files, err := Files(plans...)
	require.NoError(t, err)
	return files
}

// countFiles returns how many generated files have the given path.
func countFiles(files []*codegen.File, path string) int {
	count := 0
	for _, file := range files {
		if file.Path == path {
			count++
		}
	}
	return count
}

// findFile returns the generated file with path, or nil when no file matches.
func findFile(files []*codegen.File, path string) *codegen.File {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return nil
}

// findUserTypeData returns the render data for userType.
func findUserTypeData(types []*UserTypeData, userType expr.UserType) *UserTypeData {
	for _, data := range types {
		if data.Type == userType {
			return data
		}
	}
	return nil
}

// renderSections renders sections without writing a generated file.
func renderSections(t *testing.T, sections []*codegen.SectionTemplate) string {
	t.Helper()
	var rendered strings.Builder
	for _, section := range sections {
		require.NoError(t, section.Write(&rendered))
	}
	return rendered.String()
}

// renderServiceGolden reconstructs the former single-file declaration order
// so existing service golden assertions remain unchanged after unions move to
// their package-owned unions.go file.
func renderServiceGolden(t *testing.T, files []*codegen.File, serviceFile *codegen.File) string {
	t.Helper()
	sections := append([]*codegen.SectionTemplate(nil), serviceFile.SectionTemplates[1:]...)
	unionFile := findFile(files, filepath.Join(filepath.Dir(serviceFile.Path), "unions.go"))
	if unionFile != nil {
		insertAt := len(sections)
		for i, section := range sections {
			switch section.Name {
			case "error-init-func", "viewed-result-type-to-service-result-type",
				"service-result-type-to-viewed-result-type", "projected-type-to-service-type",
				"service-type-to-projected-type", "transform-helpers":
				insertAt = i
			}
			if insertAt != len(sections) {
				break
			}
		}
		sections = append(sections, make([]*codegen.SectionTemplate, len(unionFile.SectionTemplates)-1)...)
		copy(sections[insertAt+len(unionFile.SectionTemplates)-1:], sections[insertAt:])
		copy(sections[insertAt:], unionFile.SectionTemplates[1:])
	}
	buf := new(bytes.Buffer)
	for _, section := range sections {
		require.NoError(t, section.Write(buf))
	}
	formatted, err := format.Source(buf.Bytes())
	require.NoError(t, err, buf.String())
	return string(formatted)
}

func TestStructPkgPath_UnionJSONFieldBranchesGenerateAliases(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathUnionJSONFieldDSL)
	plan := mustServicePlan(t, root)
	require.Len(t, root.Services, 1)

	files := mustServiceFiles(t, plan)
	require.GreaterOrEqual(t, len(files), 2, "expected at least service.go + one struct:pkg:path file")

	unionFile := findFile(files, filepath.Join("gen", "types", "unions.go"))
	require.NotNil(t, unionFile, "expected package-owned union file")

	buf := new(bytes.Buffer)
	for _, s := range unionFile.SectionTemplates {
		require.NoError(t, s.Write(buf))
	}
	code := buf.String()
	require.Contains(t, code, "A ValuesA", "expected union field A to use generated alias type:\n%s", code)
	require.Contains(t, code, "B ValuesB", "expected union field B to use generated alias type:\n%s", code)

	var hasValuesAFile, hasValuesBFile bool
	for _, f := range files {
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "values_a.go")) {
			hasValuesAFile = true
		}
		if strings.HasSuffix(f.Path, filepath.Join("gen", "types", "values_b.go")) {
			hasValuesBFile = true
		}
	}
	require.True(t, hasValuesAFile, "expected generated alias file in struct:pkg:path package: gen/types/values_a.go")
	require.True(t, hasValuesBFile, "expected generated alias file in struct:pkg:path package: gen/types/values_b.go")
}

func TestStructPkgPath_ExtendedUnionGeneratedInEachOwningPackage(t *testing.T) {
	root := codegen.RunDSL(t, testdata.PkgPathExtendedUnionDSL)
	plan := mustServicePlan(t, root)
	require.Len(t, root.Services, 1)

	files := mustServiceFiles(t, plan)
	require.GreaterOrEqual(t, len(files), 2)

	var serviceFile, localUnionFile, sharedUnionFile *codegen.File
	for _, f := range files {
		switch {
		case strings.HasSuffix(f.Path, filepath.Join("gen", "pkg_path_extended_union", "service.go")):
			serviceFile = f
		case strings.HasSuffix(f.Path, filepath.Join("gen", "pkg_path_extended_union", "unions.go")):
			localUnionFile = f
		case strings.HasSuffix(f.Path, filepath.Join("gen", "types", "unions.go")):
			sharedUnionFile = f
		}
	}
	require.NotNil(t, serviceFile)
	require.NotNil(t, localUnionFile)
	require.NotNil(t, sharedUnionFile)

	render := func(file *codegen.File) string {
		buf := new(bytes.Buffer)
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(buf))
		}
		code, err := format.Source(buf.Bytes())
		require.NoError(t, err, buf.String())
		return string(code)
	}
	serviceCode := render(serviceFile)
	localUnionCode := render(localUnionFile)
	sharedUnionCode := render(sharedUnionFile)
	require.Contains(t, serviceCode, "Scope Scope")
	require.Contains(t, localUnionCode, "type Scope struct")
	require.Contains(t, sharedUnionCode, "type Scope struct")
}
