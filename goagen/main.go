package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_app"
	"github.com/raphael/goa/goagen/gen_client"
	"github.com/raphael/goa/goagen/gen_gen"
	"github.com/raphael/goa/goagen/gen_js"
	"github.com/raphael/goa/goagen/gen_main"
	"github.com/raphael/goa/goagen/gen_schema"
	"github.com/raphael/goa/goagen/gen_swagger"
	"github.com/raphael/goa/goagen/utils"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contains the list of all supported sub-commands.
var Commands = []codegen.Command{
	&BootstrapCommand{},
	genapp.NewCommand(),
	genmain.NewCommand(),
	genclient.NewCommand(),
	genswagger.NewCommand(),
	genjs.NewCommand(),
	genschema.NewCommand(),
	gengen.NewCommand(),
}

func main() {
	var (
		files            []string
		err              error
		terminatedByUser bool
	)

	// First check for the presence of `goimports`.
	_, err = exec.LookPath("goimports")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Command goimports not found. Install with:\ngo get golang.org/x/tools/cmd/goimports")
		os.Exit(1)
	}

	// Now proceed with code generation
	cleanup := func() {
		for _, f := range files {
			os.RemoveAll(f)
		}
	}

	go utils.Catch(nil, func() {
		terminatedByUser = true
	})

	files, err = command().Run()

	if terminatedByUser {
		cleanup()
		return
	}

	if err != nil {
		cleanup()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	rels := make([]string, len(files))
	cwd, err := os.Getwd()
	for i, f := range files {
		r, err := filepath.Rel(cwd, f)
		if err == nil {
			rels[i] = r
		} else {
			rels[i] = f
		}
	}
	fmt.Println(strings.Join(rels, "\n"))
}

// command parses the command line and returns the specified sub-command.
func command() codegen.Command {
	app := kingpin.New("goagen", "goa code generation tool")
	app.Version(codegen.Version)
	app.Help = help
	codegen.RegisterFlags(app)
	for _, c := range Commands {
		cmd := app.Command(c.Name(), c.Description())
		c.RegisterFlags(cmd)
	}
	if os.Args[len(os.Args)-1] == "--help" {
		args := append([]string{os.Args[0], "help"}, os.Args[1:len(os.Args)-1]...)
		os.Args = args
	}
	codegen.CommandName = kingpin.MustParse(app.Parse(os.Args[1:]))
	for _, c := range Commands {
		if codegen.CommandName == c.Name() {
			return c
		}
	}
	app.Usage(os.Args[1:])
	os.Exit(1)
	return nil
}

const help = `The goagen tool generates various artifacts from a goa service design package.

Each command supported by the tool produces a specific type of artifacts. For example
the "app" command generates the code that supports the service controllers.

The "bootstrap" command runs the "app", "main", "client" and "swagger" commands generating the
controllers supporting code and main skeleton code (if not already present) as well as a client
package and tool and the Swagger specification for the API.
`
