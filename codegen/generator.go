package codegen

// Generator is the command generator interface.
// It exposes the single Generate method called by the generator generated code.
type Generator interface {
	Generate() error
}

// NewAppGenerator instantiates an application code generator.
func NewAppGenerator() Generator {
}
