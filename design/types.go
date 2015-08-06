package design

import (
	"bytes"
	"fmt"
	"unicode"
)

type (
	// A Kind defines the JSON type that a DataType represents.
	Kind uint

	// DataType is the common interface to all types.
	DataType interface {
		Kind() Kind     // Kind
		GoType() string // go type name
	}

	// Primitive is the type for null, boolean, integer, number and string.
	Primitive Kind

	// Array is the type for a JSON array.
	Array struct {
		ElemType *AttributeDefinition
	}

	// Object is the type for a JSON object.
	// Attribute definitions have a type (DataType) and may also define validation rules.
	Object map[string]*AttributeDefinition
)

const (
	// BooleanType represents a JSON bool.
	BooleanType = iota + 1
	// IntegerType represents a JSON integer.
	IntegerType
	// NumberType represents a JSON number including integers.
	NumberType
	// StringType represents a JSON string.
	StringType
	// ArrayType represents a JSON array.
	ArrayType
	// ObjectType represents a JSON object.
	ObjectType
)

const (
	// Boolean is the type for a JSON boolean.
	Boolean = Primitive(BooleanType)

	// Integer is the type for a JSON number without a fraction or exponent part.
	Integer = Primitive(IntegerType)

	// Number is the type for any JSON number, including integers.
	Number = Primitive(NumberType)

	// String is the type for a JSON string.
	String = Primitive(StringType)
)

// Name is the human readable name of type.
func (k Kind) Name() string {
	switch Kind(k) {
	case BooleanType:
		return "boolean"
	case IntegerType:
		return "integer"
	case NumberType:
		return "number"
	case StringType:
		return "string"
	case ArrayType:
		return "array"
	case ObjectType:
		return "object"
	default:
		panic(fmt.Sprintf("goa bug: unknown type %#v", k))
	}
}

// DataType implementation

// Kind implements DataType.
func (p Primitive) Kind() Kind {
	return Kind(p)
}

// GoType implements DataType, it return a human friendly name for the primitive.
func (p Primitive) GoType() string {
	switch p.Kind() {
	case BooleanType:
		return "bool"
	case IntegerType:
		return "int"
	case NumberType:
		return "float64"
	case StringType:
		return "string"
	default:
		panic(fmt.Sprintf("goa bug: unknown primitive type %#v", p))
	}
}

// Kind implements DataType.
func (a *Array) Kind() Kind {
	return ArrayType
}

// GoType implements DataType.
func (a *Array) GoType() string {
	return "[]" + a.ElemType.Type.GoType()
}

// Struct returns go code that defines a array of items that match the arrray ElemType definition.
func (a *Array) Struct() string {
	switch t := a.ElemType.Type.(type) {
	case Primitive:
		return "[]" + t.GoType()
	case *Array:
		return "[]" + t.Struct()
	case Object:
		return "[]" + a.ElemType.Struct()
	}
	panic("goa bug: unknown array element type")
}

// Kind implements DataType.
func (o Object) Kind() Kind {
	return ObjectType
}

// GoType is not actually used as only payloads can be objects.
// Payloads are treated separatly.
func (o Object) GoType() string {
	return "map[string]interface{}"
}

// Goify makes a valid go identifier out of any string.
// It does that by removing any non letter and non digit character and by making sure the first
// character is a letter or "_".
// Goify produces a "CamelCase" version of the string, if firstUpper is true the first character
// of the identifier is uppercase otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	var b bytes.Buffer
	var firstWritten, nextUpper bool
	for i := 0; i < len(str); i++ {
		r := rune(str[i])
		if r == '_' {
			nextUpper = true
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !firstWritten {
				if firstUpper {
					r = unicode.ToUpper(r)
				} else {
					r = unicode.ToLower(r)
				}
				firstWritten = true
				nextUpper = false
			} else if nextUpper {
				r = unicode.ToUpper(r)
				nextUpper = false
			}
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "_v" // you have a better idea?
	}
	res := b.String()
	if _, ok := reserved[res]; ok {
		res += "_"
	}
	return res
}

// reserved golang keywords
var reserved = map[string]bool{
	"byte":       true,
	"complex128": true,
	"complex64":  true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"int8":       true,
	"rune":       true,
	"string":     true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uint8":      true,

	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,
}
