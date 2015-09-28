package genapp

import (
	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/meta"
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
	*goagen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := goagen.NewBaseCommand("app", "Generate application GoGenerator")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r goagen.FlagRegistry) {
	r.Flag("target", "target package").Default("app").StringVar(&TargetPackage)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"target": TargetPackage}
	gen := meta.NewGenerator(
		"app.Generate",
		[]*goagen.ImportSpec{goagen.SimpleImport("github.com/raphael/goa/goagen/app")},
		flags,
	)
	return gen.Generate()
}
