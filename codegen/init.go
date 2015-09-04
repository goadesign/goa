package codegen

import "github.com/raphael/goa/codegen/code/app"

// Generators contains the supported code and documentation generators indexed by their moniker.
var Generators map[string]Generator

// init loads the generators.
func init() {
	Generators = map[string]Generator{
		"app":    app.NewGenerator(),
		"docs":   docs.NewGenerator(),
		"test":   test.NewGenerator(),
		"client": client.NewGenerator(),
	}
}
