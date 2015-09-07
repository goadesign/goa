package documentation

import "github.com/raphael/goa/codegen/bootstrap"

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
func (c *Command) Name() string { return "docs" }

// Description of command.
func (c *Command) Description() string { return "documentation code" }
