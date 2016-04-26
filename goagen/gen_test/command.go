package gentest

import (
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
)

var (
	// AppPkg is the package path to the generated application code.
	// This is needed to get access to the payload types.
	AppPkg string

	// TargetPackage is the name of the generated Go package.
	TargetPackage string
)

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("test", "Generate Controller test helpers")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	if r.Flags().Lookup("appPkg") == nil {
		r.Flags().StringVar(&AppPkg, "appPkg", "app", "Package path to generated application code")
	}
	if r.Flags().Lookup("pkg") == nil {
		// Special case because the bootstrap command calls RegisterFlags on genapp which
		// already registers that flag.
		r.Flags().StringVar(&TargetPackage, "pkg", "app", "Name of generated Go package containing controllers supporting code (contexts, media types, user types etc.)")
	}

}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"pkg": TargetPackage, "appPkg": AppPkg}
	gen := meta.NewGenerator(
		"gentest.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/goadesign/goa/goagen/gen_test")},
		flags,
	)
	return gen.Generate()
}
