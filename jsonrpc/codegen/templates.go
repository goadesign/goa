package codegen

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Server template constants
const (
	// Server
	serverHandlerT      = "server_handler"
	serverHandlerInitT  = "server_handler_init"
	serverInitT         = "server_init"
	serverStructT       = "server_struct"
	serverServiceT      = "server_service"
	serverUseT          = "server_use"
	serverMethodNamesT  = "server_method_names"
	serverMountT        = "server_mount"
	serverEncodeErrorT  = "server_encode_error"
	mixedServerHandlerT = "mixed_server_handler"

	// Client
	clientStructT       = "client_struct"
	clientInitT         = "client_init"
	clientEndpointInitT = "client_endpoint_init"
	responseDecoderT    = "response_decoder"

	// WebSocket templates
	websocketServerStreamT        = "websocket_server_stream"
	websocketServerStreamWrapperT = "websocket_server_stream_wrapper"
	websocketServerHandlerT       = "websocket_server_handler"
	websocketServerSendT          = "websocket_server_send"
	websocketServerRecvT          = "websocket_server_recv"
	websocketServerCloseT         = "websocket_server_close"

	// JSON-RPC WebSocket client templates
	websocketClientConnT       = "websocket_client_conn"
	websocketClientStreamT     = "websocket_client_stream"
	websocketStreamErrorTypesT = "websocket_stream_error_types"

	// SSE templates
	sseServerStreamBaseT = "sse_server_stream_base"
	sseServerStreamT     = "sse_server_stream"
	sseClientStreamT     = "sse_client_stream"
	sseServerHandlerT    = "sse_server_handler"

	// Partial templates
	singleResponseP         = "single_response"
	queryTypeConversionP    = "query_type_conversion"
	elementSliceConversionP = "element_slice_conversion"
	sliceItemConversionP    = "slice_item_conversion"
)

//go:embed templates/*
var templateFS embed.FS

// jsonrpcTemplates is the shared template reader for the jsonrpc codegen package (package-private).
var jsonrpcTemplates = &template.TemplateReader{FS: templateFS}
