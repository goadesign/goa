package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	"github.com/goadesign/goa/goagen/gen_client"
	"github.com/goadesign/goa/goagen/gen_gen"
	"github.com/goadesign/goa/goagen/gen_js"
	"github.com/goadesign/goa/goagen/gen_main"
	"github.com/goadesign/goa/goagen/gen_schema"
	"github.com/goadesign/goa/goagen/gen_swagger"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
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

		// RootCmd is the base command used when goagen is called with no argument.
		RootCmd = &cobra.Command{
			Use:   "goagen",
			Short: "goa code generation tool",
			Long: `The goagen tool generates various artifacts from a goa service design package.

Each command supported by the tool produces a specific type of artifacts. For example
the "app" command generates the code that supports the service controllers.

The "bootstrap" command runs the "app", "main", "client" and "swagger" commands generating the
controllers supporting code and main skeleton code (if not already present) as well as a client
package and tool and the Swagger specification for the API.
`}
	)

	// Now proceed with code generation
	cleanup := func() {
		for _, f := range files {
			os.RemoveAll(f)
		}
	}

	go utils.Catch(nil, func() {
		terminatedByUser = true
	})

	for _, command := range Commands {
		run := command.Run
		sub := &cobra.Command{
			Use:   command.Name(),
			Short: command.Description(),
			Run: func(cmd *cobra.Command, args []string) {
				codegen.ExtraFlags = args
				files, err = run()
			},
		}
		command.RegisterFlags(sub)
		codegen.RegisterFlags(sub)
		RootCmd.AddCommand(sub)
	}
	codegen.RegisterFlags(RootCmd)
	RootCmd.Execute()

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
