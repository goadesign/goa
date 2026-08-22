// This file verifies Go transformations across primitive, composite, named,
// union, service, and transport-owned attribute contexts.
package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

type (
	transformOwnerAttributor struct {
		prefix  string
		owner   string
		scope   *NameScope
		entered *[]string
	}
)

func TestGoTransform(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	var (
		scope = NewNameScope()

		// types to test
		simple   = root.UserType("Simple")
		super    = root.UserType("Super")
		required = root.UserType("Required")
		defaultT = root.UserType("Default")

		simpleMap   = root.UserType("SimpleMap")
		requiredMap = root.UserType("RequiredMap")
		defaultMap  = root.UserType("DefaultMap")
		nestedMap   = root.UserType("NestedMap")
		typeMap     = root.UserType("TypeMap")
		arrayMap    = root.UserType("ArrayMap")

		simpleArray   = root.UserType("SimpleArray")
		requiredArray = root.UserType("RequiredArray")
		defaultArray  = root.UserType("DefaultArray")
		nestedArray   = root.UserType("NestedArray")
		typeArray     = root.UserType("TypeArray")
		mapArray      = root.UserType("MapArray")

		recursive      = root.UserType("Recursive")
		recursiveArray = root.UserType("RecursiveArray")
		recursiveMap   = root.UserType("RecursiveMap")
		composite      = root.UserType("Composite")
		customField    = root.UserType("CompositeWithCustomField")
		defaults       = root.UserType("WithDefaults")

		resultType = root.UserType("ResultType")
		rtCol      = root.UserType("ResultTypeCollection")

		simpleAlias    = root.UserType("SimpleAlias")
		nestedMapAlias = root.UserType("NestedMapAlias")
		arrayMapAlias  = root.UserType("ArrayMapAlias")
		stringAlias    = root.UserType("StringAlias")

		// primitive tyes
		stringT = expr.String

		// attribute contexts used in test cases
		defaultCtx        = NewAttributeContext(false, false, true, "", scope)
		defaultCtxPkg     = NewAttributeContext(false, false, true, "mypkg", scope)
		pointerCtx        = NewAttributeContext(true, false, false, "", scope)
		defaultPointerCtx = NewAttributeContext(true, false, true, "", scope)
	)
	tc := map[string][]struct {
		Name      string
		Source    expr.DataType
		Target    expr.DataType
		SourceCtx *AttributeContext
		TargetCtx *AttributeContext
	}{
		// source and target type use default
		"source-target-type-use-default": {
			{"simple-to-simple", simple, simple, defaultCtx, defaultCtx},
			{"simple-to-required", simple, required, defaultCtx, defaultCtx},
			{"required-to-simple", required, simple, defaultCtx, defaultCtx},
			{"simple-to-super", simple, super, defaultCtx, defaultCtx},
			{"super-to-simple", super, simple, defaultCtx, defaultCtx},
			{"simple-to-default", simple, defaultT, defaultCtx, defaultCtx},
			{"default-to-simple", defaultT, simple, defaultCtx, defaultCtx},

			// maps
			{"map-to-map", simpleMap, simpleMap, defaultCtx, defaultCtx},
			{"map-to-required-map", simpleMap, requiredMap, defaultCtx, defaultCtx},
			{"required-map-to-map", requiredMap, simpleMap, defaultCtx, defaultCtx},
			{"map-to-default-map", simpleMap, defaultMap, defaultCtx, defaultCtx},
			{"default-map-to-map", defaultMap, simpleMap, defaultCtx, defaultCtx},
			{"required-map-to-default-map", requiredMap, defaultMap, defaultCtx, defaultCtx},
			{"default-map-to-required-map", defaultMap, requiredMap, defaultCtx, defaultCtx},
			{"nested-map-to-nested-map", nestedMap, nestedMap, defaultCtx, defaultCtx},
			{"type-map-to-type-map", typeMap, typeMap, defaultCtx, defaultCtx},
			{"array-map-to-array-map", arrayMap, arrayMap, defaultCtx, defaultCtx},

			// arrays
			{"array-to-array", simpleArray, simpleArray, defaultCtx, defaultCtx},
			{"array-to-required-array", simpleArray, requiredArray, defaultCtx, defaultCtx},
			{"required-array-to-array", requiredArray, simpleArray, defaultCtx, defaultCtx},
			{"array-to-default-array", simpleArray, defaultArray, defaultCtx, defaultCtx},
			{"default-array-to-array", defaultArray, simpleArray, defaultCtx, defaultCtx},
			{"required-array-to-default-array", requiredArray, defaultArray, defaultCtx, defaultCtx},
			{"default-array-to-required-array", defaultArray, requiredArray, defaultCtx, defaultCtx},
			{"nested-array-to-nested-array", nestedArray, nestedArray, defaultCtx, defaultCtx},
			{"type-array-to-type-array", typeArray, typeArray, defaultCtx, defaultCtx},
			{"map-array-to-map-array", mapArray, mapArray, defaultCtx, defaultCtx},

			// others
			{"recursive-to-recursive", recursive, recursive, defaultCtx, defaultCtx},
			{"recursive-array-to-recursive-array", recursiveArray, recursiveArray, defaultCtx, defaultCtx},
			{"recursive-map-to-recursive-map", recursiveMap, recursiveMap, defaultCtx, defaultCtx},
			{"composite-to-custom-field", composite, customField, defaultCtx, defaultCtx},
			{"custom-field-to-composite", customField, composite, defaultCtx, defaultCtx},
			{"composite-to-custom-field-pkg", composite, customField, defaultCtx, defaultCtxPkg},
			{"result-type-to-result-type", resultType, resultType, defaultCtx, defaultCtx},
			{"result-type-collection-to-result-type-collection", rtCol, rtCol, defaultCtx, defaultCtx},
			{"defaults-to-defaults-types", defaults, defaults, defaultCtx, defaultCtx},

			// alias
			{"simple-alias-to-simple", simpleAlias, simple, defaultCtx, defaultCtx},
			{"simple-to-simple-alias", simple, simpleAlias, defaultCtx, defaultCtx},
			{"nested-map-alias-to-nested-map", nestedMapAlias, nestedMap, defaultCtx, defaultCtx},
			{"nested-map-to-nested-map-alias", nestedMap, nestedMapAlias, defaultCtx, defaultCtx},
			{"array-map-alias-to-array-map", arrayMapAlias, arrayMap, defaultCtx, defaultCtx},
			{"array-map-to-array-map-alias", arrayMap, arrayMapAlias, defaultCtx, defaultCtx},
			{"string-to-string-alias", stringT, stringAlias, defaultCtx, defaultCtx},
			{"string-alias-to-string", stringAlias, stringT, defaultCtx, defaultCtx},
			{"string-alias-to-string-alias", stringAlias, stringAlias, defaultCtx, defaultCtx},
		},

		// source type uses pointers for all fields, target type uses default
		"source-type-all-ptrs-target-type-uses-default": {
			{"simple-to-simple", simple, simple, pointerCtx, defaultCtx},
			{"simple-to-required", simple, required, pointerCtx, defaultCtx},
			{"required-to-simple", required, simple, pointerCtx, defaultCtx},
			{"simple-to-super", simple, super, pointerCtx, defaultCtx},
			{"super-to-simple", super, simple, pointerCtx, defaultCtx},
			{"simple-to-default", simple, defaultT, pointerCtx, defaultCtx},
			{"default-to-simple", defaultT, simple, pointerCtx, defaultCtx},

			// maps
			{"required-map-to-map", requiredMap, simpleMap, pointerCtx, defaultCtx},
			{"default-map-to-map", defaultMap, simpleMap, pointerCtx, defaultCtx},
			{"required-map-to-default-map", requiredMap, defaultMap, pointerCtx, defaultCtx},
			{"default-map-to-required-map", defaultMap, requiredMap, pointerCtx, defaultCtx},

			// arrays
			{"default-array-to-array", defaultArray, simpleArray, pointerCtx, defaultCtx},
			{"required-array-to-default-array", requiredArray, defaultArray, pointerCtx, defaultCtx},
			{"default-array-to-required-array", defaultArray, requiredArray, pointerCtx, defaultCtx},

			// others
			{"custom-field-to-composite", customField, composite, pointerCtx, defaultCtx},

			// alias
			{"simple-alias-to-simple", simpleAlias, simple, pointerCtx, defaultCtx},
			{"simple-to-simple-alias", simple, simpleAlias, pointerCtx, defaultCtx},
		},

		// source type uses default, target type uses pointers for all fields
		"source-type-uses-default-target-type-all-ptrs": {
			{"simple-to-simple", simple, simple, defaultCtx, pointerCtx},
			{"simple-to-required", simple, required, defaultCtx, pointerCtx},
			{"required-to-simple", required, simple, defaultCtx, pointerCtx},
			{"simple-to-default", simple, defaultT, defaultCtx, pointerCtx},
			{"default-to-simple", defaultT, simple, defaultCtx, pointerCtx},

			// maps
			{"map-to-default-map", simpleMap, defaultMap, defaultCtx, pointerCtx},

			// arrays
			{"array-to-default-array", simpleArray, defaultArray, defaultCtx, pointerCtx},

			// alias
			{"simple-alias-to-simple", simpleAlias, simple, defaultCtx, pointerCtx},
			{"simple-to-simple-alias", simple, simpleAlias, defaultCtx, pointerCtx},

			// others
			{"recursive-to-recursive", recursive, recursive, defaultCtx, pointerCtx},
			{"composite-to-custom-field", composite, customField, defaultCtx, pointerCtx},
		},

		// target type uses default and pointers for all fields
		"target-type-uses-default-all-ptrs": {
			{"simple-to-simple", simple, simple, defaultCtx, defaultPointerCtx},
		},
	}
	for name, cases := range tc {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.Name, func(t *testing.T) {
					require.NotNil(t, c.Source)
					require.NotNil(t, c.Target)
					code, _, err := GoTransform(&expr.AttributeExpr{Type: c.Source}, &expr.AttributeExpr{Type: c.Target}, "source", "target", c.SourceCtx, c.TargetCtx, "", true)
					require.NoError(t, err)
					code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
					testutil.AssertGo(t, "testdata/golden/go_transform_"+name+"_"+c.Name+".go.golden", code)
				})
			}
		})
	}
}

func TestGoTransformServiceUnionField(t *testing.T) {
	union := &expr.Union{
		TypeName: "Scope",
		Values: []*expr.NamedAttributeExpr{
			{Name: "description", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "aliases", Attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
		},
	}
	object := &expr.Object{
		{Name: "scope", Attribute: &expr.AttributeExpr{Type: union}},
	}
	attribute := &expr.AttributeExpr{Type: object}
	scope := NewNameScope()
	ctx := NewAttributeContext(false, false, true, "", scope)

	code, _, err := GoTransform(attribute, attribute, "source", "target", ctx, ctx, "", true)
	require.NoError(t, err)
	require.Contains(t, code, `if source.Scope.Kind() != "" {`)
	require.NotContains(t, code, `if source.Scope != nil {`)
	require.NotContains(t, code, "scopeValue")
	require.NotContains(t, code, "target.Scope = &")
}

func TestGoTransformUnionAcrossTransportBoundary(t *testing.T) {
	union := &expr.Union{
		TypeName: "Scope",
		Values: []*expr.NamedAttributeExpr{
			{Name: "description", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "aliases", Attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
		},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "scope", Attribute: &expr.AttributeExpr{Type: union}},
	}}
	scope := NewNameScope()
	serviceCtx := NewAttributeContext(false, false, true, "", scope)
	transportCtx := serviceCtx.Dup()
	transportCtx.UnionPointer = true

	serviceToTransport, _, err := GoTransform(
		attribute, attribute, "source", "target", serviceCtx, transportCtx, "", true,
	)
	require.NoError(t, err)
	require.Contains(t, serviceToTransport, `if source.Scope.Kind() != "" {`)
	require.Contains(t, serviceToTransport, "var scopeValue Scope")
	require.Contains(t, serviceToTransport, "target.Scope = &scopeValue")

	transportToService, _, err := GoTransform(
		attribute, attribute, "source", "target", transportCtx, serviceCtx, "", true,
	)
	require.NoError(t, err)
	require.Contains(t, transportToService, `if source.Scope != nil {`)
	require.NotContains(t, transportToService, "scopeValue")
	require.NotContains(t, transportToService, "target.Scope = &")
}

func TestGoTransformEntersSourceAndTargetOwnersIndependently(t *testing.T) {
	source := transformOwnerTestType("SourceEnvelope", "SourceChoice", "source/types")
	target := transformOwnerTestType("TargetEnvelope", "TargetSelection", "target/models")
	sourceOwner := newTransformOwnerAttributor("source")
	targetOwner := newTransformOwnerAttributor("target")

	_, helpers, err := GoTransform(
		&expr.AttributeExpr{Type: source},
		&expr.AttributeExpr{Type: target},
		"source",
		"target",
		&AttributeContext{UseDefault: true, Scope: sourceOwner},
		&AttributeContext{UseDefault: true, Scope: targetOwner},
		"",
		true,
	)
	require.NoError(t, err)
	require.NotEmpty(t, helpers)
	require.Contains(t, helpers[0].ParamTypeRef, "sourceSourceChoiceContainer.SourceChoiceContainer")
	require.Contains(t, helpers[0].ResultTypeRef, "targetTargetSelectionContainer.TargetSelectionContainer")
	require.Contains(t, *sourceOwner.entered, "sourceSourceEnvelope")
	require.Contains(t, *sourceOwner.entered, "sourceSourceChoiceContainer")
	require.Contains(t, *targetOwner.entered, "targetTargetEnvelope")
	require.Contains(t, *targetOwner.entered, "targetTargetSelectionContainer")

	reverseSource := newTransformOwnerAttributor("source")
	reverseTarget := newTransformOwnerAttributor("target")
	_, reverseHelpers, err := GoTransform(
		&expr.AttributeExpr{Type: target},
		&expr.AttributeExpr{Type: source},
		"source",
		"target",
		&AttributeContext{UseDefault: true, Scope: reverseTarget},
		&AttributeContext{UseDefault: true, Scope: reverseSource},
		"",
		true,
	)
	require.NoError(t, err)
	require.NotEmpty(t, reverseHelpers)
	require.Contains(t, reverseHelpers[0].ParamTypeRef, "targetTargetSelectionContainer.TargetSelectionContainer")
	require.Contains(t, reverseHelpers[0].ResultTypeRef, "sourceSourceChoiceContainer.SourceChoiceContainer")
}

func TestTransformPlanUsesRetainedHelperIdentityDuringRender(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	deep := root.UserType("Deep")
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: deep},
		&expr.AttributeExpr{Type: deep},
	)
	require.NoError(t, err)

	planned := plan.Helpers()
	require.Len(t, planned, 2)
	declarations := make(map[TransformHelperID]*NameDeclaration, len(planned))
	plannedByID := make(map[TransformHelperID]TransformHelper, len(planned))
	packageCatalog := newGeneratedPackage("test", "example.com/test", "gen")
	for index, helper := range planned {
		declaration := NewExactName(NameFunction, fmt.Sprintf("canonicalHelper%d", index+1))
		require.NoError(t, packageCatalog.DeclareName(declaration))
		declarations[helper.ID] = declaration
		plannedByID[helper.ID] = helper
		require.NoError(t, plan.BindHelperDeclaration(helper.ID, declaration))
	}
	require.NoError(t, packageCatalog.freeze())

	attrs := &TransformAttrs{
		SourceCtx: NewAttributeContext(false, false, true, "", NewNameScope()),
		TargetCtx: NewAttributeContext(false, false, true, "", NewNameScope()),
	}
	require.NoError(t, plan.BindContexts(attrs.SourceCtx, attrs.TargetCtx))
	code, helpers, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Len(t, helpers, len(planned))
	rendered := code
	for index, helper := range helpers {
		require.Equal(t, planned[index].ID, helper.ID)
		require.Same(t, declarations[helper.ID], helper.Declaration)
		require.Same(t, plannedByID[helper.ID].Source, plan.Helpers()[index].Source)
		require.Same(t, plannedByID[helper.ID].Target, plan.Helpers()[index].Target)
		require.Empty(t, helper.Name)
		rendered += helper.Code
	}
	for _, helper := range helpers {
		require.Contains(t, rendered, helper.Declaration.Name())
	}
}

func TestTransformPlanRetainsSameTypeSiblingOccurrences(t *testing.T) {
	plan := siblingTransformPlan(t)
	require.Len(t, plan.Helpers(), 2)
	require.NotEqual(t, plan.Helpers()[0].ID, plan.Helpers()[1].ID)
}

func TestTransformPlanRejectsOneDeclarationForDifferentHelpers(t *testing.T) {
	plan := siblingTransformPlan(t)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	declaration := NewExactName(NameFunction, "transformRecursive")

	require.NoError(t, plan.BindHelperDeclaration(helpers[0].ID, declaration))
	err := plan.BindHelperDeclaration(helpers[1].ID, declaration)
	require.EqualError(t, err, "transform helper declaration is already bound to a different operation")
}

func TestTransformPlanRequiresEveryHelperDeclaration(t *testing.T) {
	plan := siblingTransformPlan(t)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	require.NoError(t, plan.BindHelperDeclaration(
		helpers[0].ID,
		NewExactName(NameFunction, "transformLeftRecursive"),
	))
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))

	_, _, err := plan.Render("source", "target", true)
	require.EqualError(t, err, "transform helper occurrence 2 has no declaration")
}

func TestTransformPlanBindsContextsOnce(t *testing.T) {
	plan := siblingTransformPlan(t)
	source := NewAttributeContext(false, false, true, "", NewNameScope())
	target := NewAttributeContext(false, false, true, "", NewNameScope())
	require.NoError(t, plan.BindContexts(source, target))
	require.EqualError(t, plan.BindContexts(source, target), "transform contexts are already bound")
}

func TestTransformPlanRejectsNonFunctionHelperDeclaration(t *testing.T) {
	plan := siblingTransformPlan(t)
	err := plan.BindHelperDeclaration(
		plan.Helpers()[0].ID,
		NewExactName(NameType, "TransformRecursive"),
	)
	require.EqualError(t, err, "transform helper declaration must be a function, got type")
}

func TestTransformPlanHelperEligibilityMatchesCompositeRenderers(t *testing.T) {
	shapes := map[string]func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr){
		"array": func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			return &expr.AttributeExpr{Type: &expr.Array{ElemType: source}},
				&expr.AttributeExpr{Type: &expr.Array{ElemType: target}}
		},
		"map": func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			return &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: source}},
				&expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: target}}
		},
		"union": func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			return &expr.AttributeExpr{Type: &expr.Union{TypeName: "SourceChoice", Values: []*expr.NamedAttributeExpr{{Name: "value", Attribute: source}}}},
				&expr.AttributeExpr{Type: &expr.Union{TypeName: "TargetChoice", Values: []*expr.NamedAttributeExpr{{Name: "value", Attribute: target}}}}
		},
	}
	pairs := map[string]struct {
		source, target *expr.AttributeExpr
		helpers        int
	}{
		"both-named": {
			source:  transformObjectAttribute("SourceNode", true),
			target:  transformObjectAttribute("TargetNode", true),
			helpers: 1,
		},
		"anonymous-source": {
			source: transformObjectAttribute("", false),
			target: transformObjectAttribute("TargetNode", true),
		},
		"anonymous-target": {
			source: transformObjectAttribute("SourceNode", true),
			target: transformObjectAttribute("", false),
		},
	}

	for shapeName, shape := range shapes {
		for pairName, pair := range pairs {
			t.Run(shapeName+"/"+pairName, func(t *testing.T) {
				source, target := shape(pair.source, pair.target)
				plan, err := NewTransformPlan(source, target)
				require.NoError(t, err)
				require.Len(t, plan.Helpers(), pair.helpers)

				code, helpers := renderTransformPlan(t, plan)
				require.Len(t, helpers, pair.helpers)
				if pair.helpers == 0 {
					require.NotContains(t, code, "CanonicalHelper")
				} else {
					require.Contains(t, code, "CanonicalHelper1")
				}
			})
		}
	}
}

func TestTransformPlanRetainsRequiredAndOptionalSiblingCalls(t *testing.T) {
	plan := mixedSiblingTransformPlan(t)
	code, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 2)

	var required, optional *TransformFunctionData
	for _, definition := range definitions {
		if plan.Helpers()[definition.ID.index].Required {
			required = definition
		} else {
			optional = definition
		}
	}
	require.NotNil(t, required)
	require.NotNil(t, optional)
	require.NotContains(t, required.Code, "if v == nil")
	require.Contains(t, optional.Code, "if v == nil")
	require.Contains(t, code, "target.Left = "+required.Declaration.Name()+"(source.Left)")
	require.Contains(t, code, "target.Right = "+optional.Declaration.Name()+"(source.Right)")
}

func TestTransformPlanMapKeyHelperReceivesKey(t *testing.T) {
	sourceKey := transformObjectAttribute("SourceKey", true)
	targetKey := transformObjectAttribute("TargetKey", true)
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Map{
			KeyType:  sourceKey,
			ElemType: &expr.AttributeExpr{Type: expr.String},
		}},
		&expr.AttributeExpr{Type: &expr.Map{
			KeyType:  targetKey,
			ElemType: &expr.AttributeExpr{Type: expr.String},
		}},
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)
	require.Same(t, sourceKey, plan.Helpers()[0].Source)
	require.Same(t, targetKey, plan.Helpers()[0].Target)

	code, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 1)
	bound := plan.Helpers()[0]
	require.Equal(t, bound.ID, definitions[0].ID)
	require.Same(t, bound.Declaration, definitions[0].Declaration)

	generated := fmt.Sprintf(`package transformtest

type SourceKey struct {
	Value string
}

type TargetKey struct {
	Value string
}

func transform(source map[*SourceKey]string) map[*TargetKey]string {
%s
	return target
}

func %s(v %s) %s {
%s
	return res
}
`, code, definitions[0].Declaration.Name(), definitions[0].ParamTypeRef,
		definitions[0].ResultTypeRef, definitions[0].Code)
	compileTransformSource(t, generated)
	require.Contains(t, code, definitions[0].Declaration.Name()+"(key)")
}

// siblingTransformPlan builds two nonrecursive occurrences of the same named
// recursive type. Each field must own a helper even though both types share an
// authored origin.
func siblingTransformPlan(t *testing.T) *TransformPlan {
	t.Helper()
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	fields := &expr.Object{}
	fields.Set("left", &expr.AttributeExpr{Type: recursive})
	fields.Set("right", &expr.AttributeExpr{Type: recursive})
	container := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: fields},
		TypeName:      "Container",
	}
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: container},
		&expr.AttributeExpr{Type: container},
	)
	require.NoError(t, err)
	return plan
}

// mixedSiblingTransformPlan builds required and optional occurrences of the
// same recursive named type in one transform operation.
func mixedSiblingTransformPlan(t *testing.T) *TransformPlan {
	t.Helper()
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	fields := &expr.Object{}
	fields.Set("left", &expr.AttributeExpr{Type: recursive})
	fields.Set("right", &expr.AttributeExpr{Type: recursive})
	container := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type:       fields,
			Validation: &expr.ValidationExpr{Required: []string{"left"}},
		},
		TypeName: "Container",
	}
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: container},
		&expr.AttributeExpr{Type: container},
	)
	require.NoError(t, err)
	return plan
}

// transformObjectAttribute builds either a named or anonymous object with the
// same compatible field shape.
func transformObjectAttribute(name string, named bool) *expr.AttributeExpr {
	object := &expr.Object{}
	object.Set("value", &expr.AttributeExpr{Type: expr.String})
	if !named {
		return &expr.AttributeExpr{Type: object}
	}
	return &expr.AttributeExpr{Type: &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: object},
		TypeName:      name,
	}}
}

// renderTransformPlan binds deterministic function declarations and contexts,
// then renders the retained operation and definitions.
func renderTransformPlan(t *testing.T, plan *TransformPlan) (string, []*TransformFunctionData) {
	t.Helper()
	packageCatalog := newGeneratedPackage("test", "example.com/test", "gen")
	for index, helper := range plan.Helpers() {
		declaration := NewExactName(NameFunction, fmt.Sprintf("canonicalHelper%d", index+1))
		require.NoError(t, packageCatalog.DeclareName(declaration))
		require.NoError(t, plan.BindHelperDeclaration(helper.ID, declaration))
	}
	require.NoError(t, packageCatalog.freeze())
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	code, helpers, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	return code, helpers
}

// compileTransformSource proves that a rendered transform and its retained
// helper definitions agree on concrete Go argument and result types.
func compileTransformSource(t *testing.T, source string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/transformtest\n\ngo 1.25.0\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "transform.go"), []byte(source), 0o600))
	command := exec.Command("go", "test", "./...")
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("generated transform did not compile: %v\n%s", err, output)
	}
}

func newTransformOwnerAttributor(prefix string) *transformOwnerAttributor {
	entered := make([]string, 0)
	return &transformOwnerAttributor{
		prefix:  prefix,
		scope:   NewNameScope(),
		entered: &entered,
	}
}

func (a *transformOwnerAttributor) Name(att *expr.AttributeExpr, _ string, _, _ bool) string {
	return a.owner + "." + codegenTypeName(att)
}

func (a *transformOwnerAttributor) Ref(att *expr.AttributeExpr, pkg string) string {
	name := a.Name(att, pkg, false, false)
	if expr.IsObject(att.Type) || expr.IsUnion(att.Type) {
		return "*" + name
	}
	return name
}

func (*transformOwnerAttributor) Field(_ *expr.AttributeExpr, name string, firstUpper bool) string {
	return Goify(name, firstUpper)
}

func (a *transformOwnerAttributor) Package(_ *expr.AttributeExpr) string {
	return a.owner
}

func (a *transformOwnerAttributor) Enter(att *expr.AttributeExpr) Attributor {
	entered := *a
	entered.owner = a.prefix + codegenTypeName(att)
	*a.entered = append(*a.entered, entered.owner)
	return &entered
}

func (*transformOwnerAttributor) IsSumType() bool {
	return true
}

func (a *transformOwnerAttributor) ValidatorName(att *expr.AttributeExpr, view string) string {
	return "Validate" + a.Name(att, "", false, true) + Goify(view, true)
}

func (a *transformOwnerAttributor) Scope() *NameScope {
	return a.scope
}

func transformOwnerTestType(name, unionName, location string) expr.UserType {
	union := &expr.Union{
		TypeName: unionName,
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "number", Attribute: &expr.AttributeExpr{Type: expr.Int}},
		},
	}
	container := &expr.UserTypeExpr{
		TypeName: unionName + "Container",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "choice", Attribute: &expr.AttributeExpr{Type: union}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {location}},
		},
	}
	return &expr.UserTypeExpr{
		TypeName: name,
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "inner", Attribute: &expr.AttributeExpr{Type: container}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {location}},
		},
	}
}

func codegenTypeName(att *expr.AttributeExpr) string {
	if att.Type.Name() == "object" {
		return "Object"
	}
	return Goify(att.Type.Name(), true)
}
