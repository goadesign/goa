package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_app"
	"github.com/raphael/goa/goagen/gen_gen"
	"github.com/raphael/goa/goagen/gen_main"
	"github.com/raphael/goa/goagen/gen_schema"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contains the list of all supported sub-commands.
var Commands = []codegen.Command{
	&DefaultCommand{},
	genapp.NewCommand(),
	genschema.NewCommand(),
	genmain.NewCommand(),
	gengen.NewCommand(),
}

func main() {
	files, err := command().Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
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
	codegen.CommandName = kingpin.MustParse(a.Parse(os.Args[1:]))
	for _, c := range Commands {
		if codegen.CommandName == c.Name() {
			return c
		}
	}
	a.Usage(os.Args[1:])
	os.Exit(1)
	return nil
}

const help = `The goagen tool generates various artifacts from a goa application design package (metadata).

Each sub-command supported by the tool produces a specific type of artifacts. For example
the "app" command causes goagen to generate the code that supports the application controllers.

The "default" command (also invoked when no command is provided on the command line) runs the "app",
"main" and "schema" commands generating the application code and main skeleton code if not
already present as well as its JSON hyper-schema.
`
