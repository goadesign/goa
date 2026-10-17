// This file assigns Swagger definition names. Types requesting the same name
// share a definition only when their emitted fields and schema constraints agree.
package openapiv2

import (
	"fmt"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type attributePair struct{ left, right *expr.AttributeExpr }

// assignDefinitionName records the type before its children are expanded so
// recursive children receive the same assigned name as their parent.
func (b *schemaBuilder) assignDefinitionName(requested string, typ expr.UserType) string {
	for index := 1; ; index++ {
		name := requested
		if index > 1 {
			name = fmt.Sprintf("%s%d", requested, index)
		}
		previous, exists := b.definitionShapes[name]
		if !exists {
			b.definitionShapes[name] = typ
			return name
		}
		if b.equalDefinitionTypes(previous, typ, make(map[attributePair]bool), false) {
			return name
		}
	}
}

// equalDefinitionAttributes compares formatter inputs without generated Go
// identities or generated sample values. Visited pairs stop recursive traversal.
func (b *schemaBuilder) equalDefinitionAttributes(left, right *expr.AttributeExpr, seen map[attributePair]bool) bool {
	if left == right {
		return true
	}
	pair := attributePair{left, right}
	if seen[pair] {
		return true
	}
	seen[pair] = true
	if !reflect.DeepEqual(b.definitionAttributeValues(left), b.definitionAttributeValues(right)) {
		return false
	}
	if expr.IsPrimitive(left.Type) || expr.IsPrimitive(right.Type) {
		if !expr.IsPrimitive(left.Type) || !expr.IsPrimitive(right.Type) {
			return false
		}
		lt, _ := definitionPrimitive(left.Type)
		rt, _ := definitionPrimitive(right.Type)
		return lt == rt
	}
	return b.equalDefinitionTypes(left.Type, right.Type, seen, true)
}

// definitionAttributeValues uses the same conversions as the schema formatter.
// Authored examples participate in reuse; generated samples keep the first
// definition's established value rather than creating a new definition.
func (b *schemaBuilder) definitionAttributeValues(at *expr.AttributeExpr) *openapi.Schema {
	schema := openapi.NewSchema()
	if expr.IsAlias(at.Type) {
		_, schema = definitionPrimitive(at.Type)
	}
	initSchemaValidation(schema, at)
	schema.DefaultValue = openapi.ProjectDefault(at, at.DefaultValue)
	schema.Description = b.values.Description(at.AuthoredAttribute(), at.Description)
	examples := b.values.Examples(at, at.ExtractUserExamples())
	if len(examples) > 0 {
		schema.Example = openapi.ProjectExample(at, examples[len(examples)-1].Value)
	}
	schema.Extensions = openapi.ExtensionsFromExpr(at.Meta)
	schema.AdditionalProperties = openapi.AdditionalPropertiesFromExpr(at.Meta)
	return schema
}

func (b *schemaBuilder) equalDefinitionTypes(left, right expr.DataType, seen map[attributePair]bool, compareName bool) bool {
	if expr.IsPrimitive(left) || expr.IsPrimitive(right) {
		if !expr.IsPrimitive(left) || !expr.IsPrimitive(right) {
			return false
		}
		lt, lv := definitionPrimitive(left)
		rt, rv := definitionPrimitive(right)
		return lt == rt && reflect.DeepEqual(lv, rv)
	}
	if l, ok := left.(expr.UserType); ok {
		r, ok := right.(expr.UserType)
		if !ok {
			return false
		}
		lm, leftResult := l.(*expr.ResultTypeExpr)
		rm, rightResult := r.(*expr.ResultTypeExpr)
		if leftResult != rightResult || (leftResult && lm.Identifier != rm.Identifier) {
			return false
		}
		if compareName {
			var leftName, rightName string
			if leftResult {
				leftName, rightName = lm.TypeName, rm.TypeName
			} else {
				leftName, rightName = l.(*expr.UserTypeExpr).TypeName, r.(*expr.UserTypeExpr).TypeName
			}
			if n, ok := l.Attribute().Meta["openapi:typename"]; ok {
				leftName = codegen.Goify(n[0], true)
			}
			if n, ok := r.Attribute().Meta["openapi:typename"]; ok {
				rightName = codegen.Goify(n[0], true)
			}
			if leftName != rightName {
				return false
			}
		}
		if leftResult {
			lp, err := expr.Project(lm, expr.DefaultView)
			if err != nil {
				panic(fmt.Sprintf("failed to project media type %#v: %s", lm.Identifier, err))
			}
			rp, err := expr.Project(rm, expr.DefaultView)
			if err != nil {
				panic(fmt.Sprintf("failed to project media type %#v: %s", rm.Identifier, err))
			}
			return b.equalDefinitionAttributes(lp.Attribute(), rp.Attribute(), seen)
		}
		return b.equalDefinitionAttributes(l.Attribute(), r.Attribute(), seen)
	}
	switch l := left.(type) {
	case *expr.Object:
		r, ok := right.(*expr.Object)
		if !ok {
			return false
		}
		leftFields, rightFields := visibleDefinitionFields(l), visibleDefinitionFields(r)
		if len(leftFields) != len(rightFields) {
			return false
		}
		for name, field := range leftFields {
			other, exists := rightFields[name]
			if !exists || !b.equalDefinitionAttributes(field, other, seen) {
				return false
			}
		}
		return true
	case *expr.Array:
		r, ok := right.(*expr.Array)
		return ok && b.equalDefinitionAttributes(l.ElemType, r.ElemType, seen)
	case *expr.Map:
		r, ok := right.(*expr.Map)
		if !ok {
			return false
		}
		leftTyped, rightTyped := typedDefinitionMap(l), typedDefinitionMap(r)
		return leftTyped == rightTyped && (!leftTyped || b.equalDefinitionAttributes(l.ElemType, r.ElemType, seen))
	case *expr.Union:
		r, ok := right.(*expr.Union)
		if !ok || l.Untagged != r.Untagged || len(l.Values) != len(r.Values) {
			return false
		}
		if !l.Untagged && (l.Flatten != r.Flatten || l.GetTypeKey() != r.GetTypeKey() || (!l.Flatten && l.GetValueKey() != r.GetValueKey())) {
			return false
		}
		for i, field := range l.Values {
			other := r.Values[i]
			if field.Name != other.Name || !b.equalDefinitionAttributes(field.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	}
	return false
}

// visibleDefinitionFields indexes only properties emitted in the document.
func visibleDefinitionFields(object *expr.Object) map[string]*expr.AttributeExpr {
	fields := make(map[string]*expr.AttributeExpr)
	for _, field := range *object {
		if !openapi.MustGenerate(field.Attribute.Meta) {
			continue
		}
		names := strings.Split(field.Name, ":")
		name := names[0]
		if len(names) > 1 {
			name = names[1]
		}
		fields[name] = field.Attribute
	}
	return fields
}

// typedDefinitionMap reports whether the formatter emits a schema for map values.
func typedDefinitionMap(m *expr.Map) bool {
	key := m.KeyType.Type
	for expr.IsAlias(key) {
		key = key.(expr.UserType).Attribute().Type
	}
	return key == expr.String && m.ElemType.Type != expr.Any
}

// definitionPrimitive collects the validation that the formatter places on an
// inlined primitive alias, without making its Go type name part of the schema.
func definitionPrimitive(t expr.DataType) (expr.DataType, *openapi.Schema) {
	if !expr.IsAlias(t) {
		return t, openapi.NewSchema()
	}
	at := t.(expr.UserType).Attribute()
	primitive, schema := definitionPrimitive(at.Type)
	initSchemaValidation(schema, at)
	return primitive, schema
}
