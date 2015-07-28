package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
)

var (
	pack = kingpin.Flag("package", "target package").String().Required()
	dest = kingpin.Flag("outdir", "output directory").String().Required()
)

func main() {
	dest := filepath.Join(*outdir, "autogen")
	if err := os.MkdirAll(dest, 0644); err != nil {
		kingpin.Fatalf("failed to create output directory, %s", err)
	}
	gen := &Generator{Outdir: dest, TargetPackage: *pack}
	if err := gen.WriteCode(); err != nil {
		kingpin.Fatalf(err.Error())
	}
}
