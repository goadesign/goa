package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
)

const (
	goaDSL    = "goa.design/goa.v2/design/dsl"
	restDSL   = "goa.design/goa.v2/rest/design/dsl"
	rpcDSL    = "goa.design/goa.v2/rpc/design/dsl"
	aliasFile = "aliases.go"
)

// aliasTmpl is the template used to render the aliasing functions.
var aliasTmpl = template.Must(template.New("alias").Parse(aliasT))

func main() {
	var (
		restPkg, rpcPkg, goaPkg    string
		restFuncs, rpcFuncs, funcs map[string]*ExportedFunc
		restAlias, rpcAlias        string
		names, aliases             []string
		err                        error
	)
	{
		restPkg, err = codegen.PackageSourcePath(restDSL)
		if err != nil {
			fail("could not find %s package: %s", restDSL, err)
		}
		restAlias = filepath.Join(restPkg, aliasFile)
		os.Remove(restAlias) // to avoid parsing them

		rpcPkg, err = codegen.PackageSourcePath(rpcDSL)
		if err != nil {
			fail("could not find %s package: %s", rpcDSL, err)
		}
		rpcAlias = filepath.Join(rpcPkg, aliasFile)
		os.Remove(rpcAlias)

		goaPkg, err = codegen.PackageSourcePath(goaDSL)
		if err != nil {
			fail("could not find %s package: %s", goaDSL, err)
		}

		restFuncs, err = ParseFuncs(restPkg)
		if err != nil {
			fail("could not parse functions in %s: %s", restPkg, err)
		}

		rpcFuncs, err = ParseFuncs(rpcPkg)
		if err != nil {
			fail("could not parse functions in %s: %s", rpcPkg, err)
		}

		funcs, err = ParseFuncs(goaPkg)
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

	if aliases, err = CreateAliases(names, funcs, rpcFuncs, rpcAlias); err != nil {
		fail("failed to create rpc package aliases: %s", err)
	}
	fmt.Printf("\nrpc (%d):\n", len(aliases))
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

package dsl

import goadsl "goa.design/goa.v2/design/dsl"`

	// aliasT is the source of the text template that renders alias
	// implementations.
	aliasT = `{{ .Comment }}
{{ .Declaration }} {
	{{ if .Return }}return {{ end }}goadsl.{{ .Call }}
}`
)
