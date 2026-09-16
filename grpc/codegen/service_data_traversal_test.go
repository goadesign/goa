// This file verifies that gRPC validation discovery distinguishes unrelated
// declarations while stopping recursion through copied declarations.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestCollectMessagesDistinguishesEqualNameAndUIDOrigins(t *testing.T) {
	first := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	second := grpcMessageTraversalType("Shared", "shared", expr.Int, "2")
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}
	sd := grpcTraversalServiceData()

	messages := freezeTraversalMessages(t, sd, root)
	require.Len(t, messages, 2)
	require.NotEqual(t, messages[0].VarName, messages[1].VarName)
	require.NotEqual(t, messages[0].Ref, messages[1].Ref)
	require.Contains(t, messages[0].Def, "string value = 1")
	require.Contains(t, messages[1].Def, "sint32 value = 2")
}

func TestCollectMessagesDistinguishesOneOriginWithDifferentWireShape(t *testing.T) {
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	first := expr.Dup(original).(expr.UserType)
	second := expr.Dup(original).(expr.UserType)
	expr.AsObject(second).Attribute("value").Meta["rpc:tag"] = []string{"2"}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}
	sd := grpcTraversalServiceData()

	messages := freezeTraversalMessages(t, sd, root)
	require.Len(t, messages, 2)
	require.NotEqual(t, messages[0].VarName, messages[1].VarName)
	require.Contains(t, messages[0].Def, "string value = 1")
	require.Contains(t, messages[1].Def, "string value = 2")
}

func TestCollectMessagesReusesIdenticalDeclaration(t *testing.T) {
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	first := expr.Dup(original).(expr.UserType)
	second := expr.Dup(original).(expr.UserType)
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}

	messages := freezeTraversalMessages(t, grpcTraversalServiceData(), root)
	require.Len(t, messages, 1)
}

func TestCollectMessagesDistinguishesProtoOverrides(t *testing.T) {
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	first := &expr.AttributeExpr{
		Type: expr.Dup(original),
		Meta: expr.MetaExpr{"struct:name:proto": {"FirstWire"}},
	}
	second := &expr.AttributeExpr{
		Type: expr.Dup(original),
		Meta: expr.MetaExpr{"struct:name:proto": {"SecondWire"}},
	}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: first},
		{Name: "second", Attribute: second},
	}}

	messages := freezeTraversalMessages(t, grpcTraversalServiceData(), root)
	require.Len(t, messages, 2)
	require.Equal(t, "FirstWire", messages[0].VarName)
	require.Equal(t, "SecondWire", messages[1].VarName)
}

func TestCollectMessagesDistinguishesSharedExplicitNameWithDifferentSchemas(t *testing.T) {
	first := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("First", "first", expr.String, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	second := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("Second", "second", expr.Int, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: first},
		{Name: "second", Attribute: second},
	}}

	messages := freezeTraversalMessages(t, grpcTraversalServiceData(), root)
	require.Len(t, messages, 2)
	require.Equal(t, "SharedWire", messages[0].VarName)
	require.Equal(t, "SharedWire2", messages[1].VarName)
	require.Contains(t, messages[0].Def, "string value = 1")
	require.Contains(t, messages[1].Def, "sint32 value = 1")
}

func TestCollectMessagesDistinguishesSharedExplicitNameWithDifferentOrigins(t *testing.T) {
	first := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("First", "first", expr.String, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	second := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("Second", "second", expr.String, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: first},
		{Name: "second", Attribute: second},
	}}

	messages := freezeTraversalMessages(t, grpcTraversalServiceData(), root)
	require.Len(t, messages, 2)
	require.Equal(t, "SharedWire", messages[0].VarName)
	require.Equal(t, "SharedWire2", messages[1].VarName)
}

func TestCollectMessagesUsesUnaryResultSourceForMixedResults(t *testing.T) {
	firstResult := grpcMessageTraversalType("FirstResult", "first-result", expr.String, "1")
	secondResult := grpcMessageTraversalType("SecondResult", "second-result", expr.String, "1")
	streamingResult := grpcMessageTraversalType("StreamingResult", "streaming-result", expr.String, "1")
	firstEndpoint := &expr.GRPCEndpointExpr{MethodExpr: &expr.MethodExpr{
		Result:          &expr.AttributeExpr{Type: firstResult},
		StreamingResult: &expr.AttributeExpr{Type: streamingResult},
	}}
	secondEndpoint := &expr.GRPCEndpointExpr{MethodExpr: &expr.MethodExpr{
		Result:          &expr.AttributeExpr{Type: secondResult},
		StreamingResult: &expr.AttributeExpr{Type: streamingResult},
	}}
	firstWire := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("FirstWire", "first-wire", expr.String, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	secondWire := &expr.AttributeExpr{
		Type: grpcMessageTraversalType("SecondWire", "second-wire", expr.String, "1"),
		Meta: expr.MetaExpr{"struct:name:proto": {"SharedWire"}},
	}
	sd := grpcTraversalServiceData()
	sd.protobuf = newProtobufPackageCatalog(sd.PkgName)
	require.NoError(t, sd.protobuf.collectMessage(firstWire, protobufRootMessageSource(firstWire, firstEndpoint, nil, protobufResponseMessage)))
	require.NoError(t, sd.protobuf.collectMessage(secondWire, protobufRootMessageSource(secondWire, secondEndpoint, nil, protobufResponseMessage)))
	planTestProtobufCatalog(t, sd)

	messages := sd.protobuf.freezeMessages(sd)
	require.Len(t, messages, 2)
	require.Equal(t, "SharedWire", messages[0].VarName)
	require.Equal(t, "SharedWire2", messages[1].VarName)
}

func TestCollectMessagesStopsAtRecursiveCopy(t *testing.T) {
	message := grpcMessageTraversalType("Recursive", "recursive", expr.String, "1")
	object := expr.AsObject(message)
	*object = append(*object, &expr.NamedAttributeExpr{
		Name: "next",
		Attribute: &expr.AttributeExpr{
			Type: message,
			Meta: expr.MetaExpr{"rpc:tag": {"2"}},
		},
	})

	messages := freezeTraversalMessages(t, grpcTraversalServiceData(), &expr.AttributeExpr{Type: expr.Dup(message)})
	require.Len(t, messages, 1)
	require.Contains(t, messages[0].Def, "Recursive next = 2")
}

// TestCollectValidationsStopsAtRecursivePath checks that a recursive field
// calls the private validator chosen for the first nested occurrence instead
// of creating one function name per recursion depth.
func TestCollectValidationsStopsAtRecursivePath(t *testing.T) {
	minimum := 2
	message := grpcValidationTraversalType("Recursive", "recursive", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	})
	object := expr.AsObject(message)
	next := &expr.AttributeExpr{
		Type: message,
		Meta: expr.MetaExpr{"rpc:tag": {"2"}},
	}
	*object = append(*object, &expr.NamedAttributeExpr{Name: "next", Attribute: next})
	root := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, root)
	sd.protobuf.collectValidation(root, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	require.Len(t, sd.validations, 2)
	rootValidation := addValidation(root, sd, true)
	nestedValidation := addValidation(next, sd, true)
	require.Equal(t, "ValidateRecursive", rootValidation.Declaration.Name())
	require.Equal(t, "validateTestAPI_TestService_Recursive_At_next", nestedValidation.Declaration.Name())
	require.Contains(t, rootValidation.Def, "validateTestAPI_TestService_Recursive_At_next(message.Next)")
	require.Contains(t, nestedValidation.Def, "validateTestAPI_TestService_Recursive_At_next(next.Next)")
	require.Contains(t, nestedValidation.Def, `InvalidLengthError("next.value"`)
}

func TestProtoBufMessageNameRequiresFrozenDeclaration(t *testing.T) {
	message := grpcMessageTraversalType("Unbound", "unbound", expr.String, "1")
	sd := grpcTraversalServiceData()

	require.Panics(t, func() {
		protoBufMessageName(&expr.AttributeExpr{Type: message}, sd)
	})
}

func TestProtoBufMessageNameIgnoresLateScopeAllocations(t *testing.T) {
	message := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	attribute := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	messages := freezeTraversalMessages(t, sd, attribute)
	require.Len(t, messages, 1)

	sd.Scope.HashedUnique(grpcMessageTraversalType("Other", "other", expr.Int, "1"), "Shared")
	require.Equal(t, messages[0].VarName, protoBufMessageName(attribute, sd))
}

// TestProtobufCopiesRequireRegistration checks that a copied protobuf value
// uses names only after the copy is connected to the original value.
func TestProtobufCopiesRequireRegistration(t *testing.T) {
	minimum := 2
	state := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "State",
		Values: []*expr.NamedAttributeExpr{
			{
				Name: "active",
				Attribute: &expr.AttributeExpr{
					Type:       expr.String,
					Meta:       expr.MetaExpr{"rpc:tag": {"1"}},
					Validation: &expr.ValidationExpr{MinLength: &minimum},
				},
			},
		},
	}}
	message := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "state", Attribute: state},
		}},
		TypeName: "Message",
		UID:      "message",
	}
	attribute := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, attribute)
	sd.protobuf.collectValidation(attribute, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	copy := expr.DupAtt(attribute)
	copyState := expr.AsObject(copy.Type.(expr.UserType)).Attribute("state")
	branch := state.Type.(*expr.Union).Values[0].Attribute
	copyBranch := copyState.Type.(*expr.Union).Values[0].Attribute
	require.Nil(t, sd.protobuf.messageRecord(copy))
	require.Panics(t, func() {
		sd.protobuf.unionName(copyState)
	})
	_, ok := sd.protobuf.plan.wrapperName(copyBranch)
	require.False(t, ok)

	sd.protobuf.plan.bindAttributeCopy(attribute, copy)
	require.Same(t, sd.protobuf.messageRecord(attribute), sd.protobuf.messageRecord(copy))
	require.Equal(t, sd.protobuf.unionName(state), sd.protobuf.unionName(copyState))
	require.Nil(t, sd.protobuf.validationRecord(copy, validateServer))
	originalWrapper, ok := sd.protobuf.plan.wrapperName(branch)
	require.True(t, ok)
	copyWrapper, ok := sd.protobuf.plan.wrapperName(copyBranch)
	require.True(t, ok)
	require.Equal(t, originalWrapper, copyWrapper)
}

// TestProtobufValidationScopeKeepsMessageNameSeparate checks that a validation
// function collision cannot change the protobuf message name used in its body.
func TestProtobufValidationScopeKeepsMessageNameSeparate(t *testing.T) {
	minimum := 2
	message := grpcValidationTraversalType("Message", "message", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	})
	attribute := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, attribute)
	sd.protobuf.collectValidation(attribute, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	record := sd.protobuf.validationRecord(attribute, validateServer)
	require.NotNil(t, record)
	record.message.name = "RetainedMessage2"
	scope := &protobufValidationScope{
		protoBufScope: &protoBufScope{service: sd},
		catalog:       sd.protobuf,
		side:          validateServer,
		message:       record.message,
		parent:        message,
	}

	require.Equal(t, record.message.name, scope.Name(attribute, "", false, false))
}

func TestAddValidationDistinguishesRulesForOneWireDeclaration(t *testing.T) {
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	first := expr.Dup(original).(expr.UserType)
	second := expr.Dup(original).(expr.UserType)
	firstMinimum := 2
	secondMinimum := 5
	expr.AsObject(first).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &firstMinimum}
	expr.AsObject(second).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &secondMinimum}
	sd := grpcTraversalServiceData()
	firstAttribute := &expr.AttributeExpr{Type: first}
	secondAttribute := &expr.AttributeExpr{Type: second}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: firstAttribute},
		{Name: "second", Attribute: secondAttribute},
	}}
	freezeTraversalMessages(t, sd, root)
	source := grpcTraversalValidationSource()
	sd.protobuf.collectValidation(firstAttribute, validateServer, source.field("first"), "message", "message")
	sd.protobuf.collectValidation(secondAttribute, validateServer, source.field("second"), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(firstAttribute, sd, true)
	secondValidation := addValidation(secondAttribute, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.Len(t, sd.validations, 2)
	require.NotEqual(t, firstValidation.Declaration.Name(), secondValidation.Declaration.Name())
	require.Equal(t, "validateTestAPI_TestService_Shared_At_message_From_Call_Request_Field_first", firstValidation.Declaration.Name())
	require.Equal(t, "validateTestAPI_TestService_Shared_At_message_From_Call_Request_Field_second", secondValidation.Declaration.Name())
	require.Contains(t, firstValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 2, true)`)
	require.Contains(t, secondValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 5, true)`)
}

// TestGRPCValidationNamesDescribeTheirRootAndPath checks the complete private
// name for each nested validator role.
func TestGRPCValidationNamesDescribeTheirRootAndPath(t *testing.T) {
	message := &protobufMessageRecord{plannedName: "Message2"}
	source := func(role protobufValidationRole) protobufValidationSource {
		return protobufValidationSource{
			api:     "Example",
			service: "Storage",
			method:  "Store",
			role:    role,
		}
	}
	request := source(protobufRequestValidation).field("first").field("value")
	response := source(protobufResponseValidation).field("result").field("item")
	response.view = "compact"
	failure := source(protobufErrorValidation).field("cause")
	failure.error = "bad_request"
	stream := source(protobufStreamingRequestValidation).field("items").arrayElement()
	for _, test := range []struct {
		name     string
		source   protobufValidationSource
		expected string
	}{
		{"request", request, "validateExample_Storage_Message2_At_value_From_Store_Request_Field_first_Field_value"},
		{"viewed response", response, "validateExample_Storage_Message2_At_value_From_Store_Response_View_compact_Field_result_Field_item"},
		{"error", failure, "validateExample_Storage_Message2_At_value_From_Store_Error_bad_5F_request_Field_cause"},
		{"stream", stream, "validateExample_Storage_Message2_At_value_From_Store_StreamingRequest_Field_items_ArrayElement"},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual, visibility := grpcValidationName(&protobufValidationRecord{
				message:     message,
				source:      test.source,
				targetName:  "value",
				contextName: "value",
			}, true)
			require.Equal(t, test.expected, actual)
			require.Equal(t, codegen.UnexportedName, visibility)
		})
	}
}

// TestGRPCValidationNamesDistinguishFieldAndCollectionPaths checks that a
// field name cannot collide with the generated label for an array element.
func TestGRPCValidationNamesDistinguishFieldAndCollectionPaths(t *testing.T) {
	source := grpcTraversalValidationSource()
	direct := source.field("items_element")
	nested := source.field("items").arrayElement()
	message := &protobufMessageRecord{plannedName: "Message"}

	directName, _ := grpcValidationName(&protobufValidationRecord{
		message:     message,
		source:      direct,
		targetName:  "elem",
		contextName: "elem",
	}, true)
	nestedName, _ := grpcValidationName(&protobufValidationRecord{
		message:     message,
		source:      nested,
		targetName:  "elem",
		contextName: "elem",
	}, true)
	require.NotEqual(t, directName, nestedName)

	dashName, _ := grpcValidationName(&protobufValidationRecord{
		message:     message,
		source:      source.field("foo-bar"),
		targetName:  "elem",
		contextName: "elem",
	}, true)
	underscoreName, _ := grpcValidationName(&protobufValidationRecord{
		message:     message,
		source:      source.field("foo_bar"),
		targetName:  "elem",
		contextName: "elem",
	}, true)
	require.NotEqual(t, dashName, underscoreName)
}

// TestCollectValidationsKeepsRootExportedAfterNestedUse checks that endpoint
// order cannot replace a public root validator with a private nested one.
func TestCollectValidationsKeepsRootExportedAfterNestedUse(t *testing.T) {
	names := func(nestedFirst bool) (string, string) {
		minimum := 2
		child := grpcValidationTraversalType("Child", "child", &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{MinLength: &minimum},
		})
		nested := &expr.AttributeExpr{Type: expr.Dup(child)}
		root := &expr.AttributeExpr{Type: expr.Dup(child)}
		container := &expr.AttributeExpr{Type: &expr.Object{
			{Name: "nested", Attribute: nested},
			{Name: "root", Attribute: root},
		}}
		sd := grpcTraversalServiceData()
		freezeTraversalMessages(t, sd, container)
		nestedSource := grpcTraversalValidationSource().field("message")
		rootSource := grpcTraversalValidationSource()
		rootSource.method = "Root"
		if nestedFirst {
			sd.protobuf.collectValidation(nested, validateServer, nestedSource, "message", "message")
			sd.protobuf.collectValidation(root, validateServer, rootSource, "message", "message")
		} else {
			sd.protobuf.collectValidation(root, validateServer, rootSource, "message", "message")
			sd.protobuf.collectValidation(nested, validateServer, nestedSource, "message", "message")
		}
		planTraversalValidations(t, sd)
		sd.validations = sd.protobuf.freezeValidations(sd)

		rootValidation := addValidation(root, sd, true)
		nestedValidation := addValidation(nested, sd, true)
		require.NotSame(t, rootValidation, nestedValidation)
		return rootValidation.Declaration.Name(), nestedValidation.Declaration.Name()
	}

	rootFirst, nestedAfterRoot := names(false)
	rootAfterNested, nestedFirst := names(true)
	require.Equal(t, "ValidateChild", rootFirst)
	require.Equal(t, rootFirst, rootAfterNested)
	require.Equal(t, nestedAfterRoot, nestedFirst)
}

// TestNestedValidationNameIgnoresInsertedSibling checks that adding another
// use of the same message cannot renumber an existing helper.
func TestNestedValidationNameIgnoresInsertedSibling(t *testing.T) {
	name := func(withSibling bool) string {
		root := RunGRPCDSL(t, nestedValidationInsertionDSL(withSibling))
		service := CreateGRPCServices(root).Get("StableValidationNames")
		for _, validation := range service.protobuf.validators {
			if validation.side == validateServer && validation.source.path == "target" {
				return validation.declaration.Name()
			}
		}
		t.Errorf("missing target validator")
		return ""
	}

	require.Equal(t, "validatetest_20_api_StableValidationNames_Child2_At_target", name(false))
	require.Equal(t, "validatetest_20_api_StableValidationNames_Child2_At_target", name(true))
}

func TestAddValidationDistinguishesRequirednessForOneWireDeclaration(t *testing.T) {
	minimum := 2
	original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	first := expr.Dup(original).(expr.UserType)
	second := expr.Dup(original).(expr.UserType)
	first.Attribute().Validation = &expr.ValidationExpr{Required: []string{"value"}}
	expr.AsObject(first).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
	expr.AsObject(second).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
	firstAttribute := &expr.AttributeExpr{Type: first}
	secondAttribute := &expr.AttributeExpr{Type: second}
	sd := grpcTraversalServiceData()
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: firstAttribute},
		{Name: "second", Attribute: secondAttribute},
	}}
	freezeTraversalMessages(t, sd, root)
	firstSource := grpcTraversalValidationSource()
	secondSource := grpcTraversalValidationSource()
	secondSource.method = "OtherCall"
	sd.protobuf.collectValidation(firstAttribute, validateServer, firstSource, "message", "message")
	sd.protobuf.collectValidation(secondAttribute, validateServer, secondSource, "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(firstAttribute, sd, true)
	secondValidation := addValidation(secondAttribute, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.Len(t, sd.validations, 2)
	require.NotEqual(t, firstValidation.Declaration.Name(), secondValidation.Declaration.Name())
	require.Contains(t, firstValidation.Def, `MissingFieldError("value", "message")`)
	require.NotContains(t, secondValidation.Def, "MissingFieldError")
	require.Contains(t, secondValidation.Def, "if message.Value != nil")
}

func TestAddValidationDistinguishesGeneratedSide(t *testing.T) {
	minimum := 2
	message := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	expr.AsObject(message).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
	attribute := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, attribute)
	source := grpcTraversalValidationSource()
	sd.protobuf.collectValidation(attribute, validateServer, source, "message", "message")
	response := source
	response.role = protobufResponseValidation
	sd.protobuf.collectValidation(attribute, validateClient, response, "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	server := addValidation(attribute, sd, true)
	client := addValidation(attribute, sd, false)
	require.NotNil(t, server)
	require.NotNil(t, client)
	require.Len(t, sd.validations, 2)
	require.Equal(t, validateServer, server.Kind)
	require.Equal(t, validateClient, client.Kind)
}

func TestAddValidationKeepsNamesStableForIdenticalRulesAtDifferentPaths(t *testing.T) {
	minimum := 2
	message := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	expr.AsObject(message).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
	first := &expr.AttributeExpr{Type: expr.Dup(message)}
	second := &expr.AttributeExpr{Type: expr.Dup(message)}
	sd := grpcTraversalServiceData()
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: first},
		{Name: "second", Attribute: second},
	}}
	freezeTraversalMessages(t, sd, root)
	source := grpcTraversalValidationSource()
	sd.protobuf.collectValidation(first, validateServer, source.field("first"), "message", "message")
	sd.protobuf.collectValidation(second, validateServer, source.field("second"), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	require.Len(t, sd.validations, 1)
	firstValidation := addValidation(first, sd, true)
	secondValidation := addValidation(second, sd, true)
	require.Same(t, firstValidation, secondValidation)
	require.Equal(t, "validateTestAPI_TestService_Shared_At_message", firstValidation.Declaration.Name())
}

// TestSharedValidationChoosesTheSameSourceAfterTraversalOrderChanges checks
// that a shared body keeps one exact name when a different body requires the
// longer source-based spelling.
func TestSharedValidationChoosesTheSameSourceAfterTraversalOrderChanges(t *testing.T) {
	name := func(reverse bool) string {
		minimum := 2
		original := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
		first := expr.Dup(original).(expr.UserType)
		second := expr.Dup(original).(expr.UserType)
		different := expr.Dup(original).(expr.UserType)
		for _, message := range []expr.UserType{first, second} {
			message.Attribute().Validation = &expr.ValidationExpr{Required: []string{"value"}}
			expr.AsObject(message).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
		}
		expr.AsObject(different).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
		firstAttribute := &expr.AttributeExpr{Type: first}
		secondAttribute := &expr.AttributeExpr{Type: second}
		differentAttribute := &expr.AttributeExpr{Type: different}
		sd := grpcTraversalServiceData()
		root := &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: firstAttribute},
			{Name: "second", Attribute: secondAttribute},
			{Name: "different", Attribute: differentAttribute},
		}}
		freezeTraversalMessages(t, sd, root)
		source := grpcTraversalValidationSource()
		ordered := []struct {
			attribute *expr.AttributeExpr
			path      string
		}{
			{attribute: firstAttribute, path: "zeta"},
			{attribute: secondAttribute, path: "alpha"},
		}
		if reverse {
			ordered[0], ordered[1] = ordered[1], ordered[0]
		}
		for _, item := range ordered {
			sd.protobuf.collectValidation(item.attribute, validateServer, source.field(item.path), "message", "message")
		}
		sd.protobuf.collectValidation(differentAttribute, validateServer, source.field("different"), "message", "message")
		planTraversalValidations(t, sd)
		sd.validations = sd.protobuf.freezeValidations(sd)

		validation := addValidation(firstAttribute, sd, true)
		require.NotNil(t, validation)
		return validation.Declaration.Name()
	}

	want := "validateTestAPI_TestService_Shared_At_message_From_Call_Request_Field_alpha"
	require.Equal(t, want, name(false))
	require.Equal(t, want, name(true))
}

// TestAddValidationKeepsDistinctErrorPaths checks that one shared protobuf
// message gets separate validators when its callers need different field paths.
func TestAddValidationKeepsDistinctErrorPaths(t *testing.T) {
	minimum := 2
	shared := grpcValidationTraversalType("Shared2", "shared", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	})
	first := &expr.AttributeExpr{Type: shared, Meta: expr.MetaExpr{"rpc:tag": {"1"}}}
	secondElement := &expr.AttributeExpr{Type: shared}
	second := &expr.AttributeExpr{
		Type: &expr.Array{ElemType: secondElement},
		Meta: expr.MetaExpr{"rpc:tag": {"2"}},
	}
	rootType := &expr.UserTypeExpr{
		TypeName: "Root",
		UID:      "root",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: first},
			{Name: "second", Attribute: second},
		}},
	}
	root := &expr.AttributeExpr{Type: rootType}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, root)
	sd.protobuf.collectValidation(root, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(first, sd, true)
	secondValidation := addValidation(secondElement, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.NotSame(t, firstValidation, secondValidation)
	require.Equal(t, "validateTestAPI_TestService_Shared2_At_first", firstValidation.Declaration.Name())
	require.Equal(t, "validateTestAPI_TestService_Shared2_At_elem", secondValidation.Declaration.Name())
	require.Contains(t, firstValidation.Def, `InvalidLengthError("first.value"`)
	require.Contains(t, secondValidation.Def, `InvalidLengthError("elem.value"`)
	rootValidation := addValidation(root, sd, true)
	require.Equal(t, "ValidateRoot", rootValidation.Declaration.Name())
	require.Contains(t, rootValidation.Def, firstValidation.Declaration.Name()+"(message.First)")
	require.Contains(t, rootValidation.Def, secondValidation.Declaration.Name()+"(e)")
}

// TestCollectValidationsSharesRecursiveSiblings checks that identical array
// element checks share a name that does not depend on the first sibling.
func TestCollectValidationsSharesRecursiveSiblings(t *testing.T) {
	minimum := 2
	child := grpcValidationTraversalType("Child", "child", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	})
	childObject := expr.AsObject(child)
	*childObject = append(*childObject, &expr.NamedAttributeExpr{
		Name: "next",
		Attribute: &expr.AttributeExpr{
			Type: child,
			Meta: expr.MetaExpr{"rpc:tag": {"2"}},
		},
	})
	firstElement := &expr.AttributeExpr{Type: child}
	secondElement := &expr.AttributeExpr{Type: child}
	root := &expr.AttributeExpr{Type: &expr.UserTypeExpr{
		TypeName: "Root",
		UID:      "root",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: &expr.AttributeExpr{
				Type: &expr.Array{ElemType: firstElement},
				Meta: expr.MetaExpr{"rpc:tag": {"1"}},
			}},
			{Name: "second", Attribute: &expr.AttributeExpr{
				Type: &expr.Array{ElemType: secondElement},
				Meta: expr.MetaExpr{"rpc:tag": {"2"}},
			}},
		}},
	}}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(t, sd, root)
	sd.protobuf.collectValidation(root, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(firstElement, sd, true)
	secondValidation := addValidation(secondElement, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.Same(t, firstValidation, secondValidation)
	require.Equal(t, "validateTestAPI_TestService_Child_At_elem", firstValidation.Declaration.Name())
}

func TestCollectValidationsDistinguishesEqualUIDOrigins(t *testing.T) {
	minimumLength := 3
	minimum := 5.0
	first := grpcValidationTraversalType("First", "shared", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimumLength},
	})
	second := grpcValidationTraversalType("Second", "shared", &expr.AttributeExpr{
		Type:       expr.Int,
		Validation: &expr.ValidationExpr{Minimum: &minimum},
	})
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}
	sd := &ServiceData{PkgName: "pb", Scope: codegen.NewNameScope()}

	freezeTraversalMessages(t, sd, root)
	sd.protobuf.collectValidation(root, validateServer, grpcTraversalValidationSource(), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)
	names := make([]string, 0, len(sd.validations))
	for _, validation := range sd.validations {
		names = append(names, validation.SrcName)
	}
	require.ElementsMatch(t, []string{"First", "Second"}, names)
}

// grpcTraversalValidationSource describes the request used by focused
// validation tests.
func grpcTraversalValidationSource() protobufValidationSource {
	return protobufValidationSource{
		api:     "TestAPI",
		service: "TestService",
		method:  "Call",
		role:    protobufRequestValidation,
	}
}

// planTraversalValidations chooses the function names used by these focused
// message and validation tests before building the function bodies.
func planTraversalValidations(t *testing.T, sd *ServiceData) {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	client, err := generation.ClaimPackage("generated.local/gen/grpc/test/client")
	require.NoError(t, err)
	server, err := generation.ClaimPackage("generated.local/gen/grpc/test/server")
	require.NoError(t, err)
	privateNameCounts := make(map[grpcValidationNameKey]int)
	for _, record := range sd.protobuf.validators {
		if record.source.path != "" {
			privateNameCounts[grpcValidationKey(record)]++
		}
	}
	for _, record := range sd.protobuf.validators {
		pkg := client
		side := grpcClientPackage
		if record.side == validateServer {
			pkg = server
			side = grpcServerPackage
		}
		id := grpcSymbolID{
			side:      side,
			role:      grpcValidationRole,
			api:       record.source.api,
			service:   record.source.service,
			method:    record.source.method,
			subject:   record.source.error,
			view:      record.source.view,
			path:      record.source.path,
			operation: int(record.source.role),
		}
		record.declaration = grpcValidationDeclaration(
			record,
			id,
			privateNameCounts[grpcValidationKey(record)] > 1,
		)
		require.NoError(t, pkg.DeclareName(record.declaration))
	}
	require.NoError(t, generation.Freeze())
}

// grpcValidationTraversalType builds an authored message declaration with one
// constrained field so validation discovery must emit a helper for it.
func grpcValidationTraversalType(name, uid string, field *expr.AttributeExpr) *expr.UserTypeExpr {
	if field.Meta == nil {
		field.Meta = make(expr.MetaExpr)
	}
	field.Meta["rpc:tag"] = []string{"1"}
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: field},
		}},
		TypeName: name,
		UID:      uid,
	}
}

// nestedValidationInsertionDSL builds the same target field with or without an
// earlier sibling that uses the same nested type.
func nestedValidationInsertionDSL(withSibling bool) func() {
	return func() {
		child := dsl.Type("Child2", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Required("value")
		})
		dsl.Service("StableValidationNames", func() {
			dsl.Method("Store", func() {
				dsl.Payload(func() {
					if withSibling {
						dsl.Field(1, "alpha", child)
					}
					dsl.Field(2, "target", child)
					dsl.Required("target")
				})
				dsl.GRPC(func() {})
			})
		})
	}
}

// grpcMessageTraversalType builds a protobuf message declaration with one
// explicitly numbered field.
func grpcMessageTraversalType(name, uid string, fieldType expr.DataType, tag string) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: &expr.AttributeExpr{
				Type: fieldType,
				Meta: expr.MetaExpr{"rpc:tag": {tag}},
			}},
		}},
		TypeName: name,
		UID:      uid,
	}
}

// grpcTraversalServiceData supplies the protobuf package and field scope used
// by focused declaration and validator catalog tests.
func grpcTraversalServiceData() *ServiceData {
	return &ServiceData{
		Name:    "Service",
		PkgName: "servicepb",
		Scope:   codegen.NewNameScope(),
		Service: &service.Data{},
	}
}

// freezeTraversalMessages collects and freezes every message reachable from
// root in the focused test protobuf package.
func freezeTraversalMessages(t *testing.T, sd *ServiceData, root *expr.AttributeExpr) []*service.UserTypeData {
	sd.protobuf = newProtobufPackageCatalog(sd.PkgName)
	require.NoError(t, sd.protobuf.collectMessage(root, protobufMessageSource{}))
	planTestProtobufCatalog(t, sd)
	sd.Messages = sd.protobuf.freezeMessages(sd)
	return sd.Messages
}

// planTestProtobufCatalog chooses names for the messages and validation
// functions created directly by these focused tests.
func planTestProtobufCatalog(t *testing.T, sd *ServiceData) {
	t.Helper()
	require.NotEmpty(t, sd.protobuf.messages)
	message := sd.protobuf.messages[0].uses[0]
	serviceExpr := &expr.ServiceExpr{Name: "GoaCatalogTestService"}
	grpcService := &expr.GRPCServiceExpr{ServiceExpr: serviceExpr}
	method := &expr.MethodExpr{
		Name:             "Call",
		Service:          serviceExpr,
		Payload:          &expr.AttributeExpr{Type: expr.Empty},
		StreamingPayload: &expr.AttributeExpr{Type: expr.Empty},
		Result:           &expr.AttributeExpr{Type: expr.Empty},
		StreamingResult:  &expr.AttributeExpr{Type: expr.Empty},
	}
	endpoint := &expr.GRPCEndpointExpr{MethodExpr: method, Service: grpcService}
	grpcService.GRPCEndpoints = []*expr.GRPCEndpointExpr{endpoint}
	plan := &protobufServicePlan{
		expression:   grpcService,
		catalog:      sd.protobuf,
		messages:     []*protobufEndpointMessages{{request: message, response: message}},
		protoPackage: "goa_catalog_test",
		methods:      map[*expr.GRPCEndpointExpr]string{},
		names:        make(map[protocNameKey]*codegen.NameDeclaration),
		localNames:   make(map[protocNameKey]string),
		fields:       make(map[*expr.AttributeExpr]protocNameKey),
		sourceFields: make(map[*expr.AttributeExpr]string),
		sourceOneofs: make(map[*expr.AttributeExpr]string),
		wrappers:     make(map[*expr.AttributeExpr]protocNameKey),
		oneofs:       make(map[*expr.AttributeExpr]protocNameKey),
	}
	sd.protobuf.plan = plan
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	pkg, err := generation.ClaimPackage("generated.local/gen/grpc/test/pb")
	require.NoError(t, err)
	require.NoError(t, plan.chooseNames(pkg, make(map[string]struct{})))
	require.NoError(t, generation.Freeze())
}
