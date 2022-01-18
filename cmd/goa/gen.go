package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"golang.org/x/tools/go/packages"
)

// Generator is the code generation management data structure.
type Generator struct {
	// Command is the name of the command to run.
	Command string

	// DesignPath is the Go import path to the design package.
	DesignPath string

	// Output is the absolute path to the output directory.
	Output string

	// DesignVersion is the major component of the Goa version used by the design DSL.
	// DesignVersion is either 2 or 3.
	DesignVersion int

	// bin is the filename of the generated generator.
	bin string

	// tmpDir is the temporary directory used to compile the generator.
	tmpDir string

	// hasVendorDirectory is a flag to indicate whether the project uses vendoring
	hasVendorDirectory bool
}

// NewGenerator creates a Generator.
func NewGenerator(cmd string, path, output string) *Generator {
	bin := "goa"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	var version int
	var hasVendorDirectory bool
	{
		version = 2
		matched := false
		pkgs, _ := packages.Load(&packages.Config{Mode: packages.NeedFiles | packages.NeedModule}, path)
		fset := token.NewFileSet()
		p := regexp.MustCompile(`goa.design/goa/v(\d+)/dsl`)
		for _, pkg := range pkgs {
			if _, err := os.Stat(filepath.Join(pkg.Module.Dir, "vendor")); !os.IsNotExist(err) {
				hasVendorDirectory = true
			}
			for _, gof := range pkg.GoFiles {
				if bs, err := ioutil.ReadFile(gof); err == nil {
					if f, err := parser.ParseFile(fset, "", string(bs), parser.ImportsOnly); err == nil {
						for _, s := range f.Imports {
							matches := p.FindStringSubmatch(s.Path.Value)
							if len(matches) == 2 {
								matched = true
								version, _ = strconv.Atoi(matches[1]) // We know it's an integer
							}
						}
					}
				}
				if matched {
					break
				}
			}
			if matched {
				break
			}
		}
	}

	return &Generator{
		Command:            cmd,
		DesignPath:         path,
		Output:             output,
		DesignVersion:      version,
		hasVendorDirectory: hasVendorDirectory,
		bin:                bin,
	}
}

// Write writes the main file.
func (g *Generator) Write(debug bool) error {
	var tmpDir string
	{
		wd := "."
		if cwd, err := os.Getwd(); err != nil {
			wd = cwd
		}
		tmp, err := ioutil.TempDir(wd, "goa")
		if err != nil {
			return err
		}
		tmpDir = tmp
	}
	g.tmpDir = tmpDir

	var sections []*codegen.SectionTemplate
	{
		data := map[string]interface{}{
			"Command":       g.Command,
			"CleanupDirs":   cleanupDirs(g.Command, g.Output),
			"DesignVersion": g.DesignVersion,
		}
		ver := ""
		if g.DesignVersion > 2 {
			ver = "v" + strconv.Itoa(g.DesignVersion) + "/"
		}
		imports := []*codegen.ImportSpec{
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("path/filepath"),
			codegen.SimpleImport("sort"),
			codegen.SimpleImport("strconv"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("goa.design/goa/" + ver + "codegen"),
			codegen.SimpleImport("goa.design/goa/" + ver + "codegen/generator"),
			codegen.SimpleImport("goa.design/goa/" + ver + "eval"),
			codegen.NewImport("goa", "goa.design/goa/"+ver+"pkg"),
			codegen.NewImport("_", g.DesignPath),
		}
		sections = []*codegen.SectionTemplate{
			codegen.Header("Code Generator", "main", imports),
			{
				Name:   "main",
				Source: mainT,
				Data:   data,
			},
		}
	}

	f := &codegen.File{Path: "main.go", SectionTemplates: sections}
	_, err := f.Render(tmpDir)
	return err
}

// Compile compiles the generator.
func (g *Generator) Compile() error {
	// We first need to go get the generated package to make sure that all
	// dependencies are added to go.sum prior to compiling.
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, fmt.Sprintf(".%c%s", filepath.Separator, g.tmpDir))
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("expected to find one package in %s", g.tmpDir)
	}
	if !g.hasVendorDirectory {
		if err := g.runGoCmd("get", pkgs[0].PkgPath); err != nil {
			return err
		}
	}

	err = g.runGoCmd("build", "-o", g.bin)

	// If we're in vendor context we check the error string to see if it's an issue of unsatisfied dependencies
	if err != nil && g.hasVendorDirectory {
		if strings.Contains(err.Error(), "cannot find package") && strings.Contains(err.Error(), "/goa.design/goa/v3/codegen/generator") {
			return errors.New("generated code expected `goa.design/goa/v3/codegen/generator` to be present in the vendor directory, see documentation for more details")
		}
	}

	return err
}

// Run runs the compiled binary and return the output lines.
func (g *Generator) Run() ([]string, error) {
	var cmdl string
	{
		args := make([]string, len(os.Args)-1)
		gopaths := filepath.SplitList(os.Getenv("GOPATH"))
		if len(gopaths) == 0 {
			gopaths = []string{build.Default.GOPATH}
		}
		for i, a := range os.Args[1:] {
			for _, p := range gopaths {
				if strings.HasPrefix(a, p) {
					args[i] = strings.Replace(a, p, "$(GOPATH)", 1)
					break
				}
			}
			if args[i] == "" {
				args[i] = a
			}
		}
		cmdl = " " + strings.Join(args, " ")
		rawcmd := filepath.Base(os.Args[0])
		// Remove .exe suffix to avoid different output on Windows.
		rawcmd = strings.TrimSuffix(rawcmd, ".exe")

		cmdl = fmt.Sprintf("$ %s%s", rawcmd, cmdl)
	}

	args := []string{"--version=" + strconv.Itoa(g.DesignVersion), "--output=" + g.Output, "--cmd=" + cmdl}
	cmd := exec.Command(filepath.Join(g.tmpDir, g.bin), args...)
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

// Remove deletes the package files.
func (g *Generator) Remove() {
	if g.tmpDir != "" {
		os.RemoveAll(g.tmpDir)
		g.tmpDir = ""
	}
}

func (g *Generator) runGoCmd(args ...string) error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	if g.DesignVersion > 2 {
		os.Setenv("GO111MODULE", "on")
	}
	c := exec.Cmd{
		Path: gobin,
		Args: append([]string{gobin}, args...),
		Dir:  g.tmpDir,
	}
	out, err := c.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf(string(out))
		}
		return fmt.Errorf("failed to compile generator: %s", err)
	}
	return nil
}

// cleanupDirs returns the paths of the subdirectories under gendir to delete
// before generating code.
func cleanupDirs(cmd, output string) []string {
	if cmd == "gen" {
		gendirPath := filepath.Join(output, codegen.Gendir)
		gendir, err := os.Open(gendirPath)
		if err != nil {
			return nil
		}
		defer gendir.Close()
		finfos, err := gendir.Readdir(-1)
		if err != nil {
			return []string{gendirPath}
		}
		dirs := []string{}
		for _, fi := range finfos {
			if fi.IsDir() {
				dirs = append(dirs, filepath.Join(gendirPath, fi.Name()))
			}
		}
		return dirs
	}
	return nil
}

// mainT is the template for the generator main.
const mainT = `func main() {
	var (
		out     = flag.String("output", "", "")
		version = flag.String("version", "", "")
		cmdl    = flag.String("cmd", "", "")
		ver int
	)
	{
		flag.Parse()
		if *out == "" {
			fail("missing output flag")
		}
		if *version == "" {
			fail("missing version flag")
		}
		if *cmdl == "" {
			fail("missing cmd flag")
		}
		v, err := strconv.Atoi(*version)
		if err != nil {
			fail("invalid version %s", *version)
		}
		ver = v
	}

	if ver > goa.Major {
		fail("cannot run goa %s on design using goa v%s\n", goa.Version(), *version)
	}
	if err := eval.Context.Errors; err != nil {
		fail(err.Error())
	}
	if err := eval.RunDSL(); err != nil {
		fail(err.Error())
	}
{{- range .CleanupDirs }}
	if err := os.RemoveAll({{ printf "%q" . }}); err != nil {
		fail(err.Error())
	}
{{- end }}
{{- if gt .DesignVersion 2 }}
	codegen.DesignVersion = ver
{{- end }}
	outputs, err := generator.Generate(*out, {{ printf "%q" .Command }})
	if err != nil {
		fail(err.Error())
	}

	fmt.Println(strings.Join(outputs, "\n"))
}

func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, vals...)
	os.Exit(1)
}
`
