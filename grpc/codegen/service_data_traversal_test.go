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

	messages := freezeTraversalMessages(sd, root)
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

	messages := freezeTraversalMessages(sd, root)
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

	messages := freezeTraversalMessages(grpcTraversalServiceData(), root)
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

	messages := freezeTraversalMessages(grpcTraversalServiceData(), root)
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

	messages := freezeTraversalMessages(grpcTraversalServiceData(), root)
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

	messages := freezeTraversalMessages(grpcTraversalServiceData(), root)
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
	sd.protobuf.collectMessage(firstWire, protobufRootMessageSource(firstWire, firstEndpoint, nil, protobufResponseMessage), sd)
	sd.protobuf.collectMessage(secondWire, protobufRootMessageSource(secondWire, secondEndpoint, nil, protobufResponseMessage), sd)

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

	messages := freezeTraversalMessages(grpcTraversalServiceData(), &expr.AttributeExpr{Type: expr.Dup(message)})
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
	messages := freezeTraversalMessages(sd, attribute)
	require.Len(t, messages, 1)

	sd.Scope.HashedUnique(grpcMessageTraversalType("Other", "other", expr.Int, "1"), "Shared")
	require.Equal(t, messages[0].VarName, protoBufMessageName(attribute, sd))
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
	freezeTraversalMessages(sd, root)
	sd.protobuf.collectValidation(firstAttribute, validateServer, "message", "message")
	sd.protobuf.collectValidation(secondAttribute, validateServer, "message", "message")
	sd.validations = sd.protobuf.freezeValidations(sd)

	firstValidation := addValidation(firstAttribute, sd, true)
	secondValidation := addValidation(secondAttribute, sd, true)
	require.NotNil(t, firstValidation)
	require.NotNil(t, secondValidation)
	require.Len(t, sd.validations, 2)
	require.NotEqual(t, firstValidation.Name, secondValidation.Name)
	require.Contains(t, firstValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 2, true)`)
	require.Contains(t, secondValidation.Def, `InvalidLengthError("message.value", *message.Value, utf8.RuneCountInString(*message.Value), 5, true)`)
}

func TestAddValidationDistinguishesGeneratedSide(t *testing.T) {
	minimum := 2
	message := grpcMessageTraversalType("Shared", "shared", expr.String, "1")
	expr.AsObject(message).Attribute("value").Validation = &expr.ValidationExpr{MinLength: &minimum}
	attribute := &expr.AttributeExpr{Type: message}
	sd := grpcTraversalServiceData()
	freezeTraversalMessages(sd, attribute)
	sd.protobuf.collectValidation(attribute, validateServer, "message", "message")
	sd.protobuf.collectValidation(attribute, validateClient, "message", "message")
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
	freezeTraversalMessages(sd, root)
	sd.protobuf.collectValidation(first, validateServer, "message", "message")
	sd.protobuf.collectValidation(second, validateServer, "message", "message")
	sd.validations = sd.protobuf.freezeValidations(sd)

	require.Len(t, sd.validations, 1)
	require.Same(t, addValidation(first, sd, true), addValidation(second, sd, true))
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

	freezeTraversalMessages(sd, root)
	sd.protobuf.collectValidation(root, validateServer, "message", "message")
	sd.validations = sd.protobuf.freezeValidations(sd)
	var names []string
	for _, validation := range sd.validations {
		names = append(names, validation.SrcName)
	}
	require.ElementsMatch(t, []string{"First", "Second"}, names)
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
func freezeTraversalMessages(sd *ServiceData, root *expr.AttributeExpr) []*service.UserTypeData {
	sd.protobuf = newProtobufPackageCatalog(sd.PkgName)
	sd.protobuf.collectMessage(root, protobufMessageSource{}, sd)
	sd.Messages = sd.protobuf.freezeMessages(sd)
	return sd.Messages
}
