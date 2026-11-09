// This file decides whether an attribute or user type emits validation checks
// and which runtime imports a standalone validator body requires.
package codegen

import (
	"goa.design/goa/v3/expr"
)

// NeedsValidation reports whether Goa would generate at least one validation
// check for attribute with the given Go field layout.
func NeedsValidation(attribute *expr.AttributeExpr, policy GoLayoutPolicy) bool {
	return attributeNeedsValidation(attribute, policy, false, make(map[expr.UserType]struct{}))
}

// ValidationRuntimeImports returns the Goa and standard library packages used
// by validation code for attribute. Calls to validators in other generated
// packages are recorded by ValidationPlan.ImportPreferences instead.
func ValidationRuntimeImports(attribute *expr.AttributeExpr, policy GoLayoutPolicy) []GoTypeImport {
	if !NeedsValidation(attribute, policy) {
		return nil
	}
	var imports []GoTypeImport
	if validationBodyUsesUTF8(attribute, false, false, make(map[expr.UserType]struct{})) {
		imports = append(imports, GoTypeImport{Path: "unicode/utf8"})
	}
	imports = append(imports, GoTypeImport{Name: "goa", Path: GoaImport("").Path})
	return imports
}

// userTypeNeedsValidation reports whether a user-defined type or any value
// inside it needs a generated check. viewSelected reports whether userType
// already contains only the fields of a selected result view; seen stops
// recursive types.
func userTypeNeedsValidation(
	userType expr.UserType,
	policy GoLayoutPolicy,
	viewSelected bool,
	seen map[expr.UserType]struct{},
) bool {
	origin := userType.Origin()
	if _, exists := seen[origin]; exists {
		return false
	}
	seen[origin] = struct{}{}
	defer delete(seen, origin)
	return attributeNeedsValidation(userType.Attribute(), policy, viewSelected, seen)
}

// attributeNeedsValidation reports whether Goa would generate a check for the
// attribute or a value inside it. When viewSelected is false and attribute
// names a result type with an explicit view selection, only the fields in that
// selected view (and their nested views) are inspected.
func attributeNeedsValidation(
	attribute *expr.AttributeExpr,
	policy GoLayoutPolicy,
	viewSelected bool,
	seen map[expr.UserType]struct{},
) bool {
	validation := expr.EffectiveValidation(attribute)
	if validation != nil {
		if len(validation.Values) > 0 || validation.Pattern != "" ||
			validation.ExclusiveMinimum != nil || validation.Minimum != nil ||
			validation.ExclusiveMaximum != nil || validation.Maximum != nil ||
			validation.MinLength != nil || validation.MaxLength != nil {
			return true
		}
		if validation.Format != "" {
			if custom, _ := GetMetaType(attribute); custom == "" {
				return true
			}
		}
		if len(generatedRequiredValidationNames(attribute, validation, policy)) > 0 {
			return true
		}
	}
	if nested, ok := attribute.Type.(expr.UserType); ok && !expr.IsAlias(nested) {
		if !viewSelected {
			if result, ok := nested.(*expr.ResultTypeExpr); ok {
				if selected, ok := attribute.Meta.Last(expr.ViewMetaKey); ok {
					if narrowed, err := expr.Project(result, selected); err == nil {
						return userTypeNeedsValidation(narrowed, policy, true, seen)
					}
				}
			}
		}
		return userTypeNeedsValidation(nested, policy, viewSelected, seen)
	}
	switch {
	case expr.IsObject(attribute.Type):
		for _, field := range *expr.AsObject(attribute.Type) {
			if attributeNeedsValidation(field.Attribute, policy, viewSelected, seen) {
				return true
			}
		}
	case expr.IsArray(attribute.Type):
		array := expr.AsArray(attribute.Type)
		if array.NonNullableElems &&
			(IsNilable(array.ElemType.Type) || arrayElementIsPointer(array, policy.ArrayElementPointer)) {
			return true
		}
		return attributeNeedsValidation(array.ElemType, policy, viewSelected, seen)
	case expr.IsMap(attribute.Type):
		mapping := expr.AsMap(attribute.Type)
		keyPolicy := policy
		keyPolicy.Pointer = false
		valuePolicy := policy
		if expr.IsPrimitive(mapping.ElemType.Type) {
			valuePolicy.Pointer = false
		}
		return attributeNeedsValidation(mapping.KeyType, keyPolicy, viewSelected, seen) ||
			attributeNeedsValidation(mapping.ElemType, valuePolicy, viewSelected, seen)
	case expr.IsUnion(attribute.Type):
		if policy.SumType {
			return true
		}
		for _, branch := range expr.AsUnion(attribute.Type).Values {
			branchPolicy := policy
			branchPolicy.Pointer = policy.Pointer && expr.IsObject(branch.Attribute.Type)
			if attributeNeedsValidation(branch.Attribute, branchPolicy, viewSelected, seen) {
				return true
			}
		}
	}
	return false
}

// validationBodyUsesUTF8 reports whether one emitted validator counts string
// characters. A nested named object has its own validator, so the caller does
// not reserve packages used only by that separate function.
func validationBodyUsesUTF8(
	attribute *expr.AttributeExpr,
	nested, alias bool,
	seen map[expr.UserType]struct{},
) bool {
	if attribute == nil || attribute.Type == nil || attribute.Type == expr.Empty {
		return false
	}
	validation := expr.EffectiveValidation(attribute)
	if isStringDataType(attribute.Type) && validation != nil &&
		(validation.MinLength != nil || validation.MaxLength != nil) {
		return true
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if nested && !alias {
			return false
		}
		origin := actual.Origin()
		if _, exists := seen[origin]; exists {
			return false
		}
		seen[origin] = struct{}{}
		defer delete(seen, origin)
		return validationBodyUsesUTF8(actual.Attribute(), nested, alias, seen)
	case *expr.Object:
		for _, field := range *actual {
			if validationBodyUsesUTF8(field.Attribute, true, expr.IsAlias(field.Attribute.Type), seen) {
				return true
			}
		}
	case *expr.Array:
		return validationBodyUsesUTF8(actual.ElemType, true, expr.IsAlias(actual.ElemType.Type), seen)
	case *expr.Map:
		return validationBodyUsesUTF8(actual.KeyType, true, expr.IsAlias(actual.KeyType.Type), seen) ||
			validationBodyUsesUTF8(actual.ElemType, true, expr.IsAlias(actual.ElemType.Type), seen)
	case *expr.Union:
		for _, branch := range actual.Values {
			if validationBodyUsesUTF8(branch.Attribute, true, expr.IsAlias(branch.Attribute.Type), seen) {
				return true
			}
		}
	}
	return false
}

// isStringDataType reports whether dataType is a string or a named string.
func isStringDataType(dataType expr.DataType) bool {
	for {
		switch actual := dataType.(type) {
		case expr.Primitive:
			return actual == expr.String
		case expr.UserType:
			dataType = actual.Attribute().Type
		default:
			return false
		}
	}
}
