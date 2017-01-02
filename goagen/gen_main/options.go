package genmain

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

//DesignPkg Path to design package, only used to mark generated files.
func DesignPkg(designPkg string) Option {
	return func(g *Generator) {
		g.DesignPkg = designPkg
	}
}

//Target Name of generated "app" package
func Target(target string) Option {
	return func(g *Generator) {
		g.Target = target
	}
}

//Force Whether to override existing files
func Force(force bool) Option {
	return func(g *Generator) {
		g.Force = force
	}
}
