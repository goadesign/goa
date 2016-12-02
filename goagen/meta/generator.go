package meta

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/version"
)

// Generator generates the code of, compiles and runs generators.
// This extra step is necessary to compile in the end user design package so
// that generator code can iterate through it.
type Generator struct {
	// Genfunc contains the name of the generator entry point function.
	// The function signature must be:
	//
	// func <Genfunc>([]dslengine.Root) ([]string, error)
	Genfunc string

	// Imports list the imports that are specific for that generator that
	// should be added to the main Go file.
	Imports []*codegen.ImportSpec

	// Flags is the list of flags to be used when invoking the final
	// generator on the command line.
	Flags map[string]string

	// CustomFlags is the list of arguments that appear after the -- separator.
	// These arguments are appended verbatim to the final generator command line.
	CustomFlags []string

	// OutDir is the final output directory.
	OutDir string

	// DesignPkgPath is the Go import path to the design package.
	DesignPkgPath string

	debug bool
}

// NewGenerator returns a meta generator that can run an actual Generator
// given its factory method and command line flags.
func NewGenerator(genfunc string, imports []*codegen.ImportSpec, flags map[string]string, customflags []string) (*Generator, error) {
	var (
		outDir, designPkgPath string
		debug                 bool
	)

	if o, ok := flags["out"]; ok {
		outDir = o
	}
	if d, ok := flags["design"]; ok {
		designPkgPath = d
	}
	if d, ok := flags["debug"]; ok {
		var err error
		debug, err = strconv.ParseBool(d)
		if err != nil {
			return nil, fmt.Errorf("failed to parse debug flag: %s", err)
		}
	}

	return &Generator{
		Genfunc:       genfunc,
		Imports:       imports,
		Flags:         flags,
		CustomFlags:   customflags,
		OutDir:        outDir,
		DesignPkgPath: designPkgPath,
		debug:         debug,
	}, nil
}

// Generate compiles and runs the generator and returns the generated filenames.
func (m *Generator) Generate() ([]string, error) {
	// Sanity checks
	if os.Getenv("GOPATH") == "" {
		return nil, fmt.Errorf("GOPATH not set")
	}
	if m.OutDir == "" {
		return nil, fmt.Errorf("missing output directory flag")
	}
	if m.DesignPkgPath == "" {
		return nil, fmt.Errorf("missing design package flag")
	}

	// Create output directory
	if err := os.MkdirAll(m.OutDir, 0755); err != nil {
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
			err = fmt.Errorf(`invalid output directory path "%s"`, m.OutDir)
		}
		return nil, err
	}
	defer func() {
		if !m.debug {
			os.RemoveAll(tmpDir)
		}
	}()
	if m.debug {
		fmt.Printf("** Code generator source dir: %s\n", tmpDir)
	}

	pkgSourcePath, err := codegen.PackageSourcePath(m.DesignPkgPath)
	if err != nil {
		return nil, fmt.Errorf("invalid design package import path: %s", err)
	}
	pkgName, err := codegen.PackageName(pkgSourcePath)
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
	if m.debug {
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
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("github.com/goadesign/goa/dslengine"),
		codegen.NewImport("_", filepath.ToSlash(m.DesignPkgPath)),
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
		"DesignPackage": m.DesignPkgPath,
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
	var args []string
	for k, v := range m.Flags {
		if k == "debug" {
			continue
		}
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}
	sort.Strings(args)
	args = append(args, "--version="+version.String())
	args = append(args, m.CustomFlags...)
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
	// Check if there were errors while running the first DSL pass
	dslengine.FailOnError(dslengine.Errors)

	// Now run the secondary DSLs
	dslengine.FailOnError(dslengine.Run())

	files, err := {{.Genfunc}}()
	dslengine.FailOnError(err)

	// We're done
	fmt.Println(strings.Join(files, "\n"))
}`
