// This file verifies Go transformations across primitive, composite, named,
// union, service, and transport-owned attribute contexts.
package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

type (
	// transformOwnerAttributor records each nested type name selected while a
	// test plans and writes a conversion.
	transformOwnerAttributor struct {
		prefix  string
		owner   string
		scope   *NameScope
		entered *[]string
	}

	// transformIdentityAttributor records the exact attributes passed to field
	// name lookups while a plan renders its copied expressions.
	transformIdentityAttributor struct {
		Attributor
		fields *[]*expr.AttributeExpr
	}

	// transformTestOrderFactory combines a test conversion key with each exact
	// helper location supplied by the registry.
	transformTestOrderFactory string
)

// order returns one stable package name order for a helper occurrence.
func (f transformTestOrderFactory) order(location TransformHelperDefinitionLocation) PackageNameOrder {
	return testNameOrder{value: string(f) + location.encoded}
}

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

// TestGoTransformUnionKeepsNilSelectedBranch verifies that conversion leaves a
// selected nil branch for the destination validator to report.
func TestGoTransformUnionKeepsNilSelectedBranch(t *testing.T) {
	details := goTypeTestUserType("Details", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})
	union := &expr.Union{
		TypeName: "State",
		Values: []*expr.NamedAttributeExpr{
			{Name: "details", Attribute: &expr.AttributeExpr{Type: details}},
			{Name: "empty", Attribute: &expr.AttributeExpr{Type: expr.Empty}},
			{Name: "aliases", Attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
			{Name: "labels", Attribute: &expr.AttributeExpr{Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}},
			{Name: "blob", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
			{Name: "anything", Attribute: &expr.AttributeExpr{Type: expr.Any}},
			{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	code, _, err := GoTransform(
		&expr.AttributeExpr{Type: union},
		&expr.AttributeExpr{Type: union},
		"source",
		"target",
		context,
		context,
		"",
		true,
	)
	require.NoError(t, err)
	code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
	testutil.AssertGo(t, "testdata/golden/go_transform_union_nil_branch.go.golden", code)
}

// TestGoTransformRequiredPrimitiveArrayElements verifies that JSON presence
// pointers are removed after validation and added only when encoding that form.
func TestGoTransformRequiredPrimitiveArrayElements(t *testing.T) {
	alias := goTypeTestUserType("Alias", expr.String)
	array := &expr.AttributeExpr{Type: &expr.Array{
		ElemType:         &expr.AttributeExpr{Type: alias},
		NonNullableElems: true,
	}}
	scope := NewNameScope()
	service := NewAttributeContext(false, false, true, "", scope)
	jsonBody := service.Dup()
	jsonBody.ArrayElementPointer = true

	decode, _, err := GoTransform(array, array, "source", "target", jsonBody, service, "", true)
	require.NoError(t, err)
	require.Contains(t, decode, "target := make([]Alias, len(source))")
	require.Contains(t, decode, "target[i] = *val")

	encode, _, err := GoTransform(array, array, "source", "target", service, jsonBody, "", true)
	require.NoError(t, err)
	require.Contains(t, encode, "target := make([]*Alias, len(source))")
	require.Contains(t, encode, "var transformed Alias")
	require.Contains(t, encode, "target[i] = &transformed")

	ordinary := expr.DupAtt(array)
	expr.AsArray(ordinary.Type).NonNullableElems = false
	unchanged, _, err := GoTransform(ordinary, ordinary, "source", "target", jsonBody, service, "", true)
	require.NoError(t, err)
	require.NotContains(t, unchanged, "*val")
}

// TestGoTransformUsesDesignNilabilityForCustomTypes verifies that default
// handling does not infer comparability from a generated Go type spelling.
func TestGoTransformUsesDesignNilabilityForCustomTypes(t *testing.T) {
	raw := &expr.AttributeExpr{
		Type:         expr.String,
		DefaultValue: json.RawMessage("foo"),
		Meta: expr.MetaExpr{
			"struct:field:type": {"json.RawMessage", "encoding/json", "json"},
		},
	}
	defaults := goTypeTestUserType("WithRaw", &expr.Object{
		{Name: "raw", Attribute: raw},
	})
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	code, _, err := GoTransform(
		&expr.AttributeExpr{Type: defaults},
		&expr.AttributeExpr{Type: defaults},
		"source",
		"target",
		context,
		context,
		"",
		true,
	)
	require.NoError(t, err)
	require.Contains(t, code, "if target.Raw == nil")
	require.NotContains(t, code, "var zero json.RawMessage")
	compileTransformSource(t, `package transformtest

import "encoding/json"

type WithRaw struct {
	Raw json.RawMessage
}

func transform(source *WithRaw) *WithRaw {
`+code+`
	return target
}
`)
}

// TestGoTransformDefaultUsesFinalCustomTypeImportAlias verifies that the
// zero-value check uses the same package name as the generated import.
func TestGoTransformDefaultUsesFinalCustomTypeImportAlias(t *testing.T) {
	const customPath = "example.com/custom/strconv"
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	field := &expr.AttributeExpr{
		Type:         expr.String,
		DefaultValue: "ready",
		Meta: expr.MetaExpr{
			"struct:field:type": {"strconv.Token", customPath, "strconv"},
		},
	}
	withDefault := goTypeTestUserType("WithDefault", &expr.Object{
		{Name: "token", Attribute: field},
	})
	_, err := pkg.DeclareUserType(withDefault)
	require.NoError(t, err)
	require.NoError(t, pkg.DeclareImport(NewImport("strconv", customPath)))
	require.NoError(t, pkg.RequireImport(SimpleImport("strconv")))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "strconv2", pkg.ImportName(customPath))
	context := NewAttributeContext(false, false, true, "", pkg.Scope())

	code, _, err := GoTransform(
		&expr.AttributeExpr{Type: withDefault},
		&expr.AttributeExpr{Type: withDefault},
		"source",
		"target",
		context,
		context,
		"",
		true,
	)
	require.NoError(t, err)
	require.Contains(t, code, "var zero strconv2.Token")
	require.NotContains(t, code, "var zero strconv.Token")
}

// TestGoTransformReturnsFirstDefaultError verifies that a later field cannot
// hide an invalid earlier default while the generator walks an object.
func TestGoTransformReturnsFirstDefaultError(t *testing.T) {
	invalid := &expr.AttributeExpr{
		Type:         expr.String,
		DefaultValue: json.RawMessage("foo"),
		Meta: expr.MetaExpr{
			"struct:field:type": {"json.RawMessage", "example.com/not-json", "json"},
		},
	}
	valid := &expr.AttributeExpr{Type: expr.String, DefaultValue: "ready"}
	defaults := goTypeTestUserType("Defaults", &expr.Object{
		{Name: "invalid", Attribute: invalid},
		{Name: "valid", Attribute: valid},
	})
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	_, _, err := GoTransform(
		&expr.AttributeExpr{Type: defaults},
		&expr.AttributeExpr{Type: defaults},
		"source",
		"target",
		context,
		context,
		"",
		true,
	)
	require.EqualError(t, err, `render Go value for string: default for custom Go type "json.RawMessage" has Go type json.RawMessage`)
}

// TestGoTransformArrayLoopNameUsesNestingDepth verifies that brackets in a
// caller expression do not change the generated loop variable.
func TestGoTransformArrayLoopNameUsesNestingDepth(t *testing.T) {
	array := &expr.AttributeExpr{Type: &expr.Array{
		ElemType: &expr.AttributeExpr{Type: expr.String},
	}}
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	code, _, err := GoTransform(array, array, "source", "target[index]", context, context, "", false)
	require.NoError(t, err)
	require.Contains(t, code, "for i, val := range source")
	require.NotContains(t, code, "for j, val := range source")
}

// TestGoTransformUnionTemporaryUsesNestingDepth verifies that a caller's
// destination spelling does not select the local used for a union branch.
func TestGoTransformUnionTemporaryUsesNestingDepth(t *testing.T) {
	source := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "SourceChoice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}}
	target := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "TargetChoice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}}
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	code, _, err := GoTransform(source, target, "source", "target.Selected", context, context, "", false)
	require.NoError(t, err)
	require.Contains(t, code, "obj := actual")
	require.NotContains(t, code, "tmp := actual")

	nested, err := transformUnion(source, target, "source", "target.Selected", false, &TransformAttrs{
		SourceCtx:  context,
		TargetCtx:  context,
		unionDepth: 1,
	})
	require.NoError(t, err)
	require.Contains(t, nested, "tmp2 := actual")
	require.NotContains(t, nested, "obj := actual")
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

// TestGoTransformEntersArrayFieldOwners verifies that a named array element
// is resolved from the array field rather than the object that contains it.
func TestGoTransformEntersArrayFieldOwners(t *testing.T) {
	sourceComponent := goTypeTestUserType("SourceComponent", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})
	targetComponent := goTypeTestUserType("TargetComponent", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})
	source := goTypeTestUserType("SourceEnvelope", &expr.Object{
		{Name: "components", Attribute: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: sourceComponent},
		}}},
	})
	target := goTypeTestUserType("TargetEnvelope", &expr.Object{
		{Name: "components", Attribute: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: targetComponent},
		}}},
	})
	sourceOwner := newTransformOwnerAttributor("source")
	targetOwner := newTransformOwnerAttributor("target")

	code, _, err := GoTransform(
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
	require.Contains(t, code, "make([]*targetArray.TargetComponent")
	require.Contains(t, *sourceOwner.entered, "sourceArray")
	require.Contains(t, *targetOwner.entered, "targetArray")
}

func TestTransformPlanUsesRetainedHelperIdentityDuringRender(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	deep := root.UserType("Deep")
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: deep},
		&expr.AttributeExpr{Type: deep},
		"",
		nil,
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
		require.NotSame(t, plannedByID[helper.ID].Source, plan.Helpers()[index].Source)
		require.NotSame(t, plannedByID[helper.ID].Target, plan.Helpers()[index].Target)
		require.Equal(t, plannedByID[helper.ID].Source.Type.Name(), plan.Helpers()[index].Source.Type.Name())
		require.Equal(t, plannedByID[helper.ID].Target.Type.Name(), plan.Helpers()[index].Target.Type.Name())
		require.Equal(t, helper.Declaration.Name(), helper.Name)
		rendered += helper.Code
	}
	for _, helper := range helpers {
		require.Contains(t, rendered, helper.Declaration.Name())
	}
}

func TestTransformPlanKeepsDistinctCopiesWithOneOrigin(t *testing.T) {
	sourceOrigin := transformObjectAttribute("SourceNode", true).Type.(expr.UserType)
	targetOrigin := transformObjectAttribute("TargetNode", true).Type.(expr.UserType)
	sourceOuter := sourceOrigin.Dup(nil)
	sourceInner := sourceOrigin.Dup(nil)
	targetOuter := targetOrigin.Dup(nil)
	targetInner := targetOrigin.Dup(nil)
	sourceInner.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})
	targetInner.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})
	sourceOuter.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: sourceInner}},
	}})
	targetOuter.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: targetInner}},
	}})

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "root", Attribute: &expr.AttributeExpr{Type: sourceOuter}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "root", Attribute: &expr.AttributeExpr{Type: targetOuter}},
		}},
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 2)
	require.NotSame(t, plan.Helpers()[0].Source.Type, plan.Helpers()[1].Source.Type)
	require.NotSame(t, plan.Helpers()[0].Target.Type, plan.Helpers()[1].Target.Type)

	code, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 2)
	require.Contains(t, code, definitions[0].Declaration.Name()+"(source.Root)")
	require.Contains(t, definitions[0].Code, definitions[1].Declaration.Name()+"(v.Child)")
}

func TestTransformPlanClosesExactRecursiveCycle(t *testing.T) {
	sourceNode := &expr.UserTypeExpr{TypeName: "SourceNode"}
	targetNode := &expr.UserTypeExpr{TypeName: "TargetNode"}
	sourceNode.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "next", Attribute: &expr.AttributeExpr{Type: sourceNode}},
	}})
	targetNode.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "next", Attribute: &expr.AttributeExpr{Type: targetNode}},
	}})

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "root", Attribute: &expr.AttributeExpr{Type: sourceNode}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "root", Attribute: &expr.AttributeExpr{Type: targetNode}},
		}},
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)

	code, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 1)
	require.Contains(t, code, definitions[0].Declaration.Name()+"(source.Root)")
	require.Contains(t, definitions[0].Code, definitions[0].Declaration.Name()+"(v.Next)")
}

func TestTransformPlanCopiesCallerExpressions(t *testing.T) {
	sourceObject := &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}
	targetObject := &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}
	sourceType := &expr.UserTypeExpr{
		TypeName:      "Source",
		AttributeExpr: &expr.AttributeExpr{Type: sourceObject},
	}
	targetType := &expr.UserTypeExpr{
		TypeName:      "Target",
		AttributeExpr: &expr.AttributeExpr{Type: targetObject},
	}
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: sourceType},
		&expr.AttributeExpr{Type: targetType},
		"",
		nil,
	)
	require.NoError(t, err)
	sourceObject.Set("late", &expr.AttributeExpr{Type: expr.String})
	targetObject.Set("late", &expr.AttributeExpr{Type: expr.String})

	code, definitions := renderTransformPlan(t, plan)
	require.Empty(t, definitions)
	require.Contains(t, code, "Value: source.Value")
	require.NotContains(t, code, "Late")
}

func TestTransformPlanCopiesTypedCollectionDefaults(t *testing.T) {
	tests := []struct {
		name         string
		dataType     expr.DataType
		defaultValue any
		planned      string
		changed      string
	}{
		{
			name:         "slice",
			dataType:     &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}},
			defaultValue: []string{"planned"},
			planned:      `[]string{"planned"}`,
			changed:      `[]string{"changed"}`,
		},
		{
			name: "string-keyed-map",
			dataType: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.Int},
			},
			defaultValue: map[string]int{"value": 1},
			planned:      `map[string]int{"value":1}`,
			changed:      `map[string]int{"value":2}`,
		},
		{
			name: "integer-keyed-map",
			dataType: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.Int},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			},
			defaultValue: map[int]string{1: "planned"},
			planned:      `map[int]string{1:"planned"}`,
			changed:      `map[int]string{1:"changed"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceField := &expr.AttributeExpr{Type: test.dataType}
			targetField := &expr.AttributeExpr{Type: test.dataType, DefaultValue: test.defaultValue}
			source := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
				TypeName: "Source",
				AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
					{Name: "value", Attribute: sourceField},
				}},
			}}
			target := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
				TypeName: "Target",
				AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
					{Name: "value", Attribute: targetField},
				}},
			}}
			plan, err := NewTransformPlan(source, target, "", nil)
			require.NoError(t, err)

			switch value := test.defaultValue.(type) {
			case []string:
				value[0] = "changed"
			case map[string]int:
				value["value"] = 2
			case map[int]string:
				value[1] = "changed"
			default:
				t.Fatalf("missing mutation for %T", test.defaultValue)
			}
			code, definitions := renderTransformPlan(t, plan)
			require.Empty(t, definitions)
			require.Contains(t, code, test.planned)
			require.NotContains(t, code, test.changed)
		})
	}
}

func TestTransformPlanNameLookupsUseOriginalAttributes(t *testing.T) {
	sourceField := &expr.AttributeExpr{Type: expr.String}
	targetField := &expr.AttributeExpr{Type: expr.String}
	source := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: sourceField},
	}}
	target := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: targetField},
	}}
	plan, err := NewTransformPlan(source, target, "", nil)
	require.NoError(t, err)

	var sourceFields, targetFields []*expr.AttributeExpr
	sourceContext := NewAttributeContext(false, false, true, "", NewNameScope())
	sourceContext.Scope = &transformIdentityAttributor{
		Attributor: sourceContext.Scope,
		fields:     &sourceFields,
	}
	targetContext := NewAttributeContext(false, false, true, "", NewNameScope())
	targetContext.Scope = &transformIdentityAttributor{
		Attributor: targetContext.Scope,
		fields:     &targetFields,
	}
	require.NoError(t, plan.BindContexts(sourceContext, targetContext))
	_, _, err = plan.Render("source", "target", true)
	require.NoError(t, err)
	require.NotEmpty(t, sourceFields)
	require.NotEmpty(t, targetFields)
	for _, field := range sourceFields {
		require.Same(t, sourceField, field)
	}
	for _, field := range targetFields {
		require.Same(t, targetField, field)
	}
}

func TestTransformPlanCapturesStructuralHookChoices(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}}
	var unwrapCalls, fieldCalls int
	hooks := &TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
			unwrapCalls++
			return source, target, nil
		},
		FieldPairAttrs: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			fieldCalls++
			return source, target
		},
	}
	plan, err := NewTransformPlan(attribute, attribute, "", hooks)
	require.NoError(t, err)
	plannedUnwrapCalls, plannedFieldCalls := unwrapCalls, fieldCalls
	require.Positive(t, plannedUnwrapCalls)
	require.Positive(t, plannedFieldCalls)

	code, definitions := renderTransformPlan(t, plan)
	require.Empty(t, definitions)
	require.Contains(t, code, "Value: source.Value")
	require.Equal(t, plannedUnwrapCalls, unwrapCalls)
	require.Equal(t, plannedFieldCalls, fieldCalls)
}

// TestTransformPrimitiveUsesHandledHookExpression checks that a hook may
// intentionally return the source expression without triggering Goa's normal
// type conversion.
func TestTransformPrimitiveUsesHandledHookExpression(t *testing.T) {
	source := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
		TypeName:      "SourceAlias",
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
	}}
	target := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
		TypeName:      "TargetAlias",
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
	}}
	context := NewAttributeContext(false, false, false, "", NewNameScope())
	attributes := &TransformAttrs{
		SourceCtx: context,
		TargetCtx: context,
		Hooks: &TransformHooks{
			ConvertPrimitive: func(_, _ *expr.AttributeExpr, sourceVar string, _, _ bool, _ *TransformAttrs) (string, bool) {
				return sourceVar, true
			},
		},
	}

	generated, err := TransformAttribute(source, target, "source", "target", true, attributes)
	require.NoError(t, err)
	require.Equal(t, "target := source\n", generated)
}

// TestTransformAttributeDerivesWrappedPrimitivePointers checks that wrapper
// conversion follows the pointer layout defined by each type context.
func TestTransformAttributeDerivesWrappedPrimitivePointers(t *testing.T) {
	wrapper := &expr.UserTypeExpr{
		TypeName: "PrimitiveWrapper",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "field", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}},
	}
	wrapperAttribute := &expr.AttributeExpr{Type: wrapper}
	primitiveAttribute := &expr.AttributeExpr{Type: expr.String}
	valueContext := NewAttributeContext(false, false, false, "", NewNameScope())
	pointerContext := NewAttributeContext(true, false, false, "", NewNameScope())

	t.Run("target wrapper", func(t *testing.T) {
		attributes := &TransformAttrs{
			SourceCtx: valueContext,
			TargetCtx: pointerContext,
			Hooks: &TransformHooks{UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
				return source, expr.AsObject(target.Type).Attribute("field"), &WrapDirective{
					WrapTarget: true,
					Target:     target,
					FieldName:  "Field",
				}
			}},
		}

		generated, err := TransformAttribute(primitiveAttribute, wrapperAttribute, "source", "target", true, attributes)
		require.NoError(t, err)
		require.Equal(t, "target := &PrimitiveWrapper{}\ntarget.Field = new(string)\n*target.Field = source\n", generated)
	})

	t.Run("source wrapper", func(t *testing.T) {
		attributes := &TransformAttrs{
			SourceCtx: pointerContext,
			TargetCtx: valueContext,
			Hooks: &TransformHooks{UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
				return expr.AsObject(source.Type).Attribute("field"), target, &WrapDirective{FieldName: "Field"}
			}},
		}

		generated, err := TransformAttribute(wrapperAttribute, primitiveAttribute, "source", "target", true, attributes)
		require.NoError(t, err)
		require.Equal(t, "target := *source.Field\n", generated)
	})
}

func TestTransformPlanRejectsPlanningUnwrapPairMutation(t *testing.T) {
	sourceField := &expr.AttributeExpr{Type: expr.String, DefaultValue: "before"}
	targetField := &expr.AttributeExpr{Type: expr.String}
	source := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: sourceField}}}
	target := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: targetField}}}

	_, err := NewTransformPlan(source, target, "", &TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
			if object := expr.AsObject(source.Type); object != nil {
				object.Attribute("value").DefaultValue = "after"
			}
			return source, target, nil
		},
	})

	require.EqualError(t, err, "transform planning hook UnwrapPair changed the retained plan")
}

func TestTransformPlanRejectsPlanningFieldPairMutation(t *testing.T) {
	sourceField := &expr.AttributeExpr{Type: expr.String}
	targetField := &expr.AttributeExpr{Type: expr.String}
	source := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: sourceField}}}
	target := &expr.AttributeExpr{Type: &expr.Object{{Name: "value", Attribute: targetField}}}

	_, err := NewTransformPlan(source, target, "", &TransformHooks{
		FieldPairAttrs: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
			target.Meta = expr.MetaExpr{"mutated": {"yes"}}
			return source, target
		},
	})

	require.EqualError(t, err, "transform planning hook FieldPairAttrs changed the retained plan")
}

func TestTransformPlanRejectsPlanningUnionHelperMutation(t *testing.T) {
	sourceBranch := &expr.AttributeExpr{Type: expr.String}
	targetBranch := &expr.AttributeExpr{Type: expr.String}
	source := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{{Name: "value", Attribute: sourceBranch}}}}
	target := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{{Name: "value", Attribute: targetBranch}}}}

	_, err := NewTransformPlan(source, target, "", &TransformHooks{
		TransformUnion: func(_ *expr.AttributeExpr, _ *expr.AttributeExpr, _, _ string, _ bool, _, _ *expr.AttributeExpr, _ *TransformAttrs) (string, error) {
			return "", nil
		},
		PlanUnionHelpers: func(source, target *expr.AttributeExpr, _ func(*expr.AttributeExpr, *expr.AttributeExpr)) {
			expr.AsUnion(source.Type).Values[0].Attribute.DefaultValue = "after"
			expr.AsUnion(target.Type).Values[0].Attribute.Meta = expr.MetaExpr{"mutated": {"yes"}}
		},
	})

	require.EqualError(t, err, "transform planning hook PlanUnionHelpers changed the retained plan")
}

func TestTransformPlanRejectsDerivedUnionHelperBranches(t *testing.T) {
	sourceType := transformObjectAttribute("SourceBranch", true).Type
	targetType := transformObjectAttribute("TargetBranch", true).Type
	source := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: sourceType}},
	}}}
	target := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: targetType}},
	}}}

	_, err := NewTransformPlan(source, target, "", &TransformHooks{
		TransformUnion: func(_ *expr.AttributeExpr, _ *expr.AttributeExpr, _, _ string, _ bool, _, _ *expr.AttributeExpr, _ *TransformAttrs) (string, error) {
			return "", nil
		},
		PlanUnionHelpers: func(source, target *expr.AttributeExpr, record func(*expr.AttributeExpr, *expr.AttributeExpr)) {
			record(
				expr.DupAtt(expr.AsUnion(source.Type).Values[0].Attribute),
				expr.DupAtt(expr.AsUnion(target.Type).Values[0].Attribute),
			)
		},
	})
	require.EqualError(t, err, "custom union transform helper must use retained authored branch attributes")
}

func TestGoTransformWithAttrsCallsStructuralHookOnce(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	var calls int
	attrs := &TransformAttrs{
		SourceCtx: NewAttributeContext(false, false, true, "", NewNameScope()),
		TargetCtx: NewAttributeContext(false, false, true, "", NewNameScope()),
		Hooks: &TransformHooks{
			UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
				calls++
				return source, target, nil
			},
		},
	}

	code, helpers, err := GoTransformWithAttrs(attribute, attribute, "source", "target", attrs, true)
	require.NoError(t, err)
	require.Empty(t, helpers)
	require.Equal(t, "target := source", code)
	require.Equal(t, 1, calls)
}

func TestGoTransformWithAttrsSharesRequiredAndOptionalHelper(t *testing.T) {
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
	attribute := &expr.AttributeExpr{Type: container}
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	code, helpers, err := GoTransformWithAttrs(attribute, attribute, "source", "target", &TransformAttrs{
		SourceCtx: context,
		TargetCtx: context,
	}, true)
	require.NoError(t, err)
	require.Len(t, helpers, 1)
	require.NotContains(t, helpers[0].Code, "if v == nil")
	require.Contains(t, code, "target.Left = transformRecursiveToRecursive(source.Left)")
	require.Contains(t, code, "if source.Right != nil {")
	require.Contains(t, code, "target.Right = transformRecursiveToRecursive(source.Right)")
}

func TestTransformPlanUsesCustomUnionHelperOrderForNamedArraysAndAliases(t *testing.T) {
	sourceAlias := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "SourceAlias",
	}
	targetAlias := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "TargetAlias",
	}
	sourceArray := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
		TypeName:      "SourceArray",
	}
	targetArray := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
		TypeName:      "TargetArray",
	}
	source := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{
		{Name: "alias", Attribute: &expr.AttributeExpr{Type: sourceAlias}},
		{Name: "array", Attribute: &expr.AttributeExpr{Type: sourceArray}},
	}}}
	target := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{
		{Name: "alias", Attribute: &expr.AttributeExpr{Type: targetAlias}},
		{Name: "array", Attribute: &expr.AttributeExpr{Type: targetArray}},
	}}}
	hooks := &TransformHooks{
		PlanUnionHelpers: func(source, target *expr.AttributeExpr, record func(*expr.AttributeExpr, *expr.AttributeExpr)) {
			sourceUnion, targetUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
			record(sourceUnion.Values[1].Attribute, targetUnion.Values[1].Attribute)
			record(sourceUnion.Values[0].Attribute, targetUnion.Values[0].Attribute)
		},
		TransformUnion: func(source, target *expr.AttributeExpr, _, _ string, _ bool, _, _ *expr.AttributeExpr, attrs *TransformAttrs) (string, error) {
			sourceUnion, targetUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
			arrayName := TransformHelperName(sourceUnion.Values[1].Attribute, targetUnion.Values[1].Attribute, attrs)
			aliasName := TransformHelperName(sourceUnion.Values[0].Attribute, targetUnion.Values[0].Attribute, attrs)
			return arrayName + "(source.Array)\n" + aliasName + "(source.Alias)\n", nil
		},
	}
	plan, err := NewTransformPlan(source, target, "", hooks)
	require.NoError(t, err)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	require.Equal(t, "SourceArray", helpers[0].Source.Type.Name())
	require.Equal(t, "SourceAlias", helpers[1].Source.Type.Name())
	arrayDeclaration := NewExactName(NameFunction, "arrayHelper")
	aliasDeclaration := NewExactName(NameFunction, "aliasHelper")
	generatedPackage := newGeneratedPackage("test", "example.com/test", "gen")
	require.NoError(t, generatedPackage.DeclareName(arrayDeclaration))
	require.NoError(t, generatedPackage.DeclareName(aliasDeclaration))
	require.NoError(t, plan.BindHelperDeclaration(helpers[0].ID, arrayDeclaration))
	require.NoError(t, plan.BindHelperDeclaration(helpers[1].ID, aliasDeclaration))
	require.NoError(t, generatedPackage.freeze())
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))

	code, definitions, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Len(t, definitions, 2)
	require.Contains(t, code, "arrayHelper")
	require.Contains(t, code, "aliasHelper")
	require.Less(t, strings.Index(code, "arrayHelper"), strings.Index(code, "aliasHelper"))
}

func TestTransformPlanCapturesUnwrappedHelperBody(t *testing.T) {
	sourceNode := transformObjectAttribute("SourceNode", true)
	targetNode := transformObjectAttribute("TargetNode", true)
	wrapper := &expr.UserTypeExpr{
		TypeName: "TargetWrapper",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "field", Attribute: targetNode},
		}},
	}
	source := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "root", Attribute: sourceNode},
	}}
	target := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "root", Attribute: &expr.AttributeExpr{Type: wrapper}},
	}}
	hooks := &TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
			if target.Type.Name() != wrapper.TypeName {
				return source, target, nil
			}
			return source, expr.AsObject(target.Type).Attribute("field"), &WrapDirective{
				WrapTarget: true,
				Target:     target,
				FieldName:  "Field",
			}
		},
	}
	plan, err := NewTransformPlan(source, target, "", hooks)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)

	code, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 1)
	require.Contains(t, code, definitions[0].Declaration.Name()+"(source.Root)")
}

func TestTransformPlanRetainsWrapperAndInlineArrayCalls(t *testing.T) {
	node := &expr.UserTypeExpr{TypeName: "Node"}
	node.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{}}
	expr.AsObject(node.Type).Set("next", &expr.AttributeExpr{Type: node})
	source := &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: node}}}
	target := expr.DupAtt(source)
	hooks := &TransformHooks{
		InlineCompositeElems: true,
		TransformArray: func(source, target *expr.Array, sourceVar, targetVar string, newVar bool, attrs *TransformAttrs) (string, error) {
			return TransformAttribute(source.ElemType, target.ElemType, sourceVar+"[0]", targetVar+"[0]", newVar, attrs)
		},
	}
	plan, err := NewTransformPlan(source, target, "copied", hooks)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)

	declaration := NewExactName(NameFunction, "copyNode")
	packageCatalog := newGeneratedPackage("test", "example.com/test", "gen")
	require.NoError(t, packageCatalog.DeclareName(declaration))
	require.NoError(t, plan.BindHelperDeclaration(plan.Helpers()[0].ID, declaration))
	require.NoError(t, packageCatalog.freeze())
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	code, helpers, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Len(t, helpers, 1)
	require.Contains(t, code+helpers[0].Code, "copyNode")

	wrapper := &expr.UserTypeExpr{
		TypeName: "WrappedNode",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			&expr.NamedAttributeExpr{Name: "field", Attribute: &expr.AttributeExpr{Type: node}},
		}},
	}
	wrapperTarget := &expr.AttributeExpr{Type: wrapper}
	wrapperHooks := &TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
			if target.Type.Name() != wrapper.TypeName {
				return source, target, nil
			}
			return source, expr.AsObject(target.Type).Attribute("field"), &WrapDirective{
				WrapTarget: true,
				Target:     target,
				FieldName:  "Field",
			}
		},
	}
	wrapperPlan, err := NewTransformPlan(&expr.AttributeExpr{Type: node}, wrapperTarget, "wrapped", wrapperHooks)
	require.NoError(t, err)
	require.Len(t, wrapperPlan.Helpers(), 1)
	wrapperDeclaration := NewExactName(NameFunction, "copyWrappedNode")
	wrapperPackage := newGeneratedPackage("wrapper", "example.com/wrapper", "gen")
	require.NoError(t, wrapperPackage.DeclareName(wrapperDeclaration))
	require.NoError(t, wrapperPlan.BindHelperDeclaration(wrapperPlan.Helpers()[0].ID, wrapperDeclaration))
	require.NoError(t, wrapperPackage.freeze())
	require.NoError(t, wrapperPlan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	wrapperCode, _, err := wrapperPlan.Render("source", "target", true)
	require.NoError(t, err)
	require.Contains(t, wrapperCode, "target := &WrappedNode{}")
	require.Contains(t, wrapperCode, "target.Field")
}

func TestTransformPlanRetainsSameTypeSiblingOccurrences(t *testing.T) {
	plan := siblingTransformPlan(t)
	require.Len(t, plan.Helpers(), 2)
	require.NotEqual(t, plan.Helpers()[0].ID, plan.Helpers()[1].ID)
}

func TestTransformPlanHelperDescriptionsCannotChangeRenderGraph(t *testing.T) {
	plan := siblingTransformPlan(t)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)

	// Helper descriptions are used to choose declarations. Mutating one must
	// never alter the private attributes that Render uses to write those
	// declarations.
	helpers[0].Source.Type = expr.String

	_, definitions := renderTransformPlan(t, plan)
	require.Len(t, definitions, 2)
	require.Equal(t, "*Recursive", definitions[0].ParamTypeRef)
}

func TestTransformPlanRejectsRenderHookMutation(t *testing.T) {
	source := &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}
	target := expr.DupAtt(source)
	plan, err := NewTransformPlan(source, target, "", &TransformHooks{
		TransformArray: func(source, _ *expr.Array, _, _ string, _ bool, _ *TransformAttrs) (string, error) {
			source.ElemType.Type = expr.Int
			return "", nil
		},
	})
	require.NoError(t, err)
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))

	_, _, err = plan.Render("source", "target", true)
	require.EqualError(t, err, "transform render hook changed the retained plan")
}

func TestTransformPlanCachesResultAgainstMutationRetainedByRenderHook(t *testing.T) {
	source := &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}
	target := expr.DupAtt(source)
	var retained *expr.Array
	plan, err := NewTransformPlan(source, target, "", &TransformHooks{
		TransformArray: func(source, _ *expr.Array, _, _ string, _ bool, _ *TransformAttrs) (string, error) {
			retained = source
			return "", nil
		},
	})
	require.NoError(t, err)
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	_, _, err = plan.Render("source", "target", true)
	require.NoError(t, err)

	retained.ElemType.Type = expr.Int
	_, _, err = plan.Render("source", "target", true)
	require.NoError(t, err)
}

func TestTransformPlanCachesRepeatedRender(t *testing.T) {
	source := &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}
	target := expr.DupAtt(source)
	var renders int
	plan, err := NewTransformPlan(source, target, "", &TransformHooks{
		TransformArray: func(_ *expr.Array, _ *expr.Array, _, _ string, _ bool, _ *TransformAttrs) (string, error) {
			renders++
			return fmt.Sprintf("render%d", renders), nil
		},
	})
	require.NoError(t, err)
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))

	code, _, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Equal(t, "render1", code)
	code, _, err = plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Equal(t, "render1", code)
	require.Equal(t, 1, renders)
}

func TestTransformPlanSharesOneDeclarationForEquivalentHelpers(t *testing.T) {
	plan := siblingTransformPlan(t)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	declaration := NewExactName(NameFunction, "transformRecursive")
	pkg := newGeneratedPackage("test", "example.com/test", "gen")
	require.NoError(t, pkg.DeclareName(declaration))
	require.NoError(t, pkg.freeze())

	require.NoError(t, plan.BindHelperDeclaration(helpers[0].ID, declaration))
	require.NoError(t, plan.BindHelperDeclaration(helpers[1].ID, declaration))
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	code, definitions, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	require.Contains(t, code, "target.Left = transformRecursive(source.Left)")
	require.Contains(t, code, "target.Right = transformRecursive(source.Right)")
}

func TestTransformPlanSharesDeclarationAcrossRequiredAndOptionalCalls(t *testing.T) {
	plan := mixedSiblingTransformPlan(t)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	declaration := NewExactName(NameFunction, "transformRecursive")
	pkg := newGeneratedPackage("test", "example.com/test", "gen")
	require.NoError(t, pkg.DeclareName(declaration))
	require.NoError(t, pkg.freeze())

	require.NoError(t, plan.BindHelperDeclaration(helpers[0].ID, declaration))
	require.NoError(t, plan.BindHelperDeclaration(helpers[1].ID, declaration))
	require.NoError(t, plan.BindContexts(
		NewAttributeContext(false, false, true, "", NewNameScope()),
		NewAttributeContext(false, false, true, "", NewNameScope()),
	))
	code, definitions, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	require.Contains(t, code, "target.Left = transformRecursive(source.Left)")
	require.Contains(t, code, "if source.Right != nil {")
	require.Contains(t, code, "target.Right = transformRecursive(source.Right)")
}

func TestTransformPlanGroupsEquivalentHelpersIntoOneDefinition(t *testing.T) {
	plan := mixedSiblingTransformPlan(t)
	definitions := plan.HelperDefinitions()
	require.Len(t, definitions, 1)
	require.NotSame(t, plan.Helpers()[0].Source, definitions[0].Source)
	require.NotSame(t, plan.Helpers()[0].Target, definitions[0].Target)

	declaration := NewExactName(NameFunction, "transformRecursive")
	require.NoError(t, plan.BindHelperDefinition(definitions[0].ID, declaration))
	for _, helper := range plan.Helpers() {
		require.Same(t, declaration, helper.Declaration)
	}
}

func TestTransformHelperRegistryGroupsEquivalentPlans(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(
		first,
		transformTestLayout(t, first.sourceCopier.Original(first.source), GoLayoutPolicy{UseDefault: true}),
		transformTestLayout(t, first.targetCopier.Original(first.target), GoLayoutPolicy{UseDefault: true}),
		transformTestOrderFactory("first").order,
	))
	require.NoError(t, registry.Collect(
		second,
		transformTestLayout(t, second.sourceCopier.Original(second.source), GoLayoutPolicy{UseDefault: true}),
		transformTestLayout(t, second.targetCopier.Original(second.target), GoLayoutPolicy{UseDefault: true}),
		transformTestOrderFactory("second").order,
	))

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	declaration := NewExactName(NameFunction, "transformRecursive")
	require.NoError(t, groups[0].Bind(declaration))
	for _, plan := range []*TransformPlan{first, second} {
		for _, helper := range plan.Helpers() {
			require.Same(t, declaration, helper.Declaration)
		}
	}
}

func TestTransformHelperRegistryGroupsSeparateTypeWrappersForOneDeclaration(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	declaration := NewExactName(NameType, "Recursive")
	generation, err := NewGeneration("example.com/generated", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage("example.com/generated")
	require.NoError(t, err)
	require.NoError(t, pkg.DeclareName(declaration))
	registry := NewTransformHelperRegistry()
	for index, plan := range []*TransformPlan{first, second} {
		require.NoError(t, registry.Collect(
			plan,
			transformTestLayoutWithDeclaration(t, plan.sourceCopier.Original(plan.source), declaration),
			transformTestLayoutWithDeclaration(t, plan.targetCopier.Original(plan.target), declaration),
			transformTestOrderFactory(fmt.Sprintf("plan-%d", index)).order,
		))
	}

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 1)
}

func TestTransformHelperRegistryKeepsSeparateDeclarationsApart(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	generation, err := NewGeneration("example.com/generated", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage("example.com/generated")
	require.NoError(t, err)
	firstDeclaration := NewPreferredName(NameType, "Recursive", ExportedName, testNameOrder{value: "first"})
	secondDeclaration := NewPreferredName(NameType, "Recursive", ExportedName, testNameOrder{value: "second"})
	require.NoError(t, pkg.DeclareName(firstDeclaration))
	require.NoError(t, pkg.DeclareName(secondDeclaration))
	registry := NewTransformHelperRegistry()
	for index, entry := range []struct {
		plan        *TransformPlan
		declaration *NameDeclaration
	}{
		{plan: first, declaration: firstDeclaration},
		{plan: second, declaration: secondDeclaration},
	} {
		require.NoError(t, registry.Collect(
			entry.plan,
			transformTestLayoutWithDeclaration(t, entry.plan.sourceCopier.Original(entry.plan.source), entry.declaration),
			transformTestLayoutWithDeclaration(t, entry.plan.targetCopier.Original(entry.plan.target), entry.declaration),
			transformTestOrderFactory(fmt.Sprintf("plan-%d", index)).order,
		))
	}

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 2)
}

func TestTransformHelperRegistryKeepsGroupSpecificLocations(t *testing.T) {
	plan := siblingTransformPlan(t)
	sourceLayout := transformTestLayout(
		t,
		plan.sourceCopier.Original(plan.source),
		GoLayoutPolicy{UseDefault: true},
	)
	targetLayout := transformTestLayout(
		t,
		plan.targetCopier.Original(plan.target),
		GoLayoutPolicy{UseDefault: true},
	)
	firstDeclaration := NewExactName(NameType, "FirstRecursive")
	secondDeclaration := NewExactName(NameType, "SecondRecursive")
	for _, layout := range []*GoTypePlan{sourceLayout, targetLayout} {
		layout.value.fields[0].declaration = firstDeclaration
		layout.value.fields[1].declaration = secondDeclaration
	}
	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(plan, sourceLayout, targetLayout, transformTestOrderFactory("plan").order))

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.NotEqual(t, groups[0].Definition().Location, groups[1].Definition().Location)
}

func TestTransformHelperRegistryKeepsOnePlanAndLocationTogether(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	firstSource := transformTestLayout(t, first.sourceCopier.Original(first.source), GoLayoutPolicy{UseDefault: true})
	firstTarget := transformTestLayout(t, first.targetCopier.Original(first.target), GoLayoutPolicy{UseDefault: true})
	secondSource := transformTestLayout(t, second.sourceCopier.Original(second.source), GoLayoutPolicy{UseDefault: true})
	secondTarget := transformTestLayout(t, second.targetCopier.Original(second.target), GoLayoutPolicy{UseDefault: true})
	alpha := NewExactName(NameType, "AlphaRecursive")
	beta := NewExactName(NameType, "BetaRecursive")
	for _, layout := range []*GoTypePlan{firstSource, firstTarget} {
		layout.value.fields[0].declaration = alpha
		layout.value.fields[1].declaration = beta
	}
	for _, layout := range []*GoTypePlan{secondSource, secondTarget} {
		layout.value.fields[0].declaration = beta
		layout.value.fields[1].declaration = alpha
	}

	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(first, firstSource, firstTarget, transformTestOrderFactory("first").order))
	require.NoError(t, registry.Collect(second, secondSource, secondTarget, transformTestOrderFactory("second").order))
	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 2)
	firstOrder := groups[0].Order().(testNameOrder)
	secondOrder := groups[1].Order().(testNameOrder)
	require.True(t, strings.HasPrefix(firstOrder.value, "first"))
	require.True(t, strings.HasPrefix(secondOrder.value, "first"))
	require.NotEqual(t, firstOrder, secondOrder)
	require.NotEqual(t, groups[0].Definition().Location, groups[1].Definition().Location)
}

func TestTransformHelperRegistrySharesFieldAndMapValueHelpers(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	container := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "direct", Attribute: &expr.AttributeExpr{Type: recursive}},
		{Name: "mapped", Attribute: &expr.AttributeExpr{Type: &expr.Map{
			KeyType:  &expr.AttributeExpr{Type: expr.String},
			ElemType: &expr.AttributeExpr{Type: recursive},
		}}},
	}}
	plan, err := NewTransformPlan(container, container, "", nil)
	require.NoError(t, err)
	require.Len(t, plan.helpers, 2)
	require.NotEqual(t, plan.helpers[0].location, plan.helpers[1].location)
	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(
		plan,
		transformTestLayout(t, container, GoLayoutPolicy{UseDefault: true}),
		transformTestLayout(t, container, GoLayoutPolicy{UseDefault: true}),
		transformTestOrderFactory("plan").order,
	))

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 1)
}

func TestTransformHelperRegistryFollowsNamedAliasChains(t *testing.T) {
	node := &expr.UserTypeExpr{TypeName: "Node"}
	node.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{
		{Name: "next", Attribute: &expr.AttributeExpr{Type: node}},
	}}
	alias := &expr.UserTypeExpr{
		TypeName:      "Alias",
		AttributeExpr: &expr.AttributeExpr{Type: node},
	}
	container := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "alias", Attribute: &expr.AttributeExpr{Type: alias}},
	}}
	plan, err := NewTransformPlan(container, container, "", nil)
	require.NoError(t, err)
	registry := NewTransformHelperRegistry()

	err = registry.Collect(
		plan,
		transformTestLayout(t, container, GoLayoutPolicy{UseDefault: true}),
		transformTestLayout(t, container, GoLayoutPolicy{UseDefault: true}),
		transformTestOrderFactory("plan").order,
	)
	require.NoError(t, err)
}

func TestTransformHelperRegistryKeepsDifferentPlansSeparate(t *testing.T) {
	tests := []struct {
		name      string
		first     func(*testing.T) *TransformPlan
		second    func(*testing.T) *TransformPlan
		firstRule GoLayoutPolicy
		otherRule GoLayoutPolicy
	}{
		{
			name:      "pointer policy",
			first:     siblingTransformPlan,
			second:    siblingTransformPlan,
			firstRule: GoLayoutPolicy{UseDefault: true},
			otherRule: GoLayoutPolicy{Pointer: true},
		},
		{
			name:   "root default",
			first:  func(t *testing.T) *TransformPlan { return singleTransformHelperPlan(t, "first", nil) },
			second: func(t *testing.T) *TransformPlan { return singleTransformHelperPlan(t, "second", nil) },
		},
		{
			name:   "required collection",
			first:  func(t *testing.T) *TransformPlan { return collectionTransformHelperPlan(t, true) },
			second: func(t *testing.T) *TransformPlan { return collectionTransformHelperPlan(t, false) },
		},
		{
			name:  "custom hook",
			first: func(t *testing.T) *TransformPlan { return singleTransformHelperPlan(t, nil, &TransformHooks{}) },
			second: func(t *testing.T) *TransformPlan {
				return singleTransformHelperPlan(t, nil, &TransformHooks{})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := test.first(t)
			second := test.second(t)
			registry := NewTransformHelperRegistry()
			for index, entry := range []struct {
				plan   *TransformPlan
				policy GoLayoutPolicy
			}{
				{plan: first, policy: test.firstRule},
				{plan: second, policy: test.otherRule},
			} {
				require.NoError(t, registry.Collect(
					entry.plan,
					transformTestLayout(t, entry.plan.sourceCopier.Original(entry.plan.source), entry.policy),
					transformTestLayout(t, entry.plan.targetCopier.Original(entry.plan.target), entry.policy),
					transformTestOrderFactory(fmt.Sprintf("plan-%d", index)).order,
				))
			}
			groups, err := registry.Finalize()
			require.NoError(t, err)
			require.Len(t, groups, 2)
		})
	}
}

func TestTransformHelperRegistryKeepsHookDefinitionsSeparate(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	source := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "left", Attribute: &expr.AttributeExpr{Type: recursive}},
		{Name: "right", Attribute: &expr.AttributeExpr{Type: recursive}},
	}}
	target := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "left", Attribute: &expr.AttributeExpr{Type: recursive}},
		{Name: "right", Attribute: &expr.AttributeExpr{Type: recursive}},
	}}
	plan, err := NewTransformPlan(source, target, "", &TransformHooks{
		SameHelperDefinition: func(_, _, _, _ *expr.AttributeExpr) bool {
			return false
		},
	})
	require.NoError(t, err)
	require.Len(t, plan.definitions, 2)
	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(
		plan,
		transformTestLayout(t, source, GoLayoutPolicy{UseDefault: true}),
		transformTestLayout(t, target, GoLayoutPolicy{UseDefault: true}),
		transformTestOrderFactory("plan").order,
	))

	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 2)
}

// collectionTransformHelperPlan builds a named object whose slice field may be
// required. Required slices emit an empty target slice when the source is nil;
// optional slices leave the target field unset.
func collectionTransformHelperPlan(t *testing.T, required bool) *TransformPlan {
	t.Helper()
	value := &expr.UserTypeExpr{
		TypeName: "CollectionValue",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "items", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}}},
		}},
	}
	if required {
		value.Attribute().Validation = &expr.ValidationExpr{Required: []string{"items"}}
	}
	container := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "value", Attribute: &expr.AttributeExpr{Type: value}},
	}}
	plan, err := NewTransformPlan(container, container, "", nil)
	require.NoError(t, err)
	return plan
}

func TestTransformHelperGroupBindIsAtomic(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	registry := NewTransformHelperRegistry()
	for index, plan := range []*TransformPlan{first, second} {
		require.NoError(t, registry.Collect(
			plan,
			transformTestLayout(t, plan.sourceCopier.Original(plan.source), GoLayoutPolicy{UseDefault: true}),
			transformTestLayout(t, plan.targetCopier.Original(plan.target), GoLayoutPolicy{UseDefault: true}),
			transformTestOrderFactory(fmt.Sprintf("plan-%d", index)).order,
		))
	}
	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	existing := NewExactName(NameFunction, "existingTransform")
	require.NoError(t, second.BindHelperDeclaration(second.Helpers()[0].ID, existing))

	err = groups[0].Bind(NewExactName(NameFunction, "sharedTransform"))
	require.ErrorContains(t, err, "already has a different declaration")
	for _, helper := range first.Helpers() {
		require.Nil(t, helper.Declaration)
	}
}

// singleTransformHelperPlan builds one helper whose root target attribute can
// carry a default and whose hook set is retained by the plan.
func singleTransformHelperPlan(t *testing.T, targetDefault any, hooks *TransformHooks) *TransformPlan {
	t.Helper()
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	sourceChild := &expr.AttributeExpr{Type: recursive}
	targetChild := &expr.AttributeExpr{Type: recursive, DefaultValue: targetDefault}
	source := &expr.AttributeExpr{Type: &expr.Object{{Name: "child", Attribute: sourceChild}}}
	target := &expr.AttributeExpr{Type: &expr.Object{{Name: "child", Attribute: targetChild}}}
	plan, err := NewTransformPlan(source, target, "", hooks)
	require.NoError(t, err)
	return plan
}

// transformTestLayout records a complete Go type before final package names
// are chosen. Equal authored type names use the same fixed declaration so the
// tests compare conversion behavior instead of Goa value addresses.
func transformTestLayout(t *testing.T, attribute *expr.AttributeExpr, policy GoLayoutPolicy) *GoTypePlan {
	t.Helper()
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:            "example.com/generated",
		Policy:           policy,
		RetainNamedValue: true,
		Bind: func(request GoTypeBindingRequest) (GoTypeBinding, error) {
			return GoTypeBinding{
				Owner: "example.com/generated",
				name:  request.Attribute.Type.Name(),
			}, nil
		},
	})
	require.NoError(t, err)
	return layout
}

// transformTestLayoutWithDeclaration builds new type wrappers around one
// package declaration, matching independent planners that resolve the same Go
// type through separate wrapper values.
func transformTestLayoutWithDeclaration(t *testing.T, attribute *expr.AttributeExpr, declaration *NameDeclaration) *GoTypePlan {
	t.Helper()
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:            "example.com/generated",
		Policy:           GoLayoutPolicy{UseDefault: true},
		RetainNamedValue: true,
		Bind: func(GoTypeBindingRequest) (GoTypeBinding, error) {
			return GoTypeBinding{
				Owner: "example.com/generated",
				Type:  &TypeDeclaration{declaration: declaration},
			}, nil
		},
	})
	require.NoError(t, err)
	return layout
}

func TestTransformHelperDefinitionLocationComparesStructuralPaths(t *testing.T) {
	root := TransformHelperDefinitionLocation{}
	emptyField := root.objectField("")
	nulField := root.objectField("\x00")
	letterField := root.objectField("a")
	require.Less(t, emptyField.Compare(nulField), 0)
	require.Less(t, nulField.Compare(letterField), 0)
	require.NotEqual(t, root.objectField("a\x00b"), root.objectField("a").objectField("b"))
	require.NotEqual(t, root.arrayElement(), root.mapKey())
	require.NotEqual(t, root.mapValue(), root.unionBranch(""))
}

func TestTransformPlanDefinitionLocationIgnoresSiblingFieldOrder(t *testing.T) {
	plan := func(reverse bool) *TransformPlan {
		root := RunDSL(t, testdata.TestTypesDSL)
		recursive := root.UserType("Recursive")
		source, target := &expr.Object{}, &expr.Object{}
		add := func(name string) {
			source.Set(name, &expr.AttributeExpr{Type: recursive})
			target.Set(name, &expr.AttributeExpr{Type: recursive})
		}
		if reverse {
			add("right")
			add("left")
		} else {
			add("left")
			add("right")
		}
		planned, err := NewTransformPlan(
			&expr.AttributeExpr{Type: source},
			&expr.AttributeExpr{Type: target},
			"",
			nil,
		)
		require.NoError(t, err)
		return planned
	}

	forward := plan(false).HelperDefinitions()
	reverse := plan(true).HelperDefinitions()
	require.Len(t, forward, 1)
	require.Len(t, reverse, 1)
	require.Equal(t, forward[0].Location, reverse[0].Location)
}

func TestTransformPlanKeepsDifferentHelperBodiesInSeparateDefinitions(t *testing.T) {
	sourceOrigin := transformObjectAttribute("SourceNode", true).Type.(expr.UserType)
	targetOrigin := transformObjectAttribute("TargetNode", true).Type.(expr.UserType)
	sourceLeft := sourceOrigin.Dup(nil)
	sourceRight := sourceOrigin.Dup(nil)
	targetLeft := targetOrigin.Dup(nil)
	targetRight := targetOrigin.Dup(nil)
	sourceLeft.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "left", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})
	sourceRight.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "right", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})
	targetLeft.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "left", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})
	targetRight.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "right", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}})

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: sourceLeft}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: sourceRight}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: targetLeft}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: targetRight}},
		}},
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.HelperDefinitions(), 2)
}

func TestTransformPlanKeepsDifferentAuthoredTypesInSeparateDefinitions(t *testing.T) {
	newType := func(uid string) expr.UserType {
		return &expr.UserTypeExpr{
			TypeName: "Node",
			UID:      uid,
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
		}
	}
	sourceLeft, sourceRight := newType("source-left"), newType("source-right")
	targetLeft, targetRight := newType("target-left"), newType("target-right")
	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: sourceLeft}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: sourceRight}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: targetLeft}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: targetRight}},
		}},
		"",
		&TransformHooks{
			SameHelperDefinition: func(_, _, _, _ *expr.AttributeExpr) bool {
				return true
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, plan.HelperDefinitions(), 2)
}

func TestTransformPlanKeepsHookVisibleRootFactsInSeparateDefinitions(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	sourceLeft := &expr.AttributeExpr{Type: recursive}
	sourceRight := &expr.AttributeExpr{Type: recursive}
	targetLeft := &expr.AttributeExpr{Type: recursive}
	targetRight := &expr.AttributeExpr{Type: recursive}
	targetRight.AddMeta("helper:deref", "value")

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: sourceLeft},
			{Name: "right", Attribute: sourceRight},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: targetLeft},
			{Name: "right", Attribute: targetRight},
		}},
		"",
		&TransformHooks{
			ObjectDeref: func(target *expr.AttributeExpr) (string, bool) {
				if len(target.Meta["helper:deref"]) > 0 {
					return "", true
				}
				return "&", true
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 2)
	require.Len(t, plan.HelperDefinitions(), 2)
}

func TestTransformPlanKeepsDifferentRootDefaultsInSeparateDefinitions(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	targetLeft := &expr.AttributeExpr{Type: recursive}
	targetRight := &expr.AttributeExpr{Type: recursive, DefaultValue: "right"}

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: recursive}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: recursive}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: targetLeft},
			{Name: "right", Attribute: targetRight},
		}},
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.HelperDefinitions(), 2)
}

func TestTransformPlanKeepsDifferentRootValidationInSeparateDefinitions(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	targetLeft := &expr.AttributeExpr{Type: recursive}
	targetRight := &expr.AttributeExpr{
		Type:       recursive,
		Validation: &expr.ValidationExpr{Required: []string{"name"}},
	}

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: recursive}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: recursive}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: targetLeft},
			{Name: "right", Attribute: targetRight},
		}},
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.HelperDefinitions(), 2)
}

func TestTransformPlanLetsHooksApproveOneHelperDefinition(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	recursive := root.UserType("Recursive")
	targetLeft := &expr.AttributeExpr{Type: recursive}
	targetRight := &expr.AttributeExpr{Type: recursive, DefaultValue: "right"}

	plan, err := NewTransformPlan(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: &expr.AttributeExpr{Type: recursive}},
			{Name: "right", Attribute: &expr.AttributeExpr{Type: recursive}},
		}},
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: targetLeft},
			{Name: "right", Attribute: targetRight},
		}},
		"",
		&TransformHooks{
			SameHelperDefinition: func(_, _, _, _ *expr.AttributeExpr) bool {
				return true
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 2)
	require.Len(t, plan.HelperDefinitions(), 1)
}

func TestTransformPlanRejectsForeignHelperDefinition(t *testing.T) {
	first := siblingTransformPlan(t)
	second := siblingTransformPlan(t)
	err := first.BindHelperDefinition(
		second.HelperDefinitions()[0].ID,
		NewExactName(NameFunction, "transformRecursive"),
	)
	require.EqualError(t, err, "transform helper definition does not belong to this plan")
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

func TestTransformPlanUsesFinalTypeLayoutNames(t *testing.T) {
	const owner = "example.com/generated"
	branch := &expr.UserTypeExpr{
		TypeName:      "LegacySince",
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
	}
	choice := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "since", Attribute: &expr.AttributeExpr{Type: branch}},
		},
	}
	attribute := &expr.AttributeExpr{Type: choice}

	generation, err := NewGeneration(owner, nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage(owner)
	require.NoError(t, err)
	choiceName := NewExactName(NameType, "Choice")
	branchName := NewExactName(NameType, "ChoiceBranchSince")
	require.NoError(t, pkg.DeclareName(choiceName))
	require.NoError(t, pkg.DeclareName(branchName))
	require.NoError(t, generation.Freeze())

	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:            owner,
		Policy:           GoLayoutPolicy{UseDefault: true, SumType: true},
		RetainNamedValue: true,
		Bind: func(request GoTypeBindingRequest) (GoTypeBinding, error) {
			switch request.Kind {
			case GoUnion:
				return GoTypeBinding{
					Owner: owner,
					Union: &UnionDeclaration{declaration: choiceName},
				}, nil
			case GoNamed:
				return GoTypeBinding{
					Owner: owner,
					Type:  &TypeDeclaration{declaration: branchName},
				}, nil
			default:
				return GoTypeBinding{}, fmt.Errorf("unexpected type kind %s", request.Kind)
			}
		},
	})
	require.NoError(t, err)

	plan, err := NewTransformPlan(attribute, attribute, "", nil)
	require.NoError(t, err)
	context := NewAttributeContext(false, false, true, "", NewNameScope())
	linked := layout.Link(owner, func(string) string { return "" })
	plannedContext, err := context.WithGoTypeLayout(linked)
	require.NoError(t, err)
	require.NoError(t, plan.BindContexts(plannedContext, plannedContext))
	code, helpers, err := plan.Render("in", "out", false)
	require.NoError(t, err)
	require.Empty(t, helpers)
	require.Contains(t, code, "u.SetSince((ChoiceBranchSince)(obj))")
	require.NotContains(t, code, "LegacySince")
}

func TestAttributeContextRejectsGoTypeLayoutWithDifferentPointerRules(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "example.com/generated",
		Policy: GoLayoutPolicy{Pointer: true, SumType: true},
	})
	require.NoError(t, err)
	context := NewAttributeContext(false, false, true, "", NewNameScope())

	_, err = context.WithGoTypeLayout(layout.Link("example.com/generated", nil))
	require.EqualError(t, err, "attach Go type layout: pointer and default rules do not match the attribute context")
}

func TestAttributeContextUsesRequestedNameForReusedField(t *testing.T) {
	shared := &expr.AttributeExpr{Type: expr.String}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: shared},
		{Name: "second", Attribute: shared},
	}}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "example.com/generated",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	context := NewAttributeContext(false, false, true, "", NewNameScope())
	context, err = context.WithGoTypeLayout(layout.Link("example.com/generated", nil))
	require.NoError(t, err)

	require.Equal(t, "First", context.Scope.Field(shared, "first", true))
	require.Equal(t, "Second", context.Scope.Field(shared, "second", true))
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
				plan, err := NewTransformPlan(source, target, "", nil)
				require.NoError(t, err)
				require.Len(t, plan.Helpers(), pair.helpers)

				code, helpers := renderTransformPlan(t, plan)
				require.Len(t, helpers, pair.helpers)
				if pair.helpers == 0 {
					require.NotContains(t, code, "canonicalHelper")
				} else {
					require.Contains(t, code, "canonicalHelper1")
				}
			})
		}
	}
}

func TestTransformPlanGuardsOptionalSiblingCallsBeforeStrictHelpers(t *testing.T) {
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
	require.NotContains(t, optional.Code, "if v == nil")
	require.Contains(t, code, "target.Left = "+required.Declaration.Name()+"(source.Left)")
	require.Contains(t, code, "if source.Right != nil {")
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
		"",
		nil,
	)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)
	require.NotSame(t, sourceKey, plan.Helpers()[0].Source)
	require.NotSame(t, targetKey, plan.Helpers()[0].Target)
	require.Equal(t, sourceKey.Type.Name(), plan.Helpers()[0].Source.Type.Name())
	require.Equal(t, targetKey.Type.Name(), plan.Helpers()[0].Target.Type.Name())

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
		"",
		nil,
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
		"",
		nil,
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

// renderTransformPlan assigns fixed function declarations and contexts, then
// renders the stored operation and definitions.
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

// compileTransformSource proves that a rendered transform and its stored
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

// newTransformOwnerAttributor creates a test type-name provider and records
// every nested type it enters.
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

func (a *transformOwnerAttributor) ValidatorCall(att *expr.AttributeExpr, view, target, _ string) string {
	name := "Validate" + a.Name(att, "", false, true) + Goify(view, true)
	return fmt.Sprintf("%s(%s)", name, target)
}

func (a *transformOwnerAttributor) Scope() *NameScope {
	return a.scope
}

// Field records the expression identity before delegating the field name.
func (a *transformIdentityAttributor) Field(attribute *expr.AttributeExpr, name string, firstUpper bool) string {
	*a.fields = append(*a.fields, attribute)
	return a.Attributor.Field(attribute, name, firstUpper)
}

// Enter keeps recording identities below the supplied object.
func (a *transformIdentityAttributor) Enter(attribute *expr.AttributeExpr) Attributor {
	return &transformIdentityAttributor{
		Attributor: a.Attributor.Enter(attribute),
		fields:     a.fields,
	}
}

// transformOwnerTestType builds a nested union type assigned to the requested
// generated package.
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

// codegenTypeName returns the Go name used by transformOwnerAttributor for one
// test type.
func codegenTypeName(att *expr.AttributeExpr) string {
	if att.Type.Name() == "object" {
		return "Object"
	}
	return Goify(att.Type.Name(), true)
}
