// Package design defines types which describe the data types used by action controllers.
// These are the data structures of the request payloads and parameters as well as the response
// payloads.
// There are primitive types corresponding to the JSON primitive types (bool, string, integer and
// number), array types which represent a collection of another type and object types corresponding
// to JSON objects (i.e. a map indexed by strings where each value may be any of the data types).
// On top of these the package also defines "user types" and "media types". Both these types are
// named objects with additional properties (a description and for media types the media type
// identifier, links and views).
package design

type (
	// A Kind defines the JSON type that a DataType represents.
	Kind uint

	// DataType is the common interface to all types.
	DataType interface {
		Kind() Kind       // Kind
		ToObject() Object // ToObject returns the underlying object if any, nil otherwise.
	}

	// DataStructure is the interface implemented by all data structure types. That is
	// attribute definitions, user types and media types.
	DataStructure interface {
		// Definition returns the data structure definition.
		Definition() *AttributeDefinition
	}

	// NamedType is the interface implemented by all named data structures. That is
	// user and media types.
	NamedType interface {
		// TypeName is the user defined name of the data structure.
		TypeName() string
	}

	// Primitive is the type for null, boolean, integer, number and string.
	Primitive Kind

	// Array is the type for a JSON array.
	Array struct {
		ElemType *AttributeDefinition
	}

	// Object is the type for a JSON object.
	Object map[string]*AttributeDefinition

	// UserTypeDefinition is the type for user defined types that are not media types
	// (e.g. payload types).
	UserTypeDefinition struct {
		// A user type is an attribute definition.
		*AttributeDefinition
		// Name of type
		Name string
		// Description is the optional description of the media type.
		Description string
	}

	// MediaTypeDefinition describes the rendering of a resource using property and link
	// definitions. A property corresponds to a single member of the media type,
	// it has a name and a type as well as optional validation rules. A link has a
	// name and a URL that points to a related resource.
	// Media types also define views which describe which members and links to render when
	// building the response body for the corresponding view.
	MediaTypeDefinition struct {
		// A media type is a type
		*UserTypeDefinition
		// Identifier is the RFC 6838 media type identifier.
		Identifier string
		// Links list the rendered links indexed by name.
		Links map[string]*LinkDefinition
		// Views list the supported views indexed by name.
		Views map[string]*ViewDefinition
	}
)

const (
	// BooleanKind represents a JSON bool.
	BooleanKind = iota + 1
	// IntegerKind represents a JSON integer.
	IntegerKind
	// NumberKind represents a JSON number including integers.
	NumberKind
	// StringKind represents a JSON string.
	StringKind
	// ArrayKind represents a JSON array.
	ArrayKind
	// ObjectKind represents a JSON object.
	ObjectKind
	// UserTypeKind represents a user type.
	UserTypeKind
	// MediaTypeKind represents a media type.
	MediaTypeKind
)

const (
	// Boolean is the type for a JSON boolean.
	Boolean = Primitive(BooleanKind)

	// Integer is the type for a JSON number without a fraction or exponent part.
	Integer = Primitive(IntegerKind)

	// Number is the type for any JSON number, including integers.
	Number = Primitive(NumberKind)

	// String is the type for a JSON string.
	String = Primitive(StringKind)
)

// DataType implementation

// Kind implements DataKind.
func (p Primitive) Kind() Kind { return Kind(p) }

// ToObject returns nil.
func (p Primitive) ToObject() Object { return nil }

// Kind implements DataKind.
func (a *Array) Kind() Kind { return ArrayKind }

// ToObject returns nil.
func (a *Array) ToObject() Object { return nil }

// Kind implements DataKind.
func (o Object) Kind() Kind { return ObjectKind }

// ToObject returns the underlying object.
func (o Object) ToObject() Object { return o }

// Kind implements DataKind.
func (u *UserTypeDefinition) Kind() Kind { return UserTypeKind }

// ToObject calls ToObject on the user type underlying data type.
func (u *UserTypeDefinition) ToObject() Object { return u.Type.ToObject() }

// Kind implements DataKind.
func (m *MediaTypeDefinition) Kind() Kind { return MediaTypeKind }

// ToObject calls ToObject on the media type underlying data type.
func (m *MediaTypeDefinition) ToObject() Object { return m.Type.ToObject() }

// DataStructure implementation

// Definition returns itself for attribute definitions.
func (a *AttributeDefinition) Definition() *AttributeDefinition {
	return a
}

// Definition returns the underlying attribute definition.
func (u *UserTypeDefinition) Definition() *AttributeDefinition {
	return u.AttributeDefinition
}

// Definition returns the underlying attribute definition.
func (m *MediaTypeDefinition) Definition() *AttributeDefinition {
	return m.AttributeDefinition
}

// NamedType implementation

// TypeName is the user type name.
func (u *UserTypeDefinition) TypeName() string {
	return u.Name
}

// TypeName is the media type name.
func (m *MediaTypeDefinition) TypeName() string {
	return m.Name
}
