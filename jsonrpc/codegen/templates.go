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
