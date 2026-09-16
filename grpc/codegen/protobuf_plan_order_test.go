// This file checks that protobuf service ordering has one clear result for
// every generation input.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestSortProtobufMessageGroupsUsesWireDetails checks that field numbers and
// explicit protobuf names give messages the same order when the input is reversed.
func TestSortProtobufMessageGroupsUsesWireDetails(t *testing.T) {
	tests := []struct {
		name   string
		groups func() []*protobufNameGroup
		value  func(*protobufNameGroup) string
		want   []string
	}{
		{
			name: "field numbers",
			groups: func() []*protobufNameGroup {
				source := protobufOrderUserType()
				return []*protobufNameGroup{
					protobufOrderGroup(source, "Message", false, "2"),
					protobufOrderGroup(source, "Message", false, "1"),
				}
			},
			value: func(group *protobufNameGroup) string {
				field := expr.AsObject(group.message.identity.attribute.Type).Attribute("value")
				return field.Meta["rpc:tag"][0]
			},
			want: []string{"1", "2"},
		},
		{
			name: "explicit names",
			groups: func() []*protobufNameGroup {
				source := protobufOrderUserType()
				return []*protobufNameGroup{
					protobufOrderGroup(source, "Zulu", true, "1"),
					protobufOrderGroup(source, "Alpha", true, "1"),
				}
			},
			value: func(group *protobufNameGroup) string {
				return group.message.identity.preferredName
			},
			want: []string{"Alpha", "Zulu"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			forward := test.groups()
			reverse := test.groups()
			reverse[0], reverse[1] = reverse[1], reverse[0]

			require.NoError(t, sortProtobufMessageGroups(forward))
			require.NoError(t, sortProtobufMessageGroups(reverse))
			require.Equal(t, test.want, protobufOrderValues(forward, test.value))
			require.Equal(t, test.want, protobufOrderValues(reverse, test.value))
		})
	}
}

// TestSortProtobufMessageGroupsRejectsEqualOrder checks that two separate
// declarations cannot receive names according to their input order.
func TestSortProtobufMessageGroupsRejectsEqualOrder(t *testing.T) {
	left := protobufOrderUserType()
	right := protobufOrderUserType()
	groups := []*protobufNameGroup{
		protobufOrderGroup(left, "Message", false, "1"),
		protobufOrderGroup(right, "Message", false, "1"),
	}

	err := sortProtobufMessageGroups(groups)
	require.EqualError(t, err, `protobuf messages named "Message" have the same source, name, and fields but come from separate declarations`)
}

// TestProtobufMessageIdentityIgnoresRequiredFields checks that validation rules
// do not split messages with the same protobuf fields.
func TestProtobufMessageIdentityIgnoresRequiredFields(t *testing.T) {
	source := protobufOrderUserType()
	left := protobufRequiredOrderGroup(source, "a")
	right := protobufRequiredOrderGroup(source, "b")

	forward := compareProtobufMessageIdentity(left.message.identity, right.message.identity)
	backward := compareProtobufMessageIdentity(right.message.identity, left.message.identity)
	require.Zero(t, forward)
	require.Zero(t, backward)
	require.True(t, sameProtobufMessageIdentity(left.message.identity, right.message.identity))
	require.False(t, sameProtobufValidationAttribute(
		left.message.identity.attribute,
		right.message.identity.attribute,
		make(map[protobufAttributePair]struct{}),
	))
}

// TestNextAvailableProtobufNameUsesRetainedSuffix checks that collision retries
// do not parse the generated name Goa stored for the group.
func TestNextAvailableProtobufNameUsesRetainedSuffix(t *testing.T) {
	group := &protobufNameGroup{preferred: "Message", name: "unrelated"}
	occupied := &protobufNameGroup{preferred: "Message", name: "Message2"}

	next := nextAvailableProtobufName(group, []*protobufNameGroup{group, occupied})
	require.Equal(t, "Message3", next)
}

// TestPlanProtobufServicesRejectsEqualOrder checks that two separate services
// cannot receive names according to their input order.
func TestPlanProtobufServicesRejectsEqualOrder(t *testing.T) {
	roots := grpcPlanRoots(t, "Shared", "Shared")
	roots[0].API.Name = "Shared API"
	roots[1].API.Name = "Shared API"
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{roots[0], roots[1]})
	require.NoError(t, err)
	plans := []*Plan{
		{
			root:        roots[0],
			expressions: roots[0].API.GRPC.Services,
			protobuf:    make(map[*expr.GRPCServiceExpr]*protobufServicePlan),
			packages: map[*expr.GRPCServiceExpr]*grpcServicePackage{
				roots[0].API.GRPC.Services[0]: {pathName: "shared"},
			},
		},
		{
			root:        roots[1],
			expressions: roots[1].API.GRPC.Services,
			protobuf:    make(map[*expr.GRPCServiceExpr]*protobufServicePlan),
			packages: map[*expr.GRPCServiceExpr]*grpcServicePackage{
				roots[1].API.GRPC.Services[0]: {pathName: "shared"},
			},
		},
	}

	err = planProtobufServices(generation, plans)
	require.EqualError(t, err, `generated protobuf package "generated.local/gen/grpc/shared/pb" has two services named "Shared" in API "Shared API"`)
}

// protobufOrderUserType creates one separate source declaration for order
// tests.
func protobufOrderUserType() *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		TypeName: "Message",
		UID:      "message",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{
				Name: "value",
				Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"rpc:tag": {"1"}},
				},
			},
		}},
	}
}

// protobufOrderGroup creates one message with the requested name and field
// number.
func protobufOrderGroup(source expr.UserType, name string, explicit bool, tag string) *protobufNameGroup {
	attribute := expr.DupAtt(source.Attribute())
	expr.AsObject(attribute.Type).Attribute("value").Meta["rpc:tag"] = []string{tag}
	return &protobufNameGroup{
		preferred: name,
		message: &protobufMessageRecord{identity: protobufMessageIdentity{
			source:        protobufMessageSource{origin: source},
			preferredName: name,
			explicitName:  explicit,
			userType:      source,
			attribute:     attribute,
		}},
	}
}

// protobufRequiredOrderGroup creates one message with a selected required field.
func protobufRequiredOrderGroup(source expr.UserType, required string) *protobufNameGroup {
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "a", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "b", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
		Validation: &expr.ValidationExpr{Required: []string{required}},
	}
	return &protobufNameGroup{
		preferred: "Message",
		message: &protobufMessageRecord{identity: protobufMessageIdentity{
			source:        protobufMessageSource{origin: source},
			preferredName: "Message",
			userType:      source,
			attribute:     attribute,
		}},
	}
}

// protobufOrderValues reads the fact checked by one order test from every
// message.
func protobufOrderValues(groups []*protobufNameGroup, value func(*protobufNameGroup) string) []string {
	values := make([]string, len(groups))
	for index, group := range groups {
		values[index] = value(group)
	}
	return values
}
