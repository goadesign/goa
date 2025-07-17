package codegen

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Server template constants
const (
	// Server
	serverHandlerT     = "server_handler"
	serverHandlerInitT = "server_handler_init"
	serverInitT        = "server_init"
	serverStructT      = "server_struct"
	serverServiceT     = "server_service"
	serverUseT         = "server_use"
	serverMethodNamesT = "server_method_names"
	serverMountT       = "server_mount"

	// Server example
	serverConfigureT = "server_configure"

	// Client
	clientStructT    = "client_struct"
	clientInitT      = "client_init"
	endpointInitT    = "endpoint_init"
	responseDecoderT = "response_decoder"

	// Partial templates
	clientTypeConversionP   = "client_type_conversion"
	clientMapConversionP    = "client_map_conversion"
	singleResponseP         = "single_response"
	queryTypeConversionP    = "query_type_conversion"
	elementSliceConversionP = "element_slice_conversion"
	sliceItemConversionP    = "slice_item_conversion"
)

//go:embed templates/*
var templateFS embed.FS

// jsonrpcTemplates is the shared template reader for the jsonrpc codegen package (package-private).
var jsonrpcTemplates = &template.TemplateReader{FS: templateFS}
