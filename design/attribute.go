package design

type (
	// AttributeExpr defines a object field with optional description, default value and
	// validations.
	AttributeExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		*eval.DSLFunc
		// Attribute type
		Type DataType
		// Attribute reference type if any
		Reference DataType
		// Optional description
		Description string
		// Optional validations
		Validation *ValidationExpr
		// Metadata is a list of key/value pairs
		Metadata *MetadataExpr
		// Optional member default value
		DefaultValue interface{}
		// Optional member example value
		Example interface{}
	}

	// ValidationExpr contains validation rules for an attribute.
	ValidationExpr struct {
		// Values represents an enum validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor76.
		Values []interface{}
		// Format represents a format validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor104.
		Format string
		// PatternValidationExpr represents a pattern validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor33
		Pattern string
		// Minimum represents an minimum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor21.
		Minimum *float64
		// Maximum represents a maximum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor17.
		Maximum *float64
		// MinLength represents an minimum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor29.
		MinLength *int
		// MaxLength represents an maximum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor26.
		MaxLength *int
		// Required list the required fields of object attributes as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor61.
		Required []string
	}
)

// DataStructure implementation

// Expr returns the field definition.
func (a *AttributeExpr) Expr() *AttributeExpr {
	return a
}

// Walk traverses the data structure recursively and calls the given function once
// on each field starting with the field returned by Expr.
func (a *AttributeExpr) Walk(walker func(*AttributeExpr) error) error {
	return walk(a, walker, make(map[string]bool))
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite recursions by
// keeping track of types that have already been walked.
func walk(at *AttributeExpr, walker func(*AttributeExpr) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut *UserTypeExpr) error {
		if _, ok := seen[ut.TypeName]; ok {
			return nil
		}
		seen[ut.TypeName] = true
		return walk(ut.AttributeExpr, walker, seen)
	}
	switch actual := at.Type.(type) {
	case Primitive:
		return nil
	case *Array:
		return walk(actual.ElemType, walker, seen)
	case *Map:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case Object:
		for _, cat := range actual {
			if err := walk(cat, walker, seen); err != nil {
				return err
			}
		}
	case *UserTypeExpr:
		return walkUt(actual)
	case *MediaTypeExpr:
		return walkUt(actual.UserTypeExpr)
	default:
		panic("unknown field type") // bug
	}
	return nil
}
