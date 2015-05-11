package design

import "fmt"

var (
	definition *APIDefinition // API definition created via DSL
)

func API(name string, dsl func()) {
	if definition != nil {
		appendError(fmt.Errorf("multiple API definitions."))
	} else {
		definition = &APIDefinition{Name: name}
		executeDSL(dsl, definition)
	}
	if len(dslErrors) > 0 {
		reportErrors()
	}
	//generate() TBD
}

// BaseParams defines the API base params
func BaseParams(attributes ...*AttributeDefinition) {
	switch c := ctxStack.current().(type) {
	case *APIDefinition:
		c.BaseParams = attributes
	default:
		incompatibleDsl("BaseParams")
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
		template := &ResponseTemplateDefinition{Name: name}
		if ok := executeDSL(dsl, template); ok {
			def.ResponseTemplates = append(def.ResponseTemplates, template)
		}
	}
}

// Title sets the API title
func Title(val string) {
	if a, ok := apiDefinition(); ok {
		a.Title = val
	}
}

// Trait defines an API trait
func Trait(name string, val func()) {
	if a, ok := apiDefinition(); ok {
		trait := &TraitDefinition{Name: name, Dsl: val}
		a.Traits = append(a.Traits, trait)
	}

}
