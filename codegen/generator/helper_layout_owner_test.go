// These tests preserve the package selected by each enclosing conversion when
// one evaluated inline union occurs beneath two different generated owners.
package generator

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestTransformHelperLayoutsPreserveEnclosingOwners(t *testing.T) {
	fixture := sharedHelperOwnerRoots(t)
	plan := mustTestPlan(t, "generated.local/gen", fixture.roots, planServiceData)
	attribute := fixture.exchange.Payload
	layout, err := plan.Service(fixture.right).MethodTypeLayout(fixture.exchange, attribute)
	require.NoError(t, err)
	transform, err := codegen.NewTransformPlan(attribute, attribute, "", nil)
	require.NoError(t, err)
	branch := expr.AsUnion(fixture.choices[0].Type).Values[0].Attribute.Type.(expr.UserType)
	selected := make(map[string]*codegen.TypeDeclaration, 2)
	for _, helper := range transform.Helpers() {
		named, ok := helper.Source.Type.(expr.UserType)
		if !ok || named.Origin() != branch.Origin() {
			continue
		}
		source, target, err := transform.HelperLayouts(helper.ID, layout, layout)
		require.NoError(t, err)
		require.Same(t, source, target)
		require.NotNil(t, source.TypeDeclaration())
		selected[source.Owner()] = source.TypeDeclaration()
		againSource, againTarget, err := transform.HelperLayouts(helper.ID, layout, layout)
		require.NoError(t, err)
		require.Same(t, source, againSource)
		require.Same(t, target, againTarget)
	}
	require.Len(t, selected, 2, "one shared branch needs a distinct selected declaration under each parent")
	for index, owner := range []string{"lefttypes", "righttypes"} {
		pkg := plan.Generation().Package("generated.local/gen/" + owner)
		declaration, err := pkg.UnionBranchType(fixture.choices[index], "record")
		require.NoError(t, err)
		require.Same(t, declaration, selected[pkg.ImportPath()])
	}
	require.NotSame(t, selected["generated.local/gen/lefttypes"], selected["generated.local/gen/righttypes"])
}

// The normal helper render path must also enter the resolver that owns nested
// union constructors. The hook observes that context, then uses the normal union
// writer; it does not change the conversion shape or select a package itself.
func TestTransformHelperConstructorsPreserveEnclosingOwners(t *testing.T) {
	fixture := sharedHelperOwnerRoots(t)
	branch := expr.AsUnion(fixture.choices[0].Type).Values[0].Attribute
	nested := expr.AsObject(branch.Type).Attribute("nested")
	defaults := make(map[string]string, 2)
	var transform *codegen.TransformPlan
	var layout *codegen.GoTypePlan
	var output *codegen.GeneratedPackage
	plan := mustTestPlan(t, "generated.local/gen", fixture.roots, planServiceData, func(plan *Plan) error {
		var err error
		layout, err = plan.Service(fixture.right).MethodTypeLayout(fixture.exchange, fixture.exchange.Payload)
		if err != nil {
			return err
		}
		output = plan.Generation().Package("generated.local/gen/right")
		hooks := &codegen.TransformHooks{
			PlanUnionHelpers: func(source, target *expr.AttributeExpr, record func(*expr.AttributeExpr, *expr.AttributeExpr)) {
				sourceUnion, targetUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
				for index, branch := range sourceUnion.Values {
					if _, named := branch.Attribute.Type.(expr.UserType); named && expr.IsObject(branch.Attribute.Type) {
						record(branch.Attribute, targetUnion.Values[index].Attribute)
					}
				}
			},
			TransformUnion: func(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, _, _ *expr.AttributeExpr, attrs *codegen.TransformAttrs) (string, error) {
				if target.AuthoredAttribute() == nested.AuthoredAttribute() {
					resolver, ok := attrs.TargetCtx.Scope.(interface {
						GoTypeLayout(*expr.AttributeExpr, codegen.GoLayoutPolicy) (codegen.LinkedGoType, error)
						UnionConstructor(*expr.AttributeExpr, string) (string, error)
					})
					require.True(t, ok)
					selected, err := resolver.GoTypeLayout(target, attrs.TargetCtx.LayoutPolicy())
					if err != nil {
						return "", err
					}
					value, err := codegen.RenderGoValue(target, map[string]any{"type": "text", "value": "fallback"}, selected, false, resolver.UnionConstructor, "selectedDefault")
					if err != nil {
						return "", err
					}
					defaults[attrs.TargetCtx.Pkg(target)] = value.Expression
				}
				ordinary := *attrs
				ordinaryHooks := *attrs.Hooks
				ordinaryHooks.TransformUnion = nil
				ordinary.Hooks = &ordinaryHooks
				return codegen.TransformAttribute(source, target, sourceVar, targetVar, newVar, &ordinary)
			},
		}
		transform, err = codegen.NewTransformPlan(fixture.exchange.Payload, fixture.exchange.Payload, "", hooks)
		if err != nil {
			return err
		}
		for _, helper := range transform.Helpers() {
			declaration := codegen.NewExactName(codegen.NameFunction, fmt.Sprintf("copyOwner%d", helper.Occurrence))
			if err := output.DeclareName(declaration); err != nil {
				return err
			}
			if err := transform.BindHelperDeclaration(helper.ID, declaration); err != nil {
				return err
			}
		}
		return nil
	})
	context := &codegen.AttributeContext{
		UseDefault: true,
		Scope:      plan.Service(fixture.right).Services().ServiceAttributor("right", output.ImportPath()),
	}
	context, err := context.WithGoTypeLayout(layout.Link(output.ImportPath(), output.ImportName))
	require.NoError(t, err)
	require.NoError(t, transform.BindContexts(context, context))
	_, helpers, err := transform.Render("input", "result", true)
	require.NoError(t, err)
	require.NotEmpty(t, helpers)
	require.Len(t, defaults, 2)
	for index, owner := range []string{"lefttypes", "righttypes"} {
		pkg := plan.Generation().Package("generated.local/gen/" + owner)
		record := expr.AsUnion(fixture.choices[index].Type).Values[0].Attribute
		constructor, err := pkg.UnionBranch(expr.AsObject(record.Type).Attribute("nested"), "text")
		require.NoError(t, err)
		qualifier := output.ImportName(pkg.ImportPath())
		require.Contains(t, defaults[qualifier], qualifier+"."+constructor.Constructor()+"(")
	}
}

func TestHTTPSharedBranchOwnersCompileAndRoundTrip(t *testing.T) {
	for _, order := range []string{"forward", "reverse"} {
		t.Run(order, func(t *testing.T) {
			fixture := sharedHelperOwnerRoots(t)
			if order == "reverse" {
				fixture.roots[0], fixture.roots[1] = fixture.roots[1], fixture.roots[0]
			}
			plan := mustTestPlan(t, "generated.local/gen", fixture.roots, planServiceData, planTransportData)
			files, err := testServiceFiles(plan)
			require.NoError(t, err)
			transports, err := testTransportFiles(plan)
			require.NoError(t, err)
			files, err = mergeFilesByPath(append(files, transports...))
			require.NoError(t, err)
			directory := t.TempDir()
			writeGeneratedModule(t, directory, "generated.local")
			for _, file := range files {
				_, err := file.Render(directory)
				require.NoError(t, err)
			}

			// Request and response helpers must name and allocate the declaration
			// selected beneath each parent, even when the branch origin is shared.
			for _, side := range []string{"server", "client"} {
				source := generatedTreeSource(t, filepath.Join(directory, "gen", "http", "right", side))
				for index, owner := range []string{"lefttypes", "righttypes"} {
					pkg := plan.Generation().Package("generated.local/gen/" + owner)
					record, err := pkg.UnionBranchType(fixture.choices[index], "record")
					require.NoError(t, err)
					require.Contains(t, source, "*"+owner+"."+record.Name())
					require.Contains(t, source, "&"+owner+"."+record.Name()+"{")
				}
			}
			writeGeneratedContractTest(t, directory, filepath.Join("gen", "http", "right", "server"), sharedHelperOwnerRoundTrip)
			runGeneratedTests(t, directory)
		})
	}
}

// These controls distinguish generated branch ownership from authored child
// placement and from ordinary copies of a parent that already has one owner.
func TestHTTPUnionOwnerControlsCompile(t *testing.T) {
	for _, scenario := range []string{"service local", "equal names", "explicit child", "shared parent copy"} {
		t.Run(scenario, func(t *testing.T) {
			transport := func(name string, input expr.UserType) {
				dsl.Service(name, func() {
					dsl.Method("Exchange", func() {
						dsl.Payload(input)
						dsl.Result(input)
						dsl.HTTP(func() { dsl.POST("/" + name) })
					})
				})
			}
			first := codegen.RunDSL(t, func() {
				dsl.API("first", func() {})
				var child expr.UserType
				if scenario == "explicit child" {
					child = dsl.Type("LocatedRecord", func() {
						dsl.Meta("struct:pkg:path", "children")
						dsl.Attribute("text", dsl.String)
						dsl.Required("text")
					})
				}
				input := dsl.Type("FirstInput", func() {
					if scenario != "service local" {
						dsl.Meta("struct:pkg:path", "firsttypes")
					}
					dsl.OneOf("choice", func() {
						dsl.TypeName("OwnerChoice")
						if child != nil {
							dsl.Attribute("record", child)
						} else {
							dsl.Attribute("record", func() {
								dsl.Attribute("text", dsl.String)
								dsl.Required("text")
							})
						}
					})
					dsl.Required("choice")
				})
				transport("first", input)
			})
			input := first.UserType("FirstInput")
			choice := expr.AsObject(input).Attribute("choice")
			second := codegen.RunDSL(t, func() {
				dsl.API("second", func() {})
				var selected expr.UserType
				if scenario == "shared parent copy" {
					selected = expr.DupAtt(&expr.AttributeExpr{Type: input}).Type.(expr.UserType)
				} else {
					selected = dsl.Type("SecondInput", func() {
						if scenario != "service local" {
							dsl.Meta("struct:pkg:path", "secondtypes")
						}
						if scenario == "equal names" {
							dsl.OneOf("choice", func() {
								dsl.TypeName("OwnerChoice")
								dsl.Attribute("record", func() {
									dsl.Attribute("text", dsl.Int)
									dsl.Required("text")
								})
							})
						} else {
							dsl.Attribute("choice", choice.Type)
						}
						dsl.Required("choice")
					})
				}
				transport("second", selected)
			})
			plan := mustTestPlan(t, "generated.local/gen", []eval.Root{first, second}, planServiceData, planTransportData)
			files, err := testServiceFiles(plan)
			require.NoError(t, err)
			transports, err := testTransportFiles(plan)
			require.NoError(t, err)
			files, err = mergeFilesByPath(append(files, transports...))
			require.NoError(t, err)
			directory := t.TempDir()
			writeGeneratedModule(t, directory, "generated.local")
			for _, file := range files {
				_, err := file.Render(directory)
				require.NoError(t, err)
			}
			runGeneratedTests(t, directory)
		})
	}
}

type helperOwnerRoots struct {
	roots    []eval.Root
	right    *expr.RootExpr
	exchange *expr.MethodExpr
	choices  [2]*expr.AttributeExpr
}

// sharedHelperOwnerRoots reuses an evaluated union, including its generated
// record branch, without giving that branch a global authored placement.
func sharedHelperOwnerRoots(t *testing.T) helperOwnerRoots {
	t.Helper()
	first := codegen.RunDSL(t, func() {
		dsl.API("left", func() {})
		input := dsl.Type("LeftInput", func() {
			dsl.Meta("struct:pkg:path", "lefttypes")
			dsl.OneOf("choice", func() {
				dsl.TypeName("SharedChoice")
				dsl.Attribute("record", func() {
					dsl.OneOf("nested", func() {
						dsl.TypeName("NestedChoice")
						dsl.Attribute("text", dsl.String, func() { dsl.MinLength(2) })
					})
					dsl.Attribute("label", dsl.String, func() { dsl.Default("fallback") })
					dsl.Attribute("items", dsl.ArrayOf(dsl.String))
					dsl.Required("items")
				})
			})
			dsl.Required("choice")
		})
		dsl.Service("left", func() {
			dsl.Method("Exchange", func() {
				dsl.Payload(input)
				dsl.Result(input)
				dsl.HTTP(func() { dsl.POST("/left") })
			})
		})
	})
	left := first.UserType("LeftInput")
	leftChoice := expr.AsObject(left).Attribute("choice")
	var exchange *expr.MethodExpr
	second := codegen.RunDSL(t, func() {
		dsl.API("right", func() {})
		right := dsl.Type("RightInput", func() {
			dsl.Meta("struct:pkg:path", "righttypes")
			dsl.Extend(left)
			dsl.Required("choice")
		})
		node := dsl.Type("OwnerNode", func() {
			dsl.Attribute("value", dsl.String, func() { dsl.MinLength(2) })
			dsl.Attribute("next", "OwnerNode")
			dsl.Attribute("labels", dsl.MapOf(dsl.String, dsl.String))
			dsl.Required("value")
		})
		pair := dsl.Type("Pair", func() {
			dsl.Attribute("left", left)
			dsl.Attribute("right", right)
			dsl.Attribute("node", node)
			dsl.Required("left", "right")
		})
		viewed := dsl.ResultType("application/vnd.owner-node", func() {
			dsl.TypeName("ViewedNode")
			dsl.Attribute("node", node)
			dsl.View("default", func() { dsl.Attribute("node") })
		})
		dsl.Service("right", func() {
			exchange = dsl.Method("Exchange", func() {
				dsl.Payload(pair)
				dsl.Result(pair)
				dsl.HTTP(func() { dsl.POST("/right") })
			})
			dsl.Method("Single", func() {
				dsl.Payload(right)
				dsl.Result(right)
				dsl.HTTP(func() { dsl.POST("/single") })
			})
			dsl.Method("ExchangeJSON", func() {
				dsl.Payload(pair)
				dsl.Result(pair)
				dsl.JSONRPC(func() {})
			})
			dsl.Method("Viewed", func() {
				dsl.Result(viewed)
				dsl.HTTP(func() { dsl.GET("/viewed") })
			})
		})
	})
	rightChoice := expr.AsObject(second.UserType("RightInput")).Attribute("choice")
	require.Same(t, leftChoice.AuthoredAttribute(), rightChoice.AuthoredAttribute(),
		"both parents must reuse one authored union")
	leftBranch := expr.AsUnion(leftChoice.Type).Values[0].Attribute.Type.(expr.UserType)
	rightBranch := expr.AsUnion(rightChoice.Type).Values[0].Attribute.Type.(expr.UserType)
	require.Same(t, leftBranch.Origin(), rightBranch.Origin())
	return helperOwnerRoots{
		roots:    []eval.Root{first, second},
		right:    second,
		exchange: exchange,
		choices:  [2]*expr.AttributeExpr{leftChoice, rightChoice},
	}
}

const sharedHelperOwnerRoundTrip = `package server

import (
	"encoding/json"
	"reflect"
	"testing"

	genclient "generated.local/gen/http/right/client"
)

func TestSharedOwnerExchange(t *testing.T) {
	for _, input := range []string{
		"{\"left\":{\"choice\":{\"type\":\"record\",\"value\":{\"items\":[]}}},\"right\":{\"choice\":{\"type\":\"record\",\"value\":{\"items\":[]}}}}",
		"{\"left\":{\"choice\":{\"type\":\"record\",\"value\":{\"nested\":{\"type\":\"text\",\"value\":\"left\"},\"items\":[\"one\"]}}},\"right\":{\"choice\":{\"type\":\"record\",\"value\":{\"nested\":{\"type\":\"text\",\"value\":\"right\"},\"items\":[\"two\"]}}},\"node\":{\"value\":\"root\",\"labels\":{\"key\":\"value\"},\"next\":{\"value\":\"child\"}}}",
	} {
		var body ExchangeRequestBody
		if err := json.Unmarshal([]byte(input), &body); err != nil { t.Fatal(err) }
		if err := ValidateExchangeRequestBody(&body); err != nil { t.Fatal(err) }
		payload := NewExchangePair(&body)
		left, ok := payload.Left.Choice.AsRecord()
		if !ok { t.Fatal("left branch changed") }
		right, ok := payload.Right.Choice.AsRecord()
		if !ok { t.Fatal("right branch changed") }
		if left.Items == nil || right.Items == nil { t.Fatal("required empty collection was lost") }
		leftText, leftOK := left.Nested.AsText()
		rightText, rightOK := right.Nested.AsText()
		if left.Label != "fallback" || right.Label != "fallback" { t.Fatal("defaults were lost") }
		if payload.Node == nil {
			if leftOK || rightOK || left.Nested.Kind() != "" || right.Nested.Kind() != "" { t.Fatal("optional absence changed") }
		} else {
			if !leftOK || !rightOK || leftText != "left" || rightText != "right" { t.Fatal("nested values crossed owners") }
			if payload.Node.Next == nil || payload.Node.Next.Value != "child" || payload.Node.Labels["key"] != "value" { t.Fatal("recursive graph changed") }
		}
		data, err := json.Marshal(genclient.NewExchangeRequestBody(payload))
		if err != nil { t.Fatal(err) }
		var accepted ExchangeRequestBody
		if err := json.Unmarshal(data, &accepted); err != nil { t.Fatal(err) }
		if err := ValidateExchangeRequestBody(&accepted); err != nil { t.Fatal(err) }
		data, err = json.Marshal(NewExchangeResponseBody(NewExchangePair(&accepted)))
		if err != nil { t.Fatal(err) }
		var received genclient.ExchangeResponseBody
		if err := json.Unmarshal(data, &received); err != nil { t.Fatal(err) }
		if err := genclient.ValidateExchangeResponseBody(&received); err != nil { t.Fatal(err) }
		result := genclient.NewExchangePairOK(&received)
		if !reflect.DeepEqual(payload, result) { t.Fatalf("owner graph changed: %#v", result) }
	}
}
`
