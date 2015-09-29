package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/codegen/gen_app"
	"github.com/raphael/goa/codegen/gen_main"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contains the list of all supported sub-commands.
var Commands []codegen.Command

// init registers all subcommands.
func init() {
	Commands = []codegen.Command{
		&AllCommand{},
		genapp.NewCommand(),
		genmain.NewCommand(),
	}
}

func main() {
	files, err := command().Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL\n")
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	rels := make([]string, len(files))
	cwd, err := os.Getwd()
	if err != nil {
		rels = files
	} else {
		for i, f := range files {
			r, err := filepath.Rel(cwd, f)
			if err == nil {
				rels[i] = r
			} else {
				rels[i] = f
			}
		}
	}
	fmt.Println(strings.Join(rels, "\n"))
}

// command parses the command line and returns the specified sub-command.
func command() codegen.Command {
	a := kingpin.New("codegen", "goa code generation tool")
	a.Version(codegen.Version)
	a.Help = help
	codegen.RegisterFlags(a)
	for _, c := range Commands {
		cmd := a.Command(c.Name(), c.Description())
		if c.Name() == "default" {
			cmd.Default()
		}
		c.RegisterFlags(cmd)
	}
	// a.Command("default", "default command, generates all known artefacts").Default()
	cmdName := kingpin.MustParse(a.Parse(os.Args[1:]))
	for _, c := range Commands {
		if cmdName == c.Name() {
			return c
		}
	}
	a.Usage(os.Args[1:])
	os.Exit(1)
	return nil
}

const help = `The codegen tool generates various artefacts from a goa application design package (metadata).

Each sub-command supported by the tool produces a specific type of artefacts. For example
the "app" command causes codegen to generate the code that supports the application controllers.

The "default" command (also invoked when no command is provided on the command line) runs all the
commands, generating all the supported artefacts.

Artefact generation skips any file or directory that already exists unless the --force flag is
also provided.
`
