package genclient

import (
	"github.com/goadesign/goa/design"
)

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

//ToolDirName Name of tool directory where CLI main is generated once
func ToolDirName(toolDirName string) Option {
	return func(g *Generator) {
		g.ToolDirName = toolDirName
	}
}

//Tool Name of CLI tool
func Tool(tool string) Option {
	return func(g *Generator) {
		g.Tool = tool
	}
}

//NoTool Whether to skip tool generation
func NoTool(noTool bool) Option {
	return func(g *Generator) {
		g.NoTool = noTool
	}
}
