package codegen

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Server template constants
const (
	serverHandlerT     = "server_handler"
	serverHandlerInitT = "server_handler_init"
	serverInitT        = "server_init"
	serverStructT      = "server_struct"
	serverServiceT     = "server_service"
	serverUseT         = "server_use"
	serverMethodNamesT = "server_method_names"
	serverMountT       = "server_mount"
)

//go:embed templates/*
var templateFS embed.FS

// jsonrpcTemplates is the shared template reader for the jsonrpc codegen package (package-private).
var jsonrpcTemplates = &template.TemplateReader{FS: templateFS}
