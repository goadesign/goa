// Package gengen provides goagen with the ability to run user provided generators (*plugins*).
//
// How to Write a goagen Plugin
//
// The only requirement is that the package exposes a global Generate function with the following
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
// They may also take advantage of Metadata. The goa design language allows defining Metadata on a
// number of definitions: API, Resource, Action, Response and Attribute (which means Type and
// MediaType as well since these definitions are attributes). A metadata field consists of a
// key/value pair where both are simple strings. The generator can use these key/value pairs to
// produce different results, see example below. Metadata has no effect on the buit-in generators.
// The Output directory is available through the codegen.OutputDir global variable.
//
// Package genresnames is an example of a goagen plugin. It creates a file "names.txt" containing
// the names of the API resources sorted in alphabetical order. If a resource has a
// metadata pair with the key "genresnames/name" then the plugin uses the metadata value instead.
//
// Invoke the plugin with:
//	goagen gen -d <Go package path to design package> --pkg-path=<Go package path to genresnames>
//
// Source code:
// 	package genresnames
//
// 	import (
// 		"io/ioutil"
// 		"os"
// 		"path/filepath"
// 		"strings"
//
// 		"gopkg.in/alecthomas/kingpin.v2"
//
// 		"github.com/raphael/goa/design"
// 		"github.com/raphael/goa/goagen/codegen"
// 	)
//
//	// Generate is the function called by goagen to generate the names file.
//	func Generate(api *design.APIDefinition) ([]string, error) {
//		// Make sure to parse the common flags so that codegen.OutputDir gets properly
//		// initialized.
//		app := kingpin.New("Resource names", "Resource name generator")
//		codegen.RegisterFlags(app)
//		if _, err := app.Parse(os.Args[1:]); err != nil {
//			panic(err)
//		}
//
//		// Now iterate through the resources to gather their names
//		names := make([]string, len(api.Resources))
//		i := 0
//		api.IterateResources(func(res *design.ResourceDefinition) error {
//			if n, ok := res.Metadata["genresnames/name"]; ok {
//				names[i] = n
//			} else {
//				names[i] = res.Name
//			}
//			i++
//			return nil
//		})
//		content := strings.Join(names, "\n")
//
//		// Write the output file and return its name
//		outputFile := filepath.Join(codegen.OutputDir, "names.txt")
//		ioutil.WriteFile(outputFile, []byte(content), 0644)
//		return []string{outputFile}, nil
//	}
package gengen
