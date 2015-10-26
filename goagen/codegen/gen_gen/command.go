package gengen

import (
	"fmt"
	"path/filepath"

	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/codegen/meta"
)

var (
	// GenPkgPath contains the path to the third party generator Go package.
	GenPkgPath string

	// GenPkgName contains the name of the third party generator Go package.
	GenPkgName string
)

// Command is the goa generic generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("gen", "Invoke third party generator")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("pkg-path", "Go package path to generator package. The package must implement the Generate global function.").Required().StringVar(&GenPkgPath)
	r.Flag("pkg-name", "Go package name of generator package. Defaults to name of inner most directory in package path.").StringVar(&GenPkgName)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	if GenPkgName == "" {
		GenPkgName = filepath.Base(GenPkgPath)
	}
	gen := meta.NewGenerator(
		fmt.Sprintf("%s.Generate", GenPkgName),
		[]*codegen.ImportSpec{codegen.SimpleImport(GenPkgPath)},
		nil,
	)
	return gen.Generate()
}
