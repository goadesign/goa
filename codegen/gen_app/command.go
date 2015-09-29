package genapp

import (
	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/codegen/meta"
)

var (
	// TargetPackage is the name of the generated Go package.
	TargetPackage string

	// OutputDir is the path to the output directory.
	OutputDir string
)

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("app", "Generate application GoGenerator")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("pkg", "target package").Default("app").StringVar(&TargetPackage)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"pkg": TargetPackage}
	gen := meta.NewGenerator(
		"app.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/codegen/app")},
		flags,
	)
	return gen.Generate()
}
