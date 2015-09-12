package docs

import "github.com/raphael/goa/goagen/bootstrap"

// Command is the goa application code generator command line data structure.
// It implements generator.Command.
type Command struct {
	*bootstrap.TBDCommand
}

// New instantiates a new command.
func New() *Command {
	t := bootstrap.NewTBDCommand("Generate documentation", "")
	return &Command{TBDCommand: t}
}

// Name of command.
func (c *Command) Name() string { return "docs" }
