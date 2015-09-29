package genmain

import (
	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/codegen/meta"
)

// AppName is the name of the generated application.
var AppName string

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("main", "Generate application main skeleton")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("name", "application name").Default("app").StringVar(&AppName)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"name": AppName}
	gen := meta.NewGenerator(
		"main.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/codegen/gen_main")},
		flags,
	)
	return gen.Generate()
}
