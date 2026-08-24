// This file builds the JSON schemas placed in one Swagger 2.0 document. Each
// build keeps its definitions and assigned type names in its own builder.
package openapiv2

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type (
	// schemaBuilder builds every schema used by one Swagger document.
	schemaBuilder struct {
		definitions     map[string]*openapi.Schema
		definitionNames map[*expr.ResultTypeExpr]string
		values          openapi.Values
	}
)

// newSchemaBuilder starts a schema build with no definitions or assigned names.
func newSchemaBuilder(values openapi.Values) *schemaBuilder {
	return &schemaBuilder{
		definitions:     make(map[string]*openapi.Schema),
		definitionNames: make(map[*expr.ResultTypeExpr]string),
		values:          values,
	}
}

// BuildAttributeSchema returns the JSON schema for at. The returned schema
// includes every named definition referenced by at.
func BuildAttributeSchema(api *expr.APIExpr, at *expr.AttributeExpr, generator *expr.ExampleGenerator) *openapi.Schema {
	builder := newSchemaBuilder(openapi.Values{})
	schema := builder.attributeTypeSchemaWithPrefix(api, at, "", generator)
	if len(builder.definitions) > 0 {
		schema.Defs = builder.definitions
	}
	return schema
}

// resultTypeRefWithPrefix returns a reference to the requested result view. It
// adds the definition to this builder the first time the result is used.
func (b *schemaBuilder) resultTypeRefWithPrefix(api *expr.APIExpr, mt *expr.ResultTypeExpr, view, prefix string, gen *expr.ExampleGenerator) string {
	projected, err := expr.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	return b.projectedResultTypeRefWithPrefix(api, mt, projected, prefix, gen)
}

// projectedResultTypeRefWithPrefix adds a result type that was already
// projected for one HTTP response.
func (b *schemaBuilder) projectedResultTypeRefWithPrefix(api *expr.APIExpr, source, projected *expr.ResultTypeExpr, prefix string, gen *expr.ExampleGenerator) string {
	var metaName string
	if n, ok := source.Meta["openapi:typename"]; ok {
		metaName = codegen.Goify(n[0], true)
	}
	name := projected.TypeName
	if metaName != "" {
		name = metaName
	}
	if assigned, ok := b.definitionNames[projected]; ok {
		// expr.Project can return the original result type. Reuse the name chosen
		// when this build first saw that result.
		name = assigned
	} else {
		if _, ok := b.definitions[name]; !ok {
			name = codegen.Goify(prefix, true) + codegen.Goify(name, true)
			if metaName != "" {
				name = metaName
			}
		}
		if projected == source {
			// Keep the chosen name here instead of changing the design result.
			b.definitionNames[projected] = name
		}
	}
	if _, ok := b.definitions[name]; !ok {
		b.generateResultTypeDefinition(api, renamedResultType(projected, name), expr.DefaultView, gen)
	}
	return fmt.Sprintf("#/$defs/%s", name)
}

// typeRefWithPrefix returns a reference to a user type. It adds the definition
// to this builder the first time the type is used.
func (b *schemaBuilder) typeRefWithPrefix(api *expr.APIExpr, ut *expr.UserTypeExpr, prefix string, gen *expr.ExampleGenerator) string {
	typeName := ut.TypeName
	if prefix != "" {
		typeName = codegen.Goify(prefix, true) + codegen.Goify(ut.TypeName, true)
	}
	if n, ok := ut.Meta["openapi:typename"]; ok {
		typeName = codegen.Goify(n[0], true)
	}
	if _, ok := b.definitions[typeName]; !ok {
		b.generateTypeDefinitionWithName(api, ut, typeName, gen)
	}
	return fmt.Sprintf("#/$defs/%s", typeName)
}

// generateResultTypeDefinition adds the requested result view unless this
// build already has a definition with the same name.
func (b *schemaBuilder) generateResultTypeDefinition(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string, gen *expr.ExampleGenerator) {
	if _, ok := b.definitions[mt.TypeName]; ok {
		return
	}
	schema := openapi.NewSchema()
	schema.Title = fmt.Sprintf("Mediatype identifier: %s", mt.Identifier)
	b.definitions[mt.TypeName] = schema
	b.buildResultTypeSchema(api, mt, view, schema, gen)
}

// generateTypeDefinitionWithName adds the user type under typeName unless this
// build already has a definition with that name.
func (b *schemaBuilder) generateTypeDefinitionWithName(api *expr.APIExpr, ut *expr.UserTypeExpr, typeName string, gen *expr.ExampleGenerator) {
	if _, ok := b.definitions[typeName]; ok {
		return
	}
	schema := openapi.NewSchema()
	schema.Title = typeName
	b.definitions[typeName] = schema
	b.buildAttributeSchema(api, schema, ut.AttributeExpr, gen.At(expr.UserTypeExampleIdentity(ut)))
}

// typeSchema builds a schema for t and adds any named definitions it uses to
// this builder.
func (b *schemaBuilder) typeSchema(api *expr.APIExpr, t expr.DataType, gen *expr.ExampleGenerator) *openapi.Schema {
	return b.typeSchemaWithPrefix(api, t, "", gen)
}

// typeSchemaWithPrefix builds a schema for t and adds prefix to new named
// definitions created while walking the type.
func (b *schemaBuilder) typeSchemaWithPrefix(api *expr.APIExpr, t expr.DataType, prefix string, gen *expr.ExampleGenerator) *openapi.Schema {
	schema := openapi.NewSchema()
	switch actual := t.(type) {
	case expr.Primitive:
		schema.Type = openapi.Type(actual.Name())
		switch actual.Kind() {
		case expr.AnyKind:
			// Leaving the type empty allows every JSON value.
			schema.Type = openapi.Type("")
		case expr.IntKind, expr.Int64Kind,
			expr.UIntKind, expr.UInt64Kind:
			schema.Type = openapi.Integer
			schema.Format = "int64"
		case expr.Int32Kind, expr.UInt32Kind:
			schema.Type = openapi.Integer
			schema.Format = "int32"
		case expr.Float32Kind:
			schema.Type = openapi.Number
			schema.Format = "float"
		case expr.Float64Kind:
			schema.Type = openapi.Number
			schema.Format = "double"
		case expr.BytesKind:
			schema.Type = openapi.String
			schema.Format = "byte"
		}
	case *expr.Array:
		schema.Type = openapi.Array
		schema.Items = openapi.NewSchema()
		b.buildAttributeSchema(api, schema.Items, actual.ElemType, gen.ArrayElement(0))
	case *expr.Object:
		schema.Type = openapi.Object
		for _, nat := range *actual {
			if !openapi.MustGenerate(nat.Attribute.Meta) {
				continue
			}
			property := openapi.NewSchema()
			b.buildAttributeSchema(api, property, nat.Attribute, gen.Member(nat.Name))
			schema.Properties[nat.Name] = property
		}
	case *expr.Map:
		schema.Type = openapi.Object
		if actual.KeyType.Type == expr.String && actual.ElemType.Type != expr.Any {
			value := openapi.NewSchema()
			schema.AdditionalProperties = b.buildAttributeSchema(api, value, actual.ElemType, gen.MapValue(0))
		} else {
			schema.AdditionalProperties = true
		}
	case *expr.Union:
		typeKey := actual.GetTypeKey()
		valueKey := actual.GetValueKey()
		schema.Type = openapi.Object
		for _, val := range actual.Values {
			valueSchema := b.typeSchemaWithPrefix(api, val.Attribute.Type, prefix, gen.UnionMember(val.Name))
			initSchemaValidation(valueSchema, val.Attribute)
			schema.AnyOf = append(schema.AnyOf, &openapi.Schema{
				Type: openapi.Object,
				Properties: map[string]*openapi.Schema{
					typeKey: {
						Type: openapi.String,
						Enum: []any{val.Name},
					},
					valueKey: valueSchema,
				},
				Required: []string{typeKey, valueKey},
			})
		}
	case *expr.UserTypeExpr:
		if expr.IsAlias(actual) {
			schema = b.typeSchemaWithPrefix(api, actual.Attribute().Type, prefix, gen.At(expr.UserTypeExampleIdentity(actual)))
			initSchemaValidation(schema, actual.Attribute())
			break
		}
		schema.Ref = b.typeRefWithPrefix(api, actual, prefix, gen)
	case *expr.ResultTypeExpr:
		schema.Ref = b.resultTypeRefWithPrefix(api, actual, expr.DefaultView, prefix, gen)
	}
	return schema
}

// attributeTypeSchemaWithPrefix builds a schema for at, including its
// validation rules, and adds prefix to new named definitions.
func (b *schemaBuilder) attributeTypeSchemaWithPrefix(api *expr.APIExpr, at *expr.AttributeExpr, prefix string, gen *expr.ExampleGenerator) *openapi.Schema {
	schema := b.typeSchemaWithPrefix(api, at.Type, prefix, gen)
	initSchemaValidation(schema, at)
	return schema
}

// buildAttributeSchema fills schema with the type, example, description, and
// validation rules from at.
func (b *schemaBuilder) buildAttributeSchema(api *expr.APIExpr, schema *openapi.Schema, at *expr.AttributeExpr, gen *expr.ExampleGenerator) *openapi.Schema {
	schema.Merge(b.typeSchemaWithPrefix(api, at.Type, "", gen))
	if schema.Ref != "" {
		return schema
	}
	schema.DefaultValue = openapi.ToStringMap(at.DefaultValue)
	if description := b.values.Description(at.AuthoredAttribute(), at.Description); description != "" {
		schema.Description = description
	}
	schema.Example = openapi.ProjectExample(at, b.values.Example(at, gen))
	schema.Extensions = openapi.ExtensionsFromExpr(at.Meta)
	if additional := openapi.AdditionalPropertiesFromExpr(at.Meta); additional != nil {
		schema.AdditionalProperties = additional
	}
	initSchemaValidation(schema, at)
	return schema
}

// initSchemaValidation copies the validation rules from at into schema.
func initSchemaValidation(schema *openapi.Schema, at *expr.AttributeExpr) {
	validation := at.Validation
	if validation == nil {
		return
	}
	schema.Enum = validation.Values
	if validation.Format != "" {
		schema.Format = string(validation.Format)
	}
	schema.Pattern = validation.Pattern
	if validation.ExclusiveMinimum != nil {
		schema.ExclusiveMinimum = validation.ExclusiveMinimum
	}
	if validation.Minimum != nil {
		schema.Minimum = validation.Minimum
	}
	if validation.ExclusiveMaximum != nil {
		schema.ExclusiveMaximum = validation.ExclusiveMaximum
	}
	if validation.Maximum != nil {
		schema.Maximum = validation.Maximum
	}
	if validation.MinLength != nil {
		if _, ok := at.Type.(*expr.Array); ok {
			schema.MinItems = validation.MinLength
		} else {
			schema.MinLength = validation.MinLength
		}
	}
	if validation.MaxLength != nil {
		if _, ok := at.Type.(*expr.Array); ok {
			schema.MaxItems = validation.MaxLength
		} else {
			schema.MaxLength = validation.MaxLength
		}
	}
	for _, name := range validation.Required {
		if attribute := at.Find(name); attribute != nil && !openapi.MustGenerate(attribute.Meta) {
			continue
		}
		schema.Required = append(schema.Required, name)
	}
}

// renamedResultType returns rt with name without changing the design result.
func renamedResultType(rt *expr.ResultTypeExpr, name string) *expr.ResultTypeExpr {
	if rt.TypeName == name {
		return rt
	}
	userType := *rt.UserTypeExpr
	userType.TypeName = name
	result := *rt
	result.UserTypeExpr = &userType
	return &result
}

// buildResultTypeSchema fills schema with the requested result view.
func (b *schemaBuilder) buildResultTypeSchema(api *expr.APIExpr, mt *expr.ResultTypeExpr, view string, schema *openapi.Schema, gen *expr.ExampleGenerator) {
	schema.Media = &openapi.Media{Type: mt.Identifier}
	projected, err := expr.Project(mt, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project media type %#v: %s", mt.Identifier, err)) // bug
	}
	b.buildAttributeSchema(api, schema, projected.AttributeExpr, gen.At(expr.UserTypeExampleIdentity(projected)))
}
