package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raphael/goa/goagen/bootstrap"
	"github.com/raphael/goa/goagen/code/app"
	"github.com/raphael/goa/goagen/code/client"
	"github.com/raphael/goa/goagen/code/test"
	"github.com/raphael/goa/goagen/docs"
	"github.com/raphael/goa/goagen/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

// init registers all subcommands.
func init() {
	bootstrap.Register(app.New())
	bootstrap.Register(client.New())
	bootstrap.Register(test.New())
	bootstrap.Register(docs.New())
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
func command() bootstrap.Command {
	a := kingpin.New("goagen", "goa code generation tool")
	a.Version(version.Version)
	a.Help = help
	for n, c := range bootstrap.Commands {
		cmd := a.Command(n, c.Description())
		c.RegisterFlags(cmd)
	}
	cmdName := kingpin.MustParse(a.Parse(os.Args[1:]))
	if cmdName == "" {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}
	cmd, ok := bootstrap.Commands[cmdName]
	if !ok {
		panic("goa: unknown command")
	}
	return cmd
}

const help = `The goagen tool generates various artefacts from a goa application design package (metadata).
Each sub-command supported by the tool matches a specific type of artefacts. For example
the "app" command causes goagen to generate the application code.
`
