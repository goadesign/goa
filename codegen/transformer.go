package codegen

import (
	"fmt"

	"goa.design/goa/expr"
)

type (
	// Transformer produces code that initializes data structure defined by
	// target from an instance of the data structure described by source. The
	// data structures can be objects, arrays or maps. The algorithm matches
	// object fields by name and ignores object fields in target that don't
	// have a match in source.
	Transformer interface {
		// Transform returns the code that initializes data structure defined by
		// target attribute from an instance of the data structure defined by
		// source. It leverages mapped attributes so that attribute names may use
		// the "name:elem" syntax to define the name of the design attribute and
		// the name of the corresponding generated field. It returns an error
		// if target is not compatible with source (different type, fields of
		// different type etc).
		Transform(source, target *ContextualAttribute, ta *TransformAttrs) (code string, err error)
		// TransformObject returns the code to initialize a target data structure
		// defined by object type from an instance of source data structure defined
		// by an object type. The algorithm matches object fields by name and
		// ignores object fields in target that don't have a match in source.
		// It returns an error if source and target are different types or have
		// fields of different types.
		TransformObject(source, target *ContextualAttribute, ta *TransformAttrs) (code string, err error)
		// TransformArray returns the code to initialize a target array from a
		// source array. It returns an error if source and target are not arrays
		// and have fields of different types in the array element.
		TransformArray(source, target *ContextualAttribute, ta *TransformAttrs) (code string, err error)
		// TransformMap returns the code to initialize a target map from a
		// source map. It returns an error if source and target are not maps
		// and have fields of different types in the map key and element.
		TransformMap(source, target *ContextualAttribute, ta *TransformAttrs) (code string, err error)
		// MakeCompatible checks whether target is compatible with the source
		// (same type, fields of different type, etc) and returns an error if
		// target cannot be made compatible to the source. If no error, it returns
		// the compatible source and target attributes with the updated transform
		// attributes to make them compatible.
		MakeCompatible(source, target *ContextualAttribute, ta *TransformAttrs, suffix string) (src, tgt *ContextualAttribute, newTA *TransformAttrs, err error)
		Converter
	}

	// Referencer refers to a type.
	Referencer interface {
		// Name returns the type name.
		Name() string
		// Ref returns the reference to the type.
		Ref() string
	}

	// Definer generates code that defines a type.
	Definer interface {
		// Def returns the code defining a type. Pointer and useDefault paramerters
		// are used to determine if the type fields must be a pointer.
		Def(pointer, useDefault bool) string
	}

	// Attributor is the interface implemented by code generators to generate
	// code for an attribute type.
	Attributor interface {
		Scoper
		Referencer
		Definer
		// Field produces a valid field name for the attribute type.
		Field(name string, firstUpper bool) string
		// Expr returns the underlying attribute expression.
		Expr() *expr.AttributeExpr
		// Dup creates a copy of the attributor by setting the underlying
		// attribute expression.
		Dup(*expr.AttributeExpr) Attributor
	}

	// Converter is the interface implemented by code generators to generate
	// code to convert source attribute type to a target attribute type.
	Converter interface {
		// ConvertType produces code to initialze target attribute type from a
		// source attribute type held by variable in sourceVar. It is not a
		// recursive function.
		ConvertType(source, target Attributor, sourceVar string) (code string)
	}

	// ContextualAttribute determines how an attribute behaves based on certain
	// properties during code generation.
	ContextualAttribute struct {
		// Attribute is the attribute expression for which the code is generated.
		Attribute Attributor
		// NonPointer if true indicates that the attribute type is not generated
		// as a pointer irrespective of whether the attribue is required or has
		// a default value.
		NonPointer bool
		// Pointer if true indicates that the attribute type is generated as a
		// pointer even if the attribute is required or has a default value.
		// Array and map types are are always non-pointers. Object types are always
		// pointers.
		Pointer bool
		// UseDefault if true indicates that attribute type must be a non-pointer
		// if it has a default value except object type which is always a pointer.
		UseDefault bool
		// Required if true indicates that the attribute is required.
		Required bool
	}

	// TransformAttrs are the attributes that help in the transformation.
	TransformAttrs struct {
		// SourceVar and TargetVar are the source and target variable names used
		// in the transformation code.
		SourceVar, TargetVar string
		// NewVar is used to determine the assignment operator to initialize
		// TargetVar.
		NewVar bool
	}

	// TransformFunctionData describes a helper function used to transform
	// user types. These are necessary to prevent potential infinite
	// recursion when a type attribute is defined recursively. For example:
	//
	//     var Recursive = Type("Recursive", func() {
	//         Attribute("r", "Recursive")
	//     }
	//
	// Transforming this type requires generating an intermediary function:
	//
	//     func recursiveToRecursive(r *Recursive) *service.Recursive {
	//         var t service.Recursive
	//         if r.R != nil {
	//             t.R = recursiveToRecursive(r.R)
	//         }
	//    }
	//
	TransformFunctionData struct {
		Name          string
		ParamTypeRef  string
		ResultTypeRef string
		Code          string
	}
)

// NewGoContextAttr returns a default Go contextual attribute that produces Go
// code.
func NewGoContextAttr(att *expr.AttributeExpr, pkg string, scope *NameScope) *ContextualAttribute {
	return &ContextualAttribute{Attribute: NewGoAttribute(att, pkg, scope)}
}

// IsPointer checks if the attribute type is a pointer. It returns false
// if attribute type is an array, map, byte array, or an interface. If Pointer
// property is true, IsPointer returns true. If NonPointer property is true,
// IsPointer returns false. If both Pointer and NonPointer are false, the
// following table shows how the attribute properties affect the return value
//
//    UseDefault | Required | IsPointer
//         T     |     T    |     T
//         F     |     F    |     T
//         T     |     F    |     F if default value exists, else T
//         F     |     T    |     T
//
func (c *ContextualAttribute) IsPointer() bool {
	if dt := c.Attribute.Expr().Type.Kind(); dt == expr.BytesKind || dt == expr.AnyKind {
		return false
	}
	if c.NonPointer {
		return false
	}
	if c.Pointer {
		return true
	}
	return !c.Required && c.DefaultValue() == nil
}

// DefaultValue returns the default value of the attribute type if UseDefault
// is true. It returns nil otherwise.
func (c *ContextualAttribute) DefaultValue() interface{} {
	if c.UseDefault {
		return c.Attribute.Expr().DefaultValue
	}
	return nil
}

// Def returns the attribute type definition.
func (c *ContextualAttribute) Def() string {
	return c.Attribute.Def(c.Pointer, c.UseDefault)
}

// Dup creates a shallow copy of the contextual attribute with the given
// attributor and its requiredness.
func (c *ContextualAttribute) Dup(attr *expr.AttributeExpr, required bool) *ContextualAttribute {
	return &ContextualAttribute{
		Attribute:  c.Attribute.Dup(attr),
		Required:   required,
		NonPointer: c.NonPointer,
		Pointer:    c.Pointer,
		UseDefault: c.UseDefault,
	}
}

// IsCompatible returns an error if a and b are not both objects, both arrays,
// both maps or both the same primitive type. actx and bctx are used to build
// the error message if any.
func IsCompatible(a, b expr.DataType, actx, bctx string) error {
	switch {
	case expr.IsObject(a):
		if !expr.IsObject(b) {
			return fmt.Errorf("%s is an object but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsArray(a):
		if !expr.IsArray(b) {
			return fmt.Errorf("%s is an array but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsMap(a):
		if !expr.IsMap(b) {
			return fmt.Errorf("%s is a hash but %s type is %s", actx, bctx, b.Name())
		}
	default:
		if a.Kind() != b.Kind() {
			return fmt.Errorf("%s is a %s but %s type is %s", actx, a.Name(), bctx, b.Name())
		}
	}
	return nil
}

// AppendHelpers takes care of only appending helper functions from newH that
// are not already in oldH.
func AppendHelpers(oldH, newH []*TransformFunctionData) []*TransformFunctionData {
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if h.Name == h2.Name {
				found = true
				break
			}
		}
		if !found {
			oldH = append(oldH, h)
		}
	}
	return oldH
}

// HelperName returns the transformation function name to initialize a target
// user type from an instance of a source user type.
func HelperName(source, target Attributor, prefix string) string {
	var (
		sname string
		tname string
	)
	{
		sname = Goify(source.Name(), true)
		tname = Goify(target.Name(), true)
		if prefix == "" {
			prefix = "transform"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}
