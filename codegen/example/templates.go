package example

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Template constants
const (
	// Client templates
	clientStartT        = "client_start"
	clientVarInitT      = "client_var_init"
	clientEndpointInitT = "client_endpoint_init"
	clientEndT          = "client_end"
	clientUsageT        = "client_usage"

	// Server templates
	serverStartT        = "server_start"
	serverLoggerT       = "server_logger"
	serverServicesT     = "server_services"
	serverInterceptorsT = "server_interceptors"
	serverEndpointsT    = "server_endpoints"
	serverInterruptsT   = "server_interrupts"
	serverHandlerT      = "server_handler"
	serverEndT          = "server_end"
)

//go:embed templates/*
var templateFS embed.FS

// exampleTemplates is the shared template reader for the example codegen package (package-private).
var exampleTemplates = &template.TemplateReader{FS: templateFS}
