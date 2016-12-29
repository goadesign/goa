package genjs

import "github.com/goadesign/goa/design"
import "time"

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

//Timeout Timeout used by JavaScript client when making requests
func Timeout(timeout time.Duration) Option {
	return func(g *Generator) {
		g.Timeout = timeout
	}
}

//Scheme Scheme used by JavaScript client
func Scheme(scheme string) Option {
	return func(g *Generator) {
		g.Scheme = scheme
	}
}

//Host addressed by JavaScript client
func Host(host string) Option {
	return func(g *Generator) {
		g.Host = host
	}
}

//NoExample Do not generate an HTML example file
func NoExample(noExample bool) Option {
	return func(g *Generator) {
		g.NoExample = noExample
	}
}
