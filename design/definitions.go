package design

import (
	"fmt"
	"time"

	regen "github.com/zach-klippenstein/goregen"
)

// Roots contains the root definition sets built by the DSLs.
// DSL implementations should append to it to ensure the DSL gets executed by the runner.
// Note that a root definition is a different concept from a "top level" definition (i.e. a
// definition that is an entry point in the DSL). In particular a root definition may include
// an arbitrary number of definition sets forming a tree of definitions.
// For example the API DSL only has one root definition (the API definition) but many top level
// definitions (API, Version, Type, MediaType etc.) all defining a definition set.
var Roots []Root

type (
	// Definition is the common interface implemented by all definitions.
	Definition interface {
		// Context is used to build error messages that refer to the definition.
		Context() string
	}

	// DefinitionSet contains DSL definitions that are executed as one unit.
	// The slice elements may implement the Validate an, Source interfaces to enable the
	// corresponding behaviors during DSL execution.
	DefinitionSet []Definition

	// Root is the interface implemented by the DSL root objects held by the Roots variable.
	// These objects contains all the definition sets created by the DSL and can be passed to
	// the engine for execution.
	Root interface {
		// IterateSets calls the given iterator passing in each definition set sorted in
		// execution order.
		IterateSets(SetIterator)
	}

	// Validate is the interface implemented by definitions that can be validated.
	// Validation is done by the DSL engine post execution.
	Validate interface {
		Definition
		// Validate returns nil if the definition contains no validation error.
		// The Validate implementation may take advantage of ValidationErrors to report
		// more than one errors at a time.
		Validate() error
	}

	// Source is the interface implemented by definitions that can be initialized via DSL.
	Source interface {
		Definition
		// DSL returns the DSL used to initialize the definition if any.
		DSL() func()
	}

	// Finalize is the interface implemented by definitions that require an additional pass
	// after the DSL has executed (e.g. to merge generated definitions or initialize default
	// values)
	Finalize interface {
		Definition
		// Finalize is run by the DSL runner once the definition DSL has executed and the
		// definition has been validated.
		Finalize()
	}

	// Versioned is implemented by potentially versioned definitions such as API resources.
	Versioned interface {
		Definition
		// Versions returns an array of supported versions if the object is versioned, nil
		// othewise.
		Versions() []string
		// SupportsVersion returns true if the object supports the given version.
		SupportsVersion(ver string) bool
		// SupportsNoVersion returns true if the object is unversioned.
		SupportsNoVersion() bool
	}

	// SetIterator is the function signature used to iterate over definition sets with
	// IterateSets.
	SetIterator func(s DefinitionSet) error

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
		Validations []ValidationDefinition
		// Metadata is a list of key/value pairs
		Metadata MetadataDefinition
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

	// MetadataDefinition is a set of key/value pairs
	MetadataDefinition map[string][]string

	// TraitDefinition defines a set of reusable properties.
	TraitDefinition struct {
		// Trait name
		Name string
		// Trait DSL
		DSLFunc func()
	}

	// ValidationDefinition is the common interface for all validation data structures.
	// It doesn't expose any method and simply exists to help with documentation.
	ValidationDefinition interface {
		Definition
	}

	// EnumValidationDefinition represents an enum validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor76.
	EnumValidationDefinition struct {
		Values []interface{}
	}

	// FormatValidationDefinition represents a format validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor104.
	FormatValidationDefinition struct {
		Format string
	}

	// PatternValidationDefinition represents a pattern validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor33
	PatternValidationDefinition struct {
		Pattern string
	}

	// MinimumValidationDefinition represents an minimum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor21.
	MinimumValidationDefinition struct {
		Min float64
	}

	// MaximumValidationDefinition represents a maximum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor17.
	MaximumValidationDefinition struct {
		Max float64
	}

	// MinLengthValidationDefinition represents an minimum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor29.
	MinLengthValidationDefinition struct {
		MinLength int
	}

	// MaxLengthValidationDefinition represents an maximum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor26.
	MaxLengthValidationDefinition struct {
		MaxLength int
	}

	// RequiredValidationDefinition represents a required validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor61.
	RequiredValidationDefinition struct {
		Names []string
	}

	// VersionIterator is the type of functions given to IterateVersions.
	VersionIterator func(v *APIVersionDefinition) error
)

// CanUse returns nil if the provider supports all the versions supported by the client or if the
// provider is unversioned.
func CanUse(client, provider Versioned) error {
	if provider.Versions() == nil {
		return nil
	}
	versions := client.Versions()
	if versions == nil {
		return fmt.Errorf("cannot use versioned %s from unversioned %s", provider.Context(),
			client.Context())
	}
	providerVersions := provider.Versions()
	if len(versions) > len(providerVersions) {
		return fmt.Errorf("cannot use %s from %s: incompatible set of supported API versions",
			provider.Context(), client.Context())
	}
	for _, v := range versions {
		found := false
		for _, pv := range providerVersions {
			if v == pv {
				found = true
			}
			break
		}
		if !found {
			return fmt.Errorf("cannot use %s from %s: incompatible set of supported API versions",
				provider.Context(), client.Context())
		}
	}
	return nil
}

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
		if req, ok := v.(*RequiredValidationDefinition); ok {
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

// Dup returns a copy of the attribute definition.
// Note: the primitive underlying types are not duplicated for simplicity.
func (a *AttributeDefinition) Dup() *AttributeDefinition {
	valDup := make([]ValidationDefinition, len(a.Validations))
	for i, v := range a.Validations {
		valDup[i] = v
	}
	dupType := a.Type
	if dupType != nil {
		dupType = dupType.Dup()
	}
	dup := AttributeDefinition{
		Type:              dupType,
		Description:       a.Description,
		APIVersions:       a.APIVersions,
		Validations:       valDup,
		Metadata:          a.Metadata,
		DefaultValue:      a.DefaultValue,
		NonZeroAttributes: a.NonZeroAttributes,
		View:              a.View,
		DSLFunc:           a.DSLFunc,
	}
	return &dup
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
		case *EnumValidationDefinition:
			count := len(actual.Values)
			i := r.Int() % count
			return actual.Values[i]
		case *FormatValidationDefinition:
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
		case *PatternValidationDefinition:
			res, err := regen.Generate(actual.Pattern)
			if err != nil {
				return r.faker.Name()
			}
			return res
		case *MinimumValidationDefinition:
			return randomLengthExample(func(res float64) bool {
				return res >= actual.Min
			})
		case *MaximumValidationDefinition:
			return randomLengthExample(func(res float64) bool {
				return res <= actual.Max
			})
		case *MinLengthValidationDefinition:
			count := actual.MinLength + (r.Int() % 3)
			return randomValidationLengthExample(count)
		case *MaxLengthValidationDefinition:
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

// Context returns the generic definition name used in error messages.
func (t *TraitDefinition) Context() string {
	if t.Name != "" {
		return fmt.Sprintf("trait %#v", t.Name)
	}
	return "unnamed trait"
}

// DSL returns the initialization DSL.
func (t *TraitDefinition) DSL() func() {
	return t.DSLFunc
}

// Context returns the generic definition name used in error messages.
func (v *EnumValidationDefinition) Context() string {
	return "enum validation"
}

// Context returns the generic definition name used in error messages.
func (f *FormatValidationDefinition) Context() string {
	return "format validation"
}

// Context returns the generic definition name used in error messages.
func (f *PatternValidationDefinition) Context() string {
	return "pattern validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinimumValidationDefinition) Context() string {
	return "min value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaximumValidationDefinition) Context() string {
	return "max value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinLengthValidationDefinition) Context() string {
	return "min length validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaxLengthValidationDefinition) Context() string {
	return "max length validation"
}

// Context returns the generic definition name used in error messages.
func (r *RequiredValidationDefinition) Context() string {
	return "required field validation"
}
