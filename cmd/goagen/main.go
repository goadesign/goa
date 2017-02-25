package main

import (
	"fmt"
	"os"

	"flag"

	"goa.design/goa.v2/pkg"

	// These are packages required by the generated code but not by goagen.
	// We list them here so that `go get` picks them up.
	_ "gopkg.in/yaml.v2"
)

func main() {
	var (
		out     = flag.String("out", ".", "Output `directory`")
		design  = flag.String("design", "", "Design Go import `package path`")
		debug   = flag.Bool("debug", false, "Print debug information")
		version = flag.Bool("version", false, "Print version information")
	)

	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println("goagen " + pkg.Version() + "\nThe goa generation tool.")
		os.Exit(0)
	}

	if len(os.Args) < flag.Nflag+2 {
		fmt.Fprint(os.Stderr, "Missing command\n")
		os.Exit(1)
	}

	cmd := command(os.Args[flag.Nflag+1])
	command.Run(*out, *design, *debug)
}

func usage() {
	fmt.Fprint(os.Stderr, `The goagen tool generates artifacts from a goa design package.

Each command supported by the tool produces a specific type of artifacts. For
example the "server" command generates the code that supports the service
server.

The "scaffold" command is a special command that generates code only once.
The generated code is intended to be consumed as if it had not been generated.
That is the code should be edited, versioned, tested and in general maintained
like any non generated code. It is generated as a convenience to help get
started.

The "bootstrap" command runs the "scaffold", "server", "client" and "openapi"
commands generating all the artifacts needed to start implementing a new service.

Usage of goagen:
`)
	flag.PrintDefaults()
}
