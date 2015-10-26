package genapp

import (
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/meta"
)

var (
	// TargetPackage is the name of the generated Go package.
	TargetPackage string
)

// Command is the goa application code generator command line data structure.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("app", "Generate application code")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("pkg", "Name of generated Go package containing controllers supporting code (contexts, media types, user types etc.)").
		Default("app").StringVar(&TargetPackage)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"pkg": TargetPackage}
	gen := meta.NewGenerator(
		"genapp.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/goagen/gen_app")},
		flags,
	)
	return gen.Generate()
}
