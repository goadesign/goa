package codegen

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// OutputDir is the path to the directory the generated files should be
	// written to.
	OutputDir string

	// Force is true if pre-existing files should be overwritten during generation.
	Force bool

	// DesignPackagePath is the path to the user Go design package.
	DesignPackagePath string

	// Debug toggles debug mode.
	// If debug mode is enabled then the generated files are not
	// cleaned up upon failure.
	// Also logs additional debug information.
	// Set this flag to true prior to calling Generate.
	Debug bool
)

type (
	// FlagRegistry is the interface implemented by kingpin.Application
	// and kingpin.CmdClause to register flags.
	FlagRegistry interface {
		// Flag defines a new flag with the given long name and help.
		Flag(name, help string) *kingpin.FlagClause
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
		// Run uses the variables initialized by kingpin.Parse and
		// defined in RegisterFlags.
		Run() ([]string, error)
	}
)

// RegisterFlags registers the global flags.
func RegisterFlags(r FlagRegistry) {
	cwd, _ := os.Getwd()
	r.Flag("out", "output directory").
		Default(cwd).
		Short('o').
		StringVar(&OutputDir)

	r.Flag("design", "design package path").
		Required().
		Short('d').
		StringVar(&DesignPackagePath)

	r.Flag("force", "overwrite existing files").
		BoolVar(&Force)

	r.Flag("debug", "enable debug mode").
		BoolVar(&Debug)
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
