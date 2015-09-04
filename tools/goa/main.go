package main

import (
	"fmt"

	"github.com/raphael/goa/codegen"
	"gopkg.in/alecthomas/kingpin.v2"
)

// debug is set via the debug command line flag.
// When true the generators emit debug code (aimed at debugging code generation).
var debug bool

func main() {
	app = kingpin.New("goa", "the goa code generation tool")
	app.Flag("debug", "Enable debug mode.").BoolVar(&debug)
	design := app.Flag("design", "Design Go package path").Required().String()
	for name, gen := range codegen.Generators {
		cmd := app.Command(gen.GoaCmd(), gen.GoaCmdDescription())
		gen.RegisterFlags(cmd)
	}
	cmd := kingpin.Parse()
	gen, _ := codegen.Generators[cmd]
	files, err := gen.Spawn()
	kingpin.FatalIfError(err, err.String())
	for i, f := range files {
		fmt.Printf("%d. %s\n", i, f)
	}
}
