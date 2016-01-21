package genclient

import (
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
)

var (
	// Signers contains the names of the request signers supported by the client.
	Signers []string

	// SignerPackages contains the Go package path to external packages containing custom
	// signers.
	SignerPackages []string

	// Version is the generated client version.
	Version string
)

// Command is the goa application code generator command line data structure.
// It implements meta.Command.
type Command struct {
	*codegen.BaseCommand
}

// NewCommand instantiates a new command.
func NewCommand() *Command {
	base := codegen.NewBaseCommand("client", "Generate API client tool and package")
	return &Command{BaseCommand: base}
}

// RegisterFlags registers the command line flags with the given registry.
func (c *Command) RegisterFlags(r codegen.FlagRegistry) {
	r.Flag("signer", `Adds support for the given request signer, e.g. "--signer goa.BasicSigner --signer goa.JWTSigner"`).
		StringsVar(&Signers)
	r.Flag("signerPkg", `Adds the given Go package path to the import directive in files using signers`).
		StringsVar(&SignerPackages)
	r.Flag("cli-version", "Generated client version").Default("1.0").StringVar(&Version)
}

// Run simply calls the meta generator.
func (c *Command) Run() ([]string, error) {
	gen := meta.NewGenerator(
		"genclient.Generate",
		[]*codegen.ImportSpec{codegen.SimpleImport("github.com/goadesign/goa/goagen/gen_client")},
		nil,
	)
	return gen.Generate()
}
