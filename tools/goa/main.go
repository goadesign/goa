package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raphael/goa/tools/goa/generator"
	"github.com/raphael/goa/tools/goa/log"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	pack   = kingpin.Arg("package", "path to application design package").Required().String()
	output = kingpin.Flag("output", "path to output directory").Short('o').String()
	target = kingpin.Flag("target", "generator target").Short('t').Default("code").Enum("code", "docs")
	debug  = kingpin.Flag("debug", "whether to print debug information").Short('d').Bool()
)

func main() {
	kingpin.Parse()
	dest := *output
	if dest == "" {
		var err error
		dest, err = os.Getwd()
		if err != nil {
			kingpin.Fatalf("failed to get current directory: %s", err)
		}
	}
	if *debug {
		generator.Debug = true
	}
	gen := generator.New(generator.Target(*target))
	files, err := gen.Generate(*pack, "autogen", dest)
	if err != nil {
		log.Crit(err.Error())
	}
	// Say something
	fmt.Println(strings.Join(files, "\n"))
}
