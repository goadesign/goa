package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/pkg"
)

// GenPackage is the code generation management data structure.
type GenPackage struct {
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

// NewGenPackage creates a GenPackage.
func NewGenPackage(cmds []string, path, output string) *GenPackage {
	bin := "goagen"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	return &GenPackage{
		Commands:   cmds,
		DesignPath: path,
		Output:     output,
		bin:        bin,
	}
}

// WriteMain writes the main file.
func (g *GenPackage) Write(gens, debug bool) error {
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

	var s codegen.SourceFile
	{
		s = codegen.SourceFile{Path: filepath.Join(tmpDir, "main.go")}
		imports := []*codegen.ImportSpec{
			codegen.SimpleImport("flag"),
			codegen.SimpleImport("fmt"),
			codegen.SimpleImport("os"),
			codegen.SimpleImport("path/filepath"),
			codegen.SimpleImport("sort"),
			codegen.SimpleImport("strings"),
			codegen.SimpleImport("goa.design/goa.v2/codegen"),
			codegen.SimpleImport("goa.design/goa.v2/eval"),
			codegen.SimpleImport("goa.design/goa.v2/pkg"),
			codegen.NewImport("_", g.DesignPath),
		}
		for _, cmd := range g.Commands {
			imports = append(imports, codegen.SimpleImport("goa.design/goa.v2/codegen/generators/"+cmd))
		}
		if err := s.WriteHeader("Code Generator", "main", imports); err != nil {
			return err
		}
	}

	return s.ExecuteTemplate("main", mainTmpl, nil, g.Commands)
}

// Compile compiles the package.
func (g *GenPackage) Compile() error {
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
func (g *GenPackage) Run() ([]string, error) {
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
func (g *GenPackage) Remove() {
	if g.tmpDir != "" {
		os.RemoveAll(g.tmpDir)
		g.tmpDir = ""
	}
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
		fail("goa DSL was run with goa version %s but compiled generator is running %s\n", *version, pkg.Version())
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

	var writers []codegen.FileWriter
	{
{{- range . }}
		ws, err := {{ . }}.Writers(roots...)
		if err != nil {
			fail(err.Error())
		}
		writers = append(writers, ws...)
{{ end }}	}

	outputs := make([]string, len(writers))
	for i, w := range writers {
		d := filepath.Dir(w.OutputPath())
		if err := os.MkdirAll(d, 0755); err != nil {
			fail(err.Error())
		}
		if err := codegen.Render(w, *out); err != nil {
			fail(err.Error())
		}
		outputs[i] = filepath.Join(*out, w.OutputPath())
	}

	sort.Strings(outputs)
	fmt.Println(strings.Join(outputs, "\n"))
}

func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, vals...)
	os.Exit(1)
}
`
