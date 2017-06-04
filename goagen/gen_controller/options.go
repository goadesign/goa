package gencontroller

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

//AppPkg Name of generated "app" package
func AppPkg(pkg string) Option {
	return func(g *Generator) {
		g.AppPkg = pkg
	}
}

//Force Whether to override existing files
func Force(force bool) Option {
	return func(g *Generator) {
		g.Force = force
	}
}

//Regen Whether to regenerate scaffolding while maintaining controller impls
func Regen(regen bool) Option {
	return func(g *Generator) {
		g.Regen = regen
	}
}

//Pkg sets the name of generated package
func Pkg(name string) Option {
	return func(g *Generator) {
		g.Pkg = name
	}
}

//Resource Name of generated the generated file
func Resource(res string) Option {
	return func(g *Generator) {
		g.Resource = res
	}
}
