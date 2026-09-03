// This file verifies deterministic HTTP wire union identity and confirms that
// detached wire expressions do not retain service package ownership.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestCollectHTTPUnionTypesDeterministicAcrossObjectOrder(t *testing.T) {
	sourceFromSignals := makeUnionForOrderTest("source",
		"physical_point",
		"synthetic_series",
	)
	sourceFromSignals.TypeName = "SignalSource"
	sourceFromInputs := makeUnionForOrderTest("source",
		"time_series",
		"energy_rates",
	)
	sourceFromInputs.TypeName = "InputSource"

	forward := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name: "alpha",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromSignals,
				},
			},
			{
				Name: "beta",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromInputs,
				},
			},
		},
	}
	reverse := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name: "beta",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromInputs,
				},
			},
			{
				Name: "alpha",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromSignals,
				},
			},
		},
	}

	forwardNames := collectHTTPUnionTypeNames(t, forward)
	reverseNames := collectHTTPUnionTypeNames(t, reverse)

	require.Len(t, forwardNames, 2)
	require.Equal(t, forwardNames, reverseNames)
}

func TestCollectHTTPUnionTypesRejectsUnrelatedSameShapedDeclarations(t *testing.T) {
	first := makeUnionForOrderTest("Value", "bool", "number")
	second := makeUnionForOrderTest("Value", "bool", "number")
	bodies := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name:      "first",
				Attribute: &expr.AttributeExpr{Type: first},
			},
			{
				Name:      "second",
				Attribute: &expr.AttributeExpr{Type: second},
			},
		},
	}

	catalog, _ := testWireTypeCatalog(t)
	catalog.collect(bodies, wireAttribute, wireTypePolicy{})
	err := catalog.Declare()
	require.ErrorContains(t, err, `declare HTTP OneOf "Value" for direct attribute`)
	require.ErrorContains(t, err, `cannot declare exact type "Value"`)
	require.ErrorContains(t, err, "set TypeName on the OneOf to a unique name")
}

func TestCollectHTTPUnionTypesRejectsBranchesWithTheSameGoName(t *testing.T) {
	union := makeUnionForOrderTest("Value", "foo-bar", "foo_bar")
	catalog, _ := testWireTypeCatalog(t)
	catalog.collect(&expr.AttributeExpr{Type: union}, wireAttribute, wireTypePolicy{})

	err := catalog.Declare()

	require.ErrorContains(t, err, `OneOf "Value" branches "foo-bar" and "foo_bar" both generate Go name "FooBar"`)
	require.ErrorContains(t, err, "rename one of the branches")
	require.NotContains(t, err.Error(), "TypeName")
}

func TestCollectHTTPUnionTypesUsesExactRootRoleNames(t *testing.T) {
	authored := &expr.AttributeExpr{Type: makeUnionForOrderTest("Scope", "site_set", "all_sites")}
	catalog, generation := testWireTypeCatalog(t)

	catalog.collect(authored, wireAttribute, wireTypePolicy{})
	catalog.collect(expr.DupAtt(authored), wireRequestBody, wireTypePolicy{request: true})
	catalog.collect(expr.DupAtt(authored), wireRequestBody, wireTypePolicy{request: true})
	catalog.collect(expr.DupAtt(authored), wireStreamPayload, wireTypePolicy{request: true})
	catalog.collect(expr.DupAtt(authored), wireResponseBody, wireTypePolicy{})
	catalog.collect(expr.DupAtt(authored), wireResponseBody, wireTypePolicy{view: "detailed"})
	linkTestWireTypeCatalog(t, generation, catalog)

	names := make([]string, 0, len(catalog.unions))
	for _, union := range catalog.unionTypes() {
		names = append(names, union.Name)
	}
	require.Equal(t, []string{
		"Scope",
		"ScopeDetailedResponseBody",
		"ScopeRequestBody",
		"ScopeResponseBody",
		"ScopeStreamingBody",
	}, names)
}

func TestCollectHTTPUnionTypesRejectsConflictingShapesInOneRootRole(t *testing.T) {
	authored := &expr.AttributeExpr{Type: makeUnionForOrderTest("Scope", "site_set", "all_sites")}
	first := expr.DupAtt(authored)
	second := expr.DupAtt(authored)
	second.Type.(*expr.Union).ValueKey = "data"
	catalog, _ := testWireTypeCatalog(t)

	catalog.collect(first, wireRequestBody, wireTypePolicy{request: true})
	catalog.collect(second, wireRequestBody, wireTypePolicy{request: true})
	err := catalog.Declare()
	require.ErrorContains(t, err, "OneOf \"Scope\" produces different RequestBody definitions")
	require.ErrorContains(t, err, "use separate OneOf declarations")
}

func TestHTTPServiceDataReusesOneAuthoredUnionAcrossMethods(t *testing.T) {
	root := expr.RunDSL(t, func() {
		payload := dsl.Type("SharedPayload", func() {
			dsl.OneOf("Value", sameShapedValueUnionDSL)
		})
		dsl.Service("values", func() {
			dsl.Method("first", func() {
				dsl.Payload(payload)
				dsl.HTTP(func() {
					dsl.POST("/first")
				})
			})
			dsl.Method("second", func() {
				dsl.Payload(payload)
				dsl.HTTP(func() {
					dsl.POST("/second")
				})
			})
		})
	})

	data := linkedHTTPPlanForRoot(t, root).services.Get("values")
	require.NotNil(t, data)
	for _, catalog := range []*wireTypeCatalog{data.serverWireTypes, data.clientWireTypes} {
		unions := catalog.unionTypes()
		emitted := make([]string, len(unions))
		for i, union := range unions {
			emitted[i] = union.Name
		}
		require.Equal(t, []string{"ValueRequestBody"}, emitted)
	}
	require.Contains(t, data.Endpoint("first").Payload.Request.ServerBody.Def, "Value *ValueRequestBody ")
	require.Contains(t, data.Endpoint("second").Payload.Request.ServerBody.Def, "Value *ValueRequestBody ")
}

func TestHTTPServiceDataReusesOneAuthoredUnionAcrossExplicitRequestBodies(t *testing.T) {
	root := expr.RunDSL(t, func() {
		payload := dsl.Type("SharedPayload", func() {
			dsl.OneOf("Choice", sameShapedValueUnionDSL)
			dsl.Attribute("note", dsl.String)
		})
		dsl.Service("values", func() {
			for _, method := range []string{"first", "second"} {
				dsl.Method(method, func() {
					dsl.Payload(payload)
					dsl.HTTP(func() {
						dsl.POST("/" + method)
						dsl.Body(func() {
							if method == "first" {
								dsl.Attribute("Choice")
								dsl.Attribute("note")
								return
							}
							dsl.Attribute("note")
							dsl.Attribute("Choice")
						})
					})
				})
			}
		})
	})

	data := linkedHTTPPlanForRoot(t, root).services.Get("values")
	for _, catalog := range []*wireTypeCatalog{data.serverWireTypes, data.clientWireTypes} {
		require.Len(t, catalog.unionTypes(), 1)
		require.Equal(t, "ChoiceRequestBody", catalog.unionTypes()[0].Name)
	}
}

func TestHTTPServiceDataReusesOneAuthoredUnionAcrossErrors(t *testing.T) {
	root := expr.RunDSL(t, func() {
		failure := dsl.Type("SharedFailure", func() {
			dsl.OneOf("Cause", sameShapedValueUnionDSL)
		})
		dsl.Service("values", func() {
			for _, method := range []string{"first", "second"} {
				dsl.Method(method, func() {
					dsl.Error("failed", failure)
					dsl.HTTP(func() {
						dsl.GET("/" + method)
						dsl.Response("failed", dsl.StatusBadRequest)
					})
				})
			}
		})
	})

	data := linkedHTTPPlanForRoot(t, root).services.Get("values")
	for _, catalog := range []*wireTypeCatalog{data.serverWireTypes, data.clientWireTypes} {
		require.Len(t, catalog.unionTypes(), 1)
		require.Equal(t, "CauseResponseBody", catalog.unionTypes()[0].Name)
	}
}

func TestHTTPServiceDataReusesUnionWhenBranchTypeAppearsEarlier(t *testing.T) {
	root := expr.RunDSL(t, unionBranchOrderDSL)

	data := linkedHTTPPlanForRoot(t, root).services.Get("values")
	for _, catalog := range []*wireTypeCatalog{data.serverWireTypes, data.clientWireTypes} {
		require.Len(t, catalog.unionTypes(), 1)
		require.Equal(t, "RelationshipResponseBody", catalog.unionTypes()[0].Name)
	}
}

func TestHTTPUnionBranchTypesDoNotDependOnSurroundingFieldOrder(t *testing.T) {
	code := renderClientTypesCode(t, unionBranchOrderDSL)
	testutil.AssertGo(t, "testdata/golden/http_union_branch_order.go.golden", code)
}

func TestMakeHTTPTypeRemovesServicePackageOwnershipFromWireCopy(t *testing.T) {
	nested := &expr.UserTypeExpr{
		TypeName: "Nested",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "choice", Attribute: &expr.AttributeExpr{Type: makeUnionForOrderTest("Choice", "text", "number")}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	outer := &expr.UserTypeExpr{
		TypeName: "Envelope",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "nested", Attribute: &expr.AttributeExpr{Type: nested}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}

	wire := makeHTTPType(&expr.AttributeExpr{Type: outer})
	wireOuter := wire.Type.(expr.UserType)
	wireNested := expr.AsObject(wireOuter.Attribute().Type).Attribute("nested").Type.(expr.UserType)

	require.NotContains(t, wireOuter.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, wireNested.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, outer.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, nested.Attribute().Meta, "struct:pkg:path")
}

func TestStreamingHTTPTypeRemovesServicePackageOwnershipFromWireCopy(t *testing.T) {
	nested := &expr.UserTypeExpr{
		TypeName: "Nested",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	outer := &expr.UserTypeExpr{
		TypeName: "Envelope",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "nested", Attribute: &expr.AttributeExpr{Type: nested}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	body := &expr.AttributeExpr{Type: outer}
	endpoint := &expr.HTTPEndpointExpr{StreamingBody: body}

	wire := new(shapedBodies).streaming(endpoint)
	wireOuter := wire.Type.(expr.UserType)
	wireNested := expr.AsObject(wireOuter.Attribute().Type).Attribute("nested").Type.(expr.UserType)

	require.NotContains(t, wireOuter.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, wireNested.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, outer.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, nested.Attribute().Meta, "struct:pkg:path")
}

func sameShapedValueUnionDSL() {
	dsl.Attribute("bool", dsl.Boolean)
	dsl.Attribute("number", dsl.Float64)
}

// unionBranchOrderDSL returns the same authored union through response objects
// that reach its nested branch type in different orders.
func unionBranchOrderDSL() {
	identifier := dsl.Type("Identifier", dsl.String)
	label := dsl.Type("Label", func() {
		dsl.Attribute("value", dsl.String)
		dsl.Required("value")
	})
	snapshot := dsl.Type("Snapshot", func() {
		dsl.Attribute("id", identifier)
		dsl.Attribute("labels", dsl.ArrayOf(label))
		dsl.Required("id", "labels")
	})
	noEdit := dsl.Type("NoEdit", func() {})
	editable := dsl.Type("Editable", func() {
		dsl.Attribute("snapshot", snapshot)
		dsl.Required("snapshot")
	})
	edit := dsl.Type("Edit", func() {
		dsl.OneOf("relationship", func() {
			dsl.Attribute("none", noEdit)
			dsl.Attribute("editable", editable)
		})
		dsl.Required("relationship")
	})
	withEarlierSnapshot := dsl.Type("WithEarlierSnapshot", func() {
		dsl.Attribute("snapshot", snapshot)
		dsl.Attribute("edit", edit)
		dsl.Required("snapshot", "edit")
	})
	withEditOnly := dsl.Type("WithEditOnly", func() {
		dsl.Attribute("edit", edit)
		dsl.Required("edit")
	})

	dsl.Service("values", func() {
		dsl.Method("with_earlier_snapshot", func() {
			dsl.Result(withEarlierSnapshot)
			dsl.HTTP(func() {
				dsl.GET("/with-earlier-snapshot")
			})
		})
		dsl.Method("with_edit_only", func() {
			dsl.Result(withEditOnly)
			dsl.HTTP(func() {
				dsl.GET("/with-edit-only")
			})
		})
	})
}

func collectHTTPUnionTypeNames(t *testing.T, att *expr.AttributeExpr) map[string]string {
	t.Helper()
	catalog, generation := testWireTypeCatalog(t)
	catalog.collect(att, wireAttribute, wireTypePolicy{})
	linkTestWireTypeCatalog(t, generation, catalog)

	names := make(map[string]string, len(catalog.unions))
	for _, record := range catalog.unions {
		names[record.attribute.Type.Name()] = record.data.Name
	}
	return names
}

func makeUnionForOrderTest(typeName string, variants ...string) *expr.Union {
	values := make([]*expr.NamedAttributeExpr, len(variants))
	for i, variant := range variants {
		values[i] = &expr.NamedAttributeExpr{
			Name: variant,
			Attribute: &expr.AttributeExpr{
				Type: expr.String,
			},
		}
	}
	return &expr.Union{TypeName: typeName, Values: values}
}
