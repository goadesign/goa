// This file verifies that validation planning and NeedsValidation honor explicit
// result view selections on nested result types, including multi-level and
// recursive views.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestNeedsValidationSelectedResultViews checks that NeedsValidation inspects
// only the fields included in a nested result type's selected view.
func TestNeedsValidationSelectedResultViews(t *testing.T) {
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}

	for _, tc := range []struct {
		name string
		dsl  func() *expr.AttributeExpr
		want bool
	}{
		{
			name: "default view unvalidated while extended view validated",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View("default") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: false,
		},
		{
			name: "extended view validated inside array",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View("extended") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
		{
			name: "default view validated while extended view omits validated field",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String, func() { dsl.MinLength(2) })
							dsl.Attribute("summary", dsl.String)
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("summary", func() { dsl.Attribute("summary") })
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("by_key", dsl.MapOf(dsl.String, label, func() {
							dsl.Elem(func() { dsl.View("summary") })
						}))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: false,
		},
		{
			name: "three-level cascaded view reaches leaf validation",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					label := dsl.ResultType("application/vnd.test.label", func() {
						dsl.TypeName("Label")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
						})
						dsl.View("default", func() { dsl.Attribute("id") })
						dsl.View("extended", func() {
							dsl.Attribute("id")
							dsl.Attribute("name")
						})
					})
					item := dsl.ResultType("application/vnd.test.item", func() {
						dsl.TypeName("Item")
						dsl.Attributes(func() {
							dsl.Attribute("labels", dsl.ArrayOf(label))
						})
						dsl.View("default", func() { dsl.Attribute("labels") })
						dsl.View("extended", func() {
							dsl.Attribute("labels", func() { dsl.View("extended") })
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("items", dsl.ArrayOf(item, func() { dsl.View("extended") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
		{
			name: "mutually recursive result views terminate and find validation",
			dsl: func() *expr.AttributeExpr {
				var target *expr.AttributeExpr
				RunDSL(t, func() {
					node := dsl.ResultType("application/vnd.test.node", func() {
						dsl.TypeName("Node")
						dsl.Attributes(func() {
							dsl.Attribute("id", dsl.String)
							dsl.Attribute("title", dsl.String, func() { dsl.MinLength(3) })
							dsl.Attribute("children", dsl.ArrayOf("Node", func() {
								dsl.View("tiny")
							}))
						})
						dsl.View("tiny", func() {
							dsl.Attribute("id")
							dsl.Attribute("children", func() { dsl.View("default") })
						})
						dsl.View("default", func() {
							dsl.Attribute("id")
							dsl.Attribute("title")
							dsl.Attribute("children")
						})
					})
					holder := dsl.Type("Holder", func() {
						dsl.Attribute("nodes", dsl.ArrayOf(node, func() { dsl.View("tiny") }))
					})
					dsl.Service("Test", func() {
						dsl.Method("Get", func() { dsl.Payload(holder) })
					})
					target = holder.Attribute()
				})
				return target
			},
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attribute := tc.dsl()
			assert.Equal(t, tc.want, NeedsValidation(attribute, policy))
		})
	}
}

// TestValidationPlanSelectedResultViews verifies that NewValidationPlan binds a
// nested result validator only when the selected view contains checks, passing
// the selected non-default view name to the binder.
func TestValidationPlanSelectedResultViews(t *testing.T) {
	const owner = "generated.local/gen/service/views"
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}

	for _, tc := range []struct {
		name         string
		selectedView string
		wantBound    bool
		wantView     string
	}{
		{
			name:         "default view without checks skips validator binding",
			selectedView: expr.DefaultView,
			wantBound:    false,
		},
		{
			name:         "extended view with checks binds extended validator",
			selectedView: "extended",
			wantBound:    true,
			wantView:     "extended",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var label *expr.ResultTypeExpr
			var holder expr.UserType
			RunDSL(t, func() {
				label = dsl.ResultType("application/vnd.test.label", func() {
					dsl.TypeName("Label")
					dsl.Attributes(func() {
						dsl.Attribute("id", dsl.String)
						dsl.Attribute("name", dsl.String, func() { dsl.MinLength(2) })
					})
					dsl.View("default", func() { dsl.Attribute("id") })
					dsl.View("extended", func() {
						dsl.Attribute("id")
						dsl.Attribute("name")
					})
				})
				holder = dsl.Type("Holder", func() {
					dsl.Attribute("labels", dsl.ArrayOf(label, func() { dsl.View(tc.selectedView) }))
				})
				dsl.Service("Test", func() {
					dsl.Method("Get", func() { dsl.Result(label); dsl.Payload(holder) })
				})
			})

			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			labelDecl := declareGoTypeTestUserType(t, generation, owner, label)
			validator := NewExactName(NameFunction, "ValidateLabelViewExtended")
			require.NoError(t, generation.Package(owner).DeclareName(validator))

			layout, err := PlanGoType(holder.Attribute(), GoTypePlanOptions{
				Owner:  owner,
				Policy: policy,
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					label: {Owner: owner, Type: labelDecl},
				}),
			})
			require.NoError(t, err)

			var bound int
			var gotView string
			plan, err := NewValidationPlan(holder.Attribute(), layout, ValidationPlanOptions{
				Required: true,
				Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
					bound++
					gotView = request.View
					return validator, nil
				},
			})
			require.NoError(t, err)
			if !tc.wantBound {
				assert.Zero(t, bound)
				assert.Empty(t, plan.ValidatorDeclarations())
				return
			}
			assert.Equal(t, 1, bound)
			assert.Equal(t, tc.wantView, gotView)
			assert.Equal(t, []*NameDeclaration{validator}, plan.ValidatorDeclarations())
		})
	}
}
