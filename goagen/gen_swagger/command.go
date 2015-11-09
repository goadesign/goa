package genswagger

import (
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/meta"
)

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("swagger", "Generate swagger representation, see http://swagger.io")
	return &Command{BaseCommand: base}
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	gen := meta.NewGenerator(
		"genswagger.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/goagen/gen_swagger")},
		nil,
	)
	return gen.Generate()
}
