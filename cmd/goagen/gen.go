package main

// GenPackage is the code generation management data structure.
type GenPackage struct {
	// Commands is the set of generators to execute.
	Commands []string

	// DesignPath is the Go import path to the design package.
	DesignPath string

	// Output is the absolute path to the output directory.
	Output string
}

// NewGenPackage creates a GenPackage.
func NewGenPackage(cmds []string, path, output string) *GenPackage {
	return &GenPackage{
		Commands:   cmds,
		DesignPath: path,
		Output:     output,
	}
}

// WriteMain writes the main file.
func (g *GenPackage) WriteMain(gens, debug bool) error {
	return nil
}

// Compile compiles the package.
func (g *GenPackage) Compile() error {
	return nil
}

// Run runs the compiled binary and return the output lines.
func (g *GenPackage) Run() ([]string, error) {
	return nil, nil
}

// Remove deletes the package files.
func (g *GenPackage) Remove() {

}
