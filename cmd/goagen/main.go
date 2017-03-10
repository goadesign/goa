package main

import (
	"fmt"
	"go/build"
	"os"
	"strings"

	"goa.design/goa.v2/pkg"

	"flag"
)

func main() {
	var (
		cmds []string
		path string
	)
	{
		if len(os.Args) == 1 {
			usage()
		}

		switch os.Args[1] {
		case "version":
			fmt.Println("goagen version " + pkg.Version())
			os.Exit(0)
		case "client", "server", "openapi":
			if len(os.Args) == 2 {
				usage()
			}
			cmds = []string{os.Args[1]}
			i := 2
			for len(os.Args) > i+1 &&
				(os.Args[i] == "client" ||
					os.Args[i] == "server" ||
					os.Args[i] == "openapi") {
				cmds = append(cmds, os.Args[i])
				i++
			}
			path = os.Args[i]
		default:
			usage()
		}
	}

	var (
		output      string
		gens, debug bool
	)
	if len(os.Args) > 3 {
		var (
			fset     = flag.NewFlagSet("default", flag.ExitOnError)
			o        = fset.String("o", "", "output `directory`")
			out      = fset.String("out", ".", "output `directory`")
			s        = fset.Bool("s", false, "Generate scaffold (does not override existing files)")
			scaffold = fset.Bool("scaffold", false, "Generate scaffold (does not override existing files)")
		)
		fset.BoolVar(&debug, "debug", false, "Print debug information")

		fset.Usage = usage
		fset.Parse(os.Args[3:])

		output = *o
		if output == "" {
			output = *out
		}

		gens = *s
		if !gens {
			gens = *scaffold
		}
	}

	if _, err := build.Import(path, ".", build.IgnoreVendor); err != nil {
		fail(err)
	}

	var (
		tmp *GenPackage
	)
	{
		tmp = NewGenPackage(cmds, path, output)
		defer tmp.Remove()
	}

	if err := tmp.Write(gens, debug); err != nil {
		fail(err)
	}

	if err := tmp.Compile(); err != nil {
		fail(err)
	}

	files, err := tmp.Run()
	if err != nil {
		fail(err)
	}

	fmt.Println(strings.Join(files, "\n"))
}

func fail(err error) {
	fmt.Fprint(os.Stderr, err.Error())
	os.Exit(1)
}

func usage() {
	fmt.Fprint(os.Stderr, `goagen is the goa code generation tool.
Learn more about goa at https://goa.design.

The tool supports multiple subcommands that generate different outputs.
The only argument is the Go import path to the service design package.

The "--scaffold" flag tells goagen to also generate the scaffold for the service
and/or the client depending on which command is being executed. The scaffold is
code that is generated once as a way to get started quickly. It should be edited
after the initial generation.

Usage:

  goagen [server] [client] [openapi] PACKAGE [--out DIRECTORY] [--scaffold] [--debug]

  goagen version

Commands:
  server
        Generate service interfaces, endpoints and server transport code.

  client
        Generate endpoints and client transport code.

  openapi
        Generate OpenAPI specification (https://www.openapis.org/).

  version
        Print version information (exclusive with other flags and commands).

Args:
  PACKAGE
        Go import path to design package

Flags:
  -o, --out DIRECTORY
        output directory, defaults to the current working directory

  -s, --scaffold
        generate scaffold (does not override existing files)

  --debug
        Print debug information (mainly intended for goa developers)

Examples:

  goagen server goa.design/cellar/design

  goagen server client openapi goa.design/cellar/design -o gen -s

`)
	os.Exit(1)
}
