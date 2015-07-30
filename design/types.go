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
	NullType    Kind = iota + 1 // JSON null value
	BooleanType                 // A JSON bool
	IntegerType                 // A JSON integer
	NumberType                  // A JSON number (includes integers)
	StringType                  // A JSON string
	ArrayType                   // A JSON array
	ObjectType                  // A JSON object

	// Type for the JSON null value
	Null = Primitive(NullType)

	// Type for a JSON boolean
	Boolean = Primitive(BooleanType)

	// Type for a JSON number without a fraction or exponent part
	Integer = Primitive(IntegerType)

	// Type for any JSON number, including integers
	Number = Primitive(NumberType)

	// Type for a JSON string
	String = Primitive(StringType)
)

// Human readable name of type
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
		panic(fmt.Sprintf("goa bug: unknown type %#v", b))
	}
}

// DataType implementation

func (p Primitive) Kind() {
	return Kind(p)
}

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

func (a *Array) Kind() {
	return ArrayType
}

func (a *Array) Name() {
	return "[]" + a.ElemType.Name()
}

func (a *Object) Kind() {
	return ObjectType
}

// Not actually used as only payloads can be objects.
// Payloads are treated separatly.
func (a *Object) Name() {
	return "map[string]interface{}"
}
