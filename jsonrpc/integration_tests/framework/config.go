package framework

// GeneratorConfig holds code generation configuration
type GeneratorConfig struct {
	// ModuleName is the Go module name for generated code
	ModuleName string
	// PackageName is the package name for generated types (default: "test")
	PackageName string
	// ServiceName is the service name (default: "Test")
	ServiceName string
}

// DefaultGeneratorConfig returns default configuration
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		ModuleName:  "testservice",
		PackageName: "test",
		ServiceName: "Test",
	}
}

// Validate checks if the configuration is valid
func (c *GeneratorConfig) Validate() error {
	// Basic validation - can be extended
	return nil
}