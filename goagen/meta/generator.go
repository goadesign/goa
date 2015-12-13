package meta

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/raphael/goa/goagen/codegen"
)

// Generator generates the code of, compiles and runs generators.
// This extra step is necessary to compile in the end user design package so
// that generator code can iterate through it.
type Generator struct {
	*codegen.GoGenerator

	// Genfunc contains the name of the generator entry point function.
	// The function signature must be:
	//
	// func Genfunc(api *design.APIDefinition) ([]string, error)
	//
	// where "api" contains the DSL generated metadata and the returned
	// string array lists the generated filenames.
	Genfunc string

	// Imports list the imports that are specific for that generator that
	// should be added to the main Go file.
	Imports []*codegen.ImportSpec

	// Flags is the list of flags to be used when invoking the final
	// generator on the command line.
	Flags map[string]string
}

// NewGenerator returns a meta generator that can run an actual Generator
// given its factory method and command line flags.
func NewGenerator(genfunc string, imports []*codegen.ImportSpec, flags map[string]string) *Generator {
	return &Generator{
		Genfunc: genfunc,
		Imports: imports,
		Flags:   flags,
	}
}

func getDesignPath() (string, error) {
	if codegen.OutputDir == "" {
		return "", fmt.Errorf("missing output directory specification")
	}
	if codegen.DesignPackagePath == "" {
		return "", fmt.Errorf("missing design package path specification")
	}
	if err := os.MkdirAll(codegen.OutputDir, 0755); err != nil {
		return "", err
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("$GOPATH not defined")
	}
	candidates := strings.Split(gopath, ":")
	for i, c := range candidates {
		candidates[i] = filepath.Join(c, "src", codegen.DesignPackagePath)
	}
	var designPath string
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			designPath = path
			break
		}
	}
	if designPath == "" {
		if len(candidates) == 1 {
			return "", fmt.Errorf(`cannot find design package at path "%s"`, candidates[0])
		}
		return "", fmt.Errorf(`cannot find design package in any of the paths %s`, strings.Join(candidates, ", "))
	}
	_, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	return designPath, nil
}

func getDesignPackageName(designPath string) (string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, designPath, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	pkgNames := make([]string, len(pkgs))
	i := 0
	for n := range pkgs {
		pkgNames[i] = n
		i++
	}
	if len(pkgs) > 1 {
		return "", fmt.Errorf("more than one Go package found in %s (%s)",
			designPath, strings.Join(pkgNames, ","))
	}
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no Go package found in %s", designPath)
	}
	return pkgNames[0], nil
}

func (m *Generator) generateToolSourceCode(gendir, pkgName string) {
	filename := filepath.Join(gendir, "main.go")
	m.GoGenerator = codegen.NewGoGenerator(filename)
	imports := append(m.Imports,
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("strings"),
		codegen.NewImport(".", "github.com/raphael/goa/design"),
		codegen.NewImport(".", "github.com/raphael/goa/design/dsl"),
		codegen.NewImport("_", codegen.DesignPackagePath),
	)
	m.WriteHeader("Code Generator", "main", imports)
	tmpl, err := template.New("generator").Parse(mainTmpl)
	if err != nil {
		panic(err) // bug
	}
	context := map[string]string{
		"Genfunc":       m.Genfunc,
		"DesignPackage": codegen.DesignPackagePath,
		"PkgName":       pkgName,
	}
	err = tmpl.Execute(m, context)
	if err != nil {
		panic(err) // bug
	}
	if codegen.Debug {
		src, _ := ioutil.ReadFile(filename)
		fmt.Printf("goagen source:\n%s\n", src)
	}
}

// Generate compiles and runs the generator and returns the generated filenames.
func (m *Generator) Generate() ([]string, error) {
	// First make sure environment is setup correctly.
	designPath, err := getDesignPath()
	if err != nil {
		return nil, err
	}

	// Create temporary directory used for generation under the output dir.
	gendir, err := ioutil.TempDir(codegen.OutputDir, "goagen")
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			err = fmt.Errorf(`invalid output directory path "%s"`, codegen.OutputDir)
		}
		return nil, err
	}
	defer func() {
		if !codegen.Debug {
			os.RemoveAll(gendir)
		}
	}()
	if codegen.Debug {
		fmt.Printf("goagen source dir: %s\n", gendir)
	}

	// Figure out design package name from its path
	pkgName, err := getDesignPackageName(designPath)
	if err != nil {
		return nil, err
	}

	// Generate tool source code.
	m.generateToolSourceCode(gendir, pkgName)

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
	if codegen.Debug {
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
	args := []string{
		fmt.Sprintf("--out=%s", codegen.OutputDir),
		fmt.Sprintf("--design=%s", codegen.DesignPackagePath),
	}
	for name, value := range m.Flags {
		args = append(args, fmt.Sprintf("--%s=%s", name, value))
	}
	cmd := exec.Command(genbin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s\n%s", err, string(out))
	}
	res := strings.Split(string(out), "\n")
	for (len(res) > 0) && (res[len(res)-1] == "") {
		res = res[:len(res)-1]
	}
	return res, nil
}

const mainTmpl = `
func main() {
	failOnError(RunDSL())
	files, err := {{.Genfunc}}(Design)
	failOnError(err)
	fmt.Println(strings.Join(files, "\n"))
}

func failOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}`
