package codegen

import "github.com/goadesign/goa/dslengine"

type generator interface {
	Generate() ([]string, error)
}

// Run runs all generators passed as parameter. Call ParseDSL first to
// fill `design.Design`.  Each `goa` generator lives in its own
// `goagen/gen_something` package in `generator.go` and has a
// `Generator` object which implements the interface required here.
//
//   codegen.Run(
//     &genapp.Generator{
//       API: design.Design,
//       Target: "app",
//     },
//     &genmain.Generator{
//       API: design.Design,
//     },
//   )
//
func Run(generators ...generator) {
	for _, generator := range generators {
		dslengine.PrintFilesOrFail(generator.Generate())
	}
}

// ParseDSL will run the DSL engine and analyze any imported `design`
// package, creating your `design.APIDefinition` along the way.
func ParseDSL() {
	// Catch any init-time errors
	dslengine.FailOnError(dslengine.Errors)

	// Catch any runtime errors, when analyzing the DSL
	dslengine.FailOnError(dslengine.Run())
}
