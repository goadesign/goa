package genapp

import "github.com/goadesign/goa/design"

//Option a generator option definition
type Option func(*Generator)

//API The API definition
func API(API *design.APIDefinition) Option {
	return func(g *Generator) {
		g.API = API
	}
}

//OutDir Path to output directory
func OutDir(outDir string) Option {
	return func(g *Generator) {
		g.OutDir = outDir
	}
}

//Target Name of generated package
func Target(target string) Option {
	return func(g *Generator) {
		g.Target = target
	}
}

//NoTest Whether to skip test generation
func NoTest(noTest bool) Option {
	return func(g *Generator) {
		g.NoTest = noTest
	}
}
