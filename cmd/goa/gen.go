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
	// Command is the name of the command to run.
	Command string

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
func NewGenerator(cmd string, path, output string) *Generator {
	bin := "goa"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	return &Generator{
		Command:    cmd,
		DesignPath: path,
		Output:     output,
		bin:        bin,
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

	var sections []*codegen.Section
	{
		data := map[string]interface{}{
			"Command":     g.Command,
			"CleanupDirs": cleanupDirs(g.Command, g.Output),
		}
		imports := []*codegen.ImportSpec{
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("path/filepath"),
			codegen.SimpleImport("sort"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("goa.design/goa.v2/codegen"),
			codegen.SimpleImport("goa.design/goa.v2/codegen/generator"),
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
		s = codegen.NewSource("main.go", sectionsFunc)
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
	var cmdl string
	{
		args := make([]string, len(os.Args)-1)
		gopaths := filepath.SplitList(os.Getenv("GOPATH"))
		for i, a := range os.Args[1:] {
			for _, p := range gopaths {
				if strings.Contains(a, p) {
					args[i] = strings.Replace(a, p, "$(GOPATH)", -1)
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

	args := []string{"--version=" + pkg.Version(), "--output=" + g.Output, "--cmd=" + cmdl}
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

// cleanupDirs returns the names of the directories to delete before generating
// code.
func cleanupDirs(cmd, output string) []string {
	if cmd == "gen" {
		return []string{filepath.Join(output, "gen")}
	}
	return nil
}

// mainTmpl is the template for the generator main.
const mainTmpl = `func main() {
	var (
		out     = flag.String("output", "", "")
		version = flag.String("version", "", "")
		cmdl    = flag.String("cmd", "", "")
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
	}

	if *version != pkg.Version() {
		fail("cannot run generator produced by goa version %s and compiled with goa version %s\n", *version, pkg.Version())
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

	var gens []generator.Genfunc
	{
		gs, err := generator.Generators({{ printf "%q" .Command }})
		if err != nil {
			fail(err.Error())
		}
		gens = gs
	}

	var genfiles []codegen.File
	for _, gen := range gens {
		fs, err := gen(roots)
		if err != nil {
			fail(err.Error())
		}
		genfiles = append(genfiles, fs...)
	}

	var w *codegen.Writer
	{
		w = &codegen.Writer{
			Dir:   *out,
			Files: make(map[string]bool),
		}
	}
{{- range .CleanupDirs }}
	if err := os.RemoveAll({{ printf "%q" . }}); err != nil {
		fail(err.Error())
	}
{{- end }}
	for _, f := range genfiles {
		if err := w.Write(f); err != nil {
			fail(err.Error())
		}
	}

	var outputs []string
	{
		outputs = make([]string, len(w.Files))
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}	
		i := 0
		for o := range w.Files {
			rel, err := filepath.Rel(cwd, o)
			if err != nil {
				rel = o
			}
			outputs[i] = rel
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
