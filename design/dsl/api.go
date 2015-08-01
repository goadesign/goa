package dsl

import (
	"fmt"

	"github.com/raphael/goa/design"
)

// API defines the top level API DSL.
//
// API("cellar", func() {
//	Title("The virtual wine cellar")
//	Description("A basic example of a CRUD API implemented with goa")
//	BasePath("/:accountID")
//	BaseParams(
//		Param("accountID", Integer,
//			"API request account. All actions operate on resources belonging to the account."),
//	)
//	ResponseTemplate("NotFound", func() {
//		Description("Resource not found")
//		Status(404)
//		MediaType("application/json")
//	})
//	Trait("Authenticated", func() {
//		Headers(func() {
//			Header("Auth-Token", String)
//			Required("Auth-Token")
//		})
//	})
// })
//
func API(name string, dsl func()) error {
	if Definition != nil {
		appendError(fmt.Errorf("multiple API definitions"))
	} else {
		Definition = &design.APIDefinition{Name: name}
		executeDSL(dsl, Definition)
	}
	if len(dslErrors) > 0 {
		reportErrors()
	}
	//generate() TBD
	return nil // Need to return something for 'var _ = ' trick
}

// BaseParams defines the API base params
func BaseParams(attributes ...*design.AttributeDefinition) {
	switch c := ctxStack.current().(type) {
	case *design.APIDefinition:
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
