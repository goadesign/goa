package design

import (
	"fmt"

	"goa.design/goa/eval"
)

type (
	// AttributeExpr defines a object field with optional description,
	// default value and validations.
	AttributeExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Attribute type
		Type DataType
		// Base types if any
		Bases []DataType
		// Attribute reference types if any
		References []DataType
		// Optional description
		Description string
		// Docs points to external documentation
		Docs *DocsExpr
		// Optional validations
		Validation *ValidationExpr
		// Metadata is a list of key/value pairs
		Metadata MetadataExpr
		// Optional member default value
		DefaultValue interface{}
		// UserExample set in DSL or computed in Finalize
		UserExamples []*ExampleExpr
	}

	// ExampleExpr represents an example.
	ExampleExpr struct {
		// Summary is the example short summary.
		Summary string
		// Description is an optional long description.
		Description string
		// Value is the example value.
		Value interface{}
	}

	// Val is the type used to provide the value of examples for attributes that are
	// objects.
	Val map[string]interface{}

	// CompositeExpr defines a generic composite expression that contains an
	// attribute.  This makes it possible for plugins to use attributes in
	// their own data structures.
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
		Format ValidationFormat
		// PatternValidationExpr represents a pattern validation as
		// described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor33
		Pattern string
		// Minimum represents an minimum value validation as described
		// at
		// http://json-schema.org/latest/json-schema-validation.html#anchor21.
		Minimum *float64
		// Maximum represents a maximum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor17.
		Maximum *float64
		// MinLength represents an minimum length validation as
		// described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor29.
		MinLength *int
		// MaxLength represents an maximum length validation as
		// described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor26.
		MaxLength *int
		// Required list the required fields of object attributes as
		// described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor61.
		Required []string
	}

	// ValidationFormat is the type used to enumerates the possible string
	// formats.
	ValidationFormat string
)

const (
	// FormatCIDR designates
	FormatCIDR ValidationFormat = "cidr"

	// FormatDateTime designates values that follow RFC3339
	FormatDateTime = "date-time"

	// FormatEmail designates values that follow RFC5322
	FormatEmail = "email"

	// FormatHostname designates
	FormatHostname = "hostname"

	// FormatIPv4 designates values that follow RFC2373 IPv4
	FormatIPv4 = "ipv4"

	// FormatIPv6 designates values that follow RFC2373 IPv6
	FormatIPv6 = "ipv6"

	// FormatIP designates values that follow RFC2373 IPv4 or IPv6
	FormatIP = "ip"

	// FormatMAC designates
	FormatMAC = "mac"

	// FormatRegexp designates
	FormatRegexp = "regexp"

	// FormatURI designates
	FormatURI = "uri"
)

// EvalName returns the name used by the DSL evaluation.
func (a *AttributeExpr) EvalName() string {
	return "attribute"
}

// validated keeps track of validated attributes to handle cyclical definitions.
var validated = make(map[*AttributeExpr]bool)

// Validate tests whether the attribute required fields exist.  Since attributes
// are unaware of their context, additional context information can be provided
// to be used in error messages.  The parent definition context is automatically
// added to error messages.
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
	verr.Merge(a.validateEnumDefault(ctx, parent))
	if o := AsObject(a.Type); o != nil {
		for _, n := range a.AllRequired() {
			if o.Attribute(n) == nil {
				verr.Add(parent, `%srequired field %q does not exist`, ctx, n)
			}
		}
		for _, nat := range *o {
			ctx = fmt.Sprintf("field %s", nat.Name)
			verr.Merge(nat.Attribute.Validate(ctx, parent))
		}
	} else {
		if ar := AsArray(a.Type); ar != nil {
			elemType := ar.ElemType
			verr.Merge(elemType.Validate(ctx, a))
		}
	}

	if views, ok := a.Metadata["view"]; ok {
		rt, ok := a.Type.(*ResultTypeExpr)
		if !ok {
			verr.Add(parent, "%sdefines a view but is not a result type", ctx)
		}
		if rt != nil {
			found := false
			name := views[0]
			for _, v := range rt.Views {
				if v.Name == name {
					found = true
					break
				}
			}
			if !found {
				verr.Add(parent, "%stype does not define view %q", ctx, name)
			}
		}
	}

	return verr
}

// Merge merges other's attributes into a overriding attributes of a with
// attributes of other with identical names.
//
// This only applies to attributes of type Object and Merge panics if the
// argument or the target is not of type Object.
func (a *AttributeExpr) Merge(other *AttributeExpr) {
	if other == nil {
		return
	}
	left := AsObject(a.Type)
	right := AsObject(other.Type)
	if left == nil || right == nil {
		panic("cannot merge non object attributes") // bug
	}
	if other.Validation != nil {
		if a.Validation == nil {
			a.Validation = other.Validation.Dup()
		} else {
			a.Validation.Merge(other.Validation)
		}
	}
	for _, nat := range *right {
		left.Set(nat.Name, nat.Attribute)
	}
}

// Inherit merges the properties of existing target type attributes with the
// argument's.  The algorithm is recursive so that child attributes are also
// merged.
func (a *AttributeExpr) Inherit(parent *AttributeExpr) {
	if !a.shouldInherit(parent) {
		return
	}

	a.inheritValidations(parent)
	a.inheritRecursive(parent)
}

// AllRequired returns the list of all required fields from the underlying
// object. This method recurses if the type is itself an attribute (i.e. a
// UserType, this happens with the Reference DSL for example).
func (a *AttributeExpr) AllRequired() []string {
	if u, ok := a.Type.(UserType); ok {
		return u.Attribute().AllRequired()
	}
	if a.Validation != nil {
		return a.Validation.Required
	}
	return nil
}

// IsRequired returns true if the given string matches the name of a required
// attribute, false otherwise. This method only applies to attributes of type
// Object.
func (a *AttributeExpr) IsRequired(attName string) bool {
	for _, name := range a.AllRequired() {
		if name == attName {
			return true
		}
	}
	return false
}

// IsRequiredNoDefault returns true if the given string matches the name of a
// required attribute and the attribute has no default value, false otherwise.
// This method only applies to attributes of type Object.
func (a *AttributeExpr) IsRequiredNoDefault(attName string) bool {
	for _, name := range a.AllRequired() {
		if name == attName {
			return AsObject(a.Type).Attribute(name).DefaultValue == nil
		}
	}
	return false
}

// IsPrimitivePointer returns true if the field generated for the given
// attribute should be a pointer to a primitive type. The receiver attribute must
// be an object.
//
// If useDefault is true and the attribute has a default value then
// IsPrimitivePointer returns false. This makes it possible to differentiate
// between request types where attributes with default values should not be
// generated using a pointer value and response types where they should.
//
//    DefaultValue UseDefault Pointer (assuming all other conditions are true)
//    Yes          True       False
//    Yes          False      True
//    No           True       True
//    No           False      True
//
func (a *AttributeExpr) IsPrimitivePointer(attName string, useDefault bool) bool {
	o := AsObject(a.Type)
	if o == nil {
		panic("checking pointer field on non-object") // bug
	}
	att := o.Attribute(attName)
	if att == nil {
		return false
	}
	if IsPrimitive(att.Type) {
		return att.Type.Kind() != BytesKind && att.Type.Kind() != AnyKind &&
			!a.IsRequired(attName) && (!a.HasDefaultValue(attName) || !useDefault)
	}
	return false
}

// HasDefaultValue returns true if the attribute with the given name has a
// default value.
func (a *AttributeExpr) HasDefaultValue(attName string) bool {
	if o := AsObject(a.Type); o != nil {
		return o.Attribute(attName).DefaultValue != nil
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

// validateEnumDefault makes sure that the attribute default value is one of the
// enum values.
func (a *AttributeExpr) validateEnumDefault(ctx string, parent eval.Expression) *eval.ValidationErrors {
	//TODO: We only do the default value and enum check just for primitive types.
	if _, ok := a.Type.(Primitive); !ok {
		return nil
	}
	verr := new(eval.ValidationErrors)
	if a.DefaultValue != nil && a.Validation != nil && a.Validation.Values != nil {
		var found bool
		for _, e := range a.Validation.Values {
			if e == a.DefaultValue {
				found = true
				break
			}
		}
		if !found {
			verr.Add(
				parent,
				"%sdefault value %#v is not one of the accepted values: %#v",
				ctx,
				a.DefaultValue,
				a.Validation.Values,
			)
		}
	}
	return verr
}

func (a *AttributeExpr) inheritRecursive(parent *AttributeExpr) {
	if !a.shouldInherit(parent) {
		return
	}
	for _, nat := range *AsObject(a.Type) {
		if patt := AsObject(parent.Type).Attribute(nat.Name); patt != nil {
			att := nat.Attribute
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
				for _, nat := range *AsObject(att.Type) {
					nat.Attribute.Inherit(AsObject(patt.Type).Attribute(nat.Name))
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
	a.Validation.AddRequired(parent.Validation.Required...)
}

func (a *AttributeExpr) shouldInherit(parent *AttributeExpr) bool {
	return a != nil && AsObject(a.Type) != nil &&
		parent != nil && AsObject(parent.Type) != nil
}

// EvalName returns the name used by the DSL evaluation.
func (a *ExampleExpr) EvalName() string {
	return `example "` + a.Summary + `"`
}

// Context returns the generic definition name used in error messages.
func (v *ValidationExpr) Context() string {
	return "validation"
}

// Validate makes sure the validation format value is valid.
func (v *ValidationExpr) Validate() error {
	verr := new(eval.ValidationErrors)

	if v.Format == "" {
		return nil
	}
	return verr
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
	v.AddRequired(other.Required...)
}

// AddRequired merges the required fields from other into v
func (v *ValidationExpr) AddRequired(required ...string) {
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

// RemoveRequired removes the given field from the list of required fields
func (v *ValidationExpr) RemoveRequired(required string) {
	for i, r := range v.Required {
		if required == r {
			v.Required = append(v.Required[:i], v.Required[i+1:]...)
			break
		}
	}
}

// HasRequiredOnly returns true if the validation only has the Required field
// with a non-zero value.
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
	var req []string
	if len(v.Required) > 0 {
		req = make([]string, len(v.Required))
		for i, r := range v.Required {
			req[i] = r
		}
	}
	return &ValidationExpr{
		Values:    v.Values,
		Format:    v.Format,
		Pattern:   v.Pattern,
		Minimum:   v.Minimum,
		Maximum:   v.Maximum,
		MinLength: v.MinLength,
		MaxLength: v.MaxLength,
		Required:  req,
	}
}
