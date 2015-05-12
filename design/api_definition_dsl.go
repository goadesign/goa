package design

import "fmt"

var (
	Definition *APIDefinition // API definition created via DSL
)

func API(name string, dsl func()) {
	if Definition != nil {
		appendError(fmt.Errorf("multiple API definitions."))
	} else {
		Definition = &APIDefinition{Name: name}
		executeDSL(dsl, Definition)
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
	if a, ok := apiDefinition(); ok {
		a.BasePath = val
	}
}

// ResponseTemplate defines a response template
func ResponseTemplate(name string, dsl func()) {
	if a, ok := apiDefinition(); ok {
		if _, ok := a.ResponseTemplates[name]; ok {
			appendError(fmt.Errorf("multiple definitions for response template %s", name))
			return
		}
		template := &ResponseTemplateDefinition{Name: name}
		if ok := executeDSL(dsl, template); ok {
			a.ResponseTemplates[name] = template
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
		if _, ok := a.Traits[name]; ok {
			appendError(fmt.Errorf("multiple definitions for trait %s", name))
			return
		}
		trait := &TraitDefinition{Name: name, Dsl: val}
		a.Traits[name] = trait
	}

}
