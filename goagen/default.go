package main

import (
	"os"

	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_app"
	"github.com/raphael/goa/goagen/gen_main"
	"github.com/raphael/goa/goagen/gen_schema"
)

//  DefaultCommands lists the commands run by default when no sub-command is provided on the
// command line.
var DefaultCommands = []codegen.Command{
	genapp.NewCommand(),
	genmain.NewCommand(),
	genschema.NewCommand(),
}

// DefaultCommand is the default command. It runs all known commands.
type DefaultCommand struct{}

// Name returns the command name.
func (a *DefaultCommand) Name() string { return "default" }

// Description returns the command description.
func (a *DefaultCommand) Description() string { return "Default command, generates all artifacts." }

// RegisterFlags registers all the sub-commands flags.
func (a *DefaultCommand) RegisterFlags(r codegen.FlagRegistry) {
	for _, c := range DefaultCommands {
		if c != a {
			c.RegisterFlags(r)
		}
	}
}

// Run runs each known command and returns all the generated files and/or errors.
func (a *DefaultCommand) Run() ([]string, error) {
	var all []string
	var err error
	for _, c := range DefaultCommands {
		if c != a {
			var files []string
			files, err = c.Run()
			if err != nil {
				break
			}
			all = append(all, files...)
		}
	}
	if err != nil {
		for _, f := range all {
			os.Remove(f)
		}
		return nil, err
	}
	return all, nil
}
