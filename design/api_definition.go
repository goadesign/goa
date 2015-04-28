package design

// APIDefinition defines the global properties of the API
type APIDefinition struct {
	Name              string              // API name
	Title             string              // API Title
	Description       string              // API description
	BasePath          string              // Common base path to all API actions
	BaseParams        []*Attribute        // Common path parameters to all API actions
	Traits            []*Trait            // Traits available to all API resources and actions
	ResponseTemplates []*ResponseTemplate // Response templates available to all API actions
}
