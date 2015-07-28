package main

import (
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	pack   = kingpin.Arg("package", "Path to application design package").String()
	output = kingpin.Flag("output").Short("o").String()
	target = kingpin.Flag("target").Short("t").Enum("code", "docs").Default("code")
	debug  = kingpin.Flag("debug").Short("d").Bool()
)

func main() {
	dest := *output
	if dest == "" {
		dest = os.Getcwd()
	}
	if *debug {
		generator.Debug = true
	}
	gen := generators.New(generators.Moniker(target))
	files, err := gen.Generate(*pack, dest)
	if err != nil {
		log.Critical(err.String())
	}
	// Say something
	mt.Println(strings.Join(files, "\n"))
}
