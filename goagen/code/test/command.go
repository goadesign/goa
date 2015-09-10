package test

import "github.com/raphael/goa/goagen/bootstrap"

// init registers the command with the bootstrap tool.
func init() {
	bootstrap.Commands = append(bootstrap.Commands, new(Command))
}

// Command is the goa application code generator command line data structure.
// It implements generator.Command.
type Command struct {
	*bootstrap.BaseCommand
}

// Name of command.
func (c *Command) Name() string { return "test" }

// Description of command.
func (c *Command) Description() string { return "test code" }
