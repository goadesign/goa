package gendocs

import "github.com/raphael/goa/goagen"

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*goagen.TBDCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	t := goagen.NewTBDCommand("Generate documentation", "")
	return &Command{TBDCommand: t}
}

// Name of command.
func (c *Command) Name() string { return "docs" }
