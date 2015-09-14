package bootstrap

import "gopkg.in/alecthomas/kingpin.v2"

// Commands contain the registered generation commands.
// Each generator package registers itself in its init function.
var Commands map[string]Command

// Command is the interrace implemented be all code generation goa commands.
// There is one command per generation target (i.e. application, documentation, test and
// client).
type Command interface {
	// Name of the command.
	Name() string

	// Description returns the description used by the goa tool help.
	Description() string

	// RegisterFlags initialize the given command flags with all the flags relevant to
	// this command.
	RegisterFlags(*kingpin.CmdClause)

	// Run generates the corresponding generator code, compiles and runs it.
	// It returns the list of generated files.
	// Run uses the variables initialized by kingpin.Parse and defined in RegisterFlags.
	Run() ([]string, error)
}

func init() {
	Commands = make(map[string]Command)
}

// Register adds a command to Commands.
func Register(cmd Command) {
	if _, ok := Commands[cmd.Name()]; ok {
		panic("goa: duplicate command ")
	}
	Commands[cmd.Name()] = cmd
}
