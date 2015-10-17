package genmetadata

import (
	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/codegen/meta"
)

// HostName is used to build the JSON schema ID of the root document.
var HostName string

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("metadata", "Generate application metadata controller")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	gen := meta.NewGenerator(
		"genmetadata.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/codegen/gen_metadata")},
		nil,
	)
	return gen.Generate()
}
