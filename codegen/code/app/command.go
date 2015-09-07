package app

import (
	"github.com/raphael/goa/codegen/bootstrap"
	"gopkg.in/alecthomas/kingpin.v2"
)

// init registers the command with the bootstrap tool.
func init() {
	bootstrap.Commands = append(bootstrap.Commands, new(Command))
}

// Command is the goa application code generator command line data structure.
// It implements generator.Command.
type Command struct {
	*bootstrap.BaseCommand
	TargetPackage string // Target package name
}

// Name of command.
func (c *Command) Name() string { return "app" }

// Description of command.
func (c *Command) Description() string { return "application code" }

// RegisterFlags registers the command line flags with the given command clause.
func (c *Command) RegisterFlags(cmd *kingpin.CmdClause) {
	c.BaseCommand.RegisterFlags(cmd)
	targetPackage := ""
	cmd.Flag("package", "target package").Required().StringVar(&targetPackage)
	c.Flags["TargetPackage"] = targetPackage
}
