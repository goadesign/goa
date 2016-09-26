package design

import (
	"fmt"

	"github.com/goadesign/goa/eval"
)

type (
	// AttributeExpr defines a object field with optional description, default value and
	// validations.
	AttributeExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Attribute type
		Type DataType
		// Attribute reference type if any
		Reference DataType
		// Optional description
		Description string
		// Optional validations
		Validation *ValidationExpr
		// Metadata is a list of key/value pairs
		Metadata MetadataExpr
		// Optional member default value
		DefaultValue interface{}
		// UserExample set in DSL or computed in Finalize
		UserExample interface{}
	}

	// CompositeExpr defines a generic composite expression that contains an attribute.
	// This makes it possible for plugins to use attributes in their own data structures.
	CompositeExpr interface {
		// Attribute returns the composite expression embedded attribute.
		Attribute() *AttributeExpr
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

// EvalName returns the name used by the DSL evaluation.
func (a *AttributeExpr) EvalName() string {
	return "attribute"
}

// validated keeps track of validated attributes to handle cyclical definitions.
var validated = make(map[*AttributeExpr]bool)

// Validate tests whether the attribute required fields exist.
// Since attributes are unaware of their context, additional context information can be provided
// to be used in error messages.
// The parent definition context is automatically added to error messages.
func (a *AttributeExpr) Validate(ctx string, parent eval.Expression) *eval.ValidationErrors {
	if validated[a] {
		return nil
	}
	validated[a] = true
	verr := new(eval.ValidationErrors)
	if a.Type == nil {
		verr.Add(parent, "attribute type is nil")
		return verr
	}
	if ctx != "" {
		ctx += " - "
	}
	// If both Default and Enum are given, make sure the Default value is one of Enum values.
	// TODO: We only do the default value and enum check just for primitive types.
	if _, ok := a.Type.(Primitive); ok {
		if a.DefaultValue != nil && a.Validation != nil && a.Validation.Values != nil {
			var found bool
			for _, e := range a.Validation.Values {
				if e == a.DefaultValue {
					found = true
					break
				}
			}
			if !found {
				verr.Add(parent, "%sdefault value %#v is not one of the accepted values: %#v", ctx, a.DefaultValue, a.Validation.Values)
			}
		}
	}
	o := a.Type.(Object)
	if o != nil {
		for _, n := range a.AllRequired() {
			found := false
			for an := range o {
				if n == an {
					found = true
					break
				}
			}
			if !found {
				verr.Add(parent, `%srequired field "%s" does not exist`, ctx, n)
			}
		}
		for n, att := range o {
			ctx = fmt.Sprintf("field %s", n)
			verr.Merge(att.Validate(ctx, parent))
		}
	} else {
		if ar, ok := a.Type.(*Array); ok {
			elemType := ar.ElemType
			verr.Merge(elemType.Validate(ctx, a))
		}
	}

	return verr
}

// Merge merges the argument attributes into the target and returns the target overriding existing
// attributes with identical names.
// This only applies to attributes of type Object and Merge panics if the
// argument or the target is not of type Object.
func (a *AttributeExpr) Merge(other *AttributeExpr) *AttributeExpr {
	if other == nil {
		return a
	}
	if a == nil {
		return other
	}
	left := a.Type.(Object)
	right := other.Type.(Object)
	if left == nil || right == nil {
		panic("cannot merge non object attributes") // bug
	}
	for n, v := range right {
		left[n] = v
	}
	return a
}

// Inherit merges the properties of existing target type attributes with the argument's.
// The algorithm is recursive so that child attributes are also merged.
func (a *AttributeExpr) Inherit(parent *AttributeExpr) {
	if !a.shouldInherit(parent) {
		return
	}

	a.inheritValidations(parent)
	a.inheritRecursive(parent)
}

// AllRequired returns the list of all required fields from the underlying object.  This method
// recurses if the type is itself an attribute (i.e. a UserType, this happens with the Reference DSL
// for example).
func (a *AttributeExpr) AllRequired() (required []string) {
	if a == nil || a.Validation == nil {
		return
	}
	required = a.Validation.Required
	if u, ok := a.Type.(UserType); ok {
		required = append(required, u.Attribute().AllRequired()...)
	}
	return
}

// IsRequired returns true if the given string matches the name of a required attribute, false
// otherwise. This method only applies to attributes of type Object.
func (a *AttributeExpr) IsRequired(attName string) bool {
	for _, name := range a.AllRequired() {
		if name == attName {
			return true
		}
	}
	return false
}

// HasDefaultValue returns true if the given attribute has a default value.
func (a *AttributeExpr) HasDefaultValue(attName string) bool {
	if o, ok := a.Type.(Object); ok {
		att := o[attName]
		return att.DefaultValue != nil
	}
	return false
}

// SetDefault sets the default for the attribute. It also converts HashVal
// and ArrayVal to map and slice respectively.
func (a *AttributeExpr) SetDefault(def interface{}) {
	switch actual := def.(type) {
	case MapVal:
		a.DefaultValue = actual.ToMap()
	case ArrayVal:
		a.DefaultValue = actual.ToSlice()
	default:
		a.DefaultValue = actual
	}
}

func (a *AttributeExpr) inheritRecursive(parent *AttributeExpr) {
	if !a.shouldInherit(parent) {
		return
	}

	for n, att := range a.Type.(Object) {
		if patt, ok := parent.Type.(Object)[n]; ok {
			if att.Description == "" {
				att.Description = patt.Description
			}
			att.inheritValidations(patt)
			if att.DefaultValue == nil {
				att.DefaultValue = patt.DefaultValue
			}
			if att.Type == nil {
				att.Type = patt.Type
			} else if att.shouldInherit(patt) {
				for _, att := range att.Type.(Object) {
					att.Inherit(patt.Type.(Object)[n])
				}
			}
		}
	}
}

func (a *AttributeExpr) inheritValidations(parent *AttributeExpr) {
	if parent.Validation == nil {
		return
	}
	if a.Validation == nil {
		a.Validation = &ValidationExpr{}
	}
	a.Validation.AddRequired(parent.Validation.Required)
}

func (a *AttributeExpr) shouldInherit(parent *AttributeExpr) bool {
	if a == nil || parent == nil {
		return false
	}
	if _, ok := a.Type.(Object); ok {
		if _, ok := parent.Type.(Object); ok {
			return true
		}
	}
	return false
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
	walkUt := func(ut UserType) error {
		if _, ok := seen[ut.Name()]; ok {
			return nil
		}
		seen[ut.Name()] = true
		return walk(ut.Attribute(), walker, seen)
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
	case UserType:
		return walkUt(actual)
	default:
		panic("unknown field type") // bug
	}
	return nil
}

// Context returns the generic definition name used in error messages.
func (v *ValidationExpr) Context() string {
	return "validation"
}

// Merge merges other into v.
func (v *ValidationExpr) Merge(other *ValidationExpr) {
	if v.Values == nil {
		v.Values = other.Values
	}
	if v.Format == "" {
		v.Format = other.Format
	}
	if v.Pattern == "" {
		v.Pattern = other.Pattern
	}
	if v.Minimum == nil || (other.Minimum != nil && *v.Minimum > *other.Minimum) {
		v.Minimum = other.Minimum
	}
	if v.Maximum == nil || (other.Maximum != nil && *v.Maximum < *other.Maximum) {
		v.Maximum = other.Maximum
	}
	if v.MinLength == nil || (other.MinLength != nil && *v.MinLength > *other.MinLength) {
		v.MinLength = other.MinLength
	}
	if v.MaxLength == nil || (other.MaxLength != nil && *v.MaxLength < *other.MaxLength) {
		v.MaxLength = other.MaxLength
	}
	v.AddRequired(other.Required)
}

// AddRequired merges the required fields from other into v
func (v *ValidationExpr) AddRequired(required []string) {
	for _, r := range required {
		found := false
		for _, rr := range v.Required {
			if r == rr {
				found = true
				break
			}
		}
		if !found {
			v.Required = append(v.Required, r)
		}
	}
}

// HasRequiredOnly returns true if the validation only has the Required field with a non-zero value.
func (v *ValidationExpr) HasRequiredOnly() bool {
	if len(v.Values) > 0 {
		return false
	}
	if v.Format != "" || v.Pattern != "" {
		return false
	}
	if (v.Minimum != nil) || (v.Maximum != nil) || (v.MaxLength != nil) {
		return false
	}
	return true
}

// Dup makes a shallow dup of the validation.
func (v *ValidationExpr) Dup() *ValidationExpr {
	return &ValidationExpr{
		Values:    v.Values,
		Format:    v.Format,
		Pattern:   v.Pattern,
		Minimum:   v.Minimum,
		Maximum:   v.Maximum,
		MinLength: v.MinLength,
		MaxLength: v.MaxLength,
		Required:  v.Required,
	}
}
