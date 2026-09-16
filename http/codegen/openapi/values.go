// This file stores alternate OpenAPI text and examples for one specification
// build. Builders read these values without changing the evaluated Goa design.
package openapi

import (
	"maps"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// storedExample keeps an immutable example value with the exact design
	// expression used to find its translated description.
	storedExample struct {
		value  *expr.ExampleExpr
		source *expr.ExampleExpr
	}

	// Values contains alternate titles, descriptions, and examples for one
	// OpenAPI build. The zero value uses the evaluated Goa design unchanged.
	// Methods that add values return a new independent Values.
	Values struct {
		titles       map[eval.Expression]string
		descriptions map[eval.Expression]string
		examples     map[*expr.AttributeExpr][]storedExample
	}
)

// WithTitle returns a copy of v that uses title for target.
func (v Values) WithTitle(target eval.Expression, title string) Values {
	result := v.copy()
	if result.titles == nil {
		result.titles = make(map[eval.Expression]string)
	}
	result.titles[target] = title
	return result
}

// WithDescription returns a copy of v that uses description for target.
func (v Values) WithDescription(target eval.Expression, description string) Values {
	result := v.copy()
	if result.descriptions == nil {
		result.descriptions = make(map[eval.Expression]string)
	}
	result.descriptions[target] = description
	return result
}

// WithExamples returns a copy of v that uses examples for attribute. Copies
// made from attribute by Goa use the same examples.
func (v Values) WithExamples(attribute *expr.AttributeExpr, examples []*expr.ExampleExpr) Values {
	result := v.copy()
	if result.examples == nil {
		result.examples = make(map[*expr.AttributeExpr][]storedExample)
	}
	result.examples[attribute.AuthoredAttribute()] = storeExamples(examples)
	return result
}

// Title returns the title stored for target or fallback when none was stored.
func (v Values) Title(target eval.Expression, fallback string) string {
	if title, ok := v.titles[target]; ok {
		return title
	}
	return fallback
}

// Description returns the description stored for target or fallback when none
// was stored.
func (v Values) Description(target eval.Expression, fallback string) string {
	if description, ok := v.descriptions[target]; ok {
		return description
	}
	return fallback
}

// Examples returns the examples stored for attribute or a copy of fallback
// when none were stored.
func (v Values) Examples(attribute *expr.AttributeExpr, fallback []*expr.ExampleExpr) []*expr.ExampleExpr {
	if examples, ok := v.examples[attribute.AuthoredAttribute()]; ok {
		return v.materializeExamples(examples)
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if examples, ok := v.examples[userType.Attribute().AuthoredAttribute()]; ok {
			return v.materializeExamples(examples)
		}
	}
	return v.materializeExamples(storeExamples(fallback))
}

// Example returns the last stored or authored example, or generates one when
// none exists. A generator configured to suppress examples returns nil.
func (v Values) Example(attribute *expr.AttributeExpr, generator *expr.ExampleGenerator) any {
	copy := *attribute
	copy.UserExamples = v.Examples(attribute, attribute.ExtractUserExamples())
	return copy.Example(generator)
}

// copy returns independent maps while retaining the immutable values they
// contain. Example lists are copied again when they are changed or read.
func (v Values) copy() Values {
	return Values{
		titles:       maps.Clone(v.titles),
		descriptions: maps.Clone(v.descriptions),
		examples:     maps.Clone(v.examples),
	}
}

// storeExamples copies a complete example list while retaining the exact
// expression used to look up each translated description.
func storeExamples(examples []*expr.ExampleExpr) []storedExample {
	if examples == nil {
		return nil
	}
	result := make([]storedExample, len(examples))
	for index, example := range examples {
		copy := *example
		copy.Value = duplicateJSONValue(example.Value)
		result[index] = storedExample{value: &copy, source: example}
	}
	return result
}

// materializeExamples returns a fresh example list with every replacement
// description applied from its exact source expression.
func (v Values) materializeExamples(examples []storedExample) []*expr.ExampleExpr {
	if examples == nil {
		return nil
	}
	result := make([]*expr.ExampleExpr, len(examples))
	for index, stored := range examples {
		copy := *stored.value
		copy.Description = v.Description(stored.source, copy.Description)
		copy.Value = duplicateJSONValue(stored.value.Value)
		result[index] = &copy
	}
	return result
}
