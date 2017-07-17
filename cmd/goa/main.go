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
		cmds   []string
		path   string
		offset int
	)
	{
		if len(os.Args) == 1 {
			usage()
		}

		switch os.Args[1] {
		case "version":
			fmt.Println("goa version " + pkg.Version())
			os.Exit(0)
		case "gen":
			if len(os.Args) == 2 {
				usage()
			}
			cmds = []string{"client", "openapi", "server"}
			path = os.Args[2]
			offset = 2
		default:
			usage()
		}
	}

	var (
		output = "."
		debug  bool
	)
	if len(os.Args) > offset+1 {
		var (
			fset = flag.NewFlagSet("default", flag.ExitOnError)
			o    = fset.String("o", "", "output `directory`")
			out  = fset.String("output", output, "output `directory`")
		)
		fset.BoolVar(&debug, "debug", false, "Print debug information")

		fset.Usage = usage
		fset.Parse(os.Args[offset+1:])

		output = *o
		if output == "" {
			output = *out
		}
	}

	gen(cmds, path, output, debug)
}

// help with tests
var (
	usage = help
	gen   = generate
)

func generate(cmds []string, path, output string, debug bool) {
	var (
		files []string
		err   error
		tmp   *Generator
	)

	if _, err = build.Import(path, ".", build.IgnoreVendor); err != nil {
		goto fail
	}

	tmp = NewGenerator(cmds, path, output)
	if !debug {
		defer tmp.Remove()
	}

	if err = tmp.Write(debug); err != nil {
		goto fail
	}

	if err = tmp.Compile(); err != nil {
		goto fail
	}

	if files, err = tmp.Run(); err != nil {
		goto fail
	}

	fmt.Println(strings.Join(files, "\n"))
	return
fail:
	fmt.Fprint(os.Stderr, err.Error())
	if !debug && tmp != nil {
		tmp.Remove()
	}
	os.Exit(1)
}

func help() {
	fmt.Fprint(os.Stderr, `goa is the code generation tool for the goa framework.
Learn more at https://goa.design.

Usage:
  goa gen PACKAGE [--out DIRECTORY] [--debug]
  goa version

Commands:
  gen
        Generate service interfaces, endpoints, transport code and OpenAPI spec.
  version
        Print version information (exclusive with other flags and commands).

Args:
  PACKAGE
        Go import path to design package

Flags:
  -o, -output DIRECTORY
        output directory, defaults to the current working directory

  -debug
        Print debug information (mainly intended for goa developers)

Example:

  goa gen goa.design/cellar/design -o gendir

`)
	os.Exit(1)
}
