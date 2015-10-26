package genapp

import (
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/codegen/meta"
)

var (
	// TargetPackage is the name of the generated Go package.
	TargetPackage string

	// AppSubDir is the name of the output directory sub-directory where application files are
	// generated.
	AppSubDir string
)

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
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
	r.Flag("pkg", "target package").Default("app").StringVar(&TargetPackage)
	r.Flag("subdir", "name of output sub-directory where application files are generated").Default("app").StringVar(&AppSubDir)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"pkg": TargetPackage}
	gen := meta.NewGenerator(
		"genapp.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/raphael/goa/goagen/codegen/gen_app")},
		flags,
	)
	return gen.Generate()
}
