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
var Commands map[string]Command

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
		// line. They are pointer to strings so that the command line parser can update the
		// values.
		Flags map[string]*string

		// Files contains the list of generated filenames.
		Files []string

		// Debug toggles debug mode.
		// If debug mode is enabled then the generated files are not cleaned up upon failure.
		// Also logs additional debug information.
		// Set this flag to true prior to calling Command.Run.
		Debug bool

		// tempDir holds the name of the temporary directory located under $GOPATH/src used
		// to compile the goagen tool.
		tempDir string
	}
)

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

// RegisterFlags registers the common command line flags with the given command clause.
func (b *BaseCommand) RegisterFlags(cmd *kingpin.CmdClause) {
	var outdir string
	cmd.Flag("out", "destination directory").Required().StringVar(&outdir)
	b.Flags = map[string]*string{
		"OutDir": &outdir,
	}
}

// Run compiles and runs the generator and returns the generated filenames.
func (b *BaseCommand) Run() ([]string, error) {
	defer func() {
		if !b.Debug && b.tempDir != "" {
			os.RemoveAll(b.tempDir)
			b.tempDir = ""
		}
	}()
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
		if _, ok := err.(*os.PathError); ok {
			return "", fmt.Errorf(`invalid $GOPATH value "%s"`, gopath)
		}
		return "", err
	}
	b.tempDir = srcDir
	designPath := filepath.Join(gopath, "src", b.DesignPackage)
	if _, err := os.Stat(designPath); err != nil {
		return "", fmt.Errorf(`cannot find design package at path "%s"`, designPath)
	}
	if b.Debug {
		fmt.Printf("goagen source dir: %s\n", srcDir)
	}
	filename := filepath.Join(srcDir, "main.go")
	w, err := NewWriter(factory, filename, b.DesignPackage)
	if err != nil {
		return "", err
	}
	if err := w.Write(); err != nil {
		return "", err
	}
	if b.Debug {
		src, _ := ioutil.ReadFile(filename)
		fmt.Printf("goagen source:\n%s\n", src)
	}
	gobin, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	c := exec.Cmd{
		Path: gobin,
		Args: []string{gobin, "build", "-o", "goagen"},
		Dir:  srcDir,
	}
	out, err := c.CombinedOutput()
	if b.Debug {
		fmt.Printf("[%s]$ %s build -o goagen\n%s\n", srcDir, gobin, out)
	}
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
		if value != nil {
			args = append(args, fmt.Sprintf("%s=%s", name, *value))
		}
	}
	cmd := exec.Command(genbin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, out)
	}
	res := strings.Split(string(out), "\n")
	if len(res) > 0 && res[len(res)-1] == "" {
		res = res[:len(res)-1]
	}
	return res, nil
}
