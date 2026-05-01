package main

import (
	"fmt"
	"go/build"
	"os"
	"strings"
	"time"

	"flag"

	goa "goa.design/goa/v3/pkg"
)

func main() {
	var (
		cmd    string
		path   string
		offset int
	)
	if len(os.Args) == 1 {
		usage()
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("Goa version " + goa.Version())
		os.Exit(0)
	case "gen", "example":
		if len(os.Args) == 2 {
			usage()
			return
		}
		cmd = os.Args[1]
		path = os.Args[2]
		offset = 2
	default:
		usage()
		return
	}

	var (
		output        = "."
		serviceOutput = ""
		debug         bool
	)
	if len(os.Args) > offset+1 {
		var (
			fset       = flag.NewFlagSet("default", flag.ExitOnError)
			o          = fset.String("o", "", "output `directory`")
			out        = fset.String("output", output, "output `directory`")
			serviceOut = fset.String("service-output", "", "service output `directory`")
		)
		fset.BoolVar(&debug, "debug", false, "Print debug information")

		fset.Usage = usage
		if err := fset.Parse(os.Args[offset+1:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		output = *o
		if output == "" {
			output = *out
		}
		serviceOutput = *serviceOut
		if serviceOutput == "" {
			serviceOutput = output
		}
	}
	if serviceOutput == "" {
		serviceOutput = output
	}

	if err := gen(cmd, path, output, serviceOutput, debug); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// help with tests
var (
	usage = help
	gen   = generate
)

func generate(cmd, path, output, serviceOutput string, debug bool) error {
	var (
		files []string
		err   error
		tmp   *Generator
		startTotal, startImport, startNewGen,
		startWrite, startCompile, startRun time.Time
	)

	startTotal = time.Now()
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] Starting goa generation\n")
	}

	startImport = time.Now()
	if _, err = build.Import(path, ".", 0); err != nil {
		goto fail
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] build.Import took %v\n", time.Since(startImport))
	}

	startNewGen = time.Now()
	tmp = NewGenerator(cmd, path, output, serviceOutput, debug)
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] NewGenerator took %v\n", time.Since(startNewGen))
	}

	startWrite = time.Now()
	if err = tmp.Write(debug); err != nil {
		goto fail
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] Write (generate main.go) took %v\n", time.Since(startWrite))
	}

	startCompile = time.Now()
	if err = tmp.Compile(debug); err != nil {
		goto fail
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] Compile (go get + go build) took %v\n", time.Since(startCompile))
	}

	startRun = time.Now()
	if files, err = tmp.Run(debug); err != nil {
		goto fail
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING] Run (execute binary) took %v\n", time.Since(startRun))
		fmt.Fprintf(os.Stderr, "[TIMING] Total generation time: %v\n", time.Since(startTotal))
	}
	fmt.Println(strings.Join(files, "\n"))
	if !debug {
		tmp.Remove()
	}
	return nil
fail:
	if !debug && tmp != nil {
		tmp.Remove()
	}
	return err
}

func help() {
	fmt.Fprint(os.Stderr, `goa is the code generation tool for the Goa framework.
Learn more at https://goa.design.

Usage:
  goa gen PACKAGE [--output DIRECTORY] [--debug]
  goa example PACKAGE [--output DIRECTORY] [--service-output DIRECTORY] [--debug]
  goa version

Commands:
  gen
        Generate service interfaces, endpoints, transport code and OpenAPI spec.
  example
        Generate example server and client tool.
  version
        Print version information.

Args:
  PACKAGE
        Go import path to design package

Flags:
  -o, -output DIRECTORY
        output directory, defaults to the current working directory

  -service-output DIRECTORY
        service example output directory, defaults to the output directory

  -debug
        Print debug information (mainly intended for Goa developers)

Example:

  goa gen goa.design/examples/cellar/design -o gendir

`)
}
