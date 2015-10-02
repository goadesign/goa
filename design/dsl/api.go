package dsl

import . "github.com/raphael/goa/design"

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
//      Type("BottlePayload", func() {
//		Attribute("name")
//	})
//	Trait("Authenticated", func() {
//		Headers(func() {
//			Header("Auth-Token", String)
//			Required("Auth-Token")
//		})
//	})
// })
//
// We can't rely on this being run first, any of the top level DSL could run
// in any order. The top level DSLs are API, Resource, MediaType and Type.
// The first one to be called executes InitDesign.
// API checks whether that has been called yet (i.e. if the global variable
// Design is initialized) and if so makes sure that if it has a name it is the
// same as the one used in the argument: API can be called multiple times as
// long as it's always to define the same API.
func API(name string, dsl func()) *APIDefinition {
	if Design == nil {
		InitDesign()
	} else if Design.Name != "" && Design.Name != name {
		ReportError("multiple API definitions: %#v and %#v", name, Design.Name)
		return nil
	}
	if !topLevelDefinition(true) {
		return nil
	}
	Design.Name = name
	Design.DSL = dsl
	return Design
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
	} else if a, ok := attributeDefinition(false); ok {
		a.Description = d
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
	if a, ok := apiDefinition(false); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.BaseParams = params
		}
	} else if r, ok := resourceDefinition(true); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			r.BaseParams = params
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
	if a, ok := apiDefinition(true); ok {
		if a.Responses == nil {
			a.Responses = make(map[string]*ResponseDefinition)
		}
		if a.ResponseTemplates == nil {
			a.ResponseTemplates = make(map[string]*ResponseTemplateDefinition)
		}
		if _, ok := a.Responses[name]; ok {
			ReportError("multiple definitions for response template %s", name)
			return
		}
		if _, ok := a.ResponseTemplates[name]; ok {
			ReportError("multiple definitions for response template %s", name)
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
					ReportError("expected one argument when invoking response template %s", name)
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
					ReportError("expected two arguments when invoking response template %s", name)
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
					ReportError("expected three arguments when invoking response template %s", name)
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
					ReportError("expected four arguments when invoking response template %s", name)
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
					ReportError("expected five arguments when invoking response template %s", name)
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
					ReportError("expected six arguments when invoking response template %s", name)
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
					ReportError("expected seven arguments when invoking response template %s", name)
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
					ReportError("expected eight arguments when invoking response template %s", name)
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
					ReportError("expected nine arguments when invoking response template %s", name)
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
			ReportError("missing trait DSL for %s", name)
		} else {
			if _, ok := a.Traits[name]; ok {
				ReportError("multiple definitions for trait %s", name)
				return
			}
			trait := &TraitDefinition{Name: name, DSL: val[0]}
			if a.Traits == nil {
				a.Traits = make(map[string]*TraitDefinition)
			}
			a.Traits[name] = trait
		}
	} else if r, ok := resourceDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.DSL, r)
		} else {
			ReportError("unknown trait %s", name)
		}
	} else if a, ok := actionDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.DSL, a)
		} else {
			ReportError("unknown trait %s", name)
		}
	} else if a, ok := attributeDefinition(false); ok {
		if trait, ok := Design.Traits[name]; ok {
			executeDSL(trait.DSL, a)
		} else {
			ReportError("unknown trait %s", name)
		}
	}
}
