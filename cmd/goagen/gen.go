package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/pkg"
)

// Generator is the code generation management data structure.
type Generator struct {
	// Commands is the set of generators to execute.
	Commands []string

	// DesignPath is the Go import path to the design package.
	DesignPath string

	// Output is the absolute path to the output directory.
	Output string

	// bin is the filename of the generated generator.
	bin string

	// tmpDir is the temporary directory used to compile the generator.
	tmpDir string
}

// NewGenerator creates a Generator.
func NewGenerator(cmds []string, path, output string) *Generator {
	bin := "goagen"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	return &Generator{
		Commands:   cmds,
		DesignPath: path,
		Output:     output,
		bin:        bin,
	}
}

// Write writes the main file.
func (g *Generator) Write(gens, debug bool) error {
	var tmpDir string
	{
		wd := "."
		if cwd, err := os.Getwd(); err != nil {
			wd = cwd
		}
		tmp, err := ioutil.TempDir(wd, "goagen")
		if err != nil {
			return err
		}
		tmpDir = tmp
	}
	g.tmpDir = tmpDir

	var sections []*codegen.Section
	{
		data := map[string]interface{}{
			"Generators": generators(g.Commands),
		}
		if gens {
			data["Scaffolds"] = scaffolds(g.Commands)
		}
		imports := []*codegen.ImportSpec{
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("path/filepath"),
			codegen.SimpleImport("sort"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("goa.design/goa.v2/codegen"),
			codegen.SimpleImport("goa.design/goa.v2/codegen/generators"),
			codegen.SimpleImport("goa.design/goa.v2/eval"),
			codegen.SimpleImport("goa.design/goa.v2/pkg"),
			codegen.NewImport("_", g.DesignPath),
		}
		sections = []*codegen.Section{
			codegen.Header("Code Generator", "main", imports),
			&codegen.Section{
				Template: template.Must(template.New("main").Parse(mainTmpl)),
				Data:     data,
			},
		}
	}

	var s codegen.File
	{
		sectionsFunc := func(_ string) []*codegen.Section {
			return sections
		}
		s = codegen.NewSource(filepath.Join(tmpDir, "main.go"), sectionsFunc)
	}

	var w *codegen.Writer
	{
		w = &codegen.Writer{
			Dir:   tmpDir,
			Files: make(map[string]bool),
		}
	}

	return w.Write(s)
}

// Compile compiles the generator.
func (g *Generator) Compile() error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf(`failed to find a go compiler, looked in "%s"`, os.Getenv("PATH"))
	}
	c := exec.Cmd{
		Path: gobin,
		Args: []string{gobin, "build", "-o", g.bin},
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

// Run runs the compiled binary and return the output lines.
func (g *Generator) Run() ([]string, error) {
	args := []string{"--version=" + pkg.Version(), "--output=" + g.Output}
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

// generators returns the names of the generator functions exposed by the
// generator package for the given commands.
func generators(commands []string) []string {
	gens := make([]string, len(commands))
	for i, c := range commands {
		switch c {
		case "server":
			gens[i] = "Server"
		case "client":
			gens[i] = "Client"
		case "openapi":
			gens[i] = "OpenAPI"
		default:
			panic("unknown command " + c) // bug
		}
	}
	return gens
}

// scaffolds returns the names of the generator functions for the scaffolds
// exposed by the generator package for the given commands.
func scaffolds(commands []string) []string {
	var scaf []string
	for _, c := range commands {
		switch c {
		case "server":
			scaf = append(scaf, "ServerScaffold")
		case "client":
			scaf = append(scaf, "ClientScaffold")
		}
	}
	return scaf
}

// mainTmpl is the template for the generator main.
const mainTmpl = `func main() {
	var (
		out     = flag.String("output", "", "")
		version = flag.String("version", "", "")
	)
	{
		flag.Parse()
		if *out == "" {
			fail("missing output flag")
		}
		if *version == "" {
			fail("missing version flag")
		}
	}

	if *version != pkg.Version() {
		fail("cannot run generator produced by goagen version %s and compiled with goa version %s\n", *version, pkg.Version())
	}
        if err := eval.Context.Errors; err != nil {
		fail(err.Error())
	}
	if err := eval.RunDSL(); err != nil {
		fail(err.Error())
	}

	var roots []eval.Root
	{
		rs, err := eval.Context.Roots()
		if err != nil {
			fail(err.Error())
		}
		roots = rs
	}

	var files []codegen.File
{{- range .Generators }}
	{
		fs, err := generator.{{ . }}(roots...)
		if err != nil {
			fail(err.Error())
		}
		files = append(files, fs...)

		// Delete previously generated directories
		dirs := make(map[string]bool)
		for _, file := range files {
			dirs[filepath.Dir(file.OutputPath())] = true
		}
		for d := range dirs {
			if _, err := os.Stat(d); err == nil {
				if err := os.RemoveAll(d); err != nil {
					fail(err.Error())	
				}
			}
		}
	}
{{ end }}
{{- range .Scaffolds }}
	{
		fs, err := generator.{{ . }}(roots...)
		if err != nil {
			fail(err.Error())
		}
		for _, f := range fs {
			if _, err := os.Stat(f.OutputPath()); os.IsNotExist(err) {
				files = append(files, f)
			}
		}
	}
{{ end }}
	var w *codegen.Writer
	{
		w = &codegen.Writer{
			Dir:   *out,
			Files: make(map[string]bool),
		}
	}
	for _, file := range files {
		if err := w.Write(file); err != nil {
			fail(err.Error())
		}
	}

	var outputs []string
	{
		outputs = make([]string, len(w.Files))
		i := 0
		for o := range w.Files {
			outputs[i] = o
			i++
		}
	}
	sort.Strings(outputs)
	fmt.Println(strings.Join(outputs, "\n"))
}

func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, vals...)
	os.Exit(1)
}
`
