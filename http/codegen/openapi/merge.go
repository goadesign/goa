package openapi

import "reflect"

// mergeItems is an internal datatype used to merge two schemas.
type mergeItems []struct {
	a, b   interface{}
	needed bool
}

// Merge does a two level deep merge of other into s.
func (s *Schema) Merge(other *Schema) {
	items := s.createMergeItems(other)
	for _, v := range items {
		if v.needed && v.b != nil {
			reflect.Indirect(reflect.ValueOf(v.a)).Set(reflect.ValueOf(v.b))
		}
	}

	for n, p := range other.Properties {
		if _, ok := s.Properties[n]; !ok {
			if s.Properties == nil {
				s.Properties = make(map[string]*Schema)
			}
			s.Properties[n] = p
		}
	}

	for n, d := range other.Definitions {
		if _, ok := s.Definitions[n]; !ok {
			s.Definitions[n] = d
		}
	}

	s.Links = append(s.Links, other.Links...)
	s.Required = append(s.Required, other.Required...)
}

func (s *Schema) createMergeItems(other *Schema) mergeItems {
	minInt := func(a, b *int) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a > *b) }
	maxInt := func(a, b *int) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a < *b) }
	minFloat64 := func(a, b *float64) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a > *b) }
	maxFloat64 := func(a, b *float64) bool { return (a == nil && b != nil) || (a != nil && b != nil && *a < *b) }

	return mergeItems{
		{&s.ID, other.ID, s.ID == ""},
		{&s.Type, other.Type, s.Type == ""},
		{&s.Ref, other.Ref, s.Ref == ""},
		{&s.Items, other.Items, s.Items == nil},
		{&s.DefaultValue, other.DefaultValue, s.DefaultValue == nil},
		{&s.Title, other.Title, s.Title == ""},
		{&s.Media, other.Media, s.Media == nil},
		{&s.ReadOnly, other.ReadOnly, !s.ReadOnly},
		{&s.PathStart, other.PathStart, s.PathStart == ""},
		{&s.Enum, other.Enum, s.Enum == nil},
		{&s.Format, other.Format, s.Format == ""},
		{&s.Pattern, other.Pattern, s.Pattern == ""},
		{&s.AdditionalProperties, other.AdditionalProperties, s.AdditionalProperties == nil},
		{&s.Minimum, other.Minimum, minFloat64(s.Minimum, other.Minimum)},
		{&s.Maximum, other.Maximum, maxFloat64(s.Maximum, other.Maximum)},
		{&s.MinLength, other.MinLength, minInt(s.MinLength, other.MinLength)},
		{&s.MaxLength, other.MaxLength, maxInt(s.MaxLength, other.MaxLength)},
		{&s.MinItems, other.MinItems, minInt(s.MinItems, other.MinItems)},
		{&s.MaxItems, other.MaxItems, maxInt(s.MaxItems, other.MaxItems)},
	}
}
