package codegen

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Client template constants
const (
	grpcClientStructT       = "client_struct"
	grpcClientInitT         = "client_init"
	grpcClientEndpointInitT = "client_endpoint_init"
	grpcRequestEncoderT     = "request_encoder"
	grpcResponseDecoderT    = "response_decoder"
)

// Server template constants
const (
	grpcServerStructTypeT    = "server_struct_type"
	grpcServerInitT          = "server_init"
	grpcServerGRPCInitT      = "server_grpc_init"
	grpcServerGRPCInterfaceT = "server_grpc_interface"
	grpcServerGRPCRegisterT  = "server_grpc_register"
	grpcServerGRPCStartT     = "server_grpc_start"
	grpcServerGRPCEndT       = "server_grpc_end"
	grpcRequestDecoderT      = "request_decoder"
	grpcResponseEncoderT     = "response_encoder"
	grpcHandlerInitT         = "grpc_handler_init"
)

// Stream template constants
const (
	grpcStreamStructTypeT = "stream_struct_type"
	grpcStreamSendT       = "stream_send"
	grpcStreamRecvT       = "stream_recv"
	grpcStreamCloseT      = "stream_close"
	grpcStreamSetViewT    = "stream_set_view"
)

// Proto template constants
const (
	grpcProtoHeaderT = "proto_header"
	grpcProtoStartT  = "proto_start"
	grpcServiceT     = "grpc_service"
	grpcMessageT     = "grpc_message"
)

// CLI template constants
const (
	grpcDoGRPCCLIT           = "do_grpc_cli"
	grpcParseEndpointT       = "parse_endpoint"
	grpcRemoteMethodBuilderT = "remote_method_builder"
)

// Transform template constants
const (
	grpcTransformGoArrayT          = "transform_go_array"
	grpcTransformGoMapT            = "transform_go_map"
	grpcTransformGoUnionFromProtoT = "transform_go_union_from_proto"
	grpcTransformGoUnionToProtoT   = "transform_go_union_to_proto"
)

// Partial template constants
const (
	grpcConvertTypeToStringP = "convert_type_to_string"
	grpcConvertStringToTypeP = "convert_string_to_type"
)

// Common template constants
const (
	grpcTypeInitT        = "type_init"
	grpcValidateT        = "validate"
	grpcTransformHelperT = "transform_helper"
)

//go:embed templates/*
var templateFS embed.FS

// grpcTemplates is the shared template reader for the grpc codegen package (package-private).
var grpcTemplates = &template.TemplateReader{FS: templateFS}
