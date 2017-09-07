package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
)

const (
	srcDSLDefault  = "goa.design/goa/dsl"
	destDSLDefault = "goa.design/goa/http/dsl"
	//rpcDSL    = "goa.design/goa/dsl/rpc"
	aliasFile = "aliases.go"
)

// aliasTmpl is the template used to render the aliasing functions.
var aliasTmpl = template.Must(template.New("alias").Parse(aliasT))
var headerTmpl = template.Must(template.New("header").Parse(headerT))

func main() {
	var (
		srcDSL  = flag.String("src", srcDSLDefault, "source DSL `package path`")
		destDSL = flag.String("dest", destDSLDefault, "destination DSL `package path`")
	)
	{
		flag.Parse()
	}

	var (
		srcPkgDir   string
		srcPkgName  string
		srcPkgPath  string
		destPkgDir  string
		destPkgName string
	)
	{
		pkg, err := build.Import(*srcDSL, ".", 0)
		if err != nil {
			fail("could not find %s package: %s", srcDSL, err)
		}
		srcPkgDir = pkg.Dir
		srcPkgName = pkg.Name
		srcPkgPath = pkg.ImportPath

		pkg, err = build.Import(*destDSL, ".", 0)
		if err != nil {
			fail("could not parse %s package: %s", destDSL, err)
		}
		destPkgDir = pkg.Dir
		destPkgName = pkg.Name
	}

	var (
		destFuncs map[string]*ExportedFunc
		funcs     map[string]*ExportedFunc
		imports   map[string]*PackageDecl
		path      string
		names     []string
		err       error
	)
	{
		path = filepath.Join(destPkgDir, aliasFile)
		os.Remove(path) // to avoid parsing them

		destFuncs, _, err = ParseFuncs(destPkgDir)
		if err != nil {
			fail("could not parse functions in %s: %s", destPkgDir, err)
		}

		funcs, imports, err = ParseFuncs(srcPkgDir)
		if err != nil {
			fail("could not parse functions in %s: %s", srcPkgDir, err)
		}
		imports[srcPkgName] = &PackageDecl{ImportPath: srcPkgPath, Name: srcPkgName}
		names = make([]string, len(funcs))
		i := 0
		for _, fn := range funcs {
			names[i] = fn.Name
			i++
		}
		sort.Strings(names)
	}

	aliases, err := CreateAliases(names, funcs, destFuncs, destPkgName, imports, path)
	if err != nil {
		fail("failed to create package aliases: %s", err)
	}
	fmt.Printf("%s (%d):\n", destPkgDir, len(aliases))
	fmt.Println("  " + strings.Join(aliases, "\n  "))
}

// CreateAliases iterates through the funcs functions and for each creates a
// function with identical name in the file dest. existing is the list of public
// functions that already exist in the destination package and thus should not be
// generated.
// The implementations of the created functions simply call the original
// functions.
func CreateAliases(names []string, funcs, existing map[string]*ExportedFunc, destPkgName string, imports map[string]*PackageDecl, path string) ([]string, error) {
	var (
		w io.Writer
		f *os.File
	)
	{
		var err error
		f, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %s",
				path, err)
		}
		defer f.Close()
		w = f
	}

	data := map[string]interface{}{"Imports": imports, "PkgName": destPkgName}
	if err := headerTmpl.Execute(w, data); err != nil {
		return nil, err
	}

	var (
		aliases []string
	)
	for i, name := range names {
		if _, ok := existing[name]; ok {
			continue
		}
		if i > 0 {
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return nil, err
			}
		}
		if err := aliasTmpl.Execute(w, funcs[name]); err != nil {
			return nil, err
		}
		aliases = append(aliases, name)
	}

	// Clean unused imports
	f.Close()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		content, _ := ioutil.ReadFile(path)
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return nil, fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
	}
	all := astutil.Imports(fset, file)
	for _, group := range all {
		for _, imp := range group {
			path := strings.Trim(imp.Path.Value, `"`)
			if !astutil.UsesImport(file, path) {
				if imp.Name != nil {
					astutil.DeleteNamedImport(fset, file, imp.Name.Name, path)
				} else {
					astutil.DeleteImport(fset, file, path)
				}
			}
		}
	}
	ast.SortImports(fset, file)
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, err
	}
	if err := format.Node(f, fset, file); err != nil {
		return nil, err
	}

	return aliases, nil
}

// fail prints a message to stderr then exits the process with status 1.
func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", vals...)
	os.Exit(1)
}

const (
	// headerT is the generated file header template.
	headerT = `//************************************************************************//
// Aliased DSL Functions
//
// Generated with aliaser
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package {{ .PkgName }}

import (
	{{- range .Imports }}
	{{ if .Name }}{{ .Name }} {{ end }}"{{ .ImportPath }}"
	{{- end }}
)

`

	// aliasT is the source of the text template that renders alias
	// implementations.
	aliasT = `{{ .Comment }}
{{ .Declaration }} {
	{{ if .Return }}return {{ end }}dsl.{{ .Call }}
}`
)
