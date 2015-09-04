package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/raphael/goa/codegen/generator"
)

type (
	// Command is the interrace implemented be all code generation goa commands.
	// There is one command per generation target (i.e. application, documentation, test and
	// client).
	Command interface {
		// Name of the command.
		Name() string

		// Description returns the description used by the goa tool help.
		Description() string

		// RegisterFlags initialize the given command flags with all the flags relevant to
		// this generator.
		RegisterFlags(*kingpin.CmdClause)

		// Run generates the corresponding generator code, compiles and runs it.
		// It returns the list of generated files.
		// Spawn uses the variables initialized by kingpin.Parse and defined in RegisterFlags.
		Run() ([]string, error)
	}

	// BaseCommand provides the commands common code.
	BaseCommand struct {
		// DesignPackage contains the (Go package) path to the user Go design package.
		DesignPackage string

		// Debug toggles debug mode.
		// If debug mode is enabled then the generated files are not cleaned up upon failure.
		// Also logs additional debug information.
		// Set this flag to true prior to calling Generator.Spawn.
		Debug bool
	}

	// AppCommand is the command used to generate and run the application code generator.
	AppCommand struct {
		*BaseCommand
	}

	// DocsCommand is the command used to generate and run the documentation generator.
	DocsCommand struct {
		*BaseCommand
	}

	// TestCommand is the command used to generate and run the test code generator.
	TestCommand struct {
		*BaseCommand
	}

	// ClientCommand is the command used to generate and run the client code generator.
	ClientCommand struct {
		*BaseCommand
	}
)

// compile compiles a generator tool using the user design package and the target generator code.
// It returns the name of the compiled tool on success (located under $GOPATH/bin), an error
// otherwise.
func (b *BaseCommand) compile(factory string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("$GOPATH not defined")
	}
	srcDir, err := ioutil.TempDir(filepath.Join(gopath, "src"), "goa")
	if err != nil {
		return "", err
	}
	if !b.Debug {
		defer os.RemoveAll(srcDir)
	}
	filename := filepath.Join(srcDir, "main.go")
	w := generator.NewWriter(factory, filename, b.DesignPackage)
	c := exec.Cmd{
		Path: "go",
		Args: []string{"build", "-o", "goagen"},
		Dir:  srcDir,
	}
	out, err := c.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return "", fmt.Errorf(string(out))
		}
		return "", fmt.Errorf("failed to compile goagen: %s", err)
	}
	return filepath.Join(srcDir, "goagen"), nil
}

// run runs the compiled generator with the given arguments.
func (b *BaseCommand) run(genbin string, args []string) ([]string, error) {
	cmd := exec.Command(genbin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}

// Name of command.
func (a *AppCommand) Name() string { return "app" }

// Description of command.
func (a *AppCommand) Description() string { return "application code" }

// RegisterFlags registers the command line flags with the given command clause.
func (a *AppCommand) RegisterFlags(cmd *kingpin.CmdClause) {
	// TBD
}

// Run compiles and runs the generator and returns the generated filenames.
func (a *AppCommand) Run() ([]string, error) {
	genbin, err := a.compile("NewAppGenerator")
	if err != nil {
		return err
	}
	return a.run(genbin, args)
}
