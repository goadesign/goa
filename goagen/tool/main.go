package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/app"
	"github.com/raphael/goa/goagen/client"
	"github.com/raphael/goa/goagen/docs"
	"github.com/raphael/goa/goagen/js"
	"github.com/raphael/goa/goagen/test"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contains the list of all supported sub-commands.
var Commands []goagen.Command

// init registers all subcommands.
func init() {
	Commands = []goagen.Command{
		app.NewCommand(),
		client.NewCommand(),
		test.NewCommand(),
		docs.NewCommand(),
		js.NewCommand(),
	}
}

func main() {
	files, err := command().Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL\n")
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(strings.Join(files, "\n"))
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
	if cmdName == "" {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}
	for _, c := range Commands {
		if cmdName == c.Name() {
			return c
		}
	}
	panic("goa: unknown command") // bug
}

const help = `The goagen tool generates various artefacts from a goa application design package (metadata).
Each sub-command supported by the tool matches a specific type of artefacts. For example
the "app" command causes goagen to generate the application GoGenerator
`
