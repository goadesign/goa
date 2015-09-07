package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Commands contain the registered generation commands.
// Each generator package registers itself in its init function.
var Commands []Command

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
		// this command.
		RegisterFlags(*kingpin.CmdClause)

		// Run generates the corresponding generator code, compiles and runs it.
		// It returns the list of generated files.
		// Run uses the variables initialized by kingpin.Parse and defined in RegisterFlags.
		Run() ([]string, error)
	}

	// BaseCommand provides the commands common code.
	BaseCommand struct {
		// DesignPackage contains the (Go package) path to the user Go design package.
		DesignPackage string

		// Factory is the function used to create instances of the corresponding generator.
		// For example NewAppGenerator.
		Factory string

		// Flags is the list of flags to be used when invoking the generator on the command
		// line.
		Flags map[string]string

		// Files contains the list of generated filenames.
		Files []string

		// Debug toggles debug mode.
		// If debug mode is enabled then the generated files are not cleaned up upon failure.
		// Also logs additional debug information.
		// Set this flag to true prior to calling Command.Run.
		Debug bool
	}
)

// RegisterFlags registers the common command line flags with the given command clause.
func (b *BaseCommand) RegisterFlags(cmd *kingpin.CmdClause) {
	outdir := ""
	cmd.Flag("out", "destination directory").Required().StringVar(&outdir)
	b.Flags = map[string]string{
		"OutDir": outdir,
	}
}

// Run compiles and runs the generator and returns the generated filenames.
func (b *BaseCommand) Run() ([]string, error) {
	genbin, err := b.compile(b.Factory)
	if err != nil {
		return nil, err
	}
	return b.spawn(genbin)
}

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
	w, err := NewWriter(factory, filename, b.DesignPackage)
	if err != nil {
		return "", err
	}
	if err := w.Write(); err != nil {
		return "", err
	}
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

// spawn runs the compiled generator using the arguments initialized by Kingpin when parsing the
// command line.
func (b *BaseCommand) spawn(genbin string) ([]string, error) {
	var args []string
	for name, value := range b.Flags {
		args = append(args, fmt.Sprintf("%s=%s", name, value))
	}
	cmd := exec.Command(genbin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}
