package generator

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/raphael/goa/tools/goa/log"
	"gopkg.in/inconshreveable/log15.v2"
)

type (
	// Generator exposes the Generate method used to generate code or documentation from
	// metadata.
	Generator struct {
		Target Target // Generator target, code or documentation.
	}

	// Target is the generator target, code or documentation for now.
	Target string
)

var (
	// Debug toggles debug mode.
	// If debug mode is enabled then the generated files are not cleaned up upon failure.
	// Also logs additional debug information.
	// Set this flag to true prior to calling New.
	Debug bool
)

const (
	// Code is the code target, causes goa to generate Go source code.
	Code Target = "code"

	// Docs is the docs targs, causes goa to generate JSON docs.
	Docs Target = "docs"
)

// New returns a generator for the given target.
func New(target Target) *Generator {
	if Debug {
		log.Log.SetHandler(log15.StdoutHandler)
	}
	return &Generator{Target: target}
}

// Generate generates the target output (either code or documentation at this time).
// This first builds a generator tool using the user design package then invokes the tool.
func (g *Generator) Generate(designPack, targetPack, dest string) ([]string, error) {
	tool, err := g.Build(designPack)
	if err != nil {
		return nil, err
	}
	return g.Run(tool, dest, targetPack)
}

// Build compiles a generator tool using the user design package and the target generator code.
// It returns the name of the compiled tool on success (located under $GOPATH/bin), an error
// otherwise.
func (g *Generator) Build(pack string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("$GOPATH not defined")
	}
	dest, err := ioutil.TempDir(filepath.Join(gopath, "src"), "goa")
	if err != nil {
		return "", err
	}
	if !Debug {
		defer os.RemoveAll(dest)
	}
	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return "", err
	}
	src := filepath.Join(sourceDir(), string(g.Target))
	err = CopyDir(src, dest)
	if err != nil {
		return "", err
	}
	c := exec.Cmd{
		Path: "go",
		Args: []string{"build", "-o", "goagen"},
		Dir:  dest,
	}
	log.Debug("run", "cmd", "go build -o goagen", "dir", dest)
	b, err := c.CombinedOutput()
	if err != nil {
		if len(b) > 0 {
			return "", fmt.Errorf(string(b))
		}
		return "", fmt.Errorf("failed to compile goagen: %s", err)
	}
	return filepath.Join(dest, "goagen"), nil
}

// Run generates the target output (either code or documentation at this time).
// This first builds a generator tool using the user design package then invokes the tool.
func (g *Generator) Run(tool, dest, pack string) ([]string, error) {
	c := exec.Cmd{
		Path: tool,
		Args: []string{"--package", pack},
		Dir:  dest,
	}
	log.Debug("run", "cmd", log15.Lazy{func() string { return tool + " --package " + pack }}, "dir", dest)
	b, err := c.CombinedOutput()
	if err != nil {
		return nil, err
	}
	files := strings.Split(string(b), "\n")
	return files, nil
}

// CopyDir creates the directory dest and recursively copies the content of source into it.
func CopyDir(source string, dest string) error {
	log.Debug("copy", "source", source, "dest", dest)
	sinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, sinfo.Mode()); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(dest)
		}
	}()
	dir, err := os.Open(source)
	if err != nil {
		return err
	}
	objects, err := dir.Readdir(-1)
	if err != nil {
		return err
	}
	for _, obj := range objects {
		s := source + "/" + obj.Name()
		d := dest + "/" + obj.Name()
		if obj.IsDir() {
			if err := CopyDir(s, d); err != nil {
				return err
			}
		} else {
			if err := CopyFile(s, d); err != nil {
				return err
			}
		}
	}
	log.Debug("copy done")
	return nil
}

// CopyFile creates the file dest and copies the content of the file source into it.
func CopyFile(source string, dest string) error {
	s, err := os.Open(source)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	if err == nil {
		sinfo, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sinfo.Mode())
		}
	}
	return err
}

// sourceDir returns the path to the directory containing the generators sources.
func sourceDir() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get caller")
	}
	return filepath.Join(path.Dir(filename), "source")
}
