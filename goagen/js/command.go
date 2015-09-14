package js

import "github.com/raphael/goa/goagen/bootstrap"

// Command is the goa application code generator command line data structure.
// It implements bootstrap.Command.
type Command struct {
	*bootstrap.TBDCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	t := bootstrap.NewTBDCommand("Generate javascript client", "")
	return &Command{TBDCommand: t}
}

// Name of command.
func (c *Command) Name() string { return "js" }
