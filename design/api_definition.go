package design

// APIDefinition defines the global properties of the API
type APIDefinition struct {
	Name              string                                 // API name
	Title             string                                 // API Title
	Description       string                                 // API description
	BasePath          string                                 // Common base path to all API actions
	BaseParams        []*AttributeDefinition                 // Common path parameters to all API actions
	Resources         map[string]*ResourceDefinition         // Exposed resources indexed by name
	Traits            map[string]*TraitDefinition            // Traits available to all API resources and actions indexed by name
	ResponseTemplates map[string]*ResponseTemplateDefinition // Response templates available to all API actions indexed by name
}

// ResponseTemplateDefinition defines a HTTP response status and optional validation rules.
type ResponseTemplateDefinition struct {
	Name        string               // Response name
	Status      int                  // HTTP status
	Description string               // Response description
	MediaType   *MediaTypeDefinition // Response body media type if any
	Headers     []*HeaderDefinition  // Response header validations
}

// TraitDefinition defines a set of reusable properties.
type TraitDefinition struct {
	Name string // Trait name
	Dsl  func() // Trait DSL
}
