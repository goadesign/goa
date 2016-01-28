package genmain

import (
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
)

var (
	// AppName is the name of the generated application.
	AppName string

	// TargetPackage is the name of the generated Go package.
	TargetPackage string

	// Force is true if pre-existing files should be overwritten during generation.
	Force bool
)

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
	r.Flags().BoolVar(&Force, "force", false, "overwrite existing files")
	r.Flags().StringVar(&AppName, "name", "API", "application name")
	if r.Flags().Lookup("pkg") == nil {
		// Special case because the bootstrap command calls RegisterFlags on genapp which
		// already registers that flag.
		r.Flags().StringVar(&TargetPackage, "pkg", "app", "Name of generated Go package containing controllers supporting code (contexts, media types, user types etc.)")
	}
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"name": AppName}
	gen := meta.NewGenerator(
		"genmain.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/goadesign/goa/goagen/gen_main")},
		flags,
	)
	return gen.Generate()
}
