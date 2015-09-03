package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/raphael/goa/codegen/code/app"
)

// Version is the generator tool version.
const Version = "0.0.1"

var (
	pack   = kingpin.Flag("package", "target package").Required().String()
	outdir = kingpin.Flag("outdir", "output directory").Required().String()
)

// Main is called by the generator tool main.
func Main() {
	kingpin.Parse()
	dest := filepath.Join(*outdir, "autogen")
	if err := os.MkdirAll(dest, 0644); err != nil {
		kingpin.Fatalf("failed to create output directory, %s", err)
	}
	gen := &app.Generator{Outdir: dest, TargetPackage: *pack}
	if err := gen.WriteCode(); err != nil {
		kingpin.Fatalf(err.Error())
	}
	for _, f := range gen.Files {
		fmt.Println(filepath.Join(dest, f))
	}
}
