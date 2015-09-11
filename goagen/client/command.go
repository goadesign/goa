package client

import (
	"github.com/raphael/goa/goagen/bootstrap"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Command is the goa application code generator command line data structure.
// It implements generator.Command.
type Command struct {
	*bootstrap.BaseCommand
	TargetPackage string // Target package name
}

// New instantiates a new command.
func New() *Command {
	return &Command{BaseCommand: new(bootstrap.BaseCommand)}
}

// Name of command.
func (c *Command) Name() string { return "client" }

// Description of command.
func (c *Command) Description() string { return "Generate API client tool." }

// RegisterFlags registers the command line flags with the given command clause.
func (c *Command) RegisterFlags(cmd *kingpin.CmdClause) {
	c.BaseCommand.RegisterFlags(cmd)
	var targetPackage string
	cmd.Flag("package", "target package").Required().StringVar(&targetPackage)
	c.Flags["TargetPackage"] = &targetPackage
}
