package meta

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goadesign/goa/goagen/codegen"
)

// Generator generates the code of, compiles and runs generators.
// This extra step is necessary to compile in the end user design package so
// that generator code can iterate through it.
type Generator struct {
	// Genfunc contains the name of the generator entry point function.
	// The function signature must be:
	//
	// func <Genfunc>(api *design.APIDefinition) ([]string, error)
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

// Generate compiles and runs the generator and returns the generated filenames.
func (m *Generator) Generate() ([]string, error) {
	if codegen.OutputDir == "" {
		return nil, fmt.Errorf("missing output directory specification")
	}

	if codegen.DesignPackagePath == "" {
		return nil, fmt.Errorf("missing design package path specification")
	}

	if os.Getenv("GOPATH") == "" {
		return nil, fmt.Errorf("GOPATH not set")
	}

	// Create output directory
	if err := os.MkdirAll(codegen.OutputDir, 0755); err != nil {
		return nil, err
	}

	// Create temporary workspace used for generation
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tmpDir, err := ioutil.TempDir(wd, "goagen")
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			err = fmt.Errorf(`invalid output directory path "%s"`, codegen.OutputDir)
		}
		return nil, err
	}
	defer func() {
		if !codegen.Debug {
			os.RemoveAll(tmpDir)
		}
	}()
	if codegen.Debug {
		fmt.Printf("** Code generator source dir: %s\n", tmpDir)
	}

	// Figure out design package name from its path
	path, err := codegen.PackageSourcePath(codegen.DesignPackagePath)
	if err != nil {
		return nil, err
	}
	pkgName, err := codegen.PackageName(path)
	if err != nil {
		return nil, err
	}

	// Generate tool source code.
	pkgPath := filepath.Join(tmpDir, pkgName)
	p, err := codegen.PackageFor(pkgPath)
	if err != nil {
		return nil, err
	}
	m.generateToolSourceCode(p)

	// Compile and run generated tool.
	if codegen.Debug {
		fmt.Printf("** Compiling with:\n%s", strings.Join(os.Environ(), "\n"))
	}
	genbin, err := p.Compile("goagen")
	if err != nil {
		return nil, err
	}
	return m.spawn(genbin)
}

func (m *Generator) generateToolSourceCode(pkg *codegen.Package) {
	file := pkg.CreateSourceFile("main.go")
	imports := append(m.Imports,
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("github.com/goadesign/goa/dslengine"),
		codegen.NewImport("_", filepath.ToSlash(codegen.DesignPackagePath)),
	)
	file.WriteHeader("Code Generator", "main", imports)
	tmpl, err := template.New("generator").Parse(mainTmpl)
	if err != nil {
		panic(err) // bug
	}
	pkgName, err := codegen.PackageName(pkg.Abs())
	if err != nil {
		panic(err)
	}
	context := map[string]string{
		"Genfunc":       m.Genfunc,
		"DesignPackage": codegen.DesignPackagePath,
		"PkgName":       pkgName,
	}
	err = tmpl.Execute(file, context)
	if err != nil {
		panic(err) // bug
	}
}

// spawn runs the compiled generator using the arguments initialized by Kingpin
// when parsing the command line.
func (m *Generator) spawn(genbin string) ([]string, error) {
	args := []string{
		fmt.Sprintf("--out=%s", codegen.OutputDir),
		fmt.Sprintf("--design=%s", codegen.DesignPackagePath),
	}
	if codegen.NoFormat {
		args = append(args, fmt.Sprintf("--noformat"))
	}
	for name, value := range m.Flags {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s=%s", name, value))
		}
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
	failOnError(dslengine.Run())
	roots := make([]interface{}, len(dslengine.Roots))
	for i, r := range dslengine.Roots {
		roots[i] = r
	}
	files, err := {{.Genfunc}}(roots)
	failOnError(err)
	fmt.Println(strings.Join(files, "\n"))
}

func failOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}`
