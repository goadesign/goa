// This file verifies that copied schemas share no mutable values with their
// source. Plugins may safely edit a copy without changing Goa's planned schema.
package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaDupCopiesEveryMutableField(t *testing.T) {
	exclusiveMinimum := 1.0
	minimum := 2.0
	exclusiveMaximum := 9.0
	maximum := 10.0
	minLength := 1
	maxLength := 8
	minItems := 2
	maxItems := 4
	original := &Schema{
		Schema:               "schema",
		ID:                   "id",
		Title:                "title",
		Type:                 Object,
		Items:                &Schema{Title: "items"},
		Properties:           map[string]*Schema{"property": {Title: "property"}},
		Defs:                 map[string]*Schema{"definition": {Title: "definition"}},
		Description:          "description",
		DefaultValue:         map[string][]string{"values": {"default"}},
		Example:              []map[string]any{{"value": "example"}},
		Media:                &Media{BinaryEncoding: "binary", Type: "media"},
		ReadOnly:             true,
		PathStart:            "/path",
		Links:                []*Link{{Title: "link", Schema: &Schema{Title: "link schema"}, TargetSchema: &Schema{Title: "target schema"}}},
		Ref:                  "#/$defs/reference",
		Enum:                 []any{[]string{"enum"}},
		Format:               "format",
		Pattern:              "pattern",
		ExclusiveMinimum:     &exclusiveMinimum,
		Minimum:              &minimum,
		ExclusiveMaximum:     &exclusiveMaximum,
		Maximum:              &maximum,
		MinLength:            &minLength,
		MaxLength:            &maxLength,
		MinItems:             &minItems,
		MaxItems:             &maxItems,
		Required:             []string{"required"},
		AdditionalProperties: &Schema{Title: "additional properties"},
		ContentMediaType:     "application/json",
		ContentSchema:        &Schema{Title: "content schema"},
		AnyOf:                []*Schema{{Title: "union branch"}},
		Extensions:           map[string]any{"x-values": []string{"extension"}},
	}

	duplicate := original.Dup()
	require.Equal(t, original, duplicate)

	duplicate.Items.Title = "changed"
	duplicate.Properties["property"].Title = "changed"
	duplicate.Defs["definition"].Title = "changed"
	duplicate.DefaultValue.(map[string][]string)["values"][0] = "changed"
	duplicate.Example.([]map[string]any)[0]["value"] = "changed"
	duplicate.Media.Type = "changed"
	duplicate.Links[0].Title = "changed"
	duplicate.Links[0].Schema.Title = "changed"
	duplicate.Links[0].TargetSchema.Title = "changed"
	duplicate.Enum[0].([]string)[0] = "changed"
	*duplicate.ExclusiveMinimum = 3
	*duplicate.Minimum = 4
	*duplicate.ExclusiveMaximum = 7
	*duplicate.Maximum = 8
	*duplicate.MinLength = 2
	*duplicate.MaxLength = 7
	*duplicate.MinItems = 1
	*duplicate.MaxItems = 3
	duplicate.Required[0] = "changed"
	duplicate.AdditionalProperties.(*Schema).Title = "changed"
	duplicate.ContentSchema.Title = "changed"
	duplicate.AnyOf[0].Title = "changed"
	duplicate.Extensions["x-values"].([]string)[0] = "changed"

	require.Equal(t, "items", original.Items.Title)
	require.Equal(t, "property", original.Properties["property"].Title)
	require.Equal(t, "definition", original.Defs["definition"].Title)
	require.Equal(t, "default", original.DefaultValue.(map[string][]string)["values"][0])
	require.Equal(t, "example", original.Example.([]map[string]any)[0]["value"])
	require.Equal(t, "media", original.Media.Type)
	require.Equal(t, "link", original.Links[0].Title)
	require.Equal(t, "link schema", original.Links[0].Schema.Title)
	require.Equal(t, "target schema", original.Links[0].TargetSchema.Title)
	require.Equal(t, "enum", original.Enum[0].([]string)[0])
	require.Equal(t, 1.0, *original.ExclusiveMinimum)
	require.Equal(t, 2.0, *original.Minimum)
	require.Equal(t, 9.0, *original.ExclusiveMaximum)
	require.Equal(t, 10.0, *original.Maximum)
	require.Equal(t, 1, *original.MinLength)
	require.Equal(t, 8, *original.MaxLength)
	require.Equal(t, 2, *original.MinItems)
	require.Equal(t, 4, *original.MaxItems)
	require.Equal(t, "required", original.Required[0])
	require.Equal(t, "additional properties", original.AdditionalProperties.(*Schema).Title)
	require.Equal(t, "content schema", original.ContentSchema.Title)
	require.Equal(t, "union branch", original.AnyOf[0].Title)
	require.Equal(t, "extension", original.Extensions["x-values"].([]string)[0])
}
