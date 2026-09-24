// These tests compile and execute conversions of finite recursive collection
// values. They also check that custom renderers keep their existing call sites
// and that helper sharing uses both retained type graphs.
package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

func TestRecursiveCollectionTransforms(t *testing.T) {
	tests := []struct {
		name    string
		shapes  []string
		helpers int
		check   string
	}{
		{
			name: "map", shapes: []string{"map"}, helpers: 1,
			check: `source := Source0{"child": {"leaf": {}}}
target := convert(source)
if target["child"]["leaf"] == nil { t.Fatal("lost empty map") }
source["child"]["extra"] = Source0{}
if _, exists := target["child"]["extra"]; exists { t.Fatal("shared source storage") }
if got := convert(nil); got == nil || len(got) != 0 { t.Fatal("changed root nil conversion") }
deep := Source0{}
for range 256 { deep = Source0{"next": deep} }
result := convert(deep)
for range 256 { result = result["next"] }
if result == nil || len(result) != 0 { t.Fatal("lost finite leaf") }`,
		},
		{
			name: "slice", shapes: []string{"slice"}, helpers: 1,
			check: `source := Source0{{{}, {}}}
target := convert(source)
if len(target) != 1 || len(target[0]) != 2 || target[0][0] == nil { t.Fatal("lost nested slice") }
source[0][0] = Source0{{}}
if len(target[0][0]) != 0 { t.Fatal("shared source storage") }
if got := convert(nil); got == nil || len(got) != 0 { t.Fatal("changed root nil conversion") }`,
		},
		{
			name: "mutual-map-slice", shapes: []string{"map", "slice"}, helpers: 2,
			check: `source := Source0{"group": {Source0{"leaf": {}}}}
target := convert(source)
if len(target["group"]) != 1 || target["group"][0]["leaf"] == nil { t.Fatal("lost mutual graph") }`,
		},
		{
			name: "mutual-slice-map", shapes: []string{"slice", "map"}, helpers: 2,
			check: `source := Source0{Source1{"group": {Source1{}}}}
target := convert(source)
if len(target) != 1 || len(target[0]["group"]) != 1 || target[0]["group"][0] == nil { t.Fatal("lost mutual graph") }`,
		},
		{
			name: "map-object", shapes: []string{"map", "object"}, helpers: 2,
			check: `child := &Source1{Next: Source0{}, Count: 0, Enabled: false}
source := Source0{"first": child, "second": child, "absent": nil}
target := convert(source)
if target["first"].Next == nil || target["first"].Count != 0 || target["first"].Enabled { t.Fatal("changed object") }
if target["absent"] != nil { t.Fatal("changed nil object element") }
if target["first"] == target["second"] { t.Fatal("conversion unexpectedly preserves shared identity") }
source["first"].Count = 7
if target["first"].Count != 0 { t.Fatal("shared source storage") }`,
		},
		{
			name: "slice-object", shapes: []string{"slice", "object"}, helpers: 2,
			check: `source := Source0{nil, &Source1{Next: Source0{}}}
target := convert(source)
if target[0] != nil || target[1].Next == nil { t.Fatal("changed nullable elements") }`,
		},
		{
			name: "object-map", shapes: []string{"object", "map"}, helpers: 2,
			check: `source := &Source0{Next: Source1{"leaf": &Source0{}}}
target := convert(source)
if target.Next["leaf"].Next != nil { t.Fatal("changed optional nil map") }`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, sourceTypes := recursiveCollectionGraph("Source", test.shapes)
			target, targetTypes := recursiveCollectionGraph("Target", test.shapes)
			plan, err := NewTransformPlan(source, target, "", nil)
			require.NoError(t, err)
			require.Len(t, plan.Helpers(), test.helpers)
			code, helpers := renderTransformPlan(t, plan)
			require.Len(t, helpers, test.helpers)
			require.Len(t, plan.HelperDefinitions(), test.helpers)

			// Render verifies that every recorded helper call was consumed.
			// A repeated request must return the same retained result.
			repeated, repeatedHelpers, err := plan.Render("source", "target", true)
			require.NoError(t, err)
			require.Equal(t, code, repeated)
			require.Equal(t, helpers, repeatedHelpers)

			scope := NewNameScope()
			var generated strings.Builder
			generated.WriteString("package fixture\n\ntype (\n")
			for _, named := range append(sourceTypes, targetTypes...) {
				fmt.Fprintf(&generated, "%s %s\n", named.Name(), scope.GoTypeDef(named.Attribute(), false, true))
			}
			generated.WriteString(")\n")
			fmt.Fprintf(&generated, "func convert(source %s) %s {\n%s\nreturn target\n}\n",
				scope.GoTypeRef(source), scope.GoTypeRef(target), code)
			for _, helper := range helpers {
				fmt.Fprintf(&generated, "func %s(v %s) %s {\n%s\nreturn res\n}\n",
					helper.Name, helper.ParamTypeRef, helper.ResultTypeRef, helper.Code)
			}
			runCollectionFixture(t, map[string]string{
				"convert.go": generated.String(),
				"convert_test.go": "package fixture\nimport \"testing\"\nfunc TestConvert(t *testing.T) {\n" +
					test.check + "\n}\n",
			})
		})
	}
}

func TestCollectionHelpersPreserveOccurrences(t *testing.T) {
	for _, shape := range []string{"map", "slice"} {
		t.Run(shape, func(t *testing.T) {
			source, _ := recursiveCollectionGraph("Source", []string{shape})
			target, _ := recursiveCollectionGraph("Target", []string{shape})
			parent := func(child *expr.AttributeExpr) *expr.AttributeExpr {
				return &expr.AttributeExpr{Type: &expr.Object{
					{Name: "required", Attribute: child},
					{Name: "optional", Attribute: child},
				}, Validation: &expr.ValidationExpr{Required: []string{"required"}}}
			}
			plan, err := NewTransformPlan(parent(source), parent(target), "", nil)
			require.NoError(t, err)
			require.Len(t, plan.Helpers(), 2)
			require.True(t, plan.Helpers()[0].Required)
			require.False(t, plan.Helpers()[1].Required)
			require.Len(t, plan.HelperDefinitions(), 1)
			pkg := newGeneratedPackage("fixture", "example.com/fixture", "gen")
			declaration := NewExactName(NameFunction, "sharedCollection")
			require.NoError(t, pkg.DeclareName(declaration))
			require.NoError(t, plan.BindHelperDefinition(plan.HelperDefinitions()[0].ID, declaration))
			require.NoError(t, pkg.freeze())
			context := NewAttributeContext(false, false, true, "", NewNameScope())
			require.NoError(t, plan.BindContexts(context, context))
			code, helpers, err := plan.Render("source", "target", true)
			require.NoError(t, err)
			require.Len(t, helpers, 1)
			require.Equal(t, 2, strings.Count(code, "sharedCollection("))
			require.NotContains(t, helpers[0].Code, "if v == nil")
		})
	}
}

func TestCollectionHelpersKeepExactCopiedPairs(t *testing.T) {
	source, _ := recursiveCollectionGraph("Source", []string{"map"})
	first, _ := recursiveCollectionGraph("Target", []string{"map"})
	second := expr.DupAtt(first)
	// This copy has the same authored origin but a different retained child.
	// The planner must visit both pairs before comparing their equivalent
	// recursive definitions.
	expr.AsMap(first.Type).ElemType = second
	plan, err := NewTransformPlan(source, first, "", nil)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 1)
	require.Same(t, first.Type.(expr.UserType).Origin(), second.Type.(expr.UserType).Origin())

	parent := func(left, right *expr.AttributeExpr) *expr.AttributeExpr {
		return &expr.AttributeExpr{Type: &expr.Object{
			{Name: "left", Attribute: left},
			{Name: "right", Attribute: right},
		}}
	}
	plan, err = NewTransformPlan(parent(source, source), parent(first, second), "", nil)
	require.NoError(t, err)
	require.Len(t, plan.Helpers(), 3)
	require.Len(t, plan.HelperDefinitions(), 1)
}

func TestCollectionHelpersKeepDifferentCopiedLayouts(t *testing.T) {
	for _, shape := range []string{"map", "slice"} {
		t.Run(shape, func(t *testing.T) {
			makeCopies := func(name string) *expr.AttributeExpr {
				inner := &expr.UserTypeExpr{
					TypeName: name, AttributeExpr: &expr.AttributeExpr{},
				}
				leaf := &expr.AttributeExpr{Type: expr.String}
				if shape == "map" {
					inner.Type = &expr.Map{KeyType: leaf, ElemType: leaf}
				} else {
					inner.Type = &expr.Array{ElemType: leaf}
				}
				outer := expr.Dup(inner).(expr.UserType)
				child := &expr.AttributeExpr{Type: inner}
				if shape == "map" {
					expr.AsMap(outer).ElemType = child
				} else {
					expr.AsArray(outer).ElemType = child
				}
				require.Same(t, inner.Origin(), outer.Origin())
				return &expr.AttributeExpr{Type: &expr.Object{
					{Name: "value", Attribute: &expr.AttributeExpr{Type: outer}},
				}}
			}
			plan, err := NewTransformPlan(makeCopies("Source"), makeCopies("Target"), "", nil)
			require.NoError(t, err)
			require.Len(t, plan.Helpers(), 2)
			require.Len(t, plan.HelperDefinitions(), 2)
			_, helpers := renderTransformPlan(t, plan)
			require.Len(t, helpers, 2)
			require.Contains(t, helpers[0].Code, helpers[1].Name+"(")
		})
	}
}

func TestCollectionHelpersUseCanonicalOwnersAndLayouts(t *testing.T) {
	const output = "example.com/fixture"
	generation, err := NewGeneration(output, nil)
	require.NoError(t, err)
	source, sourceTypes := recursiveCollectionGraph("Source", []string{"map", "object"})
	target, targetTypes := recursiveCollectionGraph("Target", []string{"map", "object"})
	sourcePolicy := GoLayoutPolicy{UseDefault: true, SumType: true}
	targetPolicies := []GoLayoutPolicy{sourcePolicy, {Pointer: true, SumType: true}}
	owners := []string{output + "/left", output + "/right"}
	files := make(map[string]string)
	layout := func(owner string, root *expr.AttributeExpr, types []expr.UserType, policy GoLayoutPolicy) *GoTypePlan {
		pkg, claimErr := generation.ClaimPackage(owner)
		require.NoError(t, claimErr)
		declarations := make(map[expr.UserType]*TypeDeclaration)
		for index, named := range types {
			name := NewExactName(NameType, fmt.Sprintf("Value%d", index))
			require.NoError(t, pkg.DeclareName(name))
			declarations[named.Origin()] = &TypeDeclaration{declaration: name}
		}
		bind := func(request GoTypeBindingRequest) (GoTypeBinding, error) {
			return GoTypeBinding{
				Owner: owner,
				Type:  declarations[request.Attribute.Type.(expr.UserType).Origin()],
			}, nil
		}
		planned, planErr := PlanGoType(root, GoTypePlanOptions{
			Owner: owner, Policy: policy, RetainNamedValue: true, Bind: bind,
		})
		require.NoError(t, planErr)
		return planned
	}
	sourceLayout := layout(output+"/source", source, sourceTypes, sourcePolicy)
	targetLayouts := make([]*GoTypePlan, len(owners))
	plans := make([]*TransformPlan, len(owners))
	registry := NewTransformHelperRegistry()
	for index, owner := range owners {
		targetLayouts[index] = layout(owner, target, targetTypes, targetPolicies[index])
		plans[index], err = NewTransformPlan(source, target, "", nil)
		require.NoError(t, err)
		require.NoError(t, registry.Collect(plans[index], sourceLayout, targetLayouts[index],
			transformTestOrderFactory(owner).order))
	}
	groups, err := registry.Finalize()
	require.NoError(t, err)
	require.Len(t, groups, 4)
	pkg, err := generation.ClaimPackage(output)
	require.NoError(t, err)
	for index, group := range groups {
		name := NewExactName(NameFunction, fmt.Sprintf("convertHelper%d", index))
		require.NoError(t, pkg.DeclareName(name))
		require.NoError(t, group.Bind(name))
	}
	require.NoError(t, generation.Freeze())
	qualifier := filepath.Base
	allLayouts := append([]*GoTypePlan{sourceLayout}, targetLayouts...)
	for _, root := range allLayouts {
		var definitions strings.Builder
		fmt.Fprintf(&definitions, "package %s\n\ntype (\n", qualifier(root.owner))
		// This graph contains the named map followed by its named object value.
		for _, named := range []*GoTypePlan{root, root.value.element} {
			linked := named.Link(root.owner, qualifier)
			fmt.Fprintf(&definitions, "%s %s\n", linked.Name(), linked.Enter(named.value).Def())
		}
		definitions.WriteString(")\n")
		files[qualifier(root.owner)+"/types.go"] = definitions.String()
	}
	var generated strings.Builder
	generated.WriteString("package fixture\nimport (\n\"example.com/fixture/source\"\n\"example.com/fixture/left\"\n\"example.com/fixture/right\"\n)\n")
	for index, plan := range plans {
		sc, contextErr := NewAttributeContext(false, false, true, "", NewNameScope()).
			WithGoTypeLayout(sourceLayout.Link(output, qualifier))
		require.NoError(t, contextErr)
		policy := targetPolicies[index]
		tc, contextErr := NewAttributeContext(policy.Pointer, false, policy.UseDefault, "", NewNameScope()).
			WithGoTypeLayout(targetLayouts[index].Link(output, qualifier))
		require.NoError(t, contextErr)
		require.NoError(t, plan.BindContexts(sc, tc))
		code, helpers, renderErr := plan.Render("source", "target", true)
		require.NoError(t, renderErr)
		fmt.Fprintf(&generated, "func convert%d(source %s) %s {\n%s\nreturn target\n}\n",
			index, sourceLayout.Link(output, qualifier).Ref(), targetLayouts[index].Link(output, qualifier).Ref(), code)
		for _, helper := range helpers {
			fmt.Fprintf(&generated, "func %s(v %s) %s {\n%s\nreturn res\n}\n",
				helper.Name, helper.ParamTypeRef, helper.ResultTypeRef, helper.Code)
		}
	}
	files["convert.go"] = generated.String()
	files["convert_test.go"] = `package fixture
import ("testing"; "example.com/fixture/source")
func TestOwners(t *testing.T) {
	value := source.Value0{"item": &source.Value1{Count: 0, Enabled: false, Next: source.Value0{}}}
	left, right := convert0(value), convert1(value)
	if left["item"].Count != 0 || left["item"].Enabled { t.Fatal("changed left values") }
	if right["item"].Count == nil || *right["item"].Count != 0 || right["item"].Enabled == nil || *right["item"].Enabled { t.Fatal("lost right pointers") }
	if left["item"].Next == nil || right["item"].Next == nil { t.Fatal("lost collection") }
}`
	runCollectionFixture(t, files)
}

func TestCustomCollectionsKeepInlineCollectionElements(t *testing.T) {
	for _, outer := range []string{"map", "slice"} {
		for _, inner := range []string{"map", "slice"} {
			for _, inline := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s-%s-inline-%t", outer, inner, inline), func(t *testing.T) {
					makeChild := func(name string) *expr.AttributeExpr {
						child := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
							TypeName: name, AttributeExpr: &expr.AttributeExpr{Type: expr.String},
						}}
						if inner == "map" {
							child.Type.(*expr.UserTypeExpr).Type = &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.String}}
						} else {
							child.Type.(*expr.UserTypeExpr).Type = &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}
						}
						return child
					}
					wrap := func(child *expr.AttributeExpr) *expr.AttributeExpr {
						if outer == "map" {
							return &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: child}}
						}
						return &expr.AttributeExpr{Type: &expr.Array{ElemType: child}}
					}
					source, target := wrap(makeChild("Source")), wrap(makeChild("Target"))
					hooks := &TransformHooks{InlineCompositeElems: inline}
					if outer == "map" {
						hooks.TransformMap = func(source, target *expr.Map, sourceVar, targetVar string, newVar bool, attrs *TransformAttrs) (string, error) {
							return TransformAttribute(source.ElemType, target.ElemType, sourceVar+"[0]", targetVar+"[0]", newVar, attrs)
						}
					} else {
						hooks.TransformArray = func(source, target *expr.Array, sourceVar, targetVar string, newVar bool, attrs *TransformAttrs) (string, error) {
							return TransformAttribute(source.ElemType, target.ElemType, sourceVar+"[0]", targetVar+"[0]", newVar, attrs)
						}
					}
					plan, err := NewTransformPlan(source, target, "", hooks)
					require.NoError(t, err)
					require.Empty(t, plan.Helpers())
					code, helpers := renderTransformPlan(t, plan)
					require.NotEmpty(t, code)
					require.Empty(t, helpers)
					testutil.AssertString(t, filepath.Join("testdata", "golden",
						fmt.Sprintf("custom_collection_%s_%s_%t.go.golden", outer, inner, inline)), code)
				})
			}
		}
	}
}

func TestMapDepthRecursiveCollections(t *testing.T) {
	for _, test := range []struct {
		name   string
		shapes []string
		want   int
	}{
		{"map", []string{"map"}, 1},
		{"slice", []string{"slice"}, 0},
		{"mixed", []string{"map", "slice"}, 1},
		{"two-maps", []string{"map", "map"}, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, _ := recursiveCollectionGraph("Value", test.shapes)
			require.Equal(t, test.want, MapDepth(&expr.Map{ElemType: root}))
		})
	}

	leaf := &expr.AttributeExpr{Type: expr.String}
	named := &expr.UserTypeExpr{TypeName: "Map", AttributeExpr: &expr.AttributeExpr{
		Type: &expr.Map{KeyType: leaf, ElemType: leaf},
	}}
	copied := expr.Dup(named).(expr.UserType)
	expr.AsMap(named.Type).ElemType = &expr.AttributeExpr{Type: copied}
	require.Same(t, named.Origin(), copied.Origin())
	require.Equal(t, 2, MapDepth(&expr.Map{ElemType: &expr.AttributeExpr{Type: named}}))

	// The longer sibling path must still visit a shared collection again.
	shared := &expr.AttributeExpr{Type: named}
	wrapped := &expr.AttributeExpr{Type: &expr.Map{KeyType: leaf, ElemType: shared}}
	object := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "short", Attribute: shared},
		{Name: "long", Attribute: wrapped},
	}}
	require.Equal(t, 3, MapDepth(&expr.Map{ElemType: object}))
}

// recursiveCollectionGraph builds a closed type cycle from named collections
// and objects. Values used by the compiled tests terminate at empty collections
// or nil optional fields; the schema itself remains recursive.
func recursiveCollectionGraph(prefix string, shapes []string) (*expr.AttributeExpr, []expr.UserType) {
	types := make([]expr.UserType, len(shapes))
	for index := range shapes {
		types[index] = &expr.UserTypeExpr{TypeName: fmt.Sprintf("%s%d", prefix, index), AttributeExpr: &expr.AttributeExpr{}}
	}
	for index, shape := range shapes {
		child := &expr.AttributeExpr{Type: types[(index+1)%len(types)]}
		switch shape {
		case "map":
			types[index].Attribute().Type = &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: child}
		case "slice":
			types[index].Attribute().Type = &expr.Array{ElemType: child}
		case "object":
			types[index].Attribute().Type = &expr.Object{
				{Name: "next", Attribute: child},
				{Name: "count", Attribute: &expr.AttributeExpr{Type: expr.Int}},
				{Name: "enabled", Attribute: &expr.AttributeExpr{Type: expr.Boolean}},
			}
			types[index].Attribute().Validation = &expr.ValidationExpr{Required: []string{"count", "enabled"}}
		default:
			panic("unknown test collection shape")
		}
	}
	return &expr.AttributeExpr{Type: types[0]}, types
}

// runCollectionFixture compiles the rendered functions and executes assertions
// against finite values in an isolated module without external dependencies.
func runCollectionFixture(t *testing.T, files map[string]string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/fixture\n\ngo 1.26.0\n"), 0o600))
	for name, source := range files {
		filename := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o750))
		require.NoError(t, os.WriteFile(filename, []byte(source), 0o600))
	}
	command := exec.Command("go", "test", "-count=1", "-timeout=30s", "./...")
	command.Dir = dir
	output, err := command.CombinedOutput()
	require.NoError(t, err, "generated fixture:\n%s\nsources:\n%v", output, files)
}
