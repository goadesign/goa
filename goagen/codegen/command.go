package codegen

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

var (
	// OutputDir is the path to the directory the generated files should be
	// written to.
	OutputDir string

	// DesignPackagePath is the path to the user Go design package.
	DesignPackagePath string

	// Debug toggles debug mode.
	// If debug mode is enabled then the generated files are not
	// cleaned up upon failure.
	// Also logs additional debug information.
	// Set this flag to true prior to calling Generate.
	Debug bool

	// NoFormat causes "goimports" to be skipped when true.
	NoFormat bool

	// CommandName is the name of the command being run.
	CommandName string

	// ExtraFlags is the list of extra, arbitrary flags to was read from the
	// command line.  These flags will be passed as command line flags to the
	// final generator.
	ExtraFlags []string
)

type (
	// FlagRegistry is the interface implemented by cobra.Command to register flags.
	FlagRegistry interface {
		// Flags returns the command flag set
		Flags() *pflag.FlagSet
	}

	// Command is the interface implemented by all generation goa commands.
	// There is one command per generation target (i.e. app, docs, etc.)
	Command interface {
		// Name of the command
		Name() string

		// Description returns the description used by the goa tool help.
		Description() string

		// RegisterFlags initializes the given registry flags with all
		// the flags relevant to this command.
		RegisterFlags(r FlagRegistry)

		// Run generates the generator code then compiles and runs it.
		// It returns the list of generated files.
		// Run uses the variables initialized by the command line defined in RegisterFlags.
		Run() ([]string, error)
	}
)

// RegisterFlags registers the global flags.
func RegisterFlags(r FlagRegistry) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	r.Flags().StringVarP(&OutputDir, "out", "o", cwd, "output directory")
	r.Flags().StringVarP(&DesignPackagePath, "design", "d", "", "design package import path")
	r.Flags().BoolVar(&Debug, "debug", false, "enable debug mode, does not cleanup temporary files.")
	r.Flags().BoolVar(&NoFormat, "noformat", false, "disable goimports, useful to goa developers for debugging.")
	r.Flags().MarkHidden("noformat")
}

// BaseCommand provides the basic logic for all commands. It implements
// the Command interface.
// Commands may then specialize to provide the specific Run behavior.
type BaseCommand struct {
	CmdName        string
	CmdDescription string
}

// NewBaseCommand instantiates a base command.
func NewBaseCommand(name, desc string) *BaseCommand {
	return &BaseCommand{CmdName: name, CmdDescription: desc}
}

// Name returns the command name.
func (b *BaseCommand) Name() string {
	return b.CmdName
}

// Description returns the command description.
func (b *BaseCommand) Description() string {
	return b.CmdDescription
}

// RegisterFlags is a dummy implementation, override in sub-command.
func (b *BaseCommand) RegisterFlags(r FlagRegistry) {}

// Run is a dummy implementation, override in sub-command.
func (b *BaseCommand) Run() ([]string, error) {
	return nil, nil
}
