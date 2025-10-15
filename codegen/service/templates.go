package service

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Template constants
const (
	// Client templates
	serviceClientT       = "service_client"
	serviceClientInitT   = "service_client_init"
	serviceClientMethodT = "service_client_method"

	// Convert templates
	convertT         = "convert"
	createT          = "create"
	transformHelperT = "transform_helper"

	// Endpoint templates
	serviceEndpointsT            = "service_endpoints"
	serviceEndpointStreamStructT = "service_endpoint_stream_struct"
	serviceRequestBodyStructT    = "service_request_body_struct"
	serviceResponseBodyStructT   = "service_response_body_struct"
	serviceEndpointsInitT        = "service_endpoints_init"
	serviceEndpointsUseT         = "service_endpoints_use"
	serviceEndpointMethodT       = "service_endpoint_method"

	// Example interceptor templates
	exampleServerInterceptorT = "example_server_interceptor"
	exampleClientInterceptorT = "example_client_interceptor"

	// Example service templates
	exampleServiceStructT     = "example_service_struct"
	exampleServiceInitT       = "example_service_init"
	exampleSecurityAuthfuncsT = "example_security_authfuncs"
	endpointT                 = "endpoint"
	jsonrpcHandleStreamT      = "jsonrpc_handle_stream"

	// Service templates
	serviceT          = "service"
	payloadT          = "payload"
	streamingPayloadT = "streaming_payload"
	resultT           = "result"
	userTypeT         = "user_type"
	unionValueMethodT = "union_value_method"
	errorT            = "error"
	errorInitT        = "error_init"
	typeInitT         = "type_init"
	returnTypeInitT   = "return_type_init"
	typeValidateT     = "type_validate"
	validateT         = "validate"
	viewedTypeMapT    = "viewed_type_map"

	// Interceptor templates
	interceptorsT                        = "interceptors"
	interceptorsTypesT                   = "interceptors_types"
	serverInterceptorsT                  = "server_interceptors"
	clientInterceptorsT                  = "client_interceptors"
	endpointWrappersT                    = "endpoint_wrappers"
	clientWrappersT                      = "client_wrappers"
	serverInterceptorStreamWrapperTypesT = "server_interceptor_stream_wrapper_types"
	clientInterceptorStreamWrapperTypesT = "client_interceptor_stream_wrapper_types"
	serverInterceptorWrappersT           = "server_interceptor_wrappers"
	clientInterceptorWrappersT           = "client_interceptor_wrappers"
	serverInterceptorStreamWrappersT     = "server_interceptor_stream_wrappers"
	clientInterceptorStreamWrappersT     = "client_interceptor_stream_wrappers"
)

//go:embed templates/*
var templateFS embed.FS

// serviceTemplates is the shared template reader for the service codegen package (package-private).
var serviceTemplates = &template.TemplateReader{FS: templateFS}
