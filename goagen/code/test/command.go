package test

import "github.com/raphael/goa/goagen/bootstrap"

// Command is the goa application code generator command line data structure.
// It implements generator.Command.
type Command struct {
	*bootstrap.BaseCommand
}

// New instantiates a new command.
func New() *Command {
	return &Command{BaseCommand: new(bootstrap.BaseCommand)}
}

// Name of command.
func (c *Command) Name() string { return "test" }

// Description of command.
func (c *Command) Description() string { return "Generate application test code." }
