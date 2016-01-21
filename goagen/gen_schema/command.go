package genschema

import (
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
)

// ServiceURL is used to build the JSON schema ID of the root document.
var ServiceURL string

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("schema", "Generate application JSON schema controller")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("url", "API base URL used to build JSON schema ID, e.g. https://www.myapi.com").
		Short('u').
		Default("http://localhost").
		StringVar(&ServiceURL)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	flags := map[string]string{"url": ServiceURL}
	gen := meta.NewGenerator(
		"genschema.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/goadesign/goa/goagen/gen_schema")},
		flags,
	)
	return gen.Generate()
}
