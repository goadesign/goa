package design

import (
	"fmt"
	"strings"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

type (
	// AttributeMapExpr defines a map of attribute names to an object
	// attributes. The map may also define additional required attributes.
	AttributeMapExpr struct {
		// Parent is the parent expression.
		// It is one of ActionExpr, ResourceExpr or APIExpr.
		Parent eval.Expression
		// Base is the attribute whose object type contains the mapped
		// attributes if any. There may not be a based type (e.g. when
		// mapping base path params in the API expression).
		Base design.UserType
		// Names maps alias names to the base type attribute names.
		Names map[string]string
		// Required lists additional required attribute names.
		Required []string
	}
)

// NewAttributeMap creates an attribute map given a base type.
func NewAttributeMap(parent eval.Expression, base design.UserType) *AttributeMapExpr {
	return &AttributeMapExpr{
		Parent: parent,
		Base:   base,
		Names:  make(map[string]string),
	}
}

// EvalName returns the generic expression name used in error messages.
func (m *AttributeMapExpr) EvalName() string {
	return m.Parent.EvalName()
}

// Validate validates the attribute map expression.
func (m *AttributeMapExpr) Validate() error {
	if m.Base == nil {
		return nil
	}
	var (
		verr = new(eval.ValidationErrors)
		o    = design.AsObject(m.Base)
	)
	if o == nil {
		verr.Add(m, "Invalid base type, must be object got %s", m.Base.Name())
		return verr
	}

	for _, attName := range m.Names {
		if _, ok := o[attName]; !ok {
			verr.Add(m, "Unknown attribute %v", attName)
		}
	}

	for _, n := range m.Required {
		if _, ok := o[n]; !ok {
			verr.Add(m, "Unknown required attribute %v", n)
		}
	}

	return verr
}

// Merge merges the parent aliases and required attribute names into m.
func (m *AttributeMapExpr) Merge(o *AttributeMapExpr) {
	if o == nil {
		return
	}
	for n, a := range o.Names {
		if _, ok := m.Names[n]; !ok {
			m.Names[n] = a
		}
	}
	if o.Required != nil {
		m.Required = append(m.Required, o.Required...)
	}
}

// Alias adds a new alias to the map.
func (m *AttributeMapExpr) Alias(alias string) error {
	parts := strings.Split(alias, ":")
	if len(parts) > 2 {
		return fmt.Errorf("Invalid syntax, only at most one : may be used")
	}
	attName := parts[0]
	aliasName := attName
	if len(parts) == 2 {
		aliasName = parts[1]
	}
	m.Names[attName] = aliasName
	return nil
}

// Attribute computes the attribute expression by deriving from the base type
// attributes and applying the alias metadata. m must have a non-nil base type.
func (m *AttributeMapExpr) Attribute() *design.AttributeExpr {
	var (
		att = design.DupAtt(m.Base.Attribute())
		o   = design.AsObject(att.Type)
	)
	for n, catt := range o {
		if alias, ok := m.Names[n]; ok {
			if catt.Metadata == nil {
				catt.Metadata = make(design.MetadataExpr)
			}
			catt.Metadata["struct:field:origin"] = []string{alias}
		}
	}
	if m.Required != nil {
		att.Validation.Required = append(att.Validation.Required, m.Required...)
	}
	return att
}

// UserType creates a user type using the attribute returned by Attribute. If
// the attribute map does not define any alias or required attribute then
// UserType returns the attribute map base type unchanged. This is to alleviate
// the generation of additional data structures when not needed. m must have a
// non-nil base type.
func (m *AttributeMapExpr) UserType(name string) design.UserType {
	if len(m.Names) == 0 && len(m.Required) == 0 {
		return m.Base
	}
	return m.Base.Dup(m.Attribute())
}
