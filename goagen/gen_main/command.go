package genmain

import (
	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/meta"
)

// AppName is the name of the generated application.
var AppName string

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*goagen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := goagen.NewBaseCommand("main", "Generate application main skeleton")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r goagen.FlagRegistry) {
	r.Flag("name", "application name").Default("app").StringVar(&AppName)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"name": AppName}
	gen := meta.NewGenerator(
		"genmain.Generate",
		[]*goagen.ImportSpec{goagen.SimpleImport("github.com/raphael/goa/goagen/gen_main")},
		flags,
	)
	return gen.Generate()
}
