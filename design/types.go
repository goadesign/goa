package design

import "fmt"

type (
	// A Kind defines the JSON type that a DataType represents.
	Kind uint

	// DataType is the common interface to all types.
	DataType interface {
		Kind() Kind   // Kind
		Name() string // go type name
	}

	// Primitive is the type for null, boolean, integer, number and string.
	Primitive Kind

	// Array is the type for a JSON array.
	Array struct {
		ElemType DataType
	}

	// Object is the type for a JSON object.
	// Attribute definitions have a type (DataType) and may also define validation rules.
	Object map[string]*AttributeDefinition
)

const (
	// NullType is the JSON null value.
	NullType Kind = iota + 1
	// BooleanType represents a JSON bool.
	BooleanType
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

	// Null is the type for the JSON null value.
	Null = Primitive(NullType)

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
	case NullType:
		return "null"
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

// Name implements DataType, it return a human friendly name for the primitive.
func (p Primitive) Name() string {
	switch p.Kind() {
	case NullType:
		return "nil"
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

// Name implements DataType.
func (a *Array) Name() string {
	return "[]" + a.ElemType.Name()
}

// Kind implements DataType.
func (a Object) Kind() Kind {
	return ObjectType
}

// Name is not actually used as only payloads can be objects.
// Payloads are treated separatly.
func (a Object) Name() string {
	return "map[string]interface{}"
}
