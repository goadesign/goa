// This file verifies HTTP and JSON-RPC conversion functions use the exact
// declarations selected while their generated packages are planned.
package codegen

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

// TestTransformHelperOrderingSupportsMoreThan255Functions catches helper name
// ordering that narrows a plan position to one byte.
func TestTransformHelperOrderingSupportsMoreThan255Functions(t *testing.T) {
	source, target := manyDistinctTransformChildren(257, false)
	catalog, generation := testWireTypeCatalog(t)
	policy := jsonBodyPolicy(true, false, false, "")
	catalog.collect(target, wireRequestBody, policy)
	catalog.collectTransform(source, target, "marshal", "many helpers", wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	})

	require.NoError(t, catalog.Declare())
	require.NoError(t, generation.Freeze())
}

// TestTransformHandleSelectsTheCollectedPlan catches structurally identical
// conversions being exchanged when rendering happens in a different order.
func TestTransformHandleSelectsTheCollectedPlan(t *testing.T) {
	source, target := manyDistinctTransformChildren(1, false)
	catalog, generation := testWireTypeCatalog(t)
	policy := jsonBodyPolicy(true, false, false, "")
	catalog.collect(target, wireRequestBody, policy)
	first := catalog.collectTransform(source, target, "marshal", "first", wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	})
	second := catalog.collectTransform(source, target, "marshal", "second", wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	})
	linkTestWireTypeCatalog(t, generation, catalog)

	_, _, err := renderTestTransform(catalog, second, "inventory")
	require.NoError(t, err)
	require.False(t, first.record.used)
	require.True(t, second.record.used)
	_, _, err = renderTestTransform(catalog, first, "inventory")
	require.NoError(t, err)
}

// TestTransformHandleRejectsAnotherCatalog catches a planned conversion being
// rendered into a package that did not claim its declarations.
func TestTransformHandleRejectsAnotherCatalog(t *testing.T) {
	source, target := manyDistinctTransformChildren(1, false)
	first, firstGeneration := testWireTypeCatalog(t)
	policy := jsonBodyPolicy(true, false, false, "")
	first.collect(target, wireRequestBody, policy)
	handle := first.collectTransform(source, target, "marshal", "first", wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	})
	linkTestWireTypeCatalog(t, firstGeneration, first)

	second, secondGeneration := testWireTypeCatalog(t)
	linkTestWireTypeCatalog(t, secondGeneration, second)
	_, _, err := renderTestTransform(second, handle, "inventory")
	require.ErrorContains(t, err, "different generated package")
}

// TestTransformDefinitionsMatchAcrossPlans proves one package declaration may
// be shared by equivalent helper definitions produced by separate plans.
func TestTransformDefinitionsMatchAcrossPlans(t *testing.T) {
	catalog, first, second := plannedMatchingTransforms(t)
	_, firstHelpers, err := renderTestTransform(catalog, first, "inventory")
	require.NoError(t, err)
	_, secondHelpers, err := renderTestTransform(catalog, second, "inventory")
	require.NoError(t, err)
	require.Same(t, firstHelpers[0].Declaration, secondHelpers[0].Declaration)
}

// TestTransformDefinitionsRejectMismatchAcrossPlans catches AppendHelpers
// hiding two different functions assigned to one package declaration.
func TestTransformDefinitionsRejectMismatchAcrossPlans(t *testing.T) {
	catalog, first, second := plannedMatchingTransforms(t)
	_, _, err := renderTestTransform(catalog, first, "inventory")
	require.NoError(t, err)
	_, _, err = renderTestTransform(catalog, second, "different")
	require.ErrorContains(t, err, "has different definitions")
	require.False(t, second.record.used)
}

// TestPlanLinkRejectsUnusedTransform catches planned conversions that no
// generated constructor or stream method renders.
func TestPlanLinkRejectsUnusedTransform(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("audit", func() {
			dsl.Method("show", func() {
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/show")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	transportService := root.API.HTTP.Service("audit")
	planned := plans[0].wireTypes[transportService]
	record := &wireTransformRecord{owner: "unused audit transform", prefix: "marshal"}
	planned.client.transforms = append(planned.client.transforms, record)
	planned.transforms.streamingResults[transportService.HTTPEndpoints[0]] = &plannedResponseTransforms{
		clientDecode: wireTransformHandle{catalog: planned.client, record: record},
	}
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())

	err = plans[0].Link()
	require.ErrorContains(t, err, "unused audit transform")
}

// TestPlanLinkReturnsForeignTransformError catches render failures escaping as
// panics instead of the error returned by Plan.Link.
func TestPlanLinkReturnsForeignTransformError(t *testing.T) {
	root, generation, servicePlan, plan := plannedTransformErrorService(t)
	endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]
	planned := plan.wireTypes[root.API.HTTP.Services[0]]
	response := planned.transforms.responses[viewedConstructorKey{endpoint: endpoint, response: endpoint.Responses[0]}]
	response.serverEncode = response.clientDecode
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())

	var linkErr error
	require.NotPanics(t, func() {
		linkErr = plan.Link()
	})
	require.ErrorContains(t, linkErr, "different generated package")
}

// TestPlanLinkReturnsReusedTransformError catches a second production render
// of one handle escaping as a panic.
func TestPlanLinkReturnsReusedTransformError(t *testing.T) {
	root, generation, servicePlan, plan := plannedTransformErrorService(t)
	serviceExpr := root.API.HTTP.Services[0]
	endpoint := serviceExpr.HTTPEndpoints[0]
	planned := plan.wireTypes[serviceExpr]
	request := planned.transforms.requests[clientBodyConstructorKey{endpoint: endpoint, role: wireRequestBody}]
	response := planned.transforms.responses[viewedConstructorKey{endpoint: endpoint, response: endpoint.Responses[0]}]
	response.serverEncode = request.serverDecode
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())

	var linkErr error
	require.NotPanics(t, func() {
		linkErr = plan.Link()
	})
	require.ErrorContains(t, linkErr, "already rendered")
}

// TestTransformHelperOrderingDoesNotDependOnTraversalOrder catches suffixes
// changing when the same child conversions are collected in reverse order.
func TestTransformHelperOrderingDoesNotDependOnTraversalOrder(t *testing.T) {
	require.Equal(t, plannedTransformHelperNames(t, false), plannedTransformHelperNames(t, true))
}

// TestTransformHelperUsesRetainedServicePackagePreference catches helper names
// derived from the HTTP output directory or a copied view type's spelling.
func TestTransformHelperUsesRetainedServicePackagePreference(t *testing.T) {
	for _, test := range []struct {
		name       string
		preference codegen.ImportSpec
		want       string
	}{
		{"service", codegen.ImportSpec{Name: "inventory", Path: "generated.local/gen/inventory"}, "InventoryChild"},
		{"views", codegen.ImportSpec{Name: "inventoryviews", Path: "generated.local/gen/inventory/views"}, "InventoryviewsChild"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source, target := manyDistinctTransformChildren(1, false)
			catalog, generation := testWireTypeCatalog(t)
			policy := jsonBodyPolicy(true, false, false, "")
			catalog.collect(target, wireRequestBody, policy)
			catalog.collectTransform(source, target, "marshal", test.name, wireTransformLayout{
				wireSide:       wireTransformTarget,
				wirePolicy:     policy,
				servicePackage: test.preference,
			})
			linkTestWireTypeCatalog(t, generation, catalog)

			require.Len(t, catalog.transformHelpers, 1)
			require.Contains(t, catalog.transformHelpers[0].declaration.Name(), test.want)
			require.NotContains(t, catalog.transformHelpers[0].declaration.Name(), "TestChild")
		})
	}
}

func TestClientTransformHelpersNameExactSourceAndTargetTypes(t *testing.T) {
	root := expr.RunDSL(t, testdata.PayloadBodyUserInnerDefaultDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	var (
		code  bytes.Buffer
		count int
	)
	for _, file := range plan.ClientFiles() {
		for _, section := range file.SectionTemplates {
			if section.Name != "client-transform-helper" {
				continue
			}
			count++
			require.NoError(t, section.Write(&code))
		}
	}
	require.Equal(t, 2, count)
	testutil.AssertGo(
		t,
		"testdata/golden/transform_helper_bidirectional-client.go.golden",
		codegen.FormatTestCode(t, "package client\n"+code.String()),
	)
}

func TestViewedTransformHelpersNameViewsPackage(t *testing.T) {
	root := expr.RunDSL(t, testdata.ExplicitBodyUserResultObjectDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	service := plan.services.Get("ServiceExplicitBodyUserResultObject")
	names := make([]string, 0, len(service.ClientTransformHelpers))
	for _, helper := range service.ClientTransformHelpers {
		names = append(names, helper.Name)
	}
	require.Contains(t, names, "unmarshalUserTypeResponseBodyToServiceexplicitbodyuserresultobjectviewsUserTypeViewOptional")
}

func TestTransformHelpersUseConciseServiceAndWireTypeNames(t *testing.T) {
	root := expr.RunDSL(t, conciseTransformHelperDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	service := plan.services.Get("Storage")
	names := make([]string, 0, len(service.ClientTransformHelpers))
	for _, helper := range service.ClientTransformHelpers {
		names = append(names, helper.Name)
	}
	require.Contains(t, names, "marshalStorageWineryToWineryRequestBody")
	require.Contains(t, names, "marshalWineryRequestBodyToStorageWinery")
}

func TestSiblingTransformHelpersShareOneDefinition(t *testing.T) {
	root := expr.RunDSL(t, testdata.ResultTypeSiblingUserTypeFieldsDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	service := plan.services.Get("ServiceResultUserTypeSibling")
	require.Len(t, service.ServerTransformHelpers, 1)
	name := "marshalServiceresultusertypesiblingviewsUserTypeViewToUserTypeResponseBodyOptional"
	require.Equal(t, name, service.ServerTransformHelpers[0].Name)
	result := service.Endpoint("MethodResultUserTypeSibling").Result
	require.NotEmpty(t, result.Responses)
	require.NotEmpty(t, result.Responses[0].ServerBody)
	require.NotNil(t, result.Responses[0].ServerBody[0].Init)
	constructor := result.Responses[0].ServerBody[0].Init.ServerCode
	require.Contains(t, constructor, name+"(res.A)")
	require.Contains(t, constructor, name+"(res.B)")

	var code bytes.Buffer
	for _, file := range plan.ServerFiles() {
		for _, section := range file.SectionTemplates {
			if section.Name == "server-transform-helper" {
				require.NoError(t, section.Write(&code))
			}
		}
	}
	testutil.AssertGo(
		t,
		"testdata/golden/transform_helper_sibling-declarations.go.golden",
		codegen.FormatTestCode(t, "package server\n"+code.String()),
	)
}

func TestTransformHelpersShareExactPackageDeclaration(t *testing.T) {
	root := expr.RunDSL(t, sharedTransformHelperDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	service := plan.services.Get("SharedHelpers")
	requiredName, optionalName := transformHelperNamesByRequired(t, service.serverWireTypes)
	require.Contains(t, service.Endpoint("First").Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")
	require.Contains(t, service.Endpoint("Second").Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")
	require.Contains(t, service.Endpoint("Optional").Payload.Request.PayloadInit.ServerCode, optionalName+"(body.Child)")
	require.NotContains(t, service.Endpoint("Optional").Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")
	var (
		code  bytes.Buffer
		count int
	)
	for _, file := range plan.ServerFiles() {
		for _, section := range file.SectionTemplates {
			if section.Name != "server-transform-helper" {
				continue
			}
			count++
			require.NoError(t, section.Write(&code))
		}
	}
	require.Equal(t, 2, count)
	testutil.AssertGo(
		t,
		"testdata/golden/transform_helper_shared-declarations.go.golden",
		codegen.FormatTestCode(t, "package server\n"+code.String()),
	)
}

func TestJSONRPCTransformHelpersUseHTTPPackageDeclarations(t *testing.T) {
	root := expr.RunDSL(t, sharedJSONRPCTransformHelperDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	serviceData := plans[0].services.Get("SharedHelpers")
	requiredName, optionalName := transformHelperNamesByRequired(t, serviceData.serverWireTypes)
	snapshot, ok := plans[0].JSONRPCService("SharedHelpers")
	require.True(t, ok)
	require.Contains(t, snapshot.Endpoints[0].Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")
	require.Contains(t, snapshot.Endpoints[1].Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")
	require.Contains(t, snapshot.Endpoints[2].Payload.Request.PayloadInit.ServerCode, optionalName+"(body.Child)")
	require.NotContains(t, snapshot.Endpoints[2].Payload.Request.PayloadInit.ServerCode, requiredName+"(body.Child)")

	first := jsonRPCTransformHelpers(snapshot.ServerCodecFile())
	second := jsonRPCTransformHelpers(snapshot.ServerCodecFile())
	require.Len(t, first, 2)
	require.Len(t, second, 2)
	require.ElementsMatch(t, []string{requiredName, optionalName}, []string{first[0].Name, first[1].Name})
	first[0].Name = "changed"
	fresh := jsonRPCTransformHelpers(snapshot.ServerCodecFile())
	require.Equal(t, second[0].Name, fresh[0].Name)
	require.NotEqual(t, first[0].Name, fresh[0].Name)
}

// jsonRPCTransformHelpers returns the copied conversion functions from one
// JSON-RPC codec file so the test can change one copy without changing another.
func jsonRPCTransformHelpers(file *codegen.File) []*jsonRPCTransformFunctionData {
	var helpers []*jsonRPCTransformFunctionData
	for _, section := range file.SectionTemplates {
		if section.Name != "server-transform-helper" {
			continue
		}
		helpers = append(helpers, section.Data.(*jsonRPCTransformFunctionData))
	}
	return helpers
}

// transformHelperNamesByRequired returns the two function names from the test
// catalog according to whether they accept a missing source value.
func transformHelperNamesByRequired(t *testing.T, catalog *wireTypeCatalog) (string, string) {
	t.Helper()
	var required, optional string
	for _, helper := range catalog.transformHelpers {
		if helper.identity.required {
			required = helper.declaration.Name()
		} else {
			optional = helper.declaration.Name()
		}
	}
	require.NotEmpty(t, required)
	require.NotEmpty(t, optional)
	return required, optional
}

// sharedTransformHelperDSL uses one named child in two required fields and one
// optional field so functions share only when missing values behave the same.
func sharedTransformHelperDSL() {
	sharedTransformHelperDesign(false)
}

// sharedJSONRPCTransformHelperDSL applies the same service types to JSON-RPC.
func sharedJSONRPCTransformHelperDSL() {
	sharedTransformHelperDesign(true)
}

// conciseTransformHelperDSL names the service and nested type like a normal
// application so generated functions should use those public names directly.
func conciseTransformHelperDSL() {
	winery := dsl.ResultType("application/vnd.transform-helper.winery", func() {
		dsl.TypeName("Winery")
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	bottle := dsl.Type("Bottle", func() {
		dsl.Attribute("winery", winery)
		dsl.Required("winery")
	})
	dsl.Service("Storage", func() {
		dsl.Method("Create", func() {
			dsl.Payload(bottle)
			dsl.HTTP(func() {
				dsl.POST("/")
			})
		})
	})
}

// sharedTransformHelperDesign creates the service used to test normal HTTP and
// JSON-RPC package generation.
func sharedTransformHelperDesign(jsonrpc bool) {
	child := dsl.Type("SharedChild", func() {
		dsl.Attribute("value", dsl.String)
		dsl.Required("value")
	})
	dsl.Service("SharedHelpers", func() {
		for _, method := range []struct {
			name     string
			path     string
			required bool
		}{
			{name: "First", path: "/first", required: true},
			{name: "Second", path: "/second", required: true},
			{name: "Optional", path: "/optional"},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(func() {
					dsl.Attribute("child", child)
					if method.required {
						dsl.Required("child")
					}
				})
				if jsonrpc {
					dsl.JSONRPC(func() {})
				} else {
					dsl.HTTP(func() {
						dsl.POST(method.path)
					})
				}
			})
		}
	})
}

// manyDistinctTransformChildren builds one object conversion with more helper
// functions than fit in one byte. Every child requests the same preferred type
// name but has a different field, so declaration ordering must use its complete
// source and target type identity.
func manyDistinctTransformChildren(count int, reverse bool) (*expr.AttributeExpr, *expr.AttributeExpr) {
	sourceObject := make(expr.Object, 0, count)
	targetObject := make(expr.Object, 0, count)
	required := make([]string, count)
	for position := range count {
		index := position
		if reverse {
			index = count - position - 1
		}
		field := fmt.Sprintf("field_%d", index)
		value := fmt.Sprintf("value_%d", index)
		sourceChild := &expr.UserTypeExpr{
			TypeName: "Child",
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: value, Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
		}
		targetChild := &expr.UserTypeExpr{
			TypeName: "Child",
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: value, Attribute: &expr.AttributeExpr{Type: expr.String}},
			}},
		}
		sourceObject = append(sourceObject, &expr.NamedAttributeExpr{Name: field, Attribute: &expr.AttributeExpr{Type: sourceChild}})
		targetObject = append(targetObject, &expr.NamedAttributeExpr{Name: field, Attribute: &expr.AttributeExpr{Type: targetChild}})
		required[index] = field
	}
	return &expr.AttributeExpr{
			Type:       &sourceObject,
			Validation: &expr.ValidationExpr{Required: required},
		}, &expr.AttributeExpr{
			Type:       &targetObject,
			Validation: &expr.ValidationExpr{Required: append([]string(nil), required...)},
		}
}

// plannedTransformHelperNames returns each distinct child field and the helper
// name assigned to its conversion.
func plannedTransformHelperNames(t *testing.T, reverse bool) map[string]string {
	t.Helper()
	source, target := manyDistinctTransformChildren(3, reverse)
	catalog, generation := testWireTypeCatalog(t)
	policy := jsonBodyPolicy(true, false, false, "")
	catalog.collect(target, wireRequestBody, policy)
	catalog.collectTransform(source, target, "marshal", "ordered helpers", wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	})
	linkTestWireTypeCatalog(t, generation, catalog)

	names := make(map[string]string, len(catalog.transformHelpers))
	for _, helper := range catalog.transformHelpers {
		object := expr.AsObject(helper.identity.source.attribute.Type)
		names[(*object)[0].Name] = helper.declaration.Name()
	}
	return names
}

// plannedMatchingTransforms returns two independent plans whose child helpers
// share one declaration in the generated package.
func plannedMatchingTransforms(
	t *testing.T,
) (*wireTypeCatalog, wireTransformHandle, wireTransformHandle) {
	t.Helper()
	source, target := manyDistinctTransformChildren(1, false)
	catalog, generation := testWireTypeCatalog(t)
	policy := jsonBodyPolicy(true, false, false, "")
	catalog.collect(target, wireRequestBody, policy)
	layout := wireTransformLayout{
		wireSide:       wireTransformTarget,
		wirePolicy:     policy,
		servicePackage: testServicePackage(),
	}
	first := catalog.collectTransform(source, target, "marshal", "first", layout)
	second := catalog.collectTransform(source, target, "marshal", "second", layout)
	linkTestWireTypeCatalog(t, generation, catalog)
	return catalog, first, second
}

// renderTestTransform renders one service-to-wire conversion using the given
// service package qualifier.
func renderTestTransform(
	catalog *wireTypeCatalog,
	handle wireTransformHandle,
	servicePackage string,
) (string, []*codegen.TransformFunctionData, error) {
	serviceContext := codegen.NewAttributeContext(false, false, true, servicePackage, codegen.NewNameScope())
	wireContext := jsonBodyContext(catalog, catalog.scope, true, false)
	return catalog.renderTransform(handle, handle.record.target, "source", "target", serviceContext, wireContext)
}

// testServicePackage is the retained service package used by direct wire
// catalog tests.
func testServicePackage() codegen.ImportSpec {
	return codegen.ImportSpec{Name: "inventory", Path: "generated.local/gen/inventory"}
}

// plannedTransformErrorService returns an unlinked plan with request and
// response conversions that tests can replace with an invalid handle.
func plannedTransformErrorService(
	t *testing.T,
) (*expr.RootExpr, *codegen.Generation, *service.Plan, *Plan) {
	t.Helper()
	root := expr.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		result := dsl.Type("Result", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("transform errors", func() {
			dsl.Method("show", func() {
				dsl.Payload(payload)
				dsl.Result(result)
				dsl.HTTP(func() {
					dsl.POST("/show")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	return root, generation, servicePlan, plans[0]
}
