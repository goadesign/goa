package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

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
func NewGenerator(cmd, path, output string, debug bool) *Generator {
	bin := "goa"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	var version int
	var hasVendorDirectory bool
	{
		version = 2
		matched := false
		startPkgLoad := time.Now()
		pkgs, _ := packages.Load(&packages.Config{Mode: packages.NeedFiles | packages.NeedModule}, path)
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]   packages.Load (design files) took %v\n", time.Since(startPkgLoad))
		}
		fset := token.NewFileSet()
		p := regexp.MustCompile(`goa.design/goa/v(\d+)/dsl`)
		for _, pkg := range pkgs {
			// Nil check in case packages.Load can't get module info
			if pkg.Module != nil {
				if _, err := os.Stat(filepath.Join(pkg.Module.Dir, "vendor")); !os.IsNotExist(err) {
					hasVendorDirectory = true
				}
			}
			for _, gof := range pkg.GoFiles {
				if bs, err := os.ReadFile(gof); err == nil {
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
func (g *Generator) Write(_ bool) error {
	var tmpDir string
	{
		wd := "."
		if cwd, err := os.Getwd(); err != nil {
			wd = cwd
		}
		tmp, err := os.MkdirTemp(wd, "goa")
		if err != nil {
			return err
		}
		tmpDir = tmp
	}
	g.tmpDir = tmpDir

	var sections []*codegen.SectionTemplate
	{
		data := map[string]any{
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
			codegen.SimpleImport("runtime/debug"),
			codegen.SimpleImport("sort"),
			codegen.SimpleImport("strconv"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("time"),
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
func (g *Generator) Compile(debug bool) error {
	// We first need to go get the generated package to make sure that all
	// dependencies are added to go.sum prior to compiling.
	startLoad := time.Now()
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, fmt.Sprintf(".%c%s", filepath.Separator, g.tmpDir))
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("expected to find one package in %s", g.tmpDir)
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   packages.Load (temp dir) took %v\n", time.Since(startLoad))
	}

	if !g.hasVendorDirectory {
		startGet := time.Now()
		if err := g.runGoCmd("get", pkgs[0].PkgPath); err != nil {
			return err
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[TIMING]   go get took %v\n", time.Since(startGet))
		}
	}

	startBuild := time.Now()
	err = g.runGoCmd("build", "-o", g.bin)
	if debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   go build took %v\n", time.Since(startBuild))
	}

	// If we're in vendor context we check the error string to see if it's an issue of unsatisfied dependencies
	if err != nil && g.hasVendorDirectory {
		if strings.Contains(err.Error(), "cannot find package") && strings.Contains(err.Error(), "/goa.design/goa/v3/codegen/generator") {
			return errors.New("generated code expected `goa.design/goa/v3/codegen/generator` to be present in the vendor directory, see documentation for more details")
		}
	}

	return err
}

// Run runs the compiled binary and return the output lines.
func (g *Generator) Run(debug bool) ([]string, error) {
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

	args := []string{"--version=" + strconv.Itoa(g.DesignVersion), "--output=" + g.Output, "--cmd=" + cmdl, "--debug=" + strconv.FormatBool(debug)}
	cmd := exec.Command(filepath.Join(g.tmpDir, g.bin), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w\n%s", err, string(out))
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
		_ = os.RemoveAll(g.tmpDir)
		g.tmpDir = ""
	}
}

func (g *Generator) runGoCmd(args ...string) error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	if g.DesignVersion > 2 {
		_ = os.Setenv("GO111MODULE", "on")
	}
	c := exec.Cmd{
		Path: gobin,
		Args: append([]string{gobin}, args...),
		Dir:  g.tmpDir,
	}
	out, err := c.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", out)
		}
		return fmt.Errorf("failed to compile generator: %w", err)
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
		defer func() {
			err := gendir.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to close gendir: %s", err)
			}
		}()
		finfos, err := gendir.Readdir(-1)
		if err != nil {
			return []string{gendirPath}
		}
		var dirs []string

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
		debug   = flag.Bool("debug", false, "")
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

	startBinary := time.Now()
	if *debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   [binary] Starting generated binary execution\n")
	}

	if ver > goa.Major {
		fail("cannot run goa %s on design using goa v%s\n", goa.Version(), *version)
	}

	startCheckErrors := time.Now()
	if err := eval.Context.Errors; err != nil {
		fail(err.Error())
	}
	if *debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   [binary] Check eval.Context.Errors took %v\n", time.Since(startCheckErrors))
	}

	startRunDSL := time.Now()
	if err := eval.RunDSL(); err != nil {
		fail(err.Error())
	}
	if *debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   [binary] eval.RunDSL() took %v\n", time.Since(startRunDSL))
	}

{{- range .CleanupDirs }}
	if err := os.RemoveAll({{ printf "%q" . }}); err != nil {
		fail(err.Error())
	}
{{- end }}
{{- if gt .DesignVersion 2 }}
	codegen.DesignVersion = ver
{{- end }}

	startGenerate := time.Now()
	outputs, err := generate(*out, {{ printf "%q" .Command }}, *debug)
	if err != nil {
		fail(err.Error())
	}
	if *debug {
		fmt.Fprintf(os.Stderr, "[TIMING]   [binary] generator.Generate() took %v\n", time.Since(startGenerate))
		fmt.Fprintf(os.Stderr, "[TIMING]   [binary] Total binary execution took %v\n", time.Since(startBinary))
	}
	fmt.Println(strings.Join(outputs, "\n"))
}

// generate runs code generation and converts panics into a bug report
// request: Goa generators panic on internal invariant violations, and this
// recover is the single boundary turning them into actionable output. Design
// errors never reach this point; eval.RunDSL reports them before generation.
func generate(out, cmd string, dbg bool) ([]string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n\n%s\n", r, debug.Stack())
			fail("This is a bug in Goa, please report it at https://github.com/goadesign/goa/issues and include the stack trace above together with the design that triggered it.\n")
		}
	}()
	return generator.Generate(out, cmd, dbg)
}

func fail(msg string, vals ...any) {
	fmt.Fprintf(os.Stderr, msg, vals...)
	os.Exit(1)
}
`
