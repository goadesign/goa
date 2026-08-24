// This file verifies that gRPC validation discovery distinguishes unrelated
// declarations while stopping recursion through copied declarations.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
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
	sd.protobuf.collectValidation(firstAttribute, validateServer, source, "message", "message")
	sd.protobuf.collectValidation(secondAttribute, validateServer, source.child("second"), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(firstAttribute, sd, true)
	secondValidation := addValidation(secondAttribute, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.Len(t, sd.validations, 2)
	require.NotEqual(t, firstValidation.Declaration.Name(), secondValidation.Declaration.Name())
	require.Contains(t, firstValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 2, true)`)
	require.Contains(t, secondValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 5, true)`)
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

func TestAddValidationReusesIdenticalRulesOnOneSide(t *testing.T) {
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
	sd.protobuf.collectValidation(first, validateServer, source, "message", "message")
	sd.protobuf.collectValidation(second, validateServer, source.child("second"), "message", "message")
	planTraversalValidations(t, sd)
	sd.validations = sd.protobuf.freezeValidations(sd)

	require.Len(t, sd.validations, 1)
	require.Same(t, addValidation(first, sd, true), addValidation(second, sd, true))
}

// TestAddValidationKeepsDistinctErrorPaths checks that one shared protobuf
// message gets separate validators when its callers need different field paths.
func TestAddValidationKeepsDistinctErrorPaths(t *testing.T) {
	minimum := 2
	shared := grpcValidationTraversalType("Shared", "shared", &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minimum},
	})
	first := &expr.AttributeExpr{Type: shared, Meta: expr.MetaExpr{"rpc:tag": {"1"}}}
	second := &expr.AttributeExpr{Type: shared, Meta: expr.MetaExpr{"rpc:tag": {"2"}}}
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
	secondValidation := addValidation(second, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.NotSame(t, firstValidation, secondValidation)
	require.Contains(t, firstValidation.Def, `InvalidLengthError("first.value"`)
	require.Contains(t, secondValidation.Def, `InvalidLengthError("second.value"`)
	rootValidation := addValidation(root, sd, true)
	require.Contains(t, rootValidation.Def, firstValidation.Declaration.Name()+"(message.First)")
	require.Contains(t, rootValidation.Def, secondValidation.Declaration.Name()+"(message.Second)")
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
		preferred := "Validate" + record.message.plannedName
		if record.source.view != "" && record.source.view != expr.DefaultView {
			preferred += codegen.Goify(record.source.view, true)
		}
		record.declaration = codegen.NewPreferredName(
			codegen.NameFunction,
			preferred,
			codegen.ExportedName,
			grpcSymbolOrder(id),
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
