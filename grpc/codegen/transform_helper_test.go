// This file verifies gRPC conversions share strict private functions when the
// generated source and target types are the same.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
)

// TestTransformHelpersShareAcrossDocumentedUses checks that field prose and
// examples do not split one recursive conversion used by several methods and
// union branches. A nearly identical child with different required-field
// pointer rules keeps a separate function and call sites.
func TestTransformHelpersShareAcrossDocumentedUses(t *testing.T) {
	root := RunGRPCDSL(t, sharedGRPCTransformHelperMethodsDSL)
	services := CreateGRPCServices(root)
	service := services.Get("SharedMethodHelpers")
	require.NotEmpty(t, service.clientTransformHelpers)
	require.NotEmpty(t, service.serverTransformHelpers)

	var clientShared, clientNear *codegen.TransformFunctionData
	clientSharedDeclarations := make(map[*codegen.NameDeclaration]struct{})
	for _, helper := range service.clientTransformHelpers {
		if !strings.Contains(helper.Declaration.Name(), "ToProto") {
			continue
		}
		switch helper.ParamTypeRef {
		case "*sharedmethodhelpers.SharedMethodChild":
			clientShared = helper
			clientSharedDeclarations[helper.Declaration] = struct{}{}
		case "*sharedmethodhelpers.NearMethodChild":
			clientNear = helper
		}
	}
	require.NotNil(t, clientShared)
	require.NotNil(t, clientNear)
	require.Len(t, clientSharedDeclarations, 1)
	require.NotSame(t, clientShared.Declaration, clientNear.Declaration)
	require.Contains(t, clientShared.Code, clientShared.Declaration.Name()+"(v.Next)")
	require.Contains(t, clientNear.Code, clientNear.Declaration.Name()+"(v.Next)")

	var serverShared, serverNear *codegen.TransformFunctionData
	serverSharedDeclarations := make(map[*codegen.NameDeclaration]struct{})
	for _, helper := range service.serverTransformHelpers {
		switch helper.ParamTypeRef {
		case "*shared_method_helperspb.SharedMethodChild":
			serverShared = helper
			serverSharedDeclarations[helper.Declaration] = struct{}{}
		case "*shared_method_helperspb.NearMethodChild":
			serverNear = helper
		}
	}
	require.NotNil(t, serverShared)
	require.NotNil(t, serverNear)
	require.Len(t, serverSharedDeclarations, 1)
	require.NotSame(t, serverShared.Declaration, serverNear.Declaration)
	require.Contains(t, serverShared.Code, serverShared.Declaration.Name()+"(v.Next)")
	require.Contains(t, serverNear.Code, serverNear.Declaration.Name()+"(v.Next)")

	for _, endpoint := range service.Endpoints[:2] {
		require.Contains(t, endpoint.Request.ClientConvert.Init.Code,
			"message.Child = "+clientShared.Declaration.Name()+"(payload.Child)")
		require.Contains(t, endpoint.Request.ServerConvert.Init.Code,
			"v.Child = "+serverShared.Declaration.Name()+"(message.Child)")
	}
	var inspect *EndpointData
	for _, endpoint := range service.Endpoints {
		if endpoint.Method.Name == "Inspect" {
			inspect = endpoint
			break
		}
	}
	require.NotNil(t, inspect)
	require.Contains(t, inspect.Request.ClientConvert.Init.Code,
		"message.Child = "+clientNear.Declaration.Name()+"(payload.Child)")
	require.Contains(t, inspect.Request.ServerConvert.Init.Code,
		"v.Child = "+serverNear.Declaration.Name()+"(message.Child)")

	client := codegen.SectionsCode(t, clientTypeFiles(services)[0].SectionTemplates[1:])
	server := codegen.SectionsCode(t, serverTypeFiles(services)[0].SectionTemplates[1:])
	testutil.AssertGo(t, "testdata/golden/transform_helper_shared_methods_client.go.golden", client)
	testutil.AssertGo(t, "testdata/golden/transform_helper_shared_methods_server.go.golden", server)
}

// TestTransformHelpersKeepViewedAndServiceDeclarations checks that a projected
// child uses its views package declaration while the original child used by a
// normal method keeps its service package declaration. Their private helpers
// must remain separate in the final client and server packages.
func TestTransformHelpersKeepViewedAndServiceDeclarations(t *testing.T) {
	root := RunGRPCDSL(t, viewedAndServiceTransformHelperDSL)
	services := CreateGRPCServices(root)
	service := services.Get("ViewBindings")

	var clientServiceFrom, clientViewFrom *codegen.TransformFunctionData
	for _, helper := range service.clientTransformHelpers {
		switch helper.ResultTypeRef {
		case "*viewbindings.SharedDeclarationChild":
			clientServiceFrom = helper
		case "*viewbindingsviews.SharedDeclarationChildView":
			clientViewFrom = helper
		}
	}
	require.NotNil(t, clientServiceFrom)
	require.NotNil(t, clientViewFrom)
	require.NotSame(t, clientServiceFrom.Declaration, clientViewFrom.Declaration)

	var serverServiceTo, serverViewTo *codegen.TransformFunctionData
	for _, helper := range service.serverTransformHelpers {
		switch helper.ParamTypeRef {
		case "*viewbindings.SharedDeclarationChild":
			serverServiceTo = helper
		case "*viewbindingsviews.SharedDeclarationChildView":
			serverViewTo = helper
		}
	}
	require.NotNil(t, serverServiceTo)
	require.NotNil(t, serverViewTo)
	require.NotSame(t, serverServiceTo.Declaration, serverViewTo.Declaration)

	var viewed, plain *EndpointData
	for _, endpoint := range service.Endpoints {
		switch endpoint.Method.Name {
		case "Viewed":
			viewed = endpoint
		case "Plain":
			plain = endpoint
		}
	}
	require.NotNil(t, viewed)
	require.NotNil(t, plain)
	require.Contains(t, plain.Response.ClientConvert.Init.Code, clientServiceFrom.Declaration.Name()+"(message.Child)")
	require.Contains(t, plain.Response.ServerConvert.Init.Code, serverServiceTo.Declaration.Name()+"(result.Child)")
	require.Contains(t, viewed.Response.ClientConvert.Init.Code, clientViewFrom.Declaration.Name()+"(message.Child)")
	require.Contains(t, viewed.Response.ServerConvert.Init.Code, serverViewTo.Declaration.Name()+"(result.Child)")

	client := codegen.SectionsCode(t, clientTypeFiles(services)[0].SectionTemplates[1:])
	server := codegen.SectionsCode(t, serverTypeFiles(services)[0].SectionTemplates[1:])
	testutil.AssertGo(t, "testdata/golden/transform_helper_viewed_service_client.go.golden", client)
	testutil.AssertGo(t, "testdata/golden/transform_helper_viewed_service_server.go.golden", server)
}

// TestTransformHelpersFollowCollectionWrapper checks that nested conversions
// use the slice stored in a protobuf response wrapper on both transport sides.
func TestTransformHelpersFollowCollectionWrapper(t *testing.T) {
	root := RunGRPCDSL(t, wrappedCollectionTransformHelperDSL)
	services := CreateGRPCServices(root)
	service := services.Get("WrappedCollection")
	require.NotEmpty(t, service.clientTransformHelpers)
	require.NotEmpty(t, service.serverTransformHelpers)

	response := service.Endpoints[0].Response
	require.Contains(t, response.ServerConvert.Init.Code, "message.Field = make(")
	require.Contains(t, response.ClientConvert.Init.Code, "result := make(")

	client := codegen.SectionsCode(t, clientTypeFiles(services)[0].SectionTemplates[1:])
	server := codegen.SectionsCode(t, serverTypeFiles(services)[0].SectionTemplates[1:])
	testutil.AssertGo(t, "testdata/golden/transform_helper_wrapped_collection_client.go.golden", client)
	testutil.AssertGo(t, "testdata/golden/transform_helper_wrapped_collection_server.go.golden", server)
}

// TestTransformHelpersShareRequiredAndOptionalCalls checks that a required
// field and an optional field in one conversion plan share one function. The
// optional call remains inside its nil check.
func TestTransformHelpersShareRequiredAndOptionalCalls(t *testing.T) {
	root := RunGRPCDSL(t, sharedGRPCTransformHelperDSL)
	services := CreateGRPCServices(root)
	service := services.Get("SharedHelpers")
	require.Len(t, service.clientTransformHelpers, 2)
	require.Len(t, service.serverTransformHelpers, 1)
	var clientHelper *codegen.TransformFunctionData
	for _, helper := range service.clientTransformHelpers {
		if strings.HasPrefix(helper.ParamTypeRef, "*sharedhelpers.") {
			clientHelper = helper
		}
	}
	require.NotNil(t, clientHelper)
	serverHelper := service.serverTransformHelpers[0]
	require.NotContains(t, clientHelper.Code, "if v == nil")
	require.NotContains(t, serverHelper.Code, "if v == nil")

	request := service.Endpoints[0].Request
	clientCode := request.ClientConvert.Init.Code
	require.Contains(t, clientCode, "message.Left = "+clientHelper.Declaration.Name()+"(payload.Left)")
	require.Contains(t, clientCode, "if payload.Right != nil {")
	require.Contains(t, clientCode, "message.Right = "+clientHelper.Declaration.Name()+"(payload.Right)")
	serverCode := request.ServerConvert.Init.Code
	require.Contains(t, serverCode, "v.Left = "+serverHelper.Declaration.Name()+"(message.Left)")
	require.Contains(t, serverCode, "if message.Right != nil {")
	require.Contains(t, serverCode, "v.Right = "+serverHelper.Declaration.Name()+"(message.Right)")

	for _, side := range []validateKind{validateClient, validateServer} {
		validations := make(map[string]*protobufValidationRecord)
		for _, validation := range service.protobuf.validators {
			if validation.side == side {
				validations[validation.source.path] = validation
			}
		}
		rootValidation := validations[""]
		leftValidation := validations["left"]
		rightValidation := validations["right"]
		require.NotNil(t, rootValidation)
		require.NotNil(t, leftValidation)
		require.NotNil(t, rightValidation)
		require.Equal(t, "ValidateStoreRequest", rootValidation.declaration.Name())
		require.Equal(t, "validatetest_20_api_SharedHelpers_SharedChild_At_left", leftValidation.declaration.Name())
		require.Equal(t, "validatetest_20_api_SharedHelpers_SharedChild_At_right", rightValidation.declaration.Name())
		require.Contains(t, leftValidation.data.Def, `MissingFieldError("value", "left")`)
		require.Contains(t, rightValidation.data.Def, `MissingFieldError("value", "right")`)
		require.Contains(t, rootValidation.data.Def, "validatetest_20_api_SharedHelpers_SharedChild_At_left(message.Left)")
		require.Contains(t, rootValidation.data.Def, "validatetest_20_api_SharedHelpers_SharedChild_At_right(message.Right)")
	}

	client := codegen.SectionsCode(t, clientTypeFiles(services)[0].SectionTemplates[1:])
	testutil.AssertGo(t, "testdata/golden/transform_helper_shared.go.golden", client)
}

// sharedGRPCTransformHelperDSL creates one request with required and optional
// fields that use the same named type.
func sharedGRPCTransformHelperDSL() {
	child := dsl.Type("SharedChild", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Required("value")
	})
	dsl.Service("SharedHelpers", func() {
		dsl.Method("Store", func() {
			dsl.Payload(func() {
				dsl.Field(1, "left", child)
				dsl.Field(2, "right", child)
				dsl.Required("left")
			})
			dsl.GRPC(func() {})
		})
	})
}

// sharedGRPCTransformHelperMethodsDSL uses one recursive child under different
// method and union field prose. It also defines a separate recursive child
// whose value field is optional instead of required.
func sharedGRPCTransformHelperMethodsDSL() {
	child := dsl.Type("SharedMethodChild", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Field(2, "next", "SharedMethodChild")
		dsl.Required("value")
	})
	near := dsl.Type("NearMethodChild", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Field(2, "next", "NearMethodChild")
	})
	dsl.Service("SharedMethodHelpers", func() {
		for _, method := range []string{"Store", "Replace"} {
			dsl.Method(method, func() {
				dsl.Payload(func() {
					dsl.Field(1, "child", child, method+" uses the shared child.", func() {
						dsl.Example(map[string]any{"value": method})
					})
					dsl.Required("child")
				})
				dsl.GRPC(func() {})
			})
		}
		dsl.Method("Choose", func() {
			dsl.Payload(func() {
				dsl.OneOf("choice", func() {
					dsl.TypeName("DescribedChoice")
					for index, branch := range []string{"First", "Second", "Third"} {
						dsl.Field(index+1, branch, child, branch+" documented branch.", func() {
							dsl.Example(map[string]any{"value": branch})
						})
					}
				})
				dsl.Required("choice")
			})
			dsl.GRPC(func() {})
		})
		dsl.Method("Inspect", func() {
			dsl.Payload(func() {
				dsl.Field(1, "child", near)
				dsl.Required("child")
			})
			dsl.GRPC(func() {})
		})
	})
}

// viewedAndServiceTransformHelperDSL uses one named child through both a
// projected result view and an ordinary method result.
func viewedAndServiceTransformHelperDSL() {
	child := dsl.Type("SharedDeclarationChild", func() {
		dsl.Field(1, "value", dsl.String)
		dsl.Required("value")
	})
	viewedResult := dsl.ResultType("application/vnd.view-bindings", func() {
		dsl.TypeName("ViewedContainer")
		dsl.Field(1, "child", child)
		dsl.Required("child")
		dsl.View("default", func() {
			dsl.Attribute("child")
		})
	})
	dsl.Service("ViewBindings", func() {
		dsl.Method("Viewed", func() {
			dsl.Result(viewedResult)
			dsl.GRPC(func() {})
		})
		dsl.Method("Plain", func() {
			dsl.Payload(func() {
				dsl.Field(1, "child", child)
				dsl.Required("child")
			})
			dsl.Result(func() {
				dsl.Field(1, "child", child)
				dsl.Required("child")
			})
			dsl.GRPC(func() {})
		})
	})
}

// wrappedCollectionTransformHelperDSL returns a viewed slice whose item uses a
// second viewed type, so gRPC needs a nested conversion inside its Field slice.
func wrappedCollectionTransformHelperDSL() {
	child := dsl.ResultType("application/vnd.wrapped-collection-child", func() {
		dsl.TypeName("WrappedCollectionChild")
		dsl.Field(1, "value", dsl.String)
		dsl.Required("value")
		dsl.View("default", func() {
			dsl.Attribute("value")
		})
	})
	item := dsl.ResultType("application/vnd.wrapped-collection-item", func() {
		dsl.TypeName("WrappedCollectionItem")
		dsl.Field(1, "child", child)
		dsl.Required("child")
		dsl.View("default", func() {
			dsl.Attribute("child")
		})
	})
	dsl.Service("WrappedCollection", func() {
		dsl.Method("List", func() {
			dsl.Result(dsl.CollectionOf(item), func() {
				dsl.View("default")
			})
			dsl.GRPC(func() {})
		})
	})
}
