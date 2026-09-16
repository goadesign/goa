// This file reads the Go names that the supported protobuf tools assign to a
// compiled protobuf file. The gRPC planner uses these names when it records
// declarations that later files define or call.
package codegen

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type (
	// protocNameRules reads names using one supported pair of protobuf tools.
	protocNameRules struct{}

	// protocNames stores each Go name under the protobuf item and the way that
	// generated code uses it.
	protocNames struct {
		values map[protocNameKey]string
	}

	// protocNameKey identifies one Go name produced for a protobuf item.
	protocNameKey struct {
		descriptor string
		role       protocNameRole
	}

	// protocNameRole identifies one Go declaration or field produced by the
	// supported protobuf tools.
	protocNameRole uint8
)

const (
	protocMessageName protocNameRole = iota + 1
	protocFieldName
	protocEnumName
	protocEnumValueName
	protocOneofFieldName
	protocOneofInterfaceName
	protocOneofWrapperName
	protocServiceClientName
	protocServiceClientStructName
	protocServiceClientConstructorName
	protocServiceServerName
	protocServiceUnimplementedServerName
	protocServiceUnsafeServerName
	protocServiceRegisterName
	protocServiceDescriptorName
	protocMethodName
	protocMethodFullName
	protocMethodHandlerName
	protocMethodClientStreamName
	protocMethodServerStreamName
)

const protocNameVersionGo1_36GRPC1_6 = "protoc-gen-go-v1.36.12/protoc-gen-go-grpc-v1.6.2"

// newProtocNameRules returns the name reader for a supported protobuf tool
// pair. An unknown value is rejected because its Go names may differ.
func newProtocNameRules(version string) (*protocNameRules, error) {
	if version != protocNameVersionGo1_36GRPC1_6 {
		return nil, fmt.Errorf("unsupported protobuf Go naming version %q", version)
	}
	return &protocNameRules{}, nil
}

// file returns every Go name used for messages, fields, enumerations, oneofs,
// services, and methods in descriptor.
func (r *protocNameRules) file(descriptor *descriptorpb.FileDescriptorProto) (*protocNames, error) {
	request := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{descriptor.GetName()},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{descriptor},
	}
	plugin, err := (protogen.Options{}).New(request)
	if err != nil {
		return nil, fmt.Errorf("read protobuf Go names: %w", err)
	}
	if len(plugin.Files) != 1 || plugin.Files[0].Desc.Path() != descriptor.GetName() {
		return nil, fmt.Errorf("read protobuf Go names: file %q was not returned", descriptor.GetName())
	}

	names := &protocNames{values: make(map[protocNameKey]string)}
	file := plugin.Files[0]
	for _, enum := range file.Enums {
		if err := names.addEnum(enum); err != nil {
			return nil, err
		}
	}
	for _, message := range file.Messages {
		if err := names.addMessage(message); err != nil {
			return nil, err
		}
	}
	for _, service := range file.Services {
		if err := names.addService(service); err != nil {
			return nil, err
		}
	}
	return names, nil
}

// lookup returns the Go name stored for descriptor and role.
func (n *protocNames) lookup(descriptor string, role protocNameRole) (string, bool) {
	name, ok := n.values[protocNameKey{descriptor: descriptor, role: role}]
	return name, ok
}

// String returns the short role name used in test and error labels.
func (r protocNameRole) String() string {
	switch r {
	case protocMessageName:
		return "message"
	case protocFieldName:
		return "field"
	case protocEnumName:
		return "enum"
	case protocEnumValueName:
		return "enum value"
	case protocOneofFieldName:
		return "oneof field"
	case protocOneofInterfaceName:
		return "oneof interface"
	case protocOneofWrapperName:
		return "oneof wrapper"
	case protocServiceClientName:
		return "service client"
	case protocServiceClientStructName:
		return "service client struct"
	case protocServiceClientConstructorName:
		return "service client constructor"
	case protocServiceServerName:
		return "service server"
	case protocServiceUnimplementedServerName:
		return "unimplemented service server"
	case protocServiceUnsafeServerName:
		return "unsafe service server"
	case protocServiceRegisterName:
		return "service register function"
	case protocServiceDescriptorName:
		return "service description"
	case protocMethodName:
		return "method"
	case protocMethodFullName:
		return "full method name"
	case protocMethodHandlerName:
		return "method handler"
	case protocMethodClientStreamName:
		return "client stream"
	case protocMethodServerStreamName:
		return "server stream"
	default:
		panic(fmt.Sprintf("unknown protobuf Go name role %d", r))
	}
}

// addMessage stores one message, its nested declarations, fields, and oneofs.
func (n *protocNames) addMessage(message *protogen.Message) error {
	if err := n.add(string(message.Desc.FullName()), protocMessageName, message.GoIdent.GoName); err != nil {
		return err
	}
	for _, enum := range message.Enums {
		if err := n.addEnum(enum); err != nil {
			return err
		}
	}
	for _, nested := range message.Messages {
		if err := n.addMessage(nested); err != nil {
			return err
		}
	}
	for _, field := range message.Fields {
		descriptor := string(field.Desc.FullName())
		if err := n.add(descriptor, protocFieldName, field.GoName); err != nil {
			return err
		}
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			if err := n.add(descriptor, protocOneofWrapperName, field.GoIdent.GoName); err != nil {
				return err
			}
		}
	}
	for _, oneof := range message.Oneofs {
		if oneof.Desc.IsSynthetic() {
			continue
		}
		descriptor := string(oneof.Desc.FullName())
		if err := n.add(descriptor, protocOneofFieldName, oneof.GoName); err != nil {
			return err
		}
		if err := n.add(descriptor, protocOneofInterfaceName, "is"+oneof.GoIdent.GoName); err != nil {
			return err
		}
	}
	return nil
}

// addEnum stores one enumeration and all of its values.
func (n *protocNames) addEnum(enum *protogen.Enum) error {
	if err := n.add(string(enum.Desc.FullName()), protocEnumName, enum.GoIdent.GoName); err != nil {
		return err
	}
	for _, value := range enum.Values {
		if err := n.add(string(value.Desc.FullName()), protocEnumValueName, value.GoIdent.GoName); err != nil {
			return err
		}
	}
	return nil
}

// addService stores the declarations written by protoc-gen-go-grpc v1.6 for
// one service and all of its methods.
func (n *protocNames) addService(service *protogen.Service) error {
	descriptor := string(service.Desc.FullName())
	serviceName := service.GoName
	declarations := []struct {
		role protocNameRole
		name string
	}{
		{protocServiceClientName, serviceName + "Client"},
		{protocServiceClientStructName, protocGRPCUnexport(serviceName) + "Client"},
		{protocServiceClientConstructorName, "New" + serviceName + "Client"},
		{protocServiceServerName, serviceName + "Server"},
		{protocServiceUnimplementedServerName, "Unimplemented" + serviceName + "Server"},
		{protocServiceUnsafeServerName, "Unsafe" + serviceName + "Server"},
		{protocServiceRegisterName, "Register" + serviceName + "Server"},
		{protocServiceDescriptorName, serviceName + "_ServiceDesc"},
	}
	for _, declaration := range declarations {
		if err := n.add(descriptor, declaration.role, declaration.name); err != nil {
			return err
		}
	}
	for _, method := range service.Methods {
		methodDescriptor := string(method.Desc.FullName())
		methodName := method.GoName
		if err := n.add(methodDescriptor, protocMethodName, methodName); err != nil {
			return err
		}
		if err := n.add(methodDescriptor, protocMethodFullName, serviceName+"_"+methodName+"_FullMethodName"); err != nil {
			return err
		}
		if err := n.add(methodDescriptor, protocMethodHandlerName, "_"+serviceName+"_"+methodName+"_Handler"); err != nil {
			return err
		}
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			if err := n.add(methodDescriptor, protocMethodClientStreamName, serviceName+"_"+methodName+"Client"); err != nil {
				return err
			}
			if err := n.add(methodDescriptor, protocMethodServerStreamName, serviceName+"_"+methodName+"Server"); err != nil {
				return err
			}
		}
	}
	return nil
}

// add stores one name and rejects two values for the same protobuf item and
// role.
func (n *protocNames) add(descriptor string, role protocNameRole, name string) error {
	key := protocNameKey{descriptor: descriptor, role: role}
	if previous, ok := n.values[key]; ok {
		return fmt.Errorf("protobuf item %q has two %s names, %q and %q", descriptor, role, previous, name)
	}
	n.values[key] = name
	return nil
}

// protocGRPCUnexport changes the first letter exactly as protoc-gen-go-grpc
// v1.6 does when it writes the private client type.
func protocGRPCUnexport(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}
