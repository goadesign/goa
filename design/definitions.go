package design

import (
	"time"

	"github.com/goadesign/goa/dslengine"
	regen "github.com/zach-klippenstein/goregen"
)

type (
	// AttributeDefinition defines a JSON object member with optional description, default
	// value and validations.
	AttributeDefinition struct {
		// Attribute type
		Type DataType
		// Attribute reference type if any
		Reference DataType
		// Optional description
		Description string
		// Optional validation functions
		Validations []dslengine.ValidationDefinition
		// Metadata is a list of key/value pairs
		Metadata dslengine.MetadataDefinition
		// Optional member default value
		DefaultValue interface{}
		// Optional view used to render Attribute (only applies to media type attributes).
		View string
		// List of API versions that use the attribute.
		APIVersions []string
		// NonZeroAttributes lists the names of the child attributes that cannot have a
		// zero value (and thus whose presence does not need to be validated).
		NonZeroAttributes map[string]bool
		// DSLFunc contains the initialization DSL. This is used for user types.
		DSLFunc func()
	}

	// ContainerDefinition defines a generic container definition that contains attributes.
	// This makes it possible for plugins to use attributes in their own data structures.
	ContainerDefinition interface {
		// Attribute returns the container definition embedded attribute.
		Attribute() *AttributeDefinition
	}

	// VersionIterator is the type of functions given to IterateVersions.
	VersionIterator func(v *APIVersionDefinition) error
)

// Context returns the generic definition name used in error messages.
func (a *AttributeDefinition) Context() string {
	return ""
}

// AllRequired returns the list of all required fields from the underlying object.
// An attribute type can be itself an attribute (e.g. a MediaTypeDefinition or a UserTypeDefinition)
// This happens when the DSL uses references for example. So traverse the hierarchy and collect
// all the required validations.
func (a *AttributeDefinition) AllRequired() (required []string) {
	for _, v := range a.Validations {
		if req, ok := v.(*dslengine.RequiredValidationDefinition); ok {
			required = append(required, req.Names...)
		}
	}
	if ds, ok := a.Type.(DataStructure); ok {
		required = append(required, ds.Definition().AllRequired()...)
	}
	return
}

// IsRequired returns true if the given string matches the name of a required
// attribute, false otherwise.
func (a *AttributeDefinition) IsRequired(attName string) bool {
	for _, name := range a.AllRequired() {
		if name == attName {
			return true
		}
	}
	return false
}

// AllNonZero returns the complete list of all non-zero attribute name.
func (a *AttributeDefinition) AllNonZero() []string {
	nzs := make([]string, len(a.NonZeroAttributes))
	i := 0
	for n := range a.NonZeroAttributes {
		nzs[i] = n
		i++
	}
	return nzs
}

// IsNonZero returns true if the given string matches the name of a non-zero
// attribute, false otherwise.
func (a *AttributeDefinition) IsNonZero(attName string) bool {
	return a.NonZeroAttributes[attName]
}

// IsPrimitivePointer returns true if the field generated for the given attribute should be a
// pointer to a primitive type. The target attribute must be an object.
func (a *AttributeDefinition) IsPrimitivePointer(attName string) bool {
	if !a.Type.IsObject() {
		panic("checking pointer field on non-object") // bug
	}
	att := a.Type.ToObject()[attName]
	if att == nil {
		return false
	}
	if att.Type.IsPrimitive() {
		return !a.IsRequired(attName) && !a.IsNonZero(attName)
	}
	return false
}

// Example returns a random instance of the attribute that validates.
func (a *AttributeDefinition) Example(r *RandomGenerator) interface{} {
	randomValidationLengthExample := func(count int) interface{} {
		if a.Type.IsArray() {
			res := make([]interface{}, count)
			for i := 0; i < count; i++ {
				res[i] = a.Type.ToArray().ElemType.Example(r)
			}
			return res
		}
		return r.faker.Characters(count)
	}

	randomLengthExample := func(validExample func(res float64) bool) interface{} {
		if a.Type.Kind() == IntegerKind {
			res := r.Int()
			for !validExample(float64(res)) {
				res = r.Int()
			}
			return res
		}
		res := r.Float64()
		for !validExample(res) {
			res = r.Float64()
		}
		return res
	}

	for _, v := range a.Validations {
		switch actual := v.(type) {
		case *dslengine.EnumValidationDefinition:
			count := len(actual.Values)
			i := r.Int() % count
			return actual.Values[i]
		case *dslengine.FormatValidationDefinition:
			if res, ok := map[string]interface{}{
				"email":     r.faker.Email(),
				"hostname":  r.faker.DomainName() + "." + r.faker.DomainSuffix(),
				"date-time": time.Now().Format(time.RFC3339),
				"ipv4":      r.faker.IPv4Address().String(),
				"ipv6":      r.faker.IPv6Address().String(),
				"uri":       r.faker.URL(),
				"mac": func() string {
					res, err := regen.Generate(`([0-9A-F]{2}-){5}[0-9A-F]{2}`)
					if err != nil {
						return "12-34-56-78-9A-BC"
					}
					return res
				}(),
				"cidr":   "192.168.100.14/24",
				"regexp": r.faker.Characters(3) + ".*",
			}[actual.Format]; ok {
				return res
			}
			panic("unknown format") // bug
		case *dslengine.PatternValidationDefinition:
			res, err := regen.Generate(actual.Pattern)
			if err != nil {
				return r.faker.Name()
			}
			return res
		case *dslengine.MinimumValidationDefinition:
			return randomLengthExample(func(res float64) bool {
				return res >= actual.Min
			})
		case *dslengine.MaximumValidationDefinition:
			return randomLengthExample(func(res float64) bool {
				return res <= actual.Max
			})
		case *dslengine.MinLengthValidationDefinition:
			count := actual.MinLength + (r.Int() % 3)
			return randomValidationLengthExample(count)
		case *dslengine.MaxLengthValidationDefinition:
			count := actual.MaxLength - (r.Int() % 3)
			return randomValidationLengthExample(count)
		}
	}
	return a.Type.Example(r)
}

// Merge merges the argument attributes into the target and returns the target overriding existing
// attributes with identical names.
// This only applies to attributes of type Object and Merge panics if the
// argument or the target is not of type Object.
func (a *AttributeDefinition) Merge(other *AttributeDefinition) *AttributeDefinition {
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
func (a *AttributeDefinition) Inherit(parent *AttributeDefinition) {
	if !a.shouldInherit(parent) {
		return
	}

	a.inheritValidations(parent)
	a.inheritRecursive(parent)
}

// DSL returns the initialization DSL.
func (a *AttributeDefinition) DSL() func() {
	return a.DSLFunc
}

func (a *AttributeDefinition) inheritRecursive(parent *AttributeDefinition) {
	if !a.shouldInherit(parent) {
		return
	}

	for n, att := range a.Type.ToObject() {
		if patt, ok := parent.Type.ToObject()[n]; ok {
			if att.Description == "" {
				att.Description = patt.Description
			}
			att.inheritValidations(patt)
			if att.DefaultValue == nil {
				att.DefaultValue = patt.DefaultValue
			}
			if att.View == "" {
				att.View = patt.View
			}
			if att.Type == nil {
				att.Type = patt.Type
			} else if att.shouldInherit(patt) {
				for _, att := range att.Type.ToObject() {
					att.Inherit(patt.Type.ToObject()[n])
				}
			}
		}
	}
}

func (a *AttributeDefinition) inheritValidations(parent *AttributeDefinition) {
	for _, v := range parent.Validations {
		found := false
		for _, vc := range a.Validations {
			if v == vc {
				found = true
				break
			}
		}
		if !found {
			a.Validations = append(a.Validations, parent)
		}
	}
}

func (a *AttributeDefinition) shouldInherit(parent *AttributeDefinition) bool {
	return a != nil && a.Type.ToObject() != nil &&
		parent != nil && parent.Type.ToObject() != nil
}
