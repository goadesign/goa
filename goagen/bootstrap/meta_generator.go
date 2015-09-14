package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type (
	// MetaGenerator generates the code of, compiles and runs generators.
	// This extra step is necessary to compile in the end user design
	// package so that generator code can iterate through it.
	MetaGenerator struct {
		// DesignPackage contains the (Go package) path to the user Go design package.
		DesignPackage string

		// Factory is the function used to create instances of the
		// corresponding generator including the package.
		// The meta generator generates a main function which calls the
		// generator factory method then calls Generate on the resulting
		// object.
		Factory string

		// Imports list the imports that are specific for that
		// generator that should be added to the main Go file.
		Imports []string

		// Flags is the list of flags to be used when invoking the
		// generator on the command line. They are pointer to strings so
		// that the command line parser can update the values.
		Flags map[string]*string

		// Debug toggles debug mode.
		// If debug mode is enabled then the generated files are not
		// cleaned up upon failure.
		// Also logs additional debug information.
		// Set this flag to true prior to calling Generate.
		Debug bool

		// tempDir holds the name of the temporary directory located
		// under $GOPATH/src used to compile the goagen tool.
		tempDir string
	}
)

// Generate compiles and runs the generator and returns the generated filenames.
func (m *MetaGenerator) Generate() ([]string, error) {
	defer func() {
		if !m.Debug && m.tempDir != "" {
			os.RemoveAll(m.tempDir)
			m.tempDir = ""
		}
	}()
	genbin, err := m.compile()
	if err != nil {
		return nil, err
	}
	return m.spawn(genbin)
}

// compile compiles a generator tool using the user design package and the
// target generator code.
// It returns the name of the compiled tool on success
// (located under $GOPATH/bin), an error otherwise.
func (m *MetaGenerator) compile() (string, error) {
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
	m.tempDir = srcDir
	designPath := filepath.Join(gopath, "src", m.DesignPackage)
	if _, err := os.Stat(designPath); err != nil {
		return "", fmt.Errorf(`cannot find design package at path "%s"`, designPath)
	}
	if m.Debug {
		fmt.Printf("goagen source dir: %s\n", srcDir)
	}
	filename := filepath.Join(srcDir, "main.go")
	w, err := NewWriter(m.Factory, m.Imports, filename, m.DesignPackage)
	if err != nil {
		return "", err
	}
	if err := w.Write(); err != nil {
		return "", err
	}
	if m.Debug {
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
	if m.Debug {
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

// spawn runs the compiled generator using the arguments initialized by Kingpin
// when parsing the command line.
func (m *MetaGenerator) spawn(genbin string) ([]string, error) {
	var args []string
	for name, value := range m.Flags {
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
