package dsl

import (
	"fmt"

	. "github.com/raphael/goa/design"
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
	if Design != nil {
		appendError(fmt.Errorf("multiple API definitions"))
	} else {
		Design = &APIDefinition{Name: name}
		executeDSL(dsl, Design)
	}
	if len(dslErrors) > 0 {
		reportErrors()
	}
	//generate() TBD
	return nil // Need to return something for 'var _ = ' trick
}

// Description sets the description on the evaluation scope.
func Description(d string) {
	if a, ok := apiDefinition(false); ok {
		a.Description = d
	} else if r, ok := resourceDefinition(false); ok {
		r.Description = d
	} else if a, ok := actionDefinition(false); ok {
		a.Description = d
	} else if m, ok := mediaTypeDefinition(true); ok {
		m.Description = d
	}
}

// BaseParams defines the API base params
func BaseParams(dsl func()) {
	if a, ok := apiDefinition(true); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.BaseParams = params
		}
	}
}

// BasePath defines the API base path
func BasePath(val string) {
	if a, ok := apiDefinition(false); ok {
		a.BasePath = val
	} else if r, ok := resourceDefinition(true); ok {
		r.BasePath = val
	}
}

// ResponseTemplate defines a response template using either a DSL or a template function that
// can take 1 to 9 string arguments or a "...string" argument.
func ResponseTemplate(name string, p interface{}) {
	if a, ok := apiDefinition(true); ok {
		if _, ok := a.ResponseTemplates[name]; ok {
			appendError(fmt.Errorf("multiple definitions for response template %s", name))
			return
		}
		if _, ok := a.ResponseTemplateFuncs[name]; ok {
			appendError(fmt.Errorf("multiple definitions for response template %s", name))
			return
		}
		template := &ResponseTemplateDefinition{Name: name}
		if dsl, ok := p.(func()); ok {
			if ok := executeDSL(dsl, template); ok {
				a.ResponseTemplates[name] = template
			}
		} else if tmpl, ok := p.(func(v string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) == 0 {
					appendError(fmt.Errorf("expected one argument when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 2 {
					appendError(fmt.Errorf("expected two arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 3 {
					appendError(fmt.Errorf("expected three arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 4 {
					appendError(fmt.Errorf("expected four arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 5 {
					appendError(fmt.Errorf("expected five arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3], params[4])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 6 {
					appendError(fmt.Errorf("expected six arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3], params[4], params[5])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 7 {
					appendError(fmt.Errorf("expected seven arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7, v8 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 8 {
					appendError(fmt.Errorf("expected eight arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7])
				}
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7, v8, v9 string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				if len(params) < 9 {
					appendError(fmt.Errorf("expected nine arguments when invoking response template %s", name))
					return nil
				} else {
					return tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8])
				}
			}
		} else if tmpl, ok := p.(func(v ...string) *ResponseTemplateDefinition); ok {
			a.ResponseTemplateFuncs[name] = func(params ...string) *ResponseTemplateDefinition {
				return tmpl(params...)
			}
		}
	}
}

// Title sets the API title
func Title(val string) {
	if a, ok := apiDefinition(true); ok {
		a.Title = val
	}
}

// Trait defines an API trait
func Trait(name string, val ...func()) {
	if a, ok := apiDefinition(false); ok {
		if len(val) < 1 {
			appendError(fmt.Errorf("missing trait DSL for %s", name))
		} else {
			if _, ok := a.Traits[name]; ok {
				appendError(fmt.Errorf("multiple definitions for trait %s", name))
				return
			}
			trait := &TraitDefinition{Name: name, Dsl: val[0]}
			a.Traits[name] = trait
		}
	} else if r, ok := resourceDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.Dsl, r)
		} else {
			appendError(fmt.Errorf("unknown trait %s", name))
		}
	} else if a, ok := actionDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.Dsl, a)
		} else {
			appendError(fmt.Errorf("unknown trait %s", name))
		}
	} else if a, ok := attributeDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.Dsl, a)
		} else {
			appendError(fmt.Errorf("unknown trait %s", name))
		}
	}
}
