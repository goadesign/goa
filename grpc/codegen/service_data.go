// This file analyzes gRPC endpoint designs into the immutable data consumed by
// protobuf message, client, server, conversion, and validation templates.
package codegen

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// ServicesData contains the data computed from the gRPC service expressions
	// indexed by service name.
	ServicesData struct {
		*service.ServicesData
		GRPCServices  map[string]*ServiceData
		cliPlan       *grpcCLIPlan
		protobuf      map[*expr.GRPCServiceExpr]*protobufServicePlan
		tools         map[*expr.GRPCServiceExpr]*protobufToolPlan
		symbols       map[*expr.GRPCServiceExpr]*grpcSymbols
		expressions   []*expr.GRPCServiceExpr
		servicePlans  []*grpcServicePlan
		serviceByExpr map[*expr.GRPCServiceExpr]*ServiceData
		endpointPlans map[*expr.GRPCEndpointExpr]*grpcEndpointPlan
		metadataPlans map[*expr.MappedAttributeExpr][]*grpcMetadataPlan
		fileImports   map[string]*codegen.GeneratedImportPlan
		generation    *codegen.Generation
	}

	// ServiceData contains the data used to render the code related to a
	// single service.
	ServiceData struct {
		// Service contains the related service data.
		Service *service.Data
		// ClientPkgName is the final alias for the generated gRPC client package.
		ClientPkgName string
		// ServerPkgName is the final alias for the generated gRPC server package.
		ServerPkgName string
		// PkgName is the name of the generated package in *.pb.go.
		PkgName string
		// ClientProtobufPkgName is the protobuf import name in the generated client package.
		ClientProtobufPkgName string
		// ServerProtobufPkgName is the protobuf import name in the generated server package.
		ServerProtobufPkgName string
		// ClientServicePkgName is the service import name in the generated client package.
		ClientServicePkgName string
		// ServerServicePkgName is the service import name in the generated server package.
		ServerServicePkgName string
		// ProtoImports is the list of proto package imports.
		ProtoImports []string
		// Name is the service name.
		Name string
		// Description is the service description.
		Description string
		// Endpoints describes the gRPC service endpoints.
		Endpoints []*EndpointData
		// Messages describes the message data for this service.
		Messages []*service.UserTypeData
		// ServerStruct is the server type name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerStructDeclaration.Name() after planning.
		ServerStruct string
		// ServerStructDeclaration supplies the generated server type name.
		ServerStructDeclaration *codegen.NameDeclaration
		// ClientStruct is the client type name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning.
		ClientStruct string
		// ClientStructDeclaration supplies the generated client type name.
		ClientStructDeclaration *codegen.NameDeclaration
		// ServerInit is the server constructor name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerInitDeclaration.Name() after planning.
		ServerInit string
		// ServerInitDeclaration supplies the generated server constructor name.
		ServerInitDeclaration *codegen.NameDeclaration
		// ClientInit is the client constructor name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientInitDeclaration.Name() after planning.
		ClientInit string
		// ClientInitDeclaration supplies the generated client constructor name.
		ClientInitDeclaration *codegen.NameDeclaration
		// ServerInterface is the name of the gRPC server interface implemented
		// by the service.
		ServerInterface string
		// ClientInterface is the name of the gRPC client interface implemented
		// by the service.
		ClientInterface string
		// ClientInterfaceInit is the name of the client constructor function in
		// the generated pb.go package.
		ClientInterfaceInit string
		// UnimplementedServer is the generated server type embedded by Goa's
		// server implementation.
		UnimplementedServer string
		// RegisterFunction is the generated function that registers the server.
		RegisterFunction string
		// Scope records and returns unique Go names for protobuf fields and types in
		// this service package.
		Scope *codegen.NameScope

		// protobuf contains the messages and validation functions written for this
		// service.
		protobuf *protobufPackageCatalog

		// clientTransformHelpers contains recursive conversion functions written
		// in the generated client package.
		clientTransformHelpers []*codegen.TransformFunctionData
		// serverTransformHelpers contains recursive conversion functions written
		// in the generated server package.
		serverTransformHelpers []*codegen.TransformFunctionData
		// validations contain the data to generate the validation functions to
		// validate the initialized type.
		validations []*ValidationData
	}

	// EndpointData contains the data used to render the code related to
	// gRPC endpoint.
	EndpointData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// PkgName is the name of the generated package in *.pb.go.
		PkgName string
		// ServicePkgName is the name of the service package name.
		ServicePkgName string
		// ClientProtobufPkgName is the protobuf import name in the generated client package.
		ClientProtobufPkgName string
		// ServerProtobufPkgName is the protobuf import name in the generated server package.
		ServerProtobufPkgName string
		// ClientServicePkgName is the service import name in the generated client package.
		ClientServicePkgName string
		// ServerServicePkgName is the service import name in the generated server package.
		ServerServicePkgName string
		// Method is the data for the underlying method expression.
		Method *service.MethodData
		// ProtoMethodName is the method name written to the protobuf service.
		ProtoMethodName string
		// ClientMethodName is the final protobuf client method name kept for
		// existing plugins.
		//
		// Deprecated: Use ProtoMethodName.
		ClientMethodName string
		// FullMethodName is the protobuf service and method name logged when the
		// generated server starts.
		FullMethodName string
		// PayloadType is the type of the payload.
		PayloadType expr.DataType
		// PayloadRef is the fully qualified reference to the method payload.
		PayloadRef string
		// ClientPayloadRef is the payload reference in the generated client package.
		ClientPayloadRef string
		// ServerPayloadRef is the payload reference in the generated server package.
		ServerPayloadRef string
		// ResultRef is the fully qualified reference to the method result.
		ResultRef string
		// ClientResultRef is the result reference in the generated client package.
		ClientResultRef string
		// ServerResultRef is the result reference in the generated server package.
		ServerResultRef string
		// ViewedResultRef is the fully qualified reference to the viewed result.
		ViewedResultRef string
		// Request is the gRPC request data.
		Request *RequestData
		// Response is the gRPC response data.
		Response *ResponseData
		// MetadataSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request metadata.
		MetadataSchemes service.SchemesData
		// MessageSchemes lists all the security requirement schemes that
		// apply to the method and are encoded in the request message.
		MessageSchemes service.SchemesData
		// Errors describes the method gRPC errors.
		Errors []*ErrorData

		// server side

		// ServerStruct is the server type name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerStructDeclaration.Name() after planning.
		ServerStruct string
		// ServerStructDeclaration supplies the generated server type name.
		ServerStructDeclaration *codegen.NameDeclaration
		// ServerInterface is the name of the gRPC server interface implemented
		// by the service.
		ServerInterface string
		// ServerStream is the server stream data.
		ServerStream *StreamData

		// client side

		// GRPCMethodName is the Go method name written by protoc-gen-go-grpc for
		// both its client and server interfaces.
		GRPCMethodName string
		// ClientBuild is the remote call builder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientBuildDeclaration.Name() after planning.
		ClientBuild string
		// ClientBuildDeclaration supplies the generated remote call builder name.
		ClientBuildDeclaration *codegen.NameDeclaration
		// ClientEncode is the request encoder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientEncodeDeclaration.Name() after planning.
		ClientEncode string
		// ClientEncodeDeclaration supplies the generated request encoder name.
		ClientEncodeDeclaration *codegen.NameDeclaration
		// ClientDecode is the response decoder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientDecodeDeclaration.Name() after planning.
		ClientDecode string
		// ClientDecodeDeclaration supplies the generated response decoder name.
		ClientDecodeDeclaration *codegen.NameDeclaration
		// ClientStruct is the client type name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ClientStructDeclaration.Name() after planning.
		ClientStruct string
		// ClientStructDeclaration supplies the generated client type name.
		ClientStructDeclaration *codegen.NameDeclaration
		// ClientInterface is the name of the gRPC client interface implemented
		// by the service.
		ClientInterface string
		// ClientStream is the client stream data.
		ClientStream *StreamData
		// ServerHandler is the handler constructor name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerHandlerDeclaration.Name() after planning.
		ServerHandler string
		// ServerHandlerDeclaration supplies the generated handler constructor name.
		ServerHandlerDeclaration *codegen.NameDeclaration
		// ServerDecode is the request decoder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerDecodeDeclaration.Name() after planning.
		ServerDecode string
		// ServerDecodeDeclaration supplies the generated request decoder name.
		ServerDecodeDeclaration *codegen.NameDeclaration
		// ServerEncode is the response encoder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use ServerEncodeDeclaration.Name() after planning.
		ServerEncode string
		// ServerEncodeDeclaration supplies the generated response encoder name.
		ServerEncodeDeclaration *codegen.NameDeclaration
	}

	// MetadataData describes a gRPC metadata field.
	MetadataData struct {
		// Name is the name of the metadata key.
		Name string
		// AttributeName is the name of the corresponding attribute.
		AttributeName string
		// Description is the metadata description.
		Description string
		// FieldName is the name of the struct field that holds the
		// metadata value if any, empty string otherwise.
		FieldName string
		// FieldType is the type of the struct field.
		FieldType expr.DataType
		// FieldTypeRef is the field's final Go type name in the generated file.
		FieldTypeRef string
		// ServiceAttribute is the service field populated from this metadata.
		ServiceAttribute *expr.AttributeExpr
		// WireAttribute is an independent copy of the native gRPC metadata value.
		WireAttribute *expr.AttributeExpr
		// VarName is the name of the Go variable used to read or
		// convert the metadata value.
		VarName string
		// WireVarName is the local variable produced before metadata encoding.
		WireVarName string
		// EncodeCode converts the service field to WireVarName.
		EncodeCode string
		// DecodeCode converts VarName to the exact service constructor field.
		DecodeCode string
		// TypeName is the name of the type.
		TypeName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Required is true if the metadata is required.
		Required bool
		// Pointer is true if and only the metadata variable is a pointer.
		Pointer bool
		// StringSlice is true if the metadata value type is array of strings.
		StringSlice bool
		// Slice is true if the metadata value type is an array.
		Slice bool
		// MapStringSlice reports whether the metadata value is a map from strings
		// to string arrays. Valid current designs always set it to false.
		//
		// Deprecated: gRPC metadata accepts only primitive values and arrays.
		MapStringSlice bool
		// Map reports whether the metadata value is a map. Valid current designs
		// always set it to false.
		//
		// Deprecated: gRPC metadata accepts only primitive values and arrays.
		Map bool
		// Type describes the datatype of the variable value. Mainly
		// used for conversion.
		Type expr.DataType
		// Validate contains the validation code if any.
		Validate string
		// CLIPlan describes how command-line text becomes this metadata value and
		// how the generated payload builder validates it.
		CLIPlan *cli.FlagPlan
		// DefaultValue contains the default value if any.
		DefaultValue any
		// Example is an example value.
		Example any
	}

	// ErrorData contains the error information required to generate the
	// transport decode (client) and encode (server) code.
	ErrorData struct {
		// StatusCode is the response gRPC status code.
		StatusCode string
		// Name is the error name.
		Name string
		// Ref is a reference to the error type.
		Ref string
		// Response is the error response data.
		Response *ResponseData
	}

	// RequestData describes a gRPC request.
	RequestData struct {
		// ProtoMessageName is the message name written in the .proto method.
		ProtoMessageName string
		// Description is the request description.
		Description string
		// Message is the gRPC request message used by the transport. For
		// streaming payload methods with an initial payload frame, this is the
		// synthesized stream envelope.
		Message *service.UserTypeData
		// ClientMessageRef is the request message reference in the generated client package.
		ClientMessageRef string
		// ServerMessageRef is the request message reference in the generated server package.
		ServerMessageRef string
		// PayloadMessage is the gRPC message that carries the one-shot method
		// payload fields before any stream envelope wrapping.
		PayloadMessage *service.UserTypeData
		// ServerPayloadMessageRef is the one-shot payload message reference in the server package.
		ServerPayloadMessageRef string
		// StreamEnvelope describes the synthesized stream envelope when the
		// transport must carry both the one-shot payload and streaming payload
		// items through the same streamed protobuf message.
		StreamEnvelope *StreamEnvelopeData
		// LegacyDecode describes the server-side decoding of requests sent by
		// clients that speak the legacy stream protocol which predates the
		// stream envelope and carries the one-shot method payload in gRPC
		// request metadata. It is nil unless the endpoint enables the
		// "grpc:stream:compat" meta.
		LegacyDecode *LegacyDecodeData
		// Metadata is the request metadata.
		Metadata []*MetadataData
		// ServerConvert is the request data with constructor function to
		// initialize the method payload type from the generated payload type in
		// *.pb.go.
		ServerConvert *ConvertData
		// ClientConvert is the request data with constructor function to
		// initialize the generated payload type in *.pb.go from the
		// method payload.
		ClientConvert *ConvertData
		// CLIArgs is the list of arguments for the command-line client.
		// This is set only for the client side.
		CLIArgs []*InitArgData
		// CLIInitCode builds the service payload from command-line values in the
		// generated client package.
		CLIInitCode string
	}

	// StreamEnvelopeData describes a synthesized streamed protobuf envelope.
	StreamEnvelopeData struct {
		// FieldName is the protobuf oneof field name on the envelope message.
		FieldName string
		// InitialFieldName is the name of the initial payload branch field.
		InitialFieldName string
		// InitialWrapperRef is the fully qualified protobuf wrapper type for the
		// initial payload branch.
		InitialWrapperRef string
		// ClientInitialWrapperRef is the initial payload wrapper in the client package.
		ClientInitialWrapperRef string
		// ServerInitialWrapperRef is the initial payload wrapper in the server package.
		ServerInitialWrapperRef string
		// StreamItemFieldName is the name of the streaming payload item branch
		// field.
		StreamItemFieldName string
		// StreamItemWrapperRef is the fully qualified protobuf wrapper type for
		// the streaming payload item branch.
		StreamItemWrapperRef string
		// ClientStreamItemWrapperRef is the stream item wrapper in the client package.
		ClientStreamItemWrapperRef string
		// ServerStreamItemWrapperRef is the stream item wrapper in the server package.
		ServerStreamItemWrapperRef string
	}

	// LegacyDecodeData describes how generated servers decode the one-shot
	// method payload that legacy stream protocol clients send in gRPC
	// request metadata.
	LegacyDecodeData struct {
		// FuncName is the legacy decoder name kept for existing plugins.
		// Changing it does not rename generated code.
		//
		// Deprecated: Use FuncDeclaration.Name() after planning.
		FuncName string
		// FuncDeclaration supplies the generated legacy decoder name.
		FuncDeclaration *codegen.NameDeclaration
		// Metadata lists the request metadata carrying the method payload
		// along with any explicitly mapped and security metadata.
		Metadata []*MetadataData
		// ServerConvert builds the method payload from the metadata values.
		// It is nil when the payload is not an object type, in which case it
		// travels under the reserved "goa_payload" metadata key and maps
		// directly to the payload.
		ServerConvert *ConvertData
	}

	// ResponseData describes a gRPC success or error response.
	ResponseData struct {
		// ProtoMessageName is the message name written in the .proto method.
		ProtoMessageName string
		// StatusCode is the return code of the response.
		StatusCode string
		// Description is the response description.
		Description string
		// Message is the gRPC response message.
		Message *service.UserTypeData
		// ClientMessageRef is the response message reference in the generated client package.
		ClientMessageRef string
		// ServerMessageRef is the response message reference in the generated server package.
		ServerMessageRef string
		// Headers is the response header metadata.
		Headers []*MetadataData
		// Trailers is the response trailer metadata.
		Trailers []*MetadataData
		// ServerConvert is the type data with constructor function to
		// initialize the generated response type in *.pb.go from the
		// method result type or the projected result type.
		ServerConvert *ConvertData
		// ServerConverts lists the server conversion for each result view. It
		// is empty for results without views.
		ServerConverts []*ViewConvertData
		// ClientConvert is the type data with constructor function to
		// initialize the method result type or the projected result type
		// from the generated response type in *.pb.go.
		ClientConvert *ConvertData
		// ClientConverts lists the client conversion for each result view. It
		// is empty for results without views.
		ClientConverts []*ViewConvertData
	}

	// ConvertData contains the data to convert source type to a target type.
	// For request type, it contains data to transform gRPC request type to the
	// corresponding payload type (server) and vice versa (client).
	// For response type, it contains data to transform gRPC response type to the
	// corresponding result type (client) and vice versa (server).
	ConvertData struct {
		// SrcName is the fully qualified name of the source type. It is empty
		// when a streaming method builds its payload entirely from metadata.
		SrcName string
		// SrcRef is the fully qualified reference to the source type. It is empty
		// when a streaming method builds its payload entirely from metadata.
		SrcRef string
		// TgtName is the fully qualified name of the target type.
		TgtName string
		// TgtRef is the fully qualified reference to the target type.
		TgtRef string
		// Inits contain the data required to render the constructor if any
		// to transform the source type to a target type. If the source or target
		// type is a goa result type, we generate one constructor for every view
		// defined in the result type.
		Init *InitData
		// Validation contains the data required to render the validation function
		// to validate the initialized type.
		Validation *ValidationData
	}

	// ViewConvertData identifies the conversion generated for one result view.
	ViewConvertData struct {
		// View is the result view handled by Convert.
		View string
		// Convert builds the protobuf value using only fields in View.
		Convert *ConvertData
	}

	// ValidationData contains one generated validation function.
	ValidationData struct {
		// Declaration is the function name used by its definition and callers.
		Declaration *codegen.NameDeclaration
		// Name is the final validation function name kept for existing plugins.
		//
		// Deprecated: Use Declaration.Name() after planning.
		Name string
		// Def is the validation function definition.
		Def string
		// VarName is the name of the argument.
		ArgName string
		// SrcName is the fully qualified name of the type being validated.
		SrcName string
		// SrcRef is the fully qualified reference to the type being validated.
		SrcRef string
		// Kind indicates that the validation is for request (server-side),
		// response (client-side), or both (server and client side) messages.
		// It is used to generate validation code in the server and client packages.
		Kind validateKind
	}

	// InitData contains the data required to render a constructor.
	InitData struct {
		// Declaration is the constructor declaration stored in the generated
		// package and used by every call.
		Declaration *codegen.NameDeclaration
		// Name is the constructor function name.
		Name string
		// Description is the function description.
		Description string
		// Args is the list of constructor arguments.
		Args []*InitArgData
		// ReturnVarName is the name of the variable to be returned.
		ReturnVarName string
		// ReturnTypeRef is the qualified (including the package name)
		// reference to the return type.
		ReturnTypeRef string
		// ReturnTypePkg is the package where the return type is present.
		ReturnTypePkg string
		// ReturnIsStruct is true if the return type is a struct.
		ReturnIsStruct bool
		// Code is the transformation code.
		Code string
	}

	// InitArgData represents a single constructor argument.
	InitArgData struct {
		// Name is the argument name.
		Name string
		// Description is the argument description.
		Description string
		// Reference to the argument, e.g. "&body".
		Ref string
		// FieldName is the name of the data structure field that should
		// be initialized with the argument if any.
		FieldName string
		// FieldType is the type of the data structure field that should be
		// initialized with the argument if any.
		FieldType expr.DataType
		// FieldTypeRef is the field's final Go type name in the generated file.
		FieldTypeRef string
		// InitCode converts and assigns this argument to the constructor result.
		InitCode string
		// TypeName is the argument type name.
		TypeName string
		// TypeRef is the argument type reference.
		TypeRef string
		// Type is the argument type. It is never an aliased user type.
		Type expr.DataType
		// Pointer is true if a pointer to the arg should be used.
		Pointer bool
		// Required is true if the arg is required to build the payload.
		Required bool
		// DefaultValue is the default value of the arg.
		DefaultValue any
		// Validate contains the validation code for the argument
		// value if any.
		Validate string
		// CLIPlan describes how command-line text becomes this argument value and
		// how the generated payload builder validates it.
		CLIPlan *cli.FlagPlan
		// Example is a example value
		Example any
	}

	// StreamData contains data to render the stream struct type that implements
	// the service stream interface.
	StreamData struct {
		// VarName is the stream type name kept for existing plugins. Changing it
		// does not rename generated code.
		//
		// Deprecated: Use Declaration.Name() after planning.
		VarName string
		// Declaration supplies the generated stream type name.
		Declaration *codegen.NameDeclaration
		// Type is the stream type (client or server).
		Type string
		// ServiceInterface is the service interface that the struct implements.
		ServiceInterface string
		// Interface is the stream interface in *.pb.go stored in the struct.
		Interface string
		// Endpoint is the streaming endpoint data.
		Endpoint *EndpointData
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendWithContextDesc is the description for the send function with context.
		SendWithContextDesc string
		// SendRef is the fully	qualified reference to the type sent across the
		// stream.
		SendRef string
		// SendConvert is the type sent through the stream. It contains the
		// constructor to convert the service send type to the type expected by
		// the gRPC send type (in *.pb.go)
		SendConvert *ConvertData
		// SendConverts lists the server send conversion for each result view.
		// It is empty for client streams and results without views.
		SendConverts []*ViewConvertData
		// RecvConvert is the type received through the stream. It contains the
		// constructor to convert the gRPC type (in *.pb.go) to the service receive
		// type.
		RecvConvert *ConvertData
		// RecvConverts lists the client receive conversion for each result view.
		// It is empty for server streams and results without views.
		RecvConverts []*ViewConvertData
		// RecvName is the name of the receive function.
		RecvName string
		// RecvDesc is the description for the recv function.
		RecvDesc string
		// RecvWithContextName is the name of the receive function with context.
		RecvWithContextName string
		// RecvWithContextDesc is the description for the recv function with context.
		RecvWithContextDesc string
		// RecvRef is the fully	qualified reference to the type received from the
		// stream.
		RecvRef string
		// MustClose indicates whether to generate the Close() function
		// for the stream.
		MustClose bool
	}

	// validateKind is a type to determine where the validation code is generated
	// (server, client, or both)
	validateKind int
)

const (
	// pbPkgName is the directory name where the .proto file is generated and
	// compiled.
	pbPkgName = "pb"
	// validateServer generates the validation code for request messages in the
	// server package.
	validateServer validateKind = iota + 1
	// validateClient generates response and command-line request validation in
	// the client package.
	validateClient
)

// newServicesData builds the values passed to gRPC client and server templates
// from the saved service data and gRPC plan.
func newServicesData(services *service.ServicesData, plan *Plan) *ServicesData {
	if services.Root != plan.root {
		panic(fmt.Sprintf("gRPC service data does not belong to design %q", plan.root.API.Name))
	}
	data := &ServicesData{
		ServicesData:  services,
		GRPCServices:  make(map[string]*ServiceData),
		cliPlan:       plan.cli,
		protobuf:      plan.protobuf,
		tools:         plan.tools,
		symbols:       plan.symbols,
		servicePlans:  append([]*grpcServicePlan(nil), plan.servicesPlan...),
		serviceByExpr: make(map[*expr.GRPCServiceExpr]*ServiceData, len(plan.servicesPlan)),
		endpointPlans: make(map[*expr.GRPCEndpointExpr]*grpcEndpointPlan),
		metadataPlans: make(map[*expr.MappedAttributeExpr][]*grpcMetadataPlan),
		fileImports:   plan.fileImports,
		generation:    plan.generation,
	}
	data.expressions = make([]*expr.GRPCServiceExpr, len(data.servicePlans))
	for index, servicePlan := range data.servicePlans {
		data.expressions[index] = servicePlan.expression
		for _, endpointPlan := range servicePlan.endpoints {
			data.endpointPlans[endpointPlan.expression] = endpointPlan
			for mapped, metadata := range endpointPlan.metadata {
				data.metadataPlans[mapped] = metadata
			}
		}
		serviceData := data.analyze(servicePlan)
		data.GRPCServices[servicePlan.expression.Name()] = serviceData
		data.serviceByExpr[servicePlan.source] = serviceData
	}
	return data
}

// Get retrieves the transport data saved for the service with the given name.
// It returns nil if there is no service with the given name.
func (d *ServicesData) Get(name string) *ServiceData {
	return d.GRPCServices[name]
}

// exampleServiceData copies the package qualifiers used by one executable.
// Server examples import the service, protobuf, and gRPC server packages;
// command-line examples import the service package only when they receive a
// result stream.
func (d *ServicesData) exampleServiceData(source *ServiceData, outputPackage string, server bool) *ServiceData {
	data := *source
	service := *source.Service
	data.Service = &service
	data.Endpoints = make([]*EndpointData, len(source.Endpoints))
	for index, endpoint := range source.Endpoints {
		copy := *endpoint
		data.Endpoints[index] = &copy
	}
	if server {
		if len(source.Endpoints) > 0 {
			service.PkgName = d.ServiceImport(outputPackage, service.Name).Name
		}
		protobufPath := path.Join(d.GenPkg(), "grpc", service.PathName, pbPkgName)
		data.PkgName = d.PackageImport(outputPackage, protobufPath).Name
		data.ServerPkgName = d.PackageImport(outputPackage, path.Join(d.GenPkg(), "grpc", service.PathName, "server")).Name
		for _, endpoint := range data.Endpoints {
			endpoint.PkgName = data.PkgName
			endpoint.ServicePkgName = service.PkgName
		}
		return &data
	}
	if grpcServiceStreamsResult(d.servicePlan(service.Name).expression) {
		service.PkgName = d.ServiceImport(outputPackage, service.Name).Name
		for _, endpoint := range data.Endpoints {
			endpoint.ServicePkgName = service.PkgName
		}
	}
	return &data
}

// servicePlan returns the copied gRPC plan for service name.
func (d *ServicesData) servicePlan(name string) *grpcServicePlan {
	for _, plan := range d.servicePlans {
		if plan.expression.Name() == name {
			return plan
		}
	}
	panic(fmt.Sprintf("gRPC service plan %q is missing", name))
}

// Endpoint returns the endpoint data for the endpoint with the given name, nil
// if there isn't one.
func (sd *ServiceData) Endpoint(name string) *EndpointData {
	for _, ed := range sd.Endpoints {
		if ed.Method.Name == name {
			return ed
		}
	}
	return nil
}

// HasUnaryEndpoint returns true if the service has at least one unary endpoint.
func (sd *ServiceData) HasUnaryEndpoint() bool {
	for _, ed := range sd.Endpoints {
		if ed.ServerStream == nil {
			return true
		}
	}
	return false
}

// HasStreamingEndpoint returns true if the service has at least one streaming
// endpoint.
func (sd *ServiceData) HasStreamingEndpoint() bool {
	for _, ed := range sd.Endpoints {
		if ed.ServerStream != nil {
			return true
		}
	}
	return false
}

// analyze creates the data necessary to render the code of the given service.
func (d *ServicesData) analyze(servicePlan *grpcServicePlan) *ServiceData {
	gs := servicePlan.expression
	svc := d.ServicesData.Get(gs.Name())
	transportService := *svc
	transportService.ProtoImports = append([]*codegen.ImportSpec(nil), svc.ProtoImports...)
	transportService.ProtoImports = append(transportService.ProtoImports, servicePlan.protoGoImports...)
	clientPackage := path.Join(d.GenPkg(), "grpc", svc.PathName, "client")
	serverPackage := path.Join(d.GenPkg(), "grpc", svc.PathName, "server")
	clientServicePackage := ""
	if grpcAttributesHaveValues(grpcEndpointAttributes(gs.GRPCEndpoints...)) || grpcServiceHasStreaming(gs) {
		clientServicePackage = d.ServiceImport(clientPackage, svc.Name).Name
	}
	serverServicePackage := d.ServiceImport(serverPackage, svc.Name).Name
	transportService.PkgName = clientServicePackage
	svc = &transportService
	protobufPath := path.Join(d.GenPkg(), "grpc", svc.PathName, pbPkgName)
	clientProtobufPackage := d.PackageImport(clientPackage, protobufPath).Name
	serverProtobufPackage := d.PackageImport(serverPackage, protobufPath).Name
	planned := d.protobuf[gs]
	if planned == nil {
		panic(fmt.Sprintf("protobuf plan is missing for gRPC service %q", gs.Name()))
	}
	serviceDescriptor := planned.serviceFullName()
	symbols := d.symbols[gs]
	if symbols == nil {
		panic(fmt.Sprintf("Go names are missing for gRPC service %q", gs.Name()))
	}
	sd := &ServiceData{
		Service:                 svc,
		Name:                    planned.serviceName,
		Description:             svc.Description,
		PkgName:                 clientProtobufPackage,
		ClientProtobufPkgName:   clientProtobufPackage,
		ServerProtobufPkgName:   serverProtobufPackage,
		ClientServicePkgName:    clientServicePackage,
		ServerServicePkgName:    serverServicePackage,
		ProtoImports:            append([]string(nil), servicePlan.protoImports...),
		ServerStruct:            symbols.serverStruct.Name(),
		ServerStructDeclaration: symbols.serverStruct,
		ClientStruct:            symbols.clientStruct.Name(),
		ClientStructDeclaration: symbols.clientStruct,
		ServerInit:              symbols.serverInit.Name(),
		ServerInitDeclaration:   symbols.serverInit,
		ClientInit:              symbols.clientInit.Name(),
		ClientInitDeclaration:   symbols.clientInit,
		ServerInterface:         planned.name(serviceDescriptor, protocServiceServerName),
		ClientInterface:         planned.name(serviceDescriptor, protocServiceClientName),
		ClientInterfaceInit:     clientProtobufPackage + "." + planned.name(serviceDescriptor, protocServiceClientConstructorName),
		UnimplementedServer:     planned.name(serviceDescriptor, protocServiceUnimplementedServerName),
		RegisterFunction:        planned.name(serviceDescriptor, protocServiceRegisterName),
		Scope:                   servicePlan.scope,
		protobuf:                planned.catalog,
	}
	sd.protobuf.packageName = clientProtobufPackage
	finishProtobufPackage(sd)
	protobufMessages := planned.messages
	for index, e := range gs.GRPCEndpoints {
		endpointPlan := servicePlan.endpointByExpr[e]
		if endpointPlan == nil {
			panic(fmt.Sprintf("saved gRPC endpoint data is missing for %q", e.Name()))
		}
		endpointSymbols := symbols.endpoints[e]
		if endpointSymbols == nil {
			panic(fmt.Sprintf("Go names are missing for gRPC endpoint %q", e.Name()))
		}
		hasRequestMessage := !isEmpty(e.Request.Type)
		messages := protobufMessages[index]
		requestMessage := messages.request
		streamingRequest := messages.streamingRequest
		requestEnvelope := messages.requestEnvelope
		responseMessage := messages.response
		errorMessages := messages.errors
		collect := func(attribute *expr.AttributeExpr) *protobufMessageRecord {
			record := sd.protobuf.message(attribute)
			if record == nil || record.data == nil {
				panic(fmt.Sprintf("no protobuf message collected for attribute of type %q", attribute.Type.Name())) // bug
			}
			return record
		}

		var (
			clientPayloadRef string
			serverPayloadRef string
			clientResultRef  string
			serverResultRef  string
			viewedResultRef  string
		)
		md := svc.Method(e.Name())
		if e.MethodExpr.Payload.Type != expr.Empty {
			clientContext := d.serviceTypeContext(sd, "client").Enter(e.MethodExpr.Payload)
			serverContext := d.serviceTypeContext(sd, "server").Enter(e.MethodExpr.Payload)
			clientPayloadRef = clientContext.Scope.Ref(e.MethodExpr.Payload, clientContext.Pkg(e.MethodExpr.Payload))
			serverPayloadRef = serverContext.Scope.Ref(e.MethodExpr.Payload, serverContext.Pkg(e.MethodExpr.Payload))
		}
		if e.MethodExpr.Result.Type != expr.Empty {
			clientContext := d.serviceTypeContext(sd, "client").Enter(e.MethodExpr.Result)
			serverContext := d.serviceTypeContext(sd, "server").Enter(e.MethodExpr.Result)
			clientResultRef = clientContext.Scope.Ref(e.MethodExpr.Result, clientContext.Pkg(e.MethodExpr.Result))
			serverResultRef = serverContext.Scope.Ref(e.MethodExpr.Result, serverContext.Pkg(e.MethodExpr.Result))
		}
		if md.ViewedResult != nil {
			viewedResultRef = md.ViewedResult.FullRef
		}
		errors := d.buildErrorsData(e, errorMessages, sd)
		// build request data
		payloadIdentity := expr.MethodPayloadExampleIdentity(e.MethodExpr)
		resultIdentity := expr.MethodResultExampleIdentity(e.MethodExpr)
		reqMD := d.extractMetadata(e.Metadata, e.MethodExpr.Payload, sd, "server", "v", payloadIdentity)
		request := &RequestData{
			Description:   requestMessage.Description,
			Metadata:      reqMD,
			ServerConvert: d.buildRequestConvertData(requestMessage, e.MethodExpr.Payload, reqMD, e, sd, true),
			ClientConvert: d.buildRequestConvertData(requestMessage, e.MethodExpr.Payload, reqMD, e, sd, false),
		}
		if e.MethodExpr.Payload.Type != expr.Empty {
			request.CLIInitCode = d.buildCLIRequestTransform(e, requestMessage, sd)
		}
		if hasRequestMessage {
			request.PayloadMessage = collect(requestMessage).data
			request.ServerPayloadMessageRef = protoBufGoFullTypeRef(requestMessage, sd.ServerProtobufPkgName, sd)
		}
		if protobufCLIRequestNeedsMessage(requestMessage) {
			// add the request message as the first argument to the CLI
			typeName := protoBufGoFullTypeName(requestMessage, sd.PkgName, sd)
			request.CLIArgs = append(request.CLIArgs, &InitArgData{
				Name:     "message",
				Ref:      "message",
				TypeName: typeName,
				TypeRef:  protoBufGoFullTypeRef(requestMessage, sd.PkgName, sd),
				CLIPlan:  cli.NewProtobufFlagPlan(requestMessage, typeName),
				Example:  protobufCLIExample(requestMessage, d.Example(requestMessage, payloadIdentity), sd.protobuf.plan),
			})
		}
		// pass the metadata as arguments to client CLI args
		request.CLIArgs = append(request.CLIArgs, argsFromMetadata(reqMD)...)
		transportRequest := requestMessage
		switch {
		case requestEnvelope != nil:
			transportRequest = requestEnvelope
			record := collect(requestEnvelope)
			request.Message = record.data
			request.ProtoMessageName = record.protoName
			request.StreamEnvelope = buildStreamEnvelopeData(requestEnvelope, sd)
			if endpointPlan.legacyStream {
				request.LegacyDecode = d.buildLegacyDecodeData(e, sd)
			}
		case streamingRequest.Type != expr.Empty:
			transportRequest = streamingRequest
			record := collect(streamingRequest)
			request.Message = record.data
			request.ProtoMessageName = record.protoName
		default:
			record := collect(requestMessage)
			request.Message = record.data
			request.ProtoMessageName = record.protoName
		}
		request.ClientMessageRef = protoBufGoFullTypeRef(transportRequest, sd.ClientProtobufPkgName, sd)
		request.ServerMessageRef = protoBufGoFullTypeRef(transportRequest, sd.ServerProtobufPkgName, sd)

		// build response data
		serverResult, serverCtx := d.resultContext(e, sd, "server")
		clientResult, clientCtx := d.resultContext(e, sd, "client")
		hdrs := d.extractMetadata(e.Response.Headers, clientResult, sd, "client", "result", resultIdentity)
		trlrs := d.extractMetadata(e.Response.Trailers, clientResult, sd, "client", "result", resultIdentity)
		serverConverts := d.buildServerResponseConverts(responseMessage, serverResult, serverCtx, e, sd)
		clientConverts := d.buildClientResponseConverts(responseMessage, messages.responseValidations, clientResult, clientCtx, hdrs, trlrs, e, sd)
		var viewedServerConverts, viewedClientConverts []*ViewConvertData
		if _, viewed := e.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
			viewedServerConverts = serverConverts
			viewedClientConverts = clientConverts
		}
		response := &ResponseData{
			StatusCode:     statusCodeToGRPCConst(e.Response.StatusCode),
			Description:    e.Response.Description,
			Headers:        hdrs,
			Trailers:       trlrs,
			ServerConvert:  primaryViewConvert(serverConverts),
			ServerConverts: viewedServerConverts,
			ClientConvert:  primaryViewConvert(clientConverts),
			ClientConverts: viewedClientConverts,
		}
		// If the endpoint is a streaming endpoint, no message is returned
		// by gRPC. Hence, no need to set response message.
		if responseMessage.Type != expr.Empty || !e.MethodExpr.IsStreaming() {
			record := collect(responseMessage)
			response.Message = record.data
			response.ProtoMessageName = record.protoName
			response.ClientMessageRef = protoBufGoFullTypeRef(responseMessage, sd.ClientProtobufPkgName, sd)
			response.ServerMessageRef = protoBufGoFullTypeRef(responseMessage, sd.ServerProtobufPkgName, sd)
		}

		// gather security requirements
		var (
			msgSch service.SchemesData
			metSch service.SchemesData
		)
		for _, req := range e.Requirements {
			for _, sch := range req.Schemes {
				s := md.Requirements.Scheme(sch.SchemeName).Dup()
				s.In = sch.In
				switch s.In {
				case "message":
					msgSch = msgSch.Append(s)
				default:
					metSch = metSch.Append(s)
				}
			}
		}
		ed := &EndpointData{
			ServiceName:             svc.Name,
			PkgName:                 sd.PkgName,
			ServicePkgName:          svc.PkgName,
			ClientProtobufPkgName:   sd.ClientProtobufPkgName,
			ServerProtobufPkgName:   sd.ServerProtobufPkgName,
			ClientServicePkgName:    sd.ClientServicePkgName,
			ServerServicePkgName:    sd.ServerServicePkgName,
			Method:                  md,
			ProtoMethodName:         planned.methods[e],
			ClientMethodName:        planned.methods[e],
			FullMethodName:          planned.serviceFullName() + "/" + planned.methods[e],
			PayloadType:             e.MethodExpr.Payload.Type,
			PayloadRef:              clientPayloadRef,
			ClientPayloadRef:        clientPayloadRef,
			ServerPayloadRef:        serverPayloadRef,
			ResultRef:               clientResultRef,
			ClientResultRef:         clientResultRef,
			ServerResultRef:         serverResultRef,
			ViewedResultRef:         viewedResultRef,
			Request:                 request,
			Response:                response,
			MessageSchemes:          msgSch,
			MetadataSchemes:         metSch,
			Errors:                  errors,
			ServerStruct:            sd.ServerStruct,
			ServerStructDeclaration: sd.ServerStructDeclaration,
			ServerInterface:         sd.ServerInterface,
			GRPCMethodName:          planned.name(serviceDescriptor+"."+planned.methods[e], protocMethodName),
			ClientStruct:            sd.ClientStruct,
			ClientStructDeclaration: sd.ClientStructDeclaration,
			ClientInterface:         sd.ClientInterface,
		}
		ed.ClientBuild = endpointSymbols.clientBuild.Name()
		ed.ClientBuildDeclaration = endpointSymbols.clientBuild
		if endpointSymbols.clientEncode != nil {
			ed.ClientEncode = endpointSymbols.clientEncode.Name()
			ed.ClientEncodeDeclaration = endpointSymbols.clientEncode
		}
		if endpointSymbols.clientDecode != nil {
			ed.ClientDecode = endpointSymbols.clientDecode.Name()
			ed.ClientDecodeDeclaration = endpointSymbols.clientDecode
		}
		ed.ServerHandler = endpointSymbols.serverHandler.Name()
		ed.ServerHandlerDeclaration = endpointSymbols.serverHandler
		if endpointSymbols.serverDecode != nil {
			ed.ServerDecode = endpointSymbols.serverDecode.Name()
			ed.ServerDecodeDeclaration = endpointSymbols.serverDecode
		}
		ed.ServerEncode = endpointSymbols.serverEncode.Name()
		ed.ServerEncodeDeclaration = endpointSymbols.serverEncode
		sd.Endpoints = append(sd.Endpoints, ed)
		if e.MethodExpr.IsStreaming() {
			ed.ServerStream = d.buildStreamData(e, streamingRequest, responseMessage, messages.responseValidations, sd, true)
			ed.ServerStream.VarName = endpointSymbols.serverStream.Name()
			ed.ServerStream.Declaration = endpointSymbols.serverStream
			ed.ClientStream = d.buildStreamData(e, streamingRequest, responseMessage, messages.responseValidations, sd, false)
			ed.ClientStream.VarName = endpointSymbols.clientStream.Name()
			ed.ClientStream.Declaration = endpointSymbols.clientStream
		}
	}
	return sd
}

// collectProtobufPackage copies each method message and records every message
// and oneof before generated Go names are fixed.
func collectProtobufPackage(serviceExpr *expr.GRPCServiceExpr, catalog *protobufPackageCatalog) ([]*protobufEndpointMessages, error) {
	prepared := make([]*protobufEndpointMessages, len(serviceExpr.GRPCEndpoints))
	for index, endpoint := range serviceExpr.GRPCEndpoints {
		useStreamEnvelope := usesStreamEnvelope(endpoint)
		request := makeProtoBufMessage(
			endpoint.Request,
			codegen.ProtobufName(endpoint.Name()+"_request"),
			expr.GRPCRequestMessageExampleIdentity(endpoint.MethodExpr),
		)
		streamingRequest := endpoint.StreamingRequest
		if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			name := codegen.ProtobufName(endpoint.Name() + "_streaming_request")
			if useStreamEnvelope {
				name = codegen.ProtobufName(endpoint.Name() + "_stream_item")
			}
			streamingRequest = makeProtoBufMessage(
				endpoint.StreamingRequest,
				name,
				expr.GRPCStreamingRequestMessageExampleIdentity(endpoint.MethodExpr),
			)
		}
		var requestEnvelope *expr.AttributeExpr
		if useStreamEnvelope {
			requestEnvelope = makeProtoBufStreamEnvelope(
				request,
				streamingRequest,
				codegen.ProtobufName(endpoint.Name()+"_streaming_request"),
				expr.GRPCStreamingRequestMessageExampleIdentity(endpoint.MethodExpr),
			)
		}
		responseOwner := expr.GRPCResponseMessageExampleIdentity(endpoint.MethodExpr)
		if endpoint.MethodExpr.IsResultStreaming() {
			responseOwner = expr.GRPCStreamingResponseMessageExampleIdentity(endpoint.MethodExpr)
		}
		response := makeProtoBufMessage(
			endpoint.Response.Message,
			codegen.ProtobufName(endpoint.Name()+"_response"),
			responseOwner,
		)
		errors := make(map[string]*expr.AttributeExpr, len(endpoint.GRPCErrors))
		for _, grpcError := range endpoint.GRPCErrors {
			if expr.IsErrorResult(grpcError.Type) || !expr.IsObject(grpcError.Type) {
				continue
			}
			errors[grpcError.Name] = makeProtoBufMessage(
				grpcError.Response.Message,
				codegen.ProtobufName(endpoint.Name()+"_"+grpcError.Name+"_error"),
				expr.GRPCErrorMessageExampleIdentity(endpoint.MethodExpr, grpcError.ErrorExpr),
			)
		}
		prepared[index] = &protobufEndpointMessages{
			request:          request,
			streamingRequest: streamingRequest,
			requestEnvelope:  requestEnvelope,
			response:         response,
			errors:           errors,
		}
	}

	collect := catalog.collectMessage
	for index, endpoint := range serviceExpr.GRPCEndpoints {
		messages := prepared[index]
		requestSource := protobufRootMessageSource(endpoint.Request, endpoint, nil, protobufRequestMessage)
		streamingSource := protobufRootMessageSource(endpoint.StreamingRequest, endpoint, nil, protobufStreamingRequestMessage)
		responseSource := protobufRootMessageSource(endpoint.Response.Message, endpoint, nil, protobufResponseMessage)
		catalog.bindRootSource(messages.request, requestSource)
		if messages.streamingRequest.Type != expr.Empty {
			catalog.bindRootSource(messages.streamingRequest, streamingSource)
		}
		catalog.bindRootSource(messages.response, responseSource)
		for _, grpcError := range endpoint.GRPCErrors {
			message := messages.errors[grpcError.Name]
			if message == nil {
				continue
			}
			errorSource := protobufRootMessageSource(
				grpcError.Response.Message,
				endpoint,
				grpcError,
				protobufErrorMessage,
			)
			catalog.bindRootSource(message, errorSource)
			if err := collect(message, errorSource); err != nil {
				return nil, err
			}
		}
		requestNeeded := !isEmpty(endpoint.Request.Type) ||
			(messages.requestEnvelope == nil && messages.streamingRequest.Type == expr.Empty)
		if requestNeeded {
			if err := collect(messages.request, requestSource); err != nil {
				return nil, err
			}
		}
		if messages.requestEnvelope != nil {
			envelopeSource := protobufMessageSource{synthetic: protobufSyntheticMessage{
				endpoint: endpoint,
				role:     protobufStreamEnvelopeMessage,
			}}
			catalog.bindRootSource(messages.requestEnvelope, envelopeSource)
			if err := collect(messages.requestEnvelope, envelopeSource); err != nil {
				return nil, err
			}
		}
		if messages.streamingRequest.Type != expr.Empty {
			if err := collect(messages.streamingRequest, streamingSource); err != nil {
				return nil, err
			}
		}
		if messages.response.Type != expr.Empty || !endpoint.MethodExpr.IsStreaming() {
			if err := collect(messages.response, responseSource); err != nil {
				return nil, err
			}
			if _, viewed := endpoint.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
				messages.responseValidations = make(map[string]*expr.AttributeExpr)
				for _, view := range grpcResultViews(endpoint.MethodExpr) {
					selected, err := grpcResultForView(endpoint.MethodExpr.Result, view)
					if err != nil {
						return nil, err
					}
					validation := grpcProtobufValidationForView(messages.response, selected)
					catalog.bindCopiedMessageUses(messages.response, validation)
					messages.responseValidations[view] = validation
				}
			}
		}
	}
	return prepared, nil
}

// finishProtobufPackage reads the final Go names and builds the message and
// validation data used by generated clients and servers.
func finishProtobufPackage(sd *ServiceData) {
	sd.protobuf.freezeMessages(sd)
	sd.Messages = sd.protobuf.protoMessageData()

	sd.validations = sd.protobuf.freezeValidations(sd)
}

// protobufRootMessageSource connects a root message to the authored service
// declaration whose value it carries. The endpoint value is used only when
// the service value is inline or created by Goa. An explicit Message DSL may
// provide the declaration instead.
func protobufRootMessageSource(attribute *expr.AttributeExpr, endpoint *expr.GRPCEndpointExpr, grpcError *expr.GRPCErrorExpr, role protobufSyntheticRole) protobufMessageSource {
	var serviceAttribute *expr.AttributeExpr
	switch role {
	case protobufRequestMessage:
		serviceAttribute = endpoint.MethodExpr.Payload
	case protobufStreamingRequestMessage:
		serviceAttribute = endpoint.MethodExpr.StreamingPayload
	case protobufResponseMessage:
		serviceAttribute = endpoint.MethodExpr.Result
	case protobufErrorMessage:
		if methodError := endpoint.MethodExpr.Error(grpcError.Name); methodError != nil {
			serviceAttribute = methodError.AttributeExpr
		}
	}
	if serviceAttribute != nil {
		if userType, ok := serviceAttribute.Type.(expr.UserType); ok {
			return protobufMessageSource{origin: userType.Origin()}
		}
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		return protobufMessageSource{origin: userType.Origin()}
	}
	return protobufMessageSource{synthetic: protobufSyntheticMessage{
		endpoint: endpoint,
		error:    grpcError,
		role:     role,
	}}
}

// addValidation returns the validation function chosen for the given protobuf
// message in the generated server or client package.
//
// req if true indicates that the validation is generated for validating
// request (server-side) messages.
func addValidation(att *expr.AttributeExpr, sd *ServiceData, req bool) *ValidationData {
	kind := validateClient
	if req {
		kind = validateServer
	}
	return sd.protobuf.validation(att, kind)
}

// userTypeAttribute returns the attribute of the given user type.
func userTypeAttribute(ut expr.UserType) *expr.AttributeExpr {
	att := ut.Attribute()
	if rt, ok := ut.(*expr.ResultTypeExpr); ok {
		// Result type collections are wrapper user types themselves: the
		// wrappedAttrMeta marker lives directly on their attribute when the
		// collection was wrapped as a whole message (makeProtoBufMessage) and
		// on a nested wrapper user type when the collection was wrapped in
		// place as a field type (makeProtoBufMessageR).
		if len(att.Meta[wrappedAttrMeta]) > 0 || isWrappedAttr(att) {
			if expr.IsArray(unwrapAttr(att).Type) {
				// result type collection
				att = &expr.AttributeExpr{Type: expr.AsObject(rt)}
			}
		}
	}
	return att
}

// protobufCLIExample writes object field names exactly as they appear in the
// protobuf file. For example, a Goa field named "tenantID" is shown as
// "tenant_id" in command help.
func protobufCLIExample(attribute *expr.AttributeExpr, value any, plan *protobufServicePlan) any {
	return protobufCLIExampleValue(attribute, value, plan, false)
}

// protobufCLIExampleValue avoids adding the same protobuf object twice while
// visiting its field. Values inside arrays and maps are converted separately.
func protobufCLIExampleValue(attribute *expr.AttributeExpr, value any, plan *protobufServicePlan, skipWrapper bool) any {
	if value == nil {
		return nil
	}
	if !skipWrapper && isWrappedAttr(attribute) {
		field := unwrapAttr(attribute)
		fieldValue := value
		if !field.Type.IsCompatible(value) {
			var ok bool
			fieldValue, ok = namedExampleValue(value, wrappedField)
			if !ok {
				panic("protobuf CLI wrapper example has no field value")
			}
		}
		return map[string]any{
			plan.sourceFieldName(field): protobufCLIExampleValue(field, fieldValue, plan, true),
		}
	}
	if object := expr.AsObject(attribute.Type); object != nil {
		result := make(map[string]any, len(*object))
		for _, field := range *object {
			fieldValue, ok := namedExampleValue(value, field.Name)
			if !ok {
				continue
			}
			if expr.AsUnion(field.Attribute.Type) != nil {
				for name, branchValue := range protobufCLIExampleValue(field.Attribute, fieldValue, plan, false).(map[string]any) {
					result[name] = branchValue
				}
				continue
			}
			result[plan.sourceFieldName(field.Attribute)] = protobufCLIExampleValue(field.Attribute, fieldValue, plan, false)
		}
		return result
	}
	if union := expr.AsUnion(attribute.Type); union != nil {
		branchName, ok := namedExampleValue(value, union.GetTypeKey())
		if !ok {
			panic("protobuf CLI union example has no branch name")
		}
		branchValue, ok := namedExampleValue(value, union.GetValueKey())
		if !ok {
			panic("protobuf CLI union example has no branch value")
		}
		for _, branch := range union.Values {
			if branch.Name != branchName {
				continue
			}
			return map[string]any{
				plan.sourceFieldName(branch.Attribute): protobufCLIExampleValue(branch.Attribute, branchValue, plan, false),
			}
		}
		panic(fmt.Sprintf("protobuf CLI union example selects unknown branch %q", branchName))
	}
	if array := expr.AsArray(attribute.Type); array != nil {
		items := reflect.ValueOf(value)
		if items.Kind() != reflect.Array && items.Kind() != reflect.Slice {
			panic(fmt.Sprintf("protobuf CLI array example has type %T", value))
		}
		result := make([]any, items.Len())
		for index := range items.Len() {
			result[index] = protobufCLIExampleValue(array.ElemType, items.Index(index).Interface(), plan, false)
		}
		return result
	}
	if mapped := expr.AsMap(attribute.Type); mapped != nil {
		entries := reflect.ValueOf(value)
		if entries.Kind() != reflect.Map {
			panic(fmt.Sprintf("protobuf CLI map example has type %T", value))
		}
		result := make(map[string]any, entries.Len())
		for _, key := range entries.MapKeys() {
			result[fmt.Sprint(key.Interface())] = protobufCLIExampleValue(mapped.ElemType, entries.MapIndex(key).Interface(), plan, false)
		}
		return result
	}
	return value
}

// namedExampleValue returns the value stored under one Goa object or union
// field name.
func namedExampleValue(example any, name string) (any, bool) {
	fields := reflect.ValueOf(example)
	if fields.Kind() != reflect.Map || fields.Type().Key().Kind() != reflect.String {
		panic(fmt.Sprintf("protobuf CLI object example has type %T", example))
	}
	value := fields.MapIndex(reflect.ValueOf(name).Convert(fields.Type().Key()))
	if !value.IsValid() {
		return nil, false
	}
	return value.Interface(), true
}

// buildRequestConvertData builds the convert data for the server and client
// requests.
//   - server side - converts the one-shot gRPC request message (if any) and
//     gRPC metadata to the method payload type.
//   - client side - converts the method payload type to the one-shot gRPC
//     request message sent before any stream items.
//
// svr param indicates that the convert data is generated for server side.
func (d *ServicesData) buildRequestConvertData(request, payload *expr.AttributeExpr, md []*MetadataData, e *expr.GRPCEndpointExpr, sd *ServiceData, svr bool) *ConvertData {
	if svr && isEmpty(payload.Type) {
		return nil
	}
	if !svr && e.MethodExpr.IsPayloadStreaming() && isEmpty(request.Type) {
		return nil
	}
	if svr && e.MethodExpr.IsPayloadStreaming() && isEmpty(request.Type) && !expr.IsObject(payload.Type) {
		return nil
	}

	side := "client"
	protobufPackage := sd.ClientProtobufPkgName
	if svr {
		protobufPackage = sd.ServerProtobufPkgName
		side = "server"
	}
	svcCtx := d.serviceTypeContext(sd, side).Enter(payload)
	if svr {
		// server side
		data := d.buildInitData(request, payload, "message", "v", svcCtx, false, sd, expr.MethodPayloadExampleIdentity(e.MethodExpr), d.initDeclaration(e, true, grpcInitKey{role: grpcRequestInit}))
		// pass the metadata as arguments to payload constructor in server
		data.Args = append(data.Args, initArgsFromMetadata(md)...)
		conversion := &ConvertData{
			TgtName: svcCtx.Scope.Name(payload, svcCtx.Pkg(payload), svcCtx.Pointer, svcCtx.UseDefault),
			TgtRef:  svcCtx.Scope.Ref(payload, svcCtx.Pkg(payload)),
			Init:    data,
		}
		if !e.MethodExpr.IsPayloadStreaming() || !isEmpty(e.Request.Type) {
			conversion.SrcName = protoBufGoFullTypeName(request, protobufPackage, sd)
			conversion.SrcRef = protoBufGoFullTypeRef(request, protobufPackage, sd)
			conversion.Validation = addValidation(request, sd, true)
		}
		return conversion
	}

	// client side
	data := d.buildInitData(payload, request, "payload", "message", svcCtx, true, sd, expr.MethodPayloadExampleIdentity(e.MethodExpr), d.initDeclaration(e, false, grpcInitKey{role: grpcRequestInit}))
	conversion := &ConvertData{
		TgtName: protoBufGoFullTypeName(request, sd.ClientProtobufPkgName, sd),
		TgtRef:  protoBufGoFullTypeRef(request, sd.ClientProtobufPkgName, sd),
		Init:    data,
	}
	if !isEmpty(payload.Type) {
		conversion.SrcName = svcCtx.Scope.Name(payload, svcCtx.Pkg(payload), svcCtx.Pointer, svcCtx.UseDefault)
		conversion.SrcRef = svcCtx.Scope.Ref(payload, svcCtx.Pkg(payload))
	}
	return conversion
}

// buildLegacyDecodeData computes the data needed to decode requests sent by
// legacy stream protocol clients which carry the one-shot method payload in
// gRPC request metadata. The metadata layout mirrors what pre-envelope
// versions of goa generated: every payload attribute not explicitly mapped
// to metadata is carried under its own name and non-object payloads travel
// under the reserved "goa_payload" key.
func (d *ServicesData) buildLegacyDecodeData(e *expr.GRPCEndpointExpr, sd *ServiceData) *LegacyDecodeData {
	payload := e.MethodExpr.Payload
	endpointPlan := d.endpointPlans[e]
	if endpointPlan == nil || endpointPlan.legacyMetadata == nil {
		panic(fmt.Sprintf("saved legacy metadata is missing for gRPC endpoint %q", e.Name()))
	}
	owner := expr.MethodPayloadExampleIdentity(e.MethodExpr)
	md := d.extractMetadata(endpointPlan.legacyMetadata, payload, sd, "server", "v", owner)
	declaration := d.symbols[e.Service].endpoints[e].legacyDecode
	data := &LegacyDecodeData{
		FuncName:        declaration.Name(),
		FuncDeclaration: declaration,
		Metadata:        md,
	}
	if expr.IsObject(payload.Type) {
		svcCtx := d.serviceTypeContext(sd, "server").Enter(payload)
		init := d.buildInitData(&expr.AttributeExpr{Type: expr.Empty}, payload, "message", "v", svcCtx, false, sd, owner, d.initDeclaration(e, true, grpcInitKey{role: grpcLegacyRequestInit}))
		init.Args = append(init.Args, initArgsFromMetadata(md)...)
		data.ServerConvert = &ConvertData{
			TgtName: svcCtx.Scope.Name(payload, svcCtx.Pkg(payload), svcCtx.Pointer, svcCtx.UseDefault),
			TgtRef:  svcCtx.Scope.Ref(payload, svcCtx.Pkg(payload)),
			Init:    init,
		}
	}
	return data
}

// buildServerResponseConverts builds one protobuf conversion for each result
// view the server may send. Results without views have one unnamed conversion.
func (d *ServicesData) buildServerResponseConverts(response, result *expr.AttributeExpr, svcCtx *codegen.AttributeContext, e *expr.GRPCEndpointExpr, sd *ServiceData) []*ViewConvertData {
	views := []string{""}
	if _, viewed := e.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
		views = grpcResultViews(e.MethodExpr)
	}
	converts := make([]*ViewConvertData, 0, len(views))
	for _, view := range views {
		source := result
		if view != "" {
			var err error
			source, err = grpcResultForView(result, view)
			if err != nil {
				panic(err) // bug
			}
		}
		key := grpcInitKey{role: grpcResponseInit, view: view}
		converts = append(converts, &ViewConvertData{
			View:    view,
			Convert: d.buildServerResponseConvertData(response, source, svcCtx, e, sd, key),
		})
	}
	return converts
}

// buildServerResponseConvertData builds one protobuf response conversion from
// the fields selected during planning.
func (d *ServicesData) buildServerResponseConvertData(response, result *expr.AttributeExpr, svcCtx *codegen.AttributeContext, e *expr.GRPCEndpointExpr, sd *ServiceData, key grpcInitKey) *ConvertData {
	data := d.buildInitData(result, response, "result", "message", svcCtx, true, sd, expr.MethodResultExampleIdentity(e.MethodExpr), d.initDeclaration(e, true, key))
	return &ConvertData{
		SrcName: svcCtx.Scope.Name(result, svcCtx.Pkg(result), svcCtx.Pointer, svcCtx.UseDefault),
		SrcRef:  svcCtx.Scope.Ref(result, svcCtx.Pkg(result)),
		TgtName: protoBufGoFullTypeName(response, sd.ServerProtobufPkgName, sd),
		TgtRef:  protoBufGoFullTypeRef(response, sd.ServerProtobufPkgName, sd),
		Init:    data,
	}
}

// buildClientResponseConverts builds one service conversion for each result
// view the client may receive. Results without views have one unnamed
// conversion.
func (d *ServicesData) buildClientResponseConverts(response *expr.AttributeExpr, validations map[string]*expr.AttributeExpr, result *expr.AttributeExpr, svcCtx *codegen.AttributeContext, hdrs, trlrs []*MetadataData, e *expr.GRPCEndpointExpr, sd *ServiceData) []*ViewConvertData {
	if e.MethodExpr.IsStreaming() || isEmpty(e.MethodExpr.Result.Type) {
		return nil
	}
	views := []string{""}
	if _, viewed := e.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
		views = grpcResultViews(e.MethodExpr)
	}
	converts := make([]*ViewConvertData, 0, len(views))
	for _, view := range views {
		target := result
		if view != "" {
			var err error
			target, err = grpcResultForView(result, view)
			if err != nil {
				panic(err) // bug
			}
		}
		key := grpcInitKey{role: grpcResponseInit, view: view}
		data := d.buildInitData(response, target, "message", "result", svcCtx, false, sd, expr.MethodResultExampleIdentity(e.MethodExpr), d.initDeclaration(e, false, key))
		data.Args = append(data.Args, initArgsFromMetadata(hdrs)...)
		data.Args = append(data.Args, initArgsFromMetadata(trlrs)...)
		validation := response
		if validations != nil {
			validation = validations[view]
		}
		convert := &ConvertData{
			SrcName:    protoBufGoFullTypeName(response, sd.ClientProtobufPkgName, sd),
			SrcRef:     protoBufGoFullTypeRef(response, sd.ClientProtobufPkgName, sd),
			TgtName:    svcCtx.Scope.Name(target, svcCtx.Pkg(target), svcCtx.Pointer, svcCtx.UseDefault),
			TgtRef:     svcCtx.Scope.Ref(target, svcCtx.Pkg(target)),
			Init:       data,
			Validation: addValidation(validation, sd, false),
		}
		converts = append(converts, &ViewConvertData{View: view, Convert: convert})
	}
	return converts
}

// buildInitData builds the transformation code to convert source to target.
//
// source, target are the source and target attributes used in the
// transformation
// sourceVar, targetVar are the source and target variable names used in the
// transformation
// svcCtx is the attribute context for service type
// proto if true indicates the target type is a protocol buffer type
func (d *ServicesData) buildInitData(source, target *expr.AttributeExpr, sourceVar, targetVar string, svcCtx *codegen.AttributeContext, proto bool, sd *ServiceData, owner expr.ExampleIdentity, conversion *grpcConversion) *InitData {
	protobufPackage := sd.ClientProtobufPkgName
	if conversion.side == grpcServerPackage {
		protobufPackage = sd.ServerProtobufPkgName
	}
	pbCtx := protoBufTypeContext(protobufPackage, sd)
	srcCtx := pbCtx
	tgtCtx := svcCtx
	if proto {
		srcCtx = svcCtx
		tgtCtx = pbCtx
	}
	isStruct := expr.IsObject(target.Type) || expr.IsUnion(target.Type)
	if !conversion.bound {
		if err := conversion.transform.BindContexts(srcCtx, tgtCtx); err != nil {
			panic(err) // bug
		}
		conversion.bound = true
	}
	code, helpers, err := conversion.transform.Render(sourceVar, targetVar, true)
	if err != nil {
		panic(err) // bug
	}
	if conversion.side == grpcServerPackage {
		sd.serverTransformHelpers = codegen.AppendHelpers(sd.serverTransformHelpers, helpers)
	} else {
		sd.clientTransformHelpers = codegen.AppendHelpers(sd.clientTransformHelpers, helpers)
	}
	var args []*InitArgData
	if (!proto && !isEmpty(source.Type)) || (proto && !isEmpty(target.Type)) {
		args = []*InitArgData{{
			Name:     sourceVar,
			Ref:      sourceVar,
			TypeName: srcCtx.Scope.Name(source, srcCtx.Pkg(source), srcCtx.Pointer, srcCtx.UseDefault),
			TypeRef:  srcCtx.Scope.Ref(source, srcCtx.Pkg(source)),
			Example:  d.Example(source, owner),
		}}
	}
	sourceRef := "metadata values"
	if !isEmpty(source.Type) {
		sourceRef = srcCtx.Scope.Ref(source, srcCtx.Pkg(source))
	}
	targetRef := tgtCtx.Scope.Ref(target, tgtCtx.Pkg(target))
	return &InitData{
		Declaration:    conversion.declaration,
		Name:           conversion.declaration.Name(),
		Description:    fmt.Sprintf("%s builds %s from %s.", conversion.declaration.Name(), targetRef, sourceRef),
		ReturnVarName:  targetVar,
		ReturnTypeRef:  targetRef,
		ReturnIsStruct: isStruct,
		ReturnTypePkg:  tgtCtx.Pkg(target),
		Code:           code,
		Args:           args,
	}
}

// buildCLIRequestTransform checks a parsed protobuf request and then renders
// the protobuf-to-payload conversion used by command-line clients.
func (d *ServicesData) buildCLIRequestTransform(
	endpoint *expr.GRPCEndpointExpr,
	request *expr.AttributeExpr,
	sd *ServiceData,
) string {
	conversion := d.symbols[endpoint.Service].endpoints[endpoint].cliPayload
	pbCtx := protoBufTypeContext(sd.ClientProtobufPkgName, sd)
	svcCtx := d.serviceTypeContext(sd, "client").Enter(endpoint.MethodExpr.Payload)
	if !conversion.bound {
		if err := conversion.transform.BindContexts(pbCtx, svcCtx); err != nil {
			panic(err) // bug
		}
		conversion.bound = true
	}
	code, helpers, err := conversion.transform.Render("message", "v", true)
	if err != nil {
		panic(err) // bug
	}
	sd.clientTransformHelpers = codegen.AppendHelpers(sd.clientTransformHelpers, helpers)
	if protobufCLIRequestNeedsMessage(request) {
		if validation := sd.protobuf.validation(request, validateClient); validation != nil {
			payloadRef := svcCtx.Scope.Ref(endpoint.MethodExpr.Payload, svcCtx.Pkg(endpoint.MethodExpr.Payload))
			code = fmt.Sprintf(
				"if err := %s(&message); err != nil {\n\tvar zero %s\n\treturn zero, err\n}\n%s",
				validation.Declaration.Name(), payloadRef, code,
			)
		}
	}
	return code
}

// protobufCLIRequestNeedsMessage reports whether the command-line client
// parses a protobuf request before building the service payload.
func protobufCLIRequestNeedsMessage(attribute *expr.AttributeExpr) bool {
	object := expr.AsObject(attribute.Type)
	return (object != nil && len(*object) > 0) || expr.IsUnion(attribute.Type)
}

// initDeclaration returns the constructor and conversion requested for one
// endpoint value. Planning records both before generated package names are
// fixed.
func (d *ServicesData) initDeclaration(endpoint *expr.GRPCEndpointExpr, server bool, key grpcInitKey) *grpcConversion {
	symbols := d.symbols[endpoint.Service].endpoints[endpoint]
	declarations := symbols.clientInits
	if server {
		declarations = symbols.serverInits
	}
	init := declarations[key]
	if init == nil {
		panic(fmt.Sprintf("constructor name is missing for gRPC endpoint %q", endpoint.Name()))
	}
	return init
}

// buildErrorsData builds the error data for all the error responses in the
// endpoint expression. The response message for each error response are
// inferred from the method's error expression if not specified explicitly.
// errorMessages maps error names to the protobuf shaped copy of their
// response message derived by analyze; errors without a custom object type
// have no entry.
func (d *ServicesData) buildErrorsData(e *expr.GRPCEndpointExpr, errorMessages map[string]*expr.AttributeExpr, sd *ServiceData) []*ErrorData {
	errors := make([]*ErrorData, 0, len(e.GRPCErrors))
	for _, v := range e.GRPCErrors {
		responseData := &ResponseData{
			StatusCode:    statusCodeToGRPCConst(v.Response.StatusCode),
			Description:   v.Response.Description,
			ServerConvert: d.buildErrorConvertData(v, e, errorMessages[v.Name], sd, true),
			ClientConvert: d.buildErrorConvertData(v, e, errorMessages[v.Name], sd, false),
		}
		svcctx := d.serviceTypeContext(sd, "server").Enter(v.AttributeExpr)
		errors = append(errors, &ErrorData{
			Name:     v.Name,
			Ref:      svcctx.Scope.Ref(v.AttributeExpr, svcctx.Pkg(v.AttributeExpr)),
			Response: responseData,
		})
	}
	return errors
}

// buildErrorConvertData builds the convert data for the given error response.
// message is the protobuf shaped copy of the error response message derived
// by analyze; it is nil for default errors and non-object error types.
func (d *ServicesData) buildErrorConvertData(ge *expr.GRPCErrorExpr, e *expr.GRPCEndpointExpr, message *expr.AttributeExpr, sd *ServiceData, svr bool) *ConvertData {
	// No need to build transformation functions for default error or non-object
	// types.
	if expr.IsErrorResult(ge.Type) || !expr.IsObject(ge.Type) {
		return nil
	}
	side := "client"
	if svr {
		side = "server"
	}
	svcCtx := d.serviceTypeContext(sd, side).Enter(ge.AttributeExpr)
	if svr {
		// server side
		owner := expr.MethodErrorExampleIdentity(e.MethodExpr, ge.ErrorExpr)
		data := d.buildInitData(ge.AttributeExpr, message, "er", "message", svcCtx, true, sd, owner, d.initDeclaration(e, true, grpcInitKey{role: grpcErrorInit, subject: ge.Name}))
		return &ConvertData{
			SrcName: svcCtx.Scope.Name(ge.AttributeExpr, svcCtx.Pkg(ge.AttributeExpr), svcCtx.Pointer, svcCtx.UseDefault),
			SrcRef:  svcCtx.Scope.Ref(ge.AttributeExpr, svcCtx.Pkg(ge.AttributeExpr)),
			TgtName: protoBufGoFullTypeName(message, sd.ServerProtobufPkgName, sd),
			TgtRef:  protoBufGoFullTypeRef(message, sd.ServerProtobufPkgName, sd),
			Init:    data,
		}
	}

	// client side
	owner := expr.MethodErrorExampleIdentity(e.MethodExpr, ge.ErrorExpr)
	data := d.buildInitData(message, ge.AttributeExpr, "message", "er", svcCtx, false, sd, owner, d.initDeclaration(e, false, grpcInitKey{role: grpcErrorInit, subject: ge.Name}))
	return &ConvertData{
		SrcName:    protoBufGoFullTypeName(message, sd.ClientProtobufPkgName, sd),
		SrcRef:     protoBufGoFullTypeRef(message, sd.ClientProtobufPkgName, sd),
		TgtName:    svcCtx.Scope.Name(ge.AttributeExpr, svcCtx.Pkg(ge.AttributeExpr), svcCtx.Pointer, svcCtx.UseDefault),
		TgtRef:     svcCtx.Scope.Ref(ge.AttributeExpr, svcCtx.Pkg(ge.AttributeExpr)),
		Init:       data,
		Validation: addValidation(message, sd, false),
	}
}

// buildServerStreamSendConverts builds one protobuf conversion for each result
// view the server may send through a stream.
func (d *ServicesData) buildServerStreamSendConverts(e *expr.GRPCEndpointExpr, response, result *expr.AttributeExpr, resultCtx *codegen.AttributeContext, sd *ServiceData) []*ViewConvertData {
	views := []string{""}
	if _, viewed := e.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
		views = grpcResultViews(e.MethodExpr)
	}
	converts := make([]*ViewConvertData, 0, len(views))
	for _, view := range views {
		source := result
		sourceVar := "result"
		if view != "" {
			var err error
			source, err = grpcResultForView(result, view)
			if err != nil {
				panic(err) // bug
			}
			sourceVar = "vresult"
		}
		key := grpcInitKey{role: grpcStreamingResponseInit, view: view}
		convert := &ConvertData{
			SrcName: resultCtx.Scope.Name(source, resultCtx.Pkg(source), resultCtx.Pointer, resultCtx.UseDefault),
			SrcRef:  resultCtx.Scope.Ref(source, resultCtx.Pkg(source)),
			TgtName: protoBufGoFullTypeName(response, sd.ServerProtobufPkgName, sd),
			TgtRef:  protoBufGoFullTypeRef(response, sd.ServerProtobufPkgName, sd),
			Init:    d.buildInitData(source, response, sourceVar, "v", resultCtx, true, sd, expr.MethodStreamingResultExampleIdentity(e.MethodExpr), d.initDeclaration(e, true, key)),
		}
		converts = append(converts, &ViewConvertData{View: view, Convert: convert})
	}
	return converts
}

// buildClientStreamRecvConverts builds one service conversion for each result
// view the client may receive through a stream.
func (d *ServicesData) buildClientStreamRecvConverts(e *expr.GRPCEndpointExpr, response *expr.AttributeExpr, validations map[string]*expr.AttributeExpr, result *expr.AttributeExpr, resultCtx *codegen.AttributeContext, sd *ServiceData) []*ViewConvertData {
	views := []string{""}
	if _, viewed := e.MethodExpr.Result.Type.(*expr.ResultTypeExpr); viewed {
		views = grpcResultViews(e.MethodExpr)
	}
	converts := make([]*ViewConvertData, 0, len(views))
	for _, view := range views {
		target := result
		targetVar := "result"
		if view != "" {
			var err error
			target, err = grpcResultForView(result, view)
			if err != nil {
				panic(err) // bug
			}
			targetVar = "vresult"
		}
		key := grpcInitKey{role: grpcStreamingResponseInit, view: view}
		validation := response
		if validations != nil {
			validation = validations[view]
		}
		convert := &ConvertData{
			SrcName:    protoBufGoFullTypeName(response, sd.ClientProtobufPkgName, sd),
			SrcRef:     protoBufGoFullTypeRef(response, sd.ClientProtobufPkgName, sd),
			TgtName:    resultCtx.Scope.Name(target, resultCtx.Pkg(target), resultCtx.Pointer, resultCtx.UseDefault),
			TgtRef:     resultCtx.Scope.Ref(target, resultCtx.Pkg(target)),
			Init:       d.buildInitData(response, target, "v", targetVar, resultCtx, false, sd, expr.MethodStreamingResultExampleIdentity(e.MethodExpr), d.initDeclaration(e, false, key)),
			Validation: addValidation(validation, sd, false),
		}
		converts = append(converts, &ViewConvertData{View: view, Convert: convert})
	}
	return converts
}

// buildStreamData builds the StreamData for the server and client streams.
//
// streamingRequest and responseMessage are the protobuf shaped copies of the
// endpoint streaming request and response message derived by analyze.
//
// svr param indicates that the stream data is built for the server.
func (d *ServicesData) buildStreamData(e *expr.GRPCEndpointExpr, streamingRequest, responseMessage *expr.AttributeExpr, responseValidations map[string]*expr.AttributeExpr, sd *ServiceData, svr bool) *StreamData {
	var (
		varn                string
		intName             string
		svcInt              string
		sendName            string
		sendDesc            string
		sendWithContextName string
		sendWithContextDesc string
		sendRef             string
		sendConvert         *ConvertData
		sendConverts        []*ViewConvertData
		recvName            string
		recvDesc            string
		recvWithContextName string
		recvWithContextDesc string
		recvRef             string
		recvConvert         *ConvertData
		recvConverts        []*ViewConvertData
		mustClose           bool
		typ                 string
	)
	ed := sd.Endpoint(e.Name())
	md := ed.Method
	side := "client"
	if svr {
		side = "server"
	}
	svcCtx := d.serviceTypeContext(sd, side).Enter(e.MethodExpr.StreamingPayload)
	result, resCtx := d.resultContext(e, sd, side)
	if svr {
		typ = "server"
		varn = md.ServerStream.VarName
		methodDescriptor := sd.protobuf.plan.serviceFullName() + "." + sd.protobuf.plan.methods[e]
		intName = sd.ServerProtobufPkgName + "." + sd.protobuf.plan.name(methodDescriptor, protocMethodServerStreamName)
		svcInt = fmt.Sprintf("%s.%s", sd.ServerServicePkgName, md.ServerStream.Interface)
		if e.MethodExpr.Result.Type != expr.Empty {
			sendName = md.ServerStream.SendName
			sendRef = ed.ResultRef
			sendWithContextName = md.ServerStream.SendWithContextName
			sendConverts = d.buildServerStreamSendConverts(e, responseMessage, result, resCtx, sd)
			sendConvert = primaryViewConvert(sendConverts)
			if md.ViewedResult == nil {
				sendConverts = nil
			}
		}
		if e.MethodExpr.StreamingPayload.Type != expr.Empty {
			recvName = md.ServerStream.RecvName
			recvWithContextName = md.ServerStream.RecvWithContextName
			recvRef = svcCtx.Scope.Ref(e.MethodExpr.StreamingPayload, svcCtx.Pkg(e.MethodExpr.StreamingPayload))
			recvConvert = &ConvertData{
				SrcName:    protoBufGoFullTypeName(streamingRequest, sd.ServerProtobufPkgName, sd),
				SrcRef:     protoBufGoFullTypeRef(streamingRequest, sd.ServerProtobufPkgName, sd),
				TgtName:    svcCtx.Scope.Name(e.MethodExpr.StreamingPayload, svcCtx.Pkg(e.MethodExpr.StreamingPayload), svcCtx.Pointer, svcCtx.UseDefault),
				TgtRef:     recvRef,
				Init:       d.buildInitData(streamingRequest, e.MethodExpr.StreamingPayload, "v", "spayload", svcCtx, false, sd, expr.MethodStreamingPayloadExampleIdentity(e.MethodExpr), d.initDeclaration(e, true, grpcInitKey{role: grpcStreamingRequestInit})),
				Validation: addValidation(streamingRequest, sd, true),
			}
		}
		mustClose = md.ServerStream.MustClose
	} else {
		typ = "client"
		varn = md.ClientStream.VarName
		methodDescriptor := sd.protobuf.plan.serviceFullName() + "." + sd.protobuf.plan.methods[e]
		intName = sd.ClientProtobufPkgName + "." + sd.protobuf.plan.name(methodDescriptor, protocMethodClientStreamName)
		svcInt = fmt.Sprintf("%s.%s", sd.ClientServicePkgName, md.ClientStream.Interface)
		if e.MethodExpr.StreamingPayload.Type != expr.Empty {
			sendName = md.ClientStream.SendName
			sendWithContextName = md.ClientStream.SendWithContextName
			sendRef = svcCtx.Scope.Ref(e.MethodExpr.StreamingPayload, svcCtx.Pkg(e.MethodExpr.StreamingPayload))
			sendConvert = &ConvertData{
				SrcName: svcCtx.Scope.Name(e.MethodExpr.StreamingPayload, svcCtx.Pkg(e.MethodExpr.StreamingPayload), svcCtx.Pointer, svcCtx.UseDefault),
				SrcRef:  sendRef,
				TgtName: protoBufGoFullTypeName(streamingRequest, sd.ClientProtobufPkgName, sd),
				TgtRef:  protoBufGoFullTypeRef(streamingRequest, sd.ClientProtobufPkgName, sd),
				Init:    d.buildInitData(e.MethodExpr.StreamingPayload, streamingRequest, "spayload", "v", svcCtx, true, sd, expr.MethodStreamingPayloadExampleIdentity(e.MethodExpr), d.initDeclaration(e, false, grpcInitKey{role: grpcStreamingRequestInit})),
			}
		}
		if e.MethodExpr.Result.Type != expr.Empty {
			recvName = md.ClientStream.RecvName
			recvWithContextName = md.ClientStream.RecvWithContextName
			recvRef = ed.ResultRef
			recvConverts = d.buildClientStreamRecvConverts(e, responseMessage, responseValidations, result, resCtx, sd)
			recvConvert = primaryViewConvert(recvConverts)
			if md.ViewedResult == nil {
				recvConverts = nil
			}
		}
		mustClose = md.ClientStream.MustClose
	}
	if sendConvert != nil {
		sendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint gRPC stream.", sendName, sendConvert.TgtName, md.Name)
		sendWithContextDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint gRPC stream with context.", sendWithContextName, sendConvert.TgtName, md.Name)
	}
	if recvConvert != nil {
		recvDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint gRPC stream.", recvName, recvConvert.SrcName, md.Name)
		recvWithContextDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint gRPC stream with context.", recvWithContextName, recvConvert.SrcName, md.Name)
	}
	return &StreamData{
		VarName:             varn,
		Type:                typ,
		Interface:           intName,
		ServiceInterface:    svcInt,
		Endpoint:            ed,
		SendName:            sendName,
		SendDesc:            sendDesc,
		SendWithContextName: sendWithContextName,
		SendWithContextDesc: sendWithContextDesc,
		SendRef:             sendRef,
		SendConvert:         sendConvert,
		SendConverts:        sendConverts,
		RecvName:            recvName,
		RecvDesc:            recvDesc,
		RecvWithContextName: recvWithContextName,
		RecvWithContextDesc: recvWithContextDesc,
		RecvRef:             recvRef,
		RecvConvert:         recvConvert,
		RecvConverts:        recvConverts,
		MustClose:           mustClose,
	}
}

// primaryViewConvert returns the conversion kept in the original single-value
// data field. A caller-selected result uses its default view. A fixed result
// has only its selected view.
func primaryViewConvert(converts []*ViewConvertData) *ConvertData {
	if len(converts) == 0 {
		return nil
	}
	if len(converts) == 1 {
		return converts[0].Convert
	}
	for _, convert := range converts {
		if convert.View == expr.DefaultView {
			return convert.Convert
		}
	}
	panic("caller-selected gRPC result views do not include the default view") // bug
}

// extractMetadata collects the request/response metadata from the given
// metadata attribute and service type (payload/result).
func (d *ServicesData) extractMetadata(a *expr.MappedAttributeExpr, service *expr.AttributeExpr, sd *ServiceData, side, decodeTarget string, owner expr.ExampleIdentity) []*MetadataData {
	plans, ok := d.metadataPlans[a]
	if !ok {
		panic("saved gRPC metadata fields are missing")
	}
	metadata := make([]*MetadataData, 0, len(plans))
	for _, plan := range plans {
		wire := plan.wire
		arr := expr.AsArray(wire.Type)
		mp := expr.AsMap(wire.Type)
		wireCtx := codegen.NewAttributeContext(false, false, true, "", plan.scope).Enter(wire)
		var cliValidation func(string) string
		if plan.validation != "" {
			cliValidation = grpcCLIValidationRenderer(wire, wireCtx, plan.name)
		}
		varn := codegen.Goify(plan.name, false)
		fieldName := plan.fieldName
		typeName := wireCtx.Scope.Name(wire, wireCtx.Pkg(wire), false, true)
		typeRef := wireCtx.Scope.Ref(wire, wireCtx.Pkg(wire))
		valueTypeRef := typeRef
		if plan.pointer {
			typeRef = "*" + typeRef
		}
		serviceVar := "payload"
		encodeSide := "client"
		if side == "client" {
			serviceVar = "result"
			encodeSide = "server"
		}
		fieldRef := serviceVar
		targetRef := decodeTarget
		if fieldName != "" {
			fieldRef += "." + fieldName
			targetRef += "." + fieldName
		}
		wireVar := varn + "Wire"
		encodeCode := d.metadataTransform(plan, fieldRef, wireVar, sd, encodeSide, true)
		decodeCode := d.metadataTransform(plan, varn, targetRef, sd, side, false)
		serviceContext := d.serviceTypeContext(sd, side).Enter(plan.serviceField)
		metadata = append(metadata, &MetadataData{
			Name:             plan.element,
			AttributeName:    plan.name,
			Description:      wire.Description,
			FieldName:        fieldName,
			FieldType:        plan.fieldType,
			FieldTypeRef:     serviceContext.Scope.Ref(plan.serviceField, serviceContext.Pkg(plan.serviceField)),
			ServiceAttribute: plan.serviceField,
			WireAttribute:    wire,
			VarName:          varn,
			WireVarName:      wireVar,
			EncodeCode:       encodeCode,
			DecodeCode:       decodeCode,
			Required:         plan.required,
			Type:             wire.Type,
			TypeName:         typeName,
			TypeRef:          typeRef,
			Pointer:          plan.pointer,
			Slice:            arr != nil,
			StringSlice:      arr != nil && arr.ElemType.Type.Kind() == expr.StringKind,
			Map:              mp != nil,
			MapStringSlice: mp != nil &&
				mp.KeyType.Type.Kind() == expr.StringKind &&
				mp.ElemType.Type.Kind() == expr.ArrayKind &&
				expr.AsArray(mp.ElemType.Type).ElemType.Type.Kind() == expr.StringKind,
			Validate:     plan.validation,
			CLIPlan:      cli.NewFlagPlan(wire, typeName, valueTypeRef, cliValidation),
			DefaultValue: wire.DefaultValue,
			Example:      d.FieldExample(wire, service, plan.name, owner),
		})
	}
	return metadata
}

// metadataTransform writes the conversion between one generated metadata
// value and its service field using the imports of the generated file.
func (d *ServicesData) metadataTransform(plan *grpcMetadataPlan, sourceVar, targetVar string, sd *ServiceData, side string, encode bool) string {
	wireCtx := codegen.NewAttributeContext(false, false, true, "", plan.scope).Enter(plan.wire)
	serviceCtx := d.serviceTypeContext(sd, side).Enter(plan.serviceField)
	sourceCtx, targetCtx := wireCtx, serviceCtx
	transform := plan.decode
	if encode {
		sourceCtx, targetCtx = serviceCtx, wireCtx
		transform = plan.encode
	}
	if err := transform.BindContexts(sourceCtx, targetCtx); err != nil {
		panic(err) // bug
	}
	valueVar := sourceVar
	if plan.pointer {
		valueVar = "*" + sourceVar
	}
	if encode {
		code, helpers, err := transform.Render(valueVar, targetVar, true)
		if err != nil {
			panic(err)
		}
		sd.appendMetadataHelpers(side, helpers)
		return code
	}
	if !plan.pointer {
		code, helpers, err := transform.Render(valueVar, targetVar, false)
		if err != nil {
			panic(err)
		}
		sd.appendMetadataHelpers(side, helpers)
		return code
	}
	converted := codegen.Goify(sourceVar, false) + "Service"
	code, helpers, err := transform.Render(valueVar, converted, true)
	if err != nil {
		panic(err)
	}
	sd.appendMetadataHelpers(side, helpers)
	return "if " + sourceVar + " != nil {\n" + code + "\n" + targetVar + " = &" + converted + "\n}\n"
}

// appendMetadataHelpers writes recursive metadata conversions on the same
// client or server side as the metadata codec that calls them.
func (sd *ServiceData) appendMetadataHelpers(side string, helpers []*codegen.TransformFunctionData) {
	if side == "server" {
		sd.serverTransformHelpers = codegen.AppendHelpers(sd.serverTransformHelpers, helpers)
	} else {
		sd.clientTransformHelpers = codegen.AppendHelpers(sd.clientTransformHelpers, helpers)
	}
}

// argsFromMetadata builds arguments that expose decoded metadata values.
func argsFromMetadata(md []*MetadataData) []*InitArgData {
	args := make([]*InitArgData, len(md))
	for i, m := range md {
		args[i] = &InitArgData{
			Name:         m.VarName,
			Ref:          m.VarName,
			FieldName:    m.FieldName,
			FieldType:    m.FieldType,
			FieldTypeRef: m.FieldTypeRef,
			TypeName:     m.TypeName,
			TypeRef:      m.TypeRef,
			Type:         m.Type,
			Pointer:      m.Pointer,
			Required:     m.Required,
			Validate:     m.Validate,
			CLIPlan:      m.CLIPlan,
			Example:      m.Example,
			DefaultValue: m.DefaultValue,
		}
	}
	return args
}

// grpcCLIValidationRenderer writes checks for the concrete value parsed from
// command-line metadata. Metadata fields use pointers to track presence, but
// the CLI has already proved presence and passes the parsed value itself.
func grpcCLIValidationRenderer(attribute *expr.AttributeExpr, context *codegen.AttributeContext, name string) func(string) string {
	valueContext := context.Dup()
	valueContext.Pointer = false
	return func(target string) string {
		return codegen.AttributeValidationCode(attribute, nil, valueContext, true, false, target, name)
	}
}

// initArgsFromMetadata adds the exact conversion that populates the service
// constructor result from each metadata argument.
func initArgsFromMetadata(md []*MetadataData) []*InitArgData {
	args := argsFromMetadata(md)
	for index, metadata := range md {
		args[index].InitCode = metadata.DecodeCode
	}
	return args
}

// usesStreamEnvelope reports whether the transport needs a typed stream
// envelope to carry both the one-shot method payload and streaming payload
// items.
func usesStreamEnvelope(e *expr.GRPCEndpointExpr) bool {
	return e.MethodExpr.IsPayloadStreaming() && !isEmpty(e.Request.Type)
}

// makeProtoBufStreamEnvelope builds the protobuf stream envelope that carries
// the initial request payload frame and subsequent stream item frames.
func makeProtoBufStreamEnvelope(request, stream *expr.AttributeExpr, tname string, owner expr.ExampleIdentity) *expr.AttributeExpr {
	initial := expr.DupAtt(request)
	initial.Meta = initial.Meta.Dup()
	initial.Meta["rpc:tag"] = []string{"1"}
	streamItem := expr.DupAtt(stream)
	streamItem.Meta = streamItem.Meta.Dup()
	streamItem.Meta["rpc:tag"] = []string{"2"}
	envelope := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{
				Name: "body",
				Attribute: &expr.AttributeExpr{
					Type: &expr.Union{
						TypeName: "body",
						Values: []*expr.NamedAttributeExpr{
							{Name: "initial_payload", Attribute: initial},
							{Name: "stream_item", Attribute: streamItem},
						},
					},
				},
			},
		},
		Validation: &expr.ValidationExpr{Required: []string{"body"}},
	}
	return makeProtoBufMessage(envelope, tname, owner)
}

// buildStreamEnvelopeData computes the generated Go names for the protobuf
// oneof field and wrapper types of the synthesized stream envelope.
func buildStreamEnvelopeData(envelope *expr.AttributeExpr, sd *ServiceData) *StreamEnvelopeData {
	body := envelope.Find("body")
	union := expr.AsUnion(body.Type)
	scope := &protoBufScope{service: sd, pkg: sd.ClientProtobufPkgName}
	serverScope := &protoBufScope{service: sd, pkg: sd.ServerProtobufPkgName}
	fieldName := scope.Field(body, "body", true)
	initialFieldName := scope.Field(union.Values[0].Attribute, union.Values[0].Name, true)
	streamItemFieldName := scope.Field(union.Values[1].Attribute, union.Values[1].Name, true)
	return &StreamEnvelopeData{
		FieldName:                  fieldName,
		InitialFieldName:           initialFieldName,
		InitialWrapperRef:          scope.OneofWrapper(union.Values[0].Attribute),
		ClientInitialWrapperRef:    scope.OneofWrapper(union.Values[0].Attribute),
		ServerInitialWrapperRef:    serverScope.OneofWrapper(union.Values[0].Attribute),
		StreamItemFieldName:        streamItemFieldName,
		StreamItemWrapperRef:       scope.OneofWrapper(union.Values[1].Attribute),
		ClientStreamItemWrapperRef: scope.OneofWrapper(union.Values[1].Attribute),
		ServerStreamItemWrapperRef: serverScope.OneofWrapper(union.Values[1].Attribute),
	}
}

// nativeMetadataAttribute copies a gRPC metadata value as a primitive or array
// of primitives. It removes named service types but keeps their default values
// and validation rules on the copy used by the transport.
func nativeMetadataAttribute(source *expr.AttributeExpr) *expr.AttributeExpr {
	if userType, ok := source.Type.(expr.UserType); ok {
		result := nativeMetadataAttribute(userType.Attribute())
		mergeNativeMetadataContract(result, source)
		return result
	}
	result := &expr.AttributeExpr{
		Description:  source.Description,
		DefaultValue: source.DefaultValue,
		UserExamples: source.UserExamples,
	}
	if source.Validation != nil {
		result.Validation = source.Validation.Dup()
	}
	if source.Meta != nil {
		result.Meta = source.Meta.Dup()
	}
	switch actual := source.Type.(type) {
	case expr.Primitive:
		result.Type = actual
	case *expr.Array:
		result.Type = &expr.Array{
			ElemType:         nativeMetadataAttribute(actual.ElemType),
			NonNullableElems: actual.NonNullableElems,
		}
	default:
		panic(fmt.Sprintf("invalid gRPC metadata type %s", source.Type.Name()))
	}
	stripMetadataServiceNames(result)
	return result
}

// mergeNativeMetadataContract applies constraints authored on an alias use to
// the detached contract inherited from the alias declaration.
func mergeNativeMetadataContract(target, source *expr.AttributeExpr) {
	if source.Description != "" {
		target.Description = source.Description
	}
	if source.DefaultValue != nil {
		target.DefaultValue = source.DefaultValue
	}
	if source.Validation != nil {
		if target.Validation == nil {
			target.Validation = source.Validation.Dup()
		} else {
			target.Validation.Merge(source.Validation)
		}
	}
	if source.Meta != nil {
		if target.Meta == nil {
			target.Meta = make(expr.MetaExpr)
		}
		for name, values := range source.Meta {
			target.Meta[name] = append([]string(nil), values...)
		}
	}
	stripMetadataServiceNames(target)
}

// stripMetadataServiceNames removes Go service declaration overrides from a
// value rendered entirely in the generated gRPC client or server package.
func stripMetadataServiceNames(attribute *expr.AttributeExpr) {
	for name := range attribute.Meta {
		if strings.HasPrefix(name, "struct:") || name == "name:original" {
			delete(attribute.Meta, name)
		}
	}
}

// serviceTypeContext returns a context that resolves service declarations from
// the generated gRPC package for side.
func (d *ServicesData) serviceTypeContext(sd *ServiceData, side string) *codegen.AttributeContext {
	outputPackage := path.Join(d.GenPkg(), "grpc", sd.Service.PathName, side)
	return &codegen.AttributeContext{
		UseDefault: true,
		Scope:      d.ServiceAttributor(sd.Service.Name, outputPackage),
	}
}

// resultContext returns the method result and the final service or view type
// names used by the generated client or server package.
func (d *ServicesData) resultContext(e *expr.GRPCEndpointExpr, sd *ServiceData, side string) (*expr.AttributeExpr, *codegen.AttributeContext) {
	md := sd.Service.Method(e.Name())
	if md.ViewedResult != nil {
		vresAtt := expr.AsObject(md.ViewedResult.Type).Attribute("projected")
		outputPackage := path.Join(d.GenPkg(), "grpc", sd.Service.PathName, side)
		return vresAtt, &codegen.AttributeContext{
			Pointer:    true,
			UseDefault: true,
			Scope:      d.ViewAttributor(sd.Service.Name, outputPackage),
		}
	}
	result := e.MethodExpr.Result
	return result, d.serviceTypeContext(sd, side).Enter(result)
}

// getPrimitive returns the primitive expression if the given expression is an alias to one
func getPrimitive(att *expr.AttributeExpr) *expr.AttributeExpr {
	if ut, ok := att.Type.(*expr.UserTypeExpr); ok {
		if _, ok := ut.Type.(expr.Primitive); ok {
			return ut.AttributeExpr
		}
		return getPrimitive(ut.AttributeExpr)
	}
	return nil
}

// unAlias returns the base attribute of the given attribute when it is an
// alias to a primitive type, the attribute itself otherwise.
func unAlias(at *expr.AttributeExpr) *expr.AttributeExpr {
	if prim := getPrimitive(at); prim != nil {
		return prim
	}
	return at
}

// isEmpty returns true if given type is empty.
func isEmpty(dt expr.DataType) bool {
	if dt == expr.Empty {
		return true
	}
	if o := expr.AsObject(dt); o != nil && len(*o) == 0 {
		return true
	}
	return false
}

// usesAnyType reports whether any of the given endpoints uses the Any type in
// its payload, result or, when includeErrors is true, error attributes. It is
// used to determine whether generated files need the structpb import.
func usesAnyType(endpoints []*expr.GRPCEndpointExpr, includeErrors bool) bool {
	for _, e := range endpoints {
		if hasAnyType(e.MethodExpr.Payload) || hasAnyType(e.MethodExpr.Result) {
			return true
		}
		if !includeErrors {
			continue
		}
		for _, er := range e.MethodExpr.Errors {
			if hasAnyType(er.AttributeExpr) {
				return true
			}
		}
	}
	return false
}

// hasAnyType reports whether the attribute uses Any without following a named
// type more than once.
func hasAnyType(att *expr.AttributeExpr) bool {
	return hasAnyTypeR(att, make(map[expr.UserType]struct{}))
}

// hasAnyTypeR walks arrays, maps, objects, unions, and named types.
func hasAnyTypeR(att *expr.AttributeExpr, seen map[expr.UserType]struct{}) bool {
	if att == nil {
		return false
	}
	if att.Type.Kind() == expr.AnyKind {
		return true
	}
	switch dt := att.Type.(type) {
	case expr.UserType:
		origin := dt.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return hasAnyTypeR(dt.Attribute(), seen)
	case *expr.Object:
		for _, nat := range *dt {
			if hasAnyTypeR(nat.Attribute, seen) {
				return true
			}
		}
	case *expr.Array:
		return hasAnyTypeR(dt.ElemType, seen)
	case *expr.Map:
		return hasAnyTypeR(dt.KeyType, seen) || hasAnyTypeR(dt.ElemType, seen)
	case *expr.Union:
		for _, nat := range dt.Values {
			if hasAnyTypeR(nat.Attribute, seen) {
				return true
			}
		}
	}
	return false
}
