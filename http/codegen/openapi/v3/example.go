package openapiv3

import "goa.design/goa/v3/expr"

type (
	// exampler is the interface used to initialize the example of an
	// OpenAPI object.
	exampler interface {
		setExample(interface{})
		setExamples(map[string]*ExampleRef)
	}
)

// initExample sets the example or examples of the given object.
func initExamples(obj exampler, attr *expr.AttributeExpr, r *expr.Random) {
	examples := attr.ExtractUserExamples()
	switch {
	case len(examples) > 1:
		refs := make(map[string]*ExampleRef, len(examples))
		for _, ex := range examples {
			example := &Example{
				Summary:     ex.Summary,
				Description: ex.Description,
				Value:       ex.Value,
			}
			refs[ex.Summary] = &ExampleRef{Value: example}
		}
		obj.setExamples(refs)
		return
	case len(examples) > 0:
		obj.setExample(examples[0].Value)
	default:
		obj.setExample(attr.Example(r))
	}
}
