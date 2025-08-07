package codegen

import (
	"embed"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/template"
)

// Server template constants
const (
	// Server
	serverHandlerT           = "server_handler"
	serverHandlerInitT       = "server_handler_init"
	serverInitT              = "server_init"
	serverStructT            = "server_struct"
	serverServiceT           = "server_service"
	serverUseT               = "server_use"
	serverMethodNamesT       = "server_method_names"
	serverMountT             = "server_mount"
	serverEncodeErrorT       = "server_encode_error"
	mixedServerHandlerT      = "mixed_server_handler"

	// Server example
	serverConfigureT = "server_configure"
	serverHttpStartT = "server_http_start"

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
	sseServerStreamT     = "sse_server_stream"
	sseClientStreamT     = "sse_client_stream"
	sseServerStreamImplT = "sse_server_stream_impl"
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

// updateHeader modifies the header of the given file to be JSON-RPC specific.
func updateHeader(f *codegen.File) {
	// Update the title
	header := f.SectionTemplates[0]
	title := strings.Replace(header.Data.(map[string]any)["Title"].(string), "HTTP", "JSON-RPC", 1)
	header.Data.(map[string]any)["Title"] = title

	// Update the imports
	imports := header.Data.(map[string]any)["Imports"].([]*codegen.ImportSpec)
	for _, i := range imports {
		i.Path = strings.Replace(i.Path, "gen/http", "gen/jsonrpc", 1)
	}
}
