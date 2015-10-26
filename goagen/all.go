package main

import (
	"os"

	"github.com/raphael/goa/goagen/codegen"
)

// AllCommand is the default command. It runs all known commands.
type AllCommand struct{}

// Name returns the command name.
func (a *AllCommand) Name() string { return "default" }

// Description returns the command description.
func (a *AllCommand) Description() string { return "Default command, generates all artifacts." }

// RegisterFlags registers all the sub-commands flags.
func (a *AllCommand) RegisterFlags(r codegen.FlagRegistry) {
	for _, c := range Commands {
		if c != a {
			c.RegisterFlags(r)
		}
	}
}

// Run runs each known command and returns all the generated files and/or errors.
func (a *AllCommand) Run() ([]string, error) {
	var all []string
	var err error
	for _, c := range Commands {
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
