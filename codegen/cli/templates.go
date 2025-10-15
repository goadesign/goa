package cli

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Template constants
const (
	usageCommandsT = "usage_commands"
	usageExamplesT = "usage_examples"
	parseFlagsT    = "parse_flags"
	commandUsageT  = "command_usage"
	buildPayloadT  = "build_payload"
)

//go:embed templates/*.go.tpl
var templateFS embed.FS

// cliTemplates is the shared template reader for the cli package.
var cliTemplates = &template.TemplateReader{FS: templateFS}