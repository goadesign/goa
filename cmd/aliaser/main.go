package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const (
	goaDSL  = "goa.design/goa.v2/dsl"
	restDSL = "goa.design/goa.v2/dsl/rest"
	//rpcDSL    = "goa.design/goa.v2/dsl/rpc"
	aliasFile = "aliases.go"
)

// aliasTmpl is the template used to render the aliasing functions.
var aliasTmpl = template.Must(template.New("alias").Parse(aliasT))

func main() {
	var (
		restPkg, goaPkg  string
		restFuncs, funcs map[string]*ExportedFunc
		restAlias        string
		names, aliases   []string
		err              error
	)
	{
		pkg, err := build.Import(goaDSL, ".", build.FindOnly)
		if err != nil {
			fail("could not find %s package: %s", goaDSL, err)
		}
		goaPkg = pkg.Dir

		pkg, err = build.Import(restDSL, ".", build.FindOnly)
		if err != nil {
			fail("could not find %s package: %s", restDSL, err)
		}
		restPkg = pkg.Dir
		restAlias = filepath.Join(restPkg, aliasFile)
		os.Remove(restAlias) // to avoid parsing them

		restFuncs, err = ParseFuncs(restPkg, "rest")
		if err != nil {
			fail("could not parse functions in %s: %s", restPkg, err)
		}

		funcs, err = ParseFuncs(goaPkg, "dsl")
		if err != nil {
			fail("could not parse functions in %s: %s", goaPkg, err)
		}
		names = make([]string, len(funcs))
		i := 0
		for _, fn := range funcs {
			names[i] = fn.Name
			i++
		}
		sort.Strings(names)
	}

	if aliases, err = CreateAliases(names, funcs, restFuncs, restAlias); err != nil {
		fail("failed to create rest package aliases: %s", err)
	}
	fmt.Printf("rest (%d):\n", len(aliases))
	fmt.Println("  " + strings.Join(aliases, "\n  "))
}

// CreateAliases iterates through the funcs functions and for each creates a
// function with identical name in the file dest. existing is the list of public
// functions that already exist in the destination package and thus should not be
// generated.
// The implementations of the created functions simply call the original
// functions.
func CreateAliases(names []string, funcs, existing map[string]*ExportedFunc, dest string) ([]string, error) {
	var (
		aliases []string
		w       io.Writer
	)
	{
		f, err := os.Create(dest)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %s",
				dest, err)
		}
		defer f.Close()
		w = f
	}
	if _, err := w.Write([]byte(header)); err != nil {
		return nil, err
	}
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

// fail prints a message to stderr then exits the process with status 1.
func fail(msg string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", vals...)
	os.Exit(1)
}

const (
	// header is the generated file header.
	header = `//************************************************************************//
// Aliased goa DSL Functions
//
// Generated with aliaser
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package rest

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/dsl"
)

`

	// aliasT is the source of the text template that renders alias
	// implementations.
	aliasT = `{{ .Comment }}
{{ .Declaration }} {
	{{ if .Return }}return {{ end }}dsl.{{ .Call }}
}`
)
