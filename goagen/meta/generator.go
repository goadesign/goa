package meta

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/raphael/goa/goagen"
)

// Generator generates the code of, compiles and runs generators.
// This extra step is necessary to compile in the end user design package so
// that generator code can iterate through it.
type Generator struct {
	*goagen.GoGenerator

	// Factory is the function used to create instances of the corresponding
	// generator including the package.
	// The meta generator generates a main function which calls the
	// generator factory method then calls Generate on the resulting object.
	Factory string

	// Imports list the imports that are specific for that generator that
	// should be added to the main Go file.
	Imports []string

	// Flags is the list of flags to be used when invoking the final
	// generator on the command line.
	Flags map[string]string
}

// NewGenerator returns a meta generator that can run an actual Generator
// given its factory method and command line flags.
func NewGenerator(factory string, imports []string, flags map[string]string) *Generator {
	return &Generator{
		Factory: factory,
		Imports: imports,
		Flags:   flags,
	}
}

// Generate compiles and runs the generator and returns the generated filenames.
func (m *Generator) Generate() ([]string, error) {
	// First make sure environment is setup correctly.
	if goagen.OutputDir == "" {
		return nil, fmt.Errorf("missing output directory specification")
	}
	if goagen.DesignPackagePath == "" {
		return nil, fmt.Errorf("missing design package path specification")
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil, fmt.Errorf("$GOPATH not defined")
	}
	designPath := filepath.Join(gopath, "src", goagen.DesignPackagePath)
	if _, err := os.Stat(designPath); err != nil {
		return nil, fmt.Errorf(`cannot find design package at path "%s"`, designPath)
	}
	_, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}

	// Create temporary directory used for generation under the output dir.
	gendir, err := ioutil.TempDir(goagen.OutputDir, "goagen")
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			err = fmt.Errorf(`invalid output directory path "%s"`, goagen.OutputDir)
		}
		return nil, err
	}
	defer func() {
		if !goagen.Debug {
			os.RemoveAll(gendir)
		}
	}()
	if goagen.Debug {
		fmt.Printf("goagen source dir: %s\n", gendir)
	}

	// Generate tool source code.
	filename := filepath.Join(gendir, "main.go")
	m.GoGenerator = goagen.NewGoGenerator(filename)
	imports := append(m.Imports, goagen.DesignPackagePath)
	m.WriteHeader("Code Generator", "main", imports)
	tmpl, err := template.New("generator").Parse(mainTmpl)
	if err != nil {
		panic(err) // bug
	}
	context := map[string]string{
		"Factory":       m.Factory,
		"DesignPackage": goagen.DesignPackagePath,
	}
	err = tmpl.Execute(m, context)
	if err != nil {
		panic(err) // bug
	}
	if goagen.Debug {
		src, _ := ioutil.ReadFile(filename)
		fmt.Printf("goagen source:\n%s\n", src)
	}

	// Compile and run generated tool.
	genbin, err := m.compile(gendir)
	if err != nil {
		return nil, err
	}
	return m.spawn(genbin)
}

// compile compiles a generator tool using the user design package and the
// target generator code.
// It returns the name of the compiled tool on success
// (located under $GOPATH/bin), an error otherwise.
func (m *Generator) compile(srcDir string) (string, error) {
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
	if goagen.Debug {
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
func (m *Generator) spawn(genbin string) ([]string, error) {
	var args []string
	for name, value := range m.Flags {
		args = append(args, fmt.Sprintf("%s=%s", name, value))
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

const mainTmpl = `
func main() {
	gen := {{.Factory}}("{{.DesignPackage}}")
	gen.Generate()
}`
