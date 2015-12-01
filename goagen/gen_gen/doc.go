// Package gengen provides goagen with the ability to run user provided generators.
//
// How to write a goagen generator package:
//
// The only requirement is that the package expose a global Generate function with the following
// signature:
//
// 	func Generate(api *design.APIDefinition) ([]string, error)
//
// where api is the API definition computed from the design DSL. On success Generate should return
// the path to the generated files. On error the error message gets displayed to the user (and
// goagen exist with status 1).
//
// The Generate method should take advantage of the APIDefinition IterateXXX methods to iterate
// through the API resources, media types and types to guarantee that the order doesn't change
// between two invokation of the function (thereby generating different output even if the design
// hasn't changed).
//
// The Output directory is available through the codegen.OutputDir global variable. Here is an
// example of a custom generator package source:
//
// Package genresnames implements a goa generator which creates a file "names.txt"
// containing the names of the API resources sorted in alphabetical order.
// invoke with "goagen -d <Go package path to design package> gen --pkg-path=<Go package path to genresnames>"
//
//	 package genresnames
//
//	 import (
//	 	"io/ioutil"
//	 	"path/filepath"
//	 	"strings"
//
//	 	"github.com/raphael/goa/goagen/codegen"
//	 	"github.com/raphael/goa/design"
//	 )
//
//	 // Generate is the function called by goagen to generate the "names.txt" file.
//	 func Generate(api *design.APIDefinition) ([]string, error) {
//	 	names := make([]string, len(api.Resources))
//	 	i := 0
//	 	api.IterateResources(func(res *design.ResourceDefinition) error {
//	 		names[i] = res.Name
//	 		i++
//	 		return nil
//	 	})
//	 	content := strings.Join(names, "\n")
//	 	outputFile := filepath.Join(codegen.OutputDir, "names.txt")
//	 	ioutil.WriteFile(outputFile, []byte(content), 0644)
//	 	return []string{outputFile}, nil
//	 }
package gengen
