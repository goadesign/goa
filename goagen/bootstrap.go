package main

import (
	"os"

	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_app"
	"github.com/raphael/goa/goagen/gen_main"
	"github.com/raphael/goa/goagen/gen_schema"
)

// BootstrapCommands lists the commands run by default when no sub-command is provided on the
// command line.
var BootstrapCommands = []codegen.Command{
	genapp.NewCommand(),
	genmain.NewCommand(),
	genschema.NewCommand(),
}

// BootstrapCommand is the default command. It runs all common commands useful to bootstrap a goa
// application.
type BootstrapCommand struct{}

// Name returns the command name.
func (a *BootstrapCommand) Name() string { return "bootstrap" }

// Description returns the command description.
func (a *BootstrapCommand) Description() string {
	return `Bootstrap command, equivalent to running "app", "main" and "schema" commands sequentially.`
}

// RegisterFlags registers all the sub-commands flags.
func (a *BootstrapCommand) RegisterFlags(r codegen.FlagRegistry) {
	for _, c := range BootstrapCommands {
		if c != a {
			c.RegisterFlags(r)
		}
	}
}

// Run runs each known command and returns all the generated files and/or errors.
func (a *BootstrapCommand) Run() ([]string, error) {
	var all []string
	var err error
	for _, c := range BootstrapCommands {
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
