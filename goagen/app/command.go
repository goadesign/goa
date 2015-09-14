package app

import (
	"github.com/raphael/goa/goagen/bootstrap"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Command is the goa application code generator command line data structure.
// It implements bootstrap.Command.
type Command struct {
	Generator     *bootstrap.MetaGenerator
	TargetPackage string // Target package name
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	gen := bootstrap.MetaGenerator{
		Factory: "app.NewGenerator",
		Imports: []string{"github.com/raphael/goa/goagen/app"},
	}
	return &Command{Generator: &gen}
}

// Name of command.
func (c *Command) Name() string { return "app" }

// Description of command.
func (c *Command) Description() string { return "Generate application code." }

// RegisterFlags registers the command line flags with the given command clause.
func (c *Command) RegisterFlags(cmd *kingpin.CmdClause) {
	gen := c.Generator
	gen.Flags = make(map[string]*string)
	var outdir string
	var targetPackage string
	cmd.Flag("out", "destination directory").Required().StringVar(&outdir)
	cmd.Flag("package", "target package").Required().StringVar(&targetPackage)
	gen.Flags["OutDir"] = &outdir
	gen.Flags["TargetPackage"] = &targetPackage
}

// Run simply calls the generator.
func (c *Command) Run() ([]string, error) {
	return c.Generator.Generate()
}
