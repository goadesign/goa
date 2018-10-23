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
var aliasConstTmpl = template.Must(template.New("aliasConst").Parse(aliasConstT))
var headerTmpl = template.Must(template.New("header").Parse(headerT))

func main() {
	var (
		srcDSL  = flag.String("src", srcDSLDefault, "source DSL `package path`")
		destDSL = flag.String("dest", destDSLDefault, "destination DSL `package path`")
		dslOnly = flag.Bool("dsl", false, "generate DSL aliases only")
	)
	{
		flag.Parse()
	}

	var (
		srcDSLPkgDir      string
		srcDSLPkgName     string
		srcDSLPkgPath     string
		destDSLPkgDir     string
		destDSLPkgName    string
		srcDesignPkgDir   string
		srcDesignPkgPath  string
		destDesignPkgDir  string
		destDesignPkgName string
	)
	{
		pkg, err := build.Import(*srcDSL, ".", 0)
		if err != nil {
			fail("could not find %s package: %s", *srcDSL, err)
		}
		srcDSLPkgDir = pkg.Dir
		srcDSLPkgName = pkg.Name
		srcDSLPkgPath = pkg.ImportPath

		pkg, err = build.Import(*destDSL, ".", 0)
		if err != nil {
			fail("could not parse %s package: %s", *destDSL, err)
		}

		destDSLPkgDir = pkg.Dir
		destDSLPkgName = pkg.Name

		srcDesign := (*srcDSL)[:strings.LastIndex(*srcDSL, "/")] + "/design"
		pkg, err = build.Import(srcDesign, ".", 0)
		if err != nil {
			fail("could not find %s package: %s", srcDesign, err)
		}
		srcDesignPkgDir = pkg.Dir
		srcDesignPkgPath = pkg.ImportPath

		destDesign := (*destDSL)[:strings.LastIndex(*destDSL, "/")] + "/design"
		pkg, err = build.Import(destDesign, ".", 0)
		if err != nil {
			fail("could not parse %s package: %s", destDesign, err)
		}
		destDesignPkgDir = pkg.Dir
		destDesignPkgName = pkg.Name
	}

	var (
		destFuncs           map[string]*ExportedFunc
		funcs               map[string]*ExportedFunc
		consts              []*ExportedConsts
		imports             map[string]*PackageDecl
		dslPath, designPath string
		names               []string
		err                 error
	)
	{
		dslPath = filepath.Join(destDSLPkgDir, aliasFile)
		os.Remove(dslPath) // to avoid parsing them
		designPath = filepath.Join(destDesignPkgDir, aliasFile)
		os.Remove(designPath) // to avoid parsing them

		consts, err = ParseConsts(srcDesignPkgDir)
		if err != nil {
			fail("could not parse constants in %s: %s", destDesignPkgDir, err)
		}

		destFuncs, _, err = ParseFuncs(destDSLPkgDir)
		if err != nil {
			fail("could not parse functions in %s: %s", destDSLPkgDir, err)
		}

		funcs, imports, err = ParseFuncs(srcDSLPkgDir)
		if err != nil {
			fail("could not parse functions in %s: %s", srcDSLPkgDir, err)
		}
		imports[srcDSLPkgName] = &PackageDecl{ImportPath: srcDSLPkgPath, Name: srcDSLPkgName}
		names = make([]string, len(funcs))
		i := 0
		for _, fn := range funcs {
			names[i] = fn.Name
			i++
		}
		sort.Strings(names)
	}

	var dslF *os.File
	{
		var err error
		dslF, err = os.Create(dslPath)
		if err != nil {
			fail("failed to create file %s: %s", dslPath, err)
		}
		defer dslF.Close()
	}

	funcAliases, err := WriteFuncAliases(dslF, names, funcs, destFuncs, destDSLPkgName, imports)
	if err != nil {
		fail("failed to create package function aliases: %s", err)
	}

	if err := dslF.Close(); err != nil {
		fail("failed to close aliases file: %s", err)
	}
	if err := CleanImports(dslPath); err != nil {
		fail("failed to clean DSL aliases imports: %s", err)
	}
	fmt.Printf("\n%s (%d func):\n  ", destDSLPkgDir, len(funcAliases))
	fmt.Println(strings.Join(funcAliases, "\n  "))

	if !*dslOnly {
		var designF *os.File
		{
			designF, err = os.Create(designPath)
			if err != nil {
				fail("failed to create file %s: %s", designPath, err)
			}
			defer designF.Close()
		}

		designImports := map[string]*PackageDecl{
			"design": {Name: "design", ImportPath: srcDesignPkgPath},
		}
		constAliases, err := WriteConstAliases(designF, consts, destDesignPkgName, designImports)
		if err != nil {
			fail("failed to create package const aliases: %s", err)
		}

		if err := designF.Close(); err != nil {
			fail("failed to close aliases file: %s", err)
		}
		if err := CleanImports(designPath); err != nil {
			fail("failed to clean design aliases imports: %s", err)
		}
		fmt.Printf("%s (%d const):\n  ", destDesignPkgDir, len(constAliases))
		fmt.Println(strings.Join(constAliases, "\n  "))
	}
}

// WriteConstAliases writes the given constant definitions to w.
func WriteConstAliases(w io.Writer, consts []*ExportedConsts, destDesignPkgName string, imports map[string]*PackageDecl) ([]string, error) {
	data := map[string]interface{}{"Imports": imports, "PkgName": destDesignPkgName, "Kind": "Constants"}
	if err := headerTmpl.Execute(w, data); err != nil {
		return nil, err
	}
	var (
		aliases []string
	)
	for i, c := range consts {
		if i > 0 {
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return nil, err
			}
		}
		if err := aliasConstTmpl.Execute(w, c); err != nil {
			return nil, err
		}
		aliases = append(aliases, c.Names...)
	}
	return aliases, nil
}

// WriteFuncAliases iterates through the funcs functions and for each creates a
// function with identical name in the file dest. existing is the list of public
// functions that already exist in the destination package and thus should not be
// generated.
// The implementations of the created functions simply call the original
// functions.
func WriteFuncAliases(w io.Writer, names []string, funcs, existing map[string]*ExportedFunc, destDSLPkgName string, imports map[string]*PackageDecl) ([]string, error) {
	data := map[string]interface{}{"Imports": imports, "PkgName": destDSLPkgName, "Kind": "Functions"}
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
	return aliases, nil
}

// CleanImports removes unused imports.
func CleanImports(path string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		content, _ := ioutil.ReadFile(path)
		var buf bytes.Buffer
		scanner.PrintError(&buf, err)
		return fmt.Errorf("%s\n========\nContent:\n%s", buf.String(), content)
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	return format.Node(f, fset, file)
}

// fail prints a message to stderr then exits the process with status 1.
func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", vals...)
	os.Exit(1)
}

const (
	// headerT is the generated file header template.
	headerT = `//************************************************************************//
// Code generated with aliaser, DO NOT EDIT.
//
// Aliased DSL {{ .Kind }}
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

	// aliasConstT is the source of the text template that renders const
	// alias implementations.
	aliasConstT = `const (
{{- range $i, $n := .Names}}
{{- if $.Comments }}
{{ index $.Comments $i }}
{{- end }}
{{ $n }} = design.{{ $n }}
{{- end }}
)
`
)
