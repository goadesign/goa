package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raphael/goa/codegen"
)

// AllCommand is the default command. It runs all known commands.
type AllCommand struct {
	// Errors contains all the generation errors if any.
	Errors []error
}

// Name returns the command name.
func (a *AllCommand) Name() string { return "default" }

// Description returns the command description.
func (a *AllCommand) Description() string { return "Default command, generates all artefacts." }

// RegisterFlags is a dummy method for the default command.
func (a *AllCommand) RegisterFlags(codegen.FlagRegistry) {}

// Run runs each known command and returns all the generated files and/or errors.
func (a *AllCommand) Run() ([]string, error) {
	var all []string
	for _, c := range Commands {
		files, err := c.Run()
		a.Errors = append(a.Errors, fmt.Errorf("ERR - %s:\n%s\n", c.Name, err.Error()))
		all = append(all, files...)
	}
	if a.Errors != nil && !codegen.Force {
		// Cleanup in case of failure
		for _, f := range all {
			os.Remove(f)
		}
	}
	if a.Errors == nil {
		return all, nil
	}
	return nil, a

}

// Error implements the error interface.
func (a *AllCommand) Error() string {
	msgs := make([]string, len(a.Errors))
	for i, e := range a.Errors {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "\n\n")
}
