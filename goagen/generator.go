package main

type (
	// Generator is the common interface for all generators. It exposes the
	// single Generate method which is the entry point to the generation
	// code. The generation code has access to both the user design package
	// under the alias "design" and the command line flags initialized by
	// the corresponding command.
	Generator interface {
		// Generate generates the output (code, documentation etc.) and
		// returns the list of generated filenames on success or an
		// error.
		Generate() ([]string, error)
	}
)
