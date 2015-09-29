package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/gen_app"
	"github.com/raphael/goa/goagen/gen_main"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contains the list of all supported sub-commands.
var Commands []goagen.Command

// init registers all subcommands.
func init() {
	Commands = []goagen.Command{
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
func command() goagen.Command {
	a := kingpin.New("goagen", "goa code generation tool")
	a.Version(goagen.Version)
	a.Help = help
	goagen.RegisterFlags(a)
	for _, c := range Commands {
		cmd := a.Command(c.Name(), c.Description())
		c.RegisterFlags(cmd)
	}
	cmdName := kingpin.MustParse(a.Parse(os.Args[1:]))
	if cmdName == "" || cmdName == "all" {
		return new(AllCommand)
	}
	for _, c := range Commands {
		if cmdName == c.Name() {
			return c
		}
	}
	a.Usage(os.Args[1:])
	os.Exit(1)
	return nil
}

const help = `The goagen tool generates various artefacts from a goa application design package (metadata).

Each sub-command supported by the tool produces a specific type of artefacts. For example
the "app" command causes goagen to generate the code that supports the application controllers.

The "default" command (also invoked when no command is provided on the command line) runs all the
commands, generating all the supported artefacts.

Artefact generation skips any file or directory that already exists unless the --force flag is
also provided.
`
