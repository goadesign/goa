package design

var (
	apiDSLSource *apiDSL // API DSL
)

// apiDSL contains the non-evaluated API DSL functions
type apiDSL struct {
	name string   // Name of API
	dsls []func() // DSL funcs
}

// API creates or amends the API DSL
func API(name string, dsl func()) {
	if apiDSL != nil {
		if apiDSL.name != name {
			fatalf("API defined with conflicting names '%s' and '%s'",
				apiDSL.name, name)
		}
	} else {
		apiDSL = &apiDSL{name: name}
	}
	apiDSL.dsls = append(apiDSL.dsls, dsl)
}

// BaseParams defines the API base params
func BaseParams(attributes ...*Attribute) {
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.BaseParams = attributes
	default:
		incompatibleDsl()
	}
}

// BasePath defines the API base path
func BasePath(val string) {
	if def, ok := apiDefinition(); ok {
		def.BasePath = val
	}
}

// ResponseTemplate defines a response template
func ResponseTemplate(name string, dsl func()) {
	if def, ok := apiDefinition(); ok {
		template := &ResponseTemplate{Name: name}
		if ok := executeDSL(dsl, template); ok {
			def.ResponseTemplates = append(def.ResponseTemplates, template)
		}
	}
}

// Title sets the API title
func Title(val string) {
	if a, ok := apiDefinition(); ok {
		s.Title = val
	}
}

// Trait defines an API trait
func Trait(name string, val func()) {
	trait := &TraitDefinition{Name: name, Definition: val}
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.Traits = append(c.Traits, trait)
	default:
		incompatibleDsl()
	}

}
