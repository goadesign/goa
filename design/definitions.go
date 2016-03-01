package design

import (
	"sort"

	"github.com/goadesign/goa/dslengine"
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
		// Optional validations
		Validation *dslengine.ValidationDefinition
		// Metadata is a list of key/value pairs
		Metadata dslengine.MetadataDefinition
		// Optional member default value
		DefaultValue interface{}
		// Optional member example value
		Example interface{}
		// Optional view used to render Attribute (only applies to media type attributes).
		View string
		// NonZeroAttributes lists the names of the child attributes that cannot have a
		// zero value (and thus whose presence does not need to be validated).
		NonZeroAttributes map[string]bool
		// DSLFunc contains the initialization DSL. This is used for user types.
		DSLFunc func()
		// isCustomExample keeps track of whether the example is given by the user, or
		// should be automatically generated for the user.
		isCustomExample bool
	}

	// ContainerDefinition defines a generic container definition that contains attributes.
	// This makes it possible for plugins to use attributes in their own data structures.
	ContainerDefinition interface {
		// Attribute returns the container definition embedded attribute.
		Attribute() *AttributeDefinition
	}
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
	if a.Validation == nil {
		return
	}
	required = a.Validation.Required
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

// GenerateExample returns a random instance of the attribute that validates.
func (a *AttributeDefinition) GenerateExample(r *RandomGenerator) interface{} {
	if example := newExampleGenerator(a, r).generate(); example != nil {
		return example
	}
	return a.Type.GenerateExample(r)
}

// SetExample sets the custom example. SetExample also handles the case when the user doesn't
// want any example or any auto-generated example.
func (a *AttributeDefinition) SetExample(example interface{}) bool {
	if example == nil {
		a.Example = nil
		a.isCustomExample = true
		return true
	}
	if a.Type == nil || a.Type.IsCompatible(example) {
		a.Example = example
		a.isCustomExample = true
		return true
	}
	return false
}

// finalizeExample goes through each Example and consolidates all of the information it knows i.e.
// a custom example or auto-generate for the user. It also tracks whether we've randomized
// the entire example; if so, we shall re-generate the random value for Array/Hash.
func (a *AttributeDefinition) finalizeExample(stack []*AttributeDefinition) (interface{}, bool) {
	if a.Example != nil || a.isCustomExample {
		return a.Example, a.isCustomExample
	}

	// note: must traverse each node to finalize the examples unless given
	switch true {
	case a.Type.IsArray():
		ary := a.Type.ToArray()
		example, isCustom := ary.ElemType.finalizeExample(stack)
		a.Example, a.isCustomExample = ary.MakeSlice([]interface{}{example}), isCustom
	case a.Type.IsHash():
		h := a.Type.ToHash()
		exampleK, isCustomK := h.KeyType.finalizeExample(stack)
		exampleV, isCustomV := h.ElemType.finalizeExample(stack)
		a.Example, a.isCustomExample = h.MakeMap(map[interface{}]interface{}{exampleK: exampleV}), isCustomK || isCustomV
	case a.Type.IsObject():
		// keep track of the type id, in case of a cyclical situation
		stack = append(stack, a)

		// ensure fixed ordering
		aObj := a.Type.ToObject()
		keys := make([]string, 0, len(aObj))
		for n := range aObj {
			keys = append(keys, n)
		}
		sort.Strings(keys)

		example, hasCustom, isCustom := map[string]interface{}{}, false, false
		for _, n := range keys {
			att := aObj[n]
			// avoid a cyclical dependency
			isCyclical := false
			if ssize := len(stack); ssize > 0 {
				aid := ""
				if mt, ok := att.Type.(*MediaTypeDefinition); ok {
					aid = mt.Identifier
				} else if ut, ok := att.Type.(*UserTypeDefinition); ok {
					aid = ut.TypeName
				}
				if aid != "" {
					for _, sa := range stack[:ssize-1] {
						if mt, ok := sa.Type.(*MediaTypeDefinition); ok {
							isCyclical = mt.Identifier == aid
						} else if ut, ok := sa.Type.(*UserTypeDefinition); ok {
							isCyclical = ut.TypeName == aid
						}
						if isCyclical {
							break
						}
					}
				}
			}
			if !isCyclical {
				example[n], isCustom = att.finalizeExample(stack)
			} else {
				// unable to generate any example and here we set
				// isCustom to avoid touching this example again
				// i.e. GenerateExample in the end of this func
				example[n], isCustom = nil, true
			}
			hasCustom = hasCustom || isCustom
		}
		a.Example, a.isCustomExample = example, hasCustom
	}
	// while none of the examples is custom, we generate a random value for the entire object
	if !a.isCustomExample {
		a.Example = a.GenerateExample(Design.RandomGenerator())
	}
	return a.Example, a.isCustomExample
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
	if parent.Validation == nil {
		return
	}
	if a.Validation == nil {
		a.Validation = &dslengine.ValidationDefinition{}
	}
	a.Validation.AddRequired(parent.Validation.Required)
}

func (a *AttributeDefinition) shouldInherit(parent *AttributeDefinition) bool {
	return a != nil && a.Type.ToObject() != nil &&
		parent != nil && parent.Type.ToObject() != nil
}
