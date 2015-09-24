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
//	BaseParams(func() {
//		Param("accountID", Integer,
//			"API request account. All actions operate on resources belonging to the account."),
//	})
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
	if Design == nil {
		InitDesign()
		Design.Name = name
	} else if Design.Name != name {
		appendError(fmt.Errorf("multiple API definitions: %#v and %#v", name, Design.Name))
		return DSLErrors
	}
	executeDSL(dsl, Design)
	return DSLErrors
}

// Description sets the description on the evaluation scope.
// Description can be called inside API, Resource, Action or MediaType.
func Description(d string) {
	if a, ok := apiDefinition(false); ok {
		a.Description = d
	} else if r, ok := resourceDefinition(false); ok {
		r.Description = d
	} else if a, ok := actionDefinition(false); ok {
		a.Description = d
	} else if m, ok := mediaTypeDefinition(false); ok {
		m.Description = d
	} else if r, ok := responseDefinition(true); ok {
		r.Description = d
	}
}

// BasePath defines the API base path, i.e. the common path prefix to all the
// API actions.
func BasePath(val string) {
	if a, ok := apiDefinition(false); ok {
		a.BasePath = val
	} else if r, ok := resourceDefinition(true); ok {
		r.BasePath = val
	}
}

// BaseParams defines the API base path parameters.
func BaseParams(dsl func()) {
	if a, ok := apiDefinition(true); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.BaseParams = params
		}
	}
}

// ResponseTemplate defines a response template that actions can use to
// specify their responses. A response template has a name and can optionally
// take additional string parameters (up to 9). These parameters can be used in
// the DSL function to define the response fields. For example the function
// could accept an argument that specifies the response media type. These
// arguments must be provided when the corresponding response is defined on an
// action. For example:
// ResponseTemplate("Success", func(mt string) *ResponseDefinition {
//	return &ResponseDefinition{
//		Status: 200,
//		MediaType: mt,
//	}
// }
func ResponseTemplate(name string, p interface{}) {
	if Design.ResponseTemplates == nil {
		Design.ResponseTemplates = make(map[string]*ResponseTemplateDefinition)
	}
	if a, ok := apiDefinition(true); ok {
		if _, ok := a.Responses[name]; ok {
			appendError(fmt.Errorf("multiple definitions for response template %s", name))
			return
		}
		if _, ok := a.ResponseTemplates[name]; ok {
			appendError(fmt.Errorf("multiple definitions for response template %s", name))
			return
		}
		if dsl, ok := p.(func()); ok {
			r := &ResponseDefinition{Name: name}
			if ok := executeDSL(dsl, r); ok {
				a.Responses[name] = r
			}
		} else if tmpl, ok := p.(func(v string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) == 0 {
					appendError(fmt.Errorf("expected one argument when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 2 {
					appendError(fmt.Errorf("expected two arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 3 {
					appendError(fmt.Errorf("expected three arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 4 {
					appendError(fmt.Errorf("expected four arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2], params[3]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 5 {
					appendError(fmt.Errorf("expected five arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2], params[3], params[4]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 6 {
					appendError(fmt.Errorf("expected six arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2], params[3], params[4], params[5]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 7 {
					appendError(fmt.Errorf("expected seven arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7, v8 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 8 {
					appendError(fmt.Errorf("expected eight arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7]) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v1, v2, v3, v4, v5, v6, v7, v8, v9 string)); ok {
			t := func(params ...string) *ResponseDefinition {
				if len(params) < 9 {
					appendError(fmt.Errorf("expected nine arguments when invoking response template %s", name))
					return nil
				}
				r := &ResponseDefinition{Name: name}
				executeDSL(func() {
					tmpl(params[0], params[1], params[2], params[3], params[4], params[5], params[6], params[7], params[8])
				}, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
			}
		} else if tmpl, ok := p.(func(v ...string)); ok {
			t := func(params ...string) *ResponseDefinition {
				r := &ResponseDefinition{Name: name}
				executeDSL(func() { tmpl(params...) }, r)
				return r
			}
			a.ResponseTemplates[name] = &ResponseTemplateDefinition{
				Name:     name,
				Template: t,
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
			if a.Traits == nil {
				a.Traits = make(map[string]*TraitDefinition)
			}
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
