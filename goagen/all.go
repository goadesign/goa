package main

import (
	"os"
	"regexp"

	"github.com/raphael/goa/codegen"
)

// AllCommand is the default command. It runs all known commands.
type AllCommand struct{}

// Name returns the command name.
func (a *AllCommand) Name() string { return "default" }

// Description returns the command description.
func (a *AllCommand) Description() string { return "Default command, generates all artefacts." }

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

// metadataRegexp is the regular expression used to match undefined Metadata variable errors.
var metadataRegexp = regexp.MustCompile(`undefined: "[^"]+"\.Metadata`)

// simplify recognizes certain error messages and makes them easier to understand to users.
func simplify(msg string) string {
	switch {
	case metadataRegexp.MatchString(msg):
		return `"Metadata" global variable not found. Please make sure to assign the API definition to a variable named "Metadata":
	var Metadata = API(...
`
	default:
		return msg
	}
}
