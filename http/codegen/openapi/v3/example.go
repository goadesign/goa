// This file adds authored or generated examples to OpenAPI 3 values.
package openapiv3

import (
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type (
	// exampler is the interface used to initialize the example of an
	// OpenAPI object.
	exampler interface {
		setExample(any)
		setExamples(map[string]*ExampleRef)
	}
)

// initExample sets the example or examples of the given object.
func initExamples(obj exampler, attr *expr.AttributeExpr, r *expr.ExampleGenerator, values openapi.Values) {
	selected := values.Example(attr, r)
	if selected == nil {
		obj.setExample(nil)
		return
	}
	examples := values.Examples(attr, attr.ExtractUserExamples())
	switch {
	case len(examples) > 1:
		refs := make(map[string]*ExampleRef, len(examples))
		for _, ex := range examples {
			example := &Example{
				Summary:     ex.Summary,
				Description: ex.Description,
				Value:       openapi.ProjectExample(attr, ex.Value),
			}
			refs[ex.Summary] = &ExampleRef{Value: example}
		}
		obj.setExamples(refs)
		return
	case len(examples) > 0:
		obj.setExample(openapi.ProjectExample(attr, examples[0].Value))
	default:
		obj.setExample(openapi.ProjectExample(attr, selected))
	}
}
