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

import (
	"fmt"
	"mime"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/goadesign/goa/dslengine"
	"github.com/satori/go.uuid"
)

// DefaultView is the name of the default view.
const DefaultView = "default"

// It returns the default view - or if not available the link view - or if not available the first
// view by alphabetical order.
type (
	// A Kind defines the JSON type that a DataType represents.
	Kind uint

	// DataType is the common interface to all types.
	DataType interface {
		// Kind of data type, one of the Kind enum.
		Kind() Kind
		// Name returns the type name.
		Name() string
		// IsPrimitive returns true if the underlying type is one of the primitive types.
		IsPrimitive() bool
		// HasAttributes returns true if the underlying type has any attributes.
		HasAttributes() bool
		// IsObject returns true if the underlying type is an object, a user type which
		// is an object or a media type whose type is an object.
		IsObject() bool
		// IsArray returns true if the underlying type is an array, a user type which
		// is an array or a media type whose type is an array.
		IsArray() bool
		// IsHash returns true if the underlying type is a hash map, a user type which
		// is a hash map or a media type whose type is a hash map.
		IsHash() bool
		// ToObject returns the underlying object if any (i.e. if IsObject returns true),
		// nil otherwise.
		ToObject() Object
		// ToArray returns the underlying array if any (i.e. if IsArray returns true),
		// nil otherwise.
		ToArray() *Array
		// ToHash returns the underlying hash map if any (i.e. if IsHash returns true),
		// nil otherwise.
		ToHash() *Hash
		// CanHaveDefault returns whether the data type can have a default value.
		CanHaveDefault() bool
		// IsCompatible checks whether val has a Go type that is
		// compatible with the data type.
		IsCompatible(val interface{}) bool
		// GenerateExample returns a random value for the given data type.
		// If the data type has validations then the example value validates them.
		// seen keeps track of the user and media types that have been traversed via
		// recursion to prevent infinite loops.
		GenerateExample(r *RandomGenerator, seen []string) interface{}
	}

	// DataStructure is the interface implemented by all data structure types.
	// That is attribute definitions, user types and media types.
	DataStructure interface {
		// Definition returns the data structure definition.
		Definition() *AttributeDefinition
		// Walk traverses the data structure recursively and calls the given function once
		// on each attribute starting with the attribute returned by Definition.
		// User type and media type attributes are traversed once even for recursive
		// definitions to avoid infinite recursion.
		// Walk stops and returns the error if the function returns a non-nil error.
		Walk(func(*AttributeDefinition) error) error
	}

	// Primitive is the type for null, boolean, integer, number, string, and time.
	Primitive Kind

	// Array is the type for a JSON array.
	Array struct {
		ElemType *AttributeDefinition
	}

	// ArrayVal is the value of an array used to specify the default value.
	ArrayVal []interface{}

	// Object is the type for a JSON object.
	Object map[string]*AttributeDefinition

	// Hash is the type for a hash map.
	Hash struct {
		KeyType  *AttributeDefinition
		ElemType *AttributeDefinition
	}

	// HashVal is the value of a hash used to specify the default value.
	HashVal map[interface{}]interface{}

	// UserTypeDefinition is the type for user defined types that are not media types
	// (e.g. payload types).
	UserTypeDefinition struct {
		// A user type is an attribute definition.
		*AttributeDefinition
		// Name of type
		TypeName string
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
		// ContentType identifies the value written to the response "Content-Type" header.
		// Defaults to Identifier.
		ContentType string
		// Links list the rendered links indexed by name.
		Links map[string]*LinkDefinition
		// Views list the supported views indexed by name.
		Views map[string]*ViewDefinition
		// Resource this media type is the canonical representation for if any
		Resource *ResourceDefinition
	}
)

const (
	// BooleanKind represents a JSON bool.
	BooleanKind Kind = iota + 1
	// IntegerKind represents a JSON integer.
	IntegerKind
	// NumberKind represents a JSON number including integers.
	NumberKind
	// StringKind represents a JSON string.
	StringKind
	// DateTimeKind represents a JSON string that is parsed as a Go time.Time
	DateTimeKind
	// UUIDKind represents a JSON string that is parsed as a Go uuid.UUID
	UUIDKind
	// AnyKind represents a generic interface{}.
	AnyKind
	// ArrayKind represents a JSON array.
	ArrayKind
	// ObjectKind represents a JSON object.
	ObjectKind
	// HashKind represents a JSON object where the keys are not known in advance.
	HashKind
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

	// DateTime is the type for a JSON string parsed as a Go time.Time
	// DateTime expects an RFC3339 formatted value.
	DateTime = Primitive(DateTimeKind)

	// UUID is the type for a JSON string parsed as a Go uuid.UUID
	// UUID expects an RFC4122 formatted value.
	UUID = Primitive(UUIDKind)

	// Any is the type for an arbitrary JSON value (interface{} in Go).
	Any = Primitive(AnyKind)
)

// DataType implementation

// Kind implements DataKind.
func (p Primitive) Kind() Kind { return Kind(p) }

// Name returns the JSON type name.
func (p Primitive) Name() string {
	switch p {
	case Boolean:
		return "boolean"
	case Integer:
		return "integer"
	case Number:
		return "number"
	case String, DateTime, UUID:
		return "string"
	case Any:
		return "any"
	default:
		panic("unknown primitive type") // bug
	}
}

// IsPrimitive returns true.
func (p Primitive) IsPrimitive() bool { return true }

// HasAttributes returns false.
func (p Primitive) HasAttributes() bool { return false }

// IsObject returns false.
func (p Primitive) IsObject() bool { return false }

// IsArray returns false.
func (p Primitive) IsArray() bool { return false }

// IsHash returns false.
func (p Primitive) IsHash() bool { return false }

// ToObject returns nil.
func (p Primitive) ToObject() Object { return nil }

// ToArray returns nil.
func (p Primitive) ToArray() *Array { return nil }

// ToHash returns nil.
func (p Primitive) ToHash() *Hash { return nil }

// CanHaveDefault returns whether the primitive can have a default value.
func (p Primitive) CanHaveDefault() (ok bool) {
	switch p {
	case Boolean, Integer, Number, String, DateTime:
		ok = true
	}
	return
}

// IsCompatible returns true if val is compatible with p.
func (p Primitive) IsCompatible(val interface{}) bool {
	if p != Boolean && p != Integer && p != Number && p != String && p != DateTime && p != UUID && p != Any {
		panic("unknown primitive type") // bug
	}
	if p == Any {
		return true
	}
	switch val.(type) {
	case bool:
		return p == Boolean
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return p == Integer || p == Number
	case float32, float64:
		return p == Number
	case string:
		if p == String {
			return true
		}
		if p == DateTime {
			_, err := time.Parse(time.RFC3339, val.(string))
			return err == nil
		}
		if p == UUID {
			_, err := uuid.FromString(val.(string))
			return err == nil
		}
	}
	return false
}

var anyPrimitive = []Primitive{Boolean, Integer, Number, DateTime, UUID}

// GenerateExample returns an instance of the given data type.
func (p Primitive) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	switch p {
	case Boolean:
		return r.Bool()
	case Integer:
		return r.Int()
	case Number:
		return r.Float64()
	case String:
		return r.String()
	case DateTime:
		return r.DateTime()
	case UUID:
		return r.UUID().String() // Generate string to can be JSON marshaled
	case Any:
		// to not make it too complicated, pick one of the primitive types
		return anyPrimitive[r.Int()%len(anyPrimitive)].GenerateExample(r, seen)
	default:
		panic("unknown primitive type") // bug
	}
}

// Kind implements DataKind.
func (a *Array) Kind() Kind { return ArrayKind }

// Name returns the type name.
func (a *Array) Name() string {
	return "array"
}

// IsPrimitive returns false.
func (a *Array) IsPrimitive() bool { return false }

// HasAttributes returns true if the array's element type is user defined.
func (a *Array) HasAttributes() bool {
	return a.ElemType.Type.HasAttributes()
}

// IsObject returns false.
func (a *Array) IsObject() bool { return false }

// IsArray returns true.
func (a *Array) IsArray() bool { return true }

// IsHash returns false.
func (a *Array) IsHash() bool { return false }

// ToObject returns nil.
func (a *Array) ToObject() Object { return nil }

// ToArray returns a.
func (a *Array) ToArray() *Array { return a }

// ToHash returns nil.
func (a *Array) ToHash() *Hash { return nil }

// CanHaveDefault returns true if the array type can have a default value.
// The array type can have a default value only if the element type can
// have a default value.
func (a *Array) CanHaveDefault() bool {
	return a.ElemType.Type.CanHaveDefault()
}

// IsCompatible returns true if val is compatible with p.
func (a *Array) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	if k != reflect.Array && k != reflect.Slice {
		return false
	}
	v := reflect.ValueOf(val)
	for i := 0; i < v.Len(); i++ {
		compat := (a.ElemType.Type != nil) && a.ElemType.Type.IsCompatible(v.Index(i).Interface())
		if !compat {
			return false
		}
	}
	return true
}

// GenerateExample produces a random array value.
func (a *Array) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = a.ElemType.Type.GenerateExample(r, seen)
	}
	return a.MakeSlice(res)
}

// MakeSlice examines the key type from the Array and create a slice with builtin type if possible.
// The idea is to avoid generating []interface{} and produce more known types.
func (a *Array) MakeSlice(s []interface{}) interface{} {
	slice := reflect.MakeSlice(toReflectType(a), 0, len(s))
	for _, item := range s {
		slice = reflect.Append(slice, reflect.ValueOf(item))
	}
	return slice.Interface()
}

// Kind implements DataKind.
func (o Object) Kind() Kind { return ObjectKind }

// Name returns the type name.
func (o Object) Name() string { return "object" }

// IsPrimitive returns false.
func (o Object) IsPrimitive() bool { return false }

// HasAttributes returns true.
func (o Object) HasAttributes() bool { return true }

// IsObject returns true.
func (o Object) IsObject() bool { return true }

// IsArray returns false.
func (o Object) IsArray() bool { return false }

// IsHash returns false.
func (o Object) IsHash() bool { return false }

// ToObject returns the underlying object.
func (o Object) ToObject() Object { return o }

// ToArray returns nil.
func (o Object) ToArray() *Array { return nil }

// ToHash returns nil.
func (o Object) ToHash() *Hash { return nil }

// CanHaveDefault returns false.
func (o Object) CanHaveDefault() bool { return false }

// Merge copies other's attributes into o overridding any pre-existing attribute with the same name.
func (o Object) Merge(other Object) {
	for n, att := range other {
		o[n] = DupAtt(att)
	}
}

// IsCompatible returns true if val is compatible with p.
func (o Object) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	return k == reflect.Map || k == reflect.Struct
}

// GenerateExample returns a random value of the object.
func (o Object) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	// ensure fixed ordering
	keys := make([]string, 0, len(o))
	for n := range o {
		keys = append(keys, n)
	}
	sort.Strings(keys)

	res := make(map[string]interface{})
	for _, n := range keys {
		att := o[n]
		res[n] = att.Type.GenerateExample(r, seen)
	}
	return res
}

// Kind implements DataKind.
func (h *Hash) Kind() Kind { return HashKind }

// Name returns the type name.
func (h *Hash) Name() string { return "hash" }

// IsPrimitive returns false.
func (h *Hash) IsPrimitive() bool { return false }

// HasAttributes returns true if the either hash's key type is user defined
// or the element type is user defined.
func (h *Hash) HasAttributes() bool {
	return h.KeyType.Type.HasAttributes() || h.ElemType.Type.HasAttributes()
}

// IsObject returns false.
func (h *Hash) IsObject() bool { return false }

// IsArray returns false.
func (h *Hash) IsArray() bool { return false }

// IsHash returns true.
func (h *Hash) IsHash() bool { return true }

// ToObject returns nil.
func (h *Hash) ToObject() Object { return nil }

// ToArray returns nil.
func (h *Hash) ToArray() *Array { return nil }

// ToHash returns the underlying hash map.
func (h *Hash) ToHash() *Hash { return h }

// CanHaveDefault returns true if the hash type can have a default value.
// The hash type can have a default value only if both the key type and
// the element type can have a default value.
func (h *Hash) CanHaveDefault() bool {
	return h.KeyType.Type.CanHaveDefault() && h.ElemType.Type.CanHaveDefault()
}

// IsCompatible returns true if val is compatible with p.
func (h *Hash) IsCompatible(val interface{}) bool {
	k := reflect.TypeOf(val).Kind()
	if k != reflect.Map {
		return false
	}
	v := reflect.ValueOf(val)
	for _, key := range v.MapKeys() {
		keyCompat := h.KeyType.Type == nil || h.KeyType.Type.IsCompatible(key.Interface())
		elemCompat := h.ElemType.Type == nil || h.ElemType.Type.IsCompatible(v.MapIndex(key).Interface())
		if !keyCompat || !elemCompat {
			return false
		}
	}
	return true
}

// GenerateExample returns a random hash value.
func (h *Hash) GenerateExample(r *RandomGenerator, seen []string) interface{} {
	count := r.Int()%3 + 1
	pair := map[interface{}]interface{}{}
	for i := 0; i < count; i++ {
		pair[h.KeyType.Type.GenerateExample(r, seen)] = h.ElemType.Type.GenerateExample(r, seen)
	}
	return h.MakeMap(pair)
}

// MakeMap examines the key type from a Hash and create a map with builtin type if possible.
// The idea is to avoid generating map[interface{}]interface{}, which cannot be handled by json.Marshal.
func (h *Hash) MakeMap(m map[interface{}]interface{}) interface{} {
	hash := reflect.MakeMap(toReflectType(h))
	for key, value := range m {
		hash.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
	return hash.Interface()
}

// AttributeIterator is the type of the function given to IterateAttributes.
type AttributeIterator func(string, *AttributeDefinition) error

// IterateAttributes calls the given iterator passing in each attribute sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateObject returns that
// error.
func (o Object) IterateAttributes(it AttributeIterator) error {
	names := make([]string, len(o))
	i := 0
	for n := range o {
		names[i] = n
		i++
	}
	sort.Strings(names)
	for _, n := range names {
		if err := it(n, o[n]); err != nil {
			return err
		}
	}
	return nil
}

// UserTypes traverses the data type recursively and collects all the user types used to
// define it. The returned map is indexed by type name.
func UserTypes(dt DataType) map[string]*UserTypeDefinition {
	collect := func(types map[string]*UserTypeDefinition) func(*AttributeDefinition) error {
		return func(at *AttributeDefinition) error {
			if u, ok := at.Type.(*UserTypeDefinition); ok {
				types[u.TypeName] = u
			} else if m, ok := at.Type.(*MediaTypeDefinition); ok {
				types[m.TypeName] = m.UserTypeDefinition
			}
			return nil
		}
	}
	switch actual := dt.(type) {
	case Primitive:
		return nil
	case *Array:
		return UserTypes(actual.ElemType.Type)
	case *Hash:
		ktypes := UserTypes(actual.KeyType.Type)
		vtypes := UserTypes(actual.ElemType.Type)
		if vtypes == nil {
			return ktypes
		}
		for n, ut := range ktypes {
			vtypes[n] = ut
		}
		return vtypes
	case Object:
		types := make(map[string]*UserTypeDefinition)
		for _, att := range actual {
			att.Walk(collect(types))
		}
		if len(types) == 0 {
			return nil
		}
		return types
	case *UserTypeDefinition:
		types := map[string]*UserTypeDefinition{actual.TypeName: actual}
		actual.Walk(collect(types))
		return types
	case *MediaTypeDefinition:
		types := map[string]*UserTypeDefinition{actual.TypeName: actual.UserTypeDefinition}
		actual.Walk(collect(types))
		return types
	default:
		panic("unknown type") // bug
	}
}

// ToSlice converts an ArrayVal to a slice.
func (a ArrayVal) ToSlice() []interface{} {
	arr := make([]interface{}, len(a))
	for i, elem := range a {
		switch actual := elem.(type) {
		case ArrayVal:
			arr[i] = actual.ToSlice()
		case HashVal:
			arr[i] = actual.ToMap()
		default:
			arr[i] = actual
		}
	}
	return arr
}

// ToMap converts a HashVal to a map.
func (h HashVal) ToMap() map[interface{}]interface{} {
	mp := make(map[interface{}]interface{}, len(h))
	for k, v := range h {
		switch actual := v.(type) {
		case ArrayVal:
			mp[k] = actual.ToSlice()
		case HashVal:
			mp[k] = actual.ToMap()
		default:
			mp[k] = actual
		}
	}
	return mp
}

// NewUserTypeDefinition creates a user type definition but does not
// execute the DSL.
func NewUserTypeDefinition(name string, dsl func()) *UserTypeDefinition {
	return &UserTypeDefinition{
		TypeName:            name,
		AttributeDefinition: &AttributeDefinition{DSLFunc: dsl},
	}
}

// Kind implements DataKind.
func (u *UserTypeDefinition) Kind() Kind { return UserTypeKind }

// Name returns the JSON type name.
func (u *UserTypeDefinition) Name() string { return u.Type.Name() }

// IsPrimitive calls IsPrimitive on the user type underlying data type.
func (u *UserTypeDefinition) IsPrimitive() bool { return u.Type != nil && u.Type.IsPrimitive() }

// HasAttributes calls the HasAttributes on the user type underlying data type.
func (u *UserTypeDefinition) HasAttributes() bool { return u.Type.HasAttributes() }

// IsObject calls IsObject on the user type underlying data type.
func (u *UserTypeDefinition) IsObject() bool { return u.Type != nil && u.Type.IsObject() }

// IsArray calls IsArray on the user type underlying data type.
func (u *UserTypeDefinition) IsArray() bool { return u.Type != nil && u.Type.IsArray() }

// IsHash calls IsHash on the user type underlying data type.
func (u *UserTypeDefinition) IsHash() bool { return u.Type != nil && u.Type.IsHash() }

// ToObject calls ToObject on the user type underlying data type.
func (u *UserTypeDefinition) ToObject() Object { return u.Type.ToObject() }

// ToArray calls ToArray on the user type underlying data type.
func (u *UserTypeDefinition) ToArray() *Array { return u.Type.ToArray() }

// ToHash calls ToHash on the user type underlying data type.
func (u *UserTypeDefinition) ToHash() *Hash { return u.Type.ToHash() }

// CanHaveDefault calls CanHaveDefault on the user type underlying data type.
func (u *UserTypeDefinition) CanHaveDefault() bool { return u.Type.CanHaveDefault() }

// IsCompatible returns true if val is compatible with u.
func (u *UserTypeDefinition) IsCompatible(val interface{}) bool {
	return u.Type == nil || u.Type.IsCompatible(val)
}

// Finalize merges base type attributes.
func (u *UserTypeDefinition) Finalize() {
	if u.Reference != nil {
		if bat := u.AttributeDefinition; bat != nil {
			u.AttributeDefinition.Inherit(bat)
		}
	}

	u.GenerateExample(Design.RandomGenerator(), nil)
}

// NewMediaTypeDefinition creates a media type definition but does not
// execute the DSL.
func NewMediaTypeDefinition(name, identifier string, dsl func()) *MediaTypeDefinition {
	return &MediaTypeDefinition{
		UserTypeDefinition: &UserTypeDefinition{
			AttributeDefinition: &AttributeDefinition{Type: Object{}, DSLFunc: dsl},
			TypeName:            name,
		},
		Identifier: identifier,
	}
}

// Kind implements DataKind.
func (m *MediaTypeDefinition) Kind() Kind { return MediaTypeKind }

// IsError returns true if the media type is implemented via a goa struct.
func (m *MediaTypeDefinition) IsError() bool {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		panic("invalid media type identifier " + m.Identifier) // bug
	}
	delete(params, "view")
	return mime.FormatMediaType(base, params) == ErrorMedia.Identifier
}

// ComputeViews returns the media type views recursing as necessary if the media type is a
// collection.
func (m *MediaTypeDefinition) ComputeViews() map[string]*ViewDefinition {
	if m.Views != nil {
		return m.Views
	}
	if m.IsArray() {
		if mt, ok := m.ToArray().ElemType.Type.(*MediaTypeDefinition); ok {
			return mt.ComputeViews()
		}
	}
	return nil
}

// Finalize sets the value of ContentType to the identifier if not set.
func (m *MediaTypeDefinition) Finalize() {
	if m.ContentType == "" {
		m.ContentType = m.Identifier
	}
	m.UserTypeDefinition.Finalize()
}

// ViewIterator is the type of the function given to IterateViews.
type ViewIterator func(*ViewDefinition) error

// IterateViews calls the given iterator passing in each attribute sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateViews returns that
// error.
func (m *MediaTypeDefinition) IterateViews(it ViewIterator) error {
	o := m.Views
	// gather names and sort them
	names := make([]string, len(o))
	i := 0
	for n := range o {
		names[i] = n
		i++
	}
	sort.Strings(names)
	// iterate
	for _, n := range names {
		if err := it(o[n]); err != nil {
			return err
		}
	}
	return nil
}

// Project creates a MediaTypeDefinition containing the fields defined in the given view.  The
// resuling media type only defines the default view and its identifier is modified to indicate that
// it was projected by adding the view as id parameter.  links is a user type of type Object where
// each key corresponds to a linked media type as defined by the media type "links" attribute.
func (m *MediaTypeDefinition) Project(view string) (*MediaTypeDefinition, *UserTypeDefinition, error) {
	canonical := m.projectCanonical(view)
	if p, ok := ProjectedMediaTypes[canonical]; ok {
		var links *UserTypeDefinition
		mLinks := ProjectedMediaTypes[canonical+"; links"]
		if mLinks != nil {
			links = mLinks.UserTypeDefinition
		}
		return p, links, nil
	}
	if m.IsArray() {
		return m.projectCollection(view)
	}
	return m.projectSingle(view, canonical)
}

func (m *MediaTypeDefinition) projectSingle(view, canonical string) (p *MediaTypeDefinition, links *UserTypeDefinition, err error) {
	v, ok := m.Views[view]
	if !ok {
		return nil, nil, fmt.Errorf("unknown view %#v", view)
	}
	viewObj := v.Type.ToObject()

	// Compute validations - view may not have all attributes
	var val *dslengine.ValidationDefinition
	if m.Validation != nil {
		names := m.Validation.Required
		var required []string
		for _, n := range names {
			if _, ok := viewObj[n]; ok {
				required = append(required, n)
			}
		}
		val = m.Validation.Dup()
		val.Required = required
	}

	// Compute description
	desc := m.Description
	if desc == "" {
		desc = m.TypeName + " media type"
	}
	desc += " (" + view + " view)"

	p = &MediaTypeDefinition{
		Identifier: m.projectIdentifier(view),
		UserTypeDefinition: &UserTypeDefinition{
			TypeName: m.projectTypeName(view),
			AttributeDefinition: &AttributeDefinition{
				Description: desc,
				Type:        Dup(v.Type),
				Validation:  val,
			},
		},
	}
	p.Views = map[string]*ViewDefinition{"default": {
		Name:                "default",
		AttributeDefinition: DupAtt(v.AttributeDefinition),
		Parent:              p,
	}}

	ProjectedMediaTypes[canonical] = p
	projectedObj := p.Type.ToObject()
	mtObj := m.Type.ToObject()
	_, hasAttNamedLinks := mtObj["links"]
	for n := range viewObj {
		if n == "links" && !hasAttNamedLinks {
			linkObj := make(Object)
			for n, link := range m.Links {
				linkView := link.View
				if linkView == "" {
					linkView = "link"
				}
				mtAtt, ok := mtObj[n]
				if !ok {
					return nil, nil, fmt.Errorf("unknown attribute %#v used in links", n)
				}
				mtt := mtAtt.Type.(*MediaTypeDefinition)
				vl, _, err := mtt.Project(linkView)
				if err != nil {
					return nil, nil, err
				}
				linkObj[n] = &AttributeDefinition{Type: vl, Validation: mtt.Validation, Metadata: mtAtt.Metadata}
			}
			lTypeName := fmt.Sprintf("%sLinks", m.TypeName)
			links = &UserTypeDefinition{
				AttributeDefinition: &AttributeDefinition{
					Description: fmt.Sprintf("%s contains links to related resources of %s.", lTypeName, m.TypeName),
					Type:        linkObj,
				},
				TypeName: lTypeName,
			}
			projectedObj[n] = &AttributeDefinition{Type: links, Description: "Links to related resources"}
			ProjectedMediaTypes[canonical+"; links"] = &MediaTypeDefinition{UserTypeDefinition: links}
		} else {
			if at := mtObj[n]; at != nil {
				at = DupAtt(at)
				if mt, ok := at.Type.(*MediaTypeDefinition); ok {
					vatt := viewObj[n]
					view := vatt.View
					if view == "" {
						view = at.View
					}
					if view == "" {
						view = DefaultView
					}
					pr, _, err := mt.Project(view)
					if err != nil {
						return nil, nil, fmt.Errorf("view %#v on field %#v cannot be computed: %s", view, n, err)
					}
					at.Type = pr
					// Force example to be generated again
					// since set of attributes has changed
					at.Example = nil
				}
				projectedObj[n] = at
			}
		}
	}
	return
}

func (m *MediaTypeDefinition) projectCollection(view string) (*MediaTypeDefinition, *UserTypeDefinition, error) {
	// Project the collection element media type
	e := m.ToArray().ElemType.Type.(*MediaTypeDefinition) // validation checked this cast would work
	pe, le, err2 := e.Project(view)
	if err2 != nil {
		return nil, nil, fmt.Errorf("collection element: %s", err2)
	}

	// Build the projected collection with the results
	desc := m.TypeName + " is the media type for an array of " + e.TypeName + " (" + view + " view)"
	p := &MediaTypeDefinition{
		Identifier: m.projectIdentifier(view),
		UserTypeDefinition: &UserTypeDefinition{
			AttributeDefinition: &AttributeDefinition{
				Description: desc,
				Type:        &Array{ElemType: &AttributeDefinition{Type: pe}},
				Example:     nil,
			},
			TypeName: pe.TypeName + "Collection",
		},
	}
	p.Views = map[string]*ViewDefinition{"default": &ViewDefinition{
		AttributeDefinition: DupAtt(pe.Views["default"].AttributeDefinition),
		Name:                "default",
		Parent:              p,
	}}

	// Run the DSL that was created by the CollectionOf function
	if !dslengine.Execute(p.DSL(), p) {
		return nil, nil, dslengine.Errors
	}

	// Build the links user type
	var links *UserTypeDefinition
	if le != nil {
		lTypeName := le.TypeName + "Array"
		links = &UserTypeDefinition{
			AttributeDefinition: &AttributeDefinition{
				Type:        &Array{ElemType: &AttributeDefinition{Type: le}},
				Description: fmt.Sprintf("%s contains links to related resources of %s.", lTypeName, m.TypeName),
			},
			TypeName: lTypeName,
		}
	}

	return p, links, nil
}

// projectIdentifier computes the projected media type identifier by adding the "view" param.  We
// need the projected media type identifier to be different so that looking up projected media types
// from ProjectedMediaTypes works correctly. It's also good for clients.
func (m *MediaTypeDefinition) projectIdentifier(view string) string {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		base = m.Identifier
	}
	params["view"] = view
	return mime.FormatMediaType(base, params)
}

// projectIdentifier computes the projected canonical media type identifier by adding the "view"
// param if the view is not the default view.
func (m *MediaTypeDefinition) projectCanonical(view string) string {
	cano := CanonicalIdentifier(m.Identifier)
	base, params, _ := mime.ParseMediaType(cano)
	if params["view"] != "" {
		return cano // Already projected
	}
	params["view"] = view
	return mime.FormatMediaType(base, params)
}

// projectTypeName appends the view name to the media type name if the view name is not "default".
func (m *MediaTypeDefinition) projectTypeName(view string) string {
	typeName := m.TypeName
	if view != "default" {
		typeName += strings.Title(view)
	}
	return typeName
}

// DataStructure implementation

// Definition returns the underlying attribute definition.
// Note that this function is "inherited" by both UserTypeDefinition and
// MediaTypeDefinition.
func (a *AttributeDefinition) Definition() *AttributeDefinition {
	return a
}

// Walk traverses the data structure recursively and calls the given function once
// on each attribute starting with the attribute returned by Definition.
func (a *AttributeDefinition) Walk(walker func(*AttributeDefinition) error) error {
	return walk(a, walker, make(map[string]bool))
}

// Walk traverses the data structure recursively and calls the given function once
// on each attribute starting with the attribute returned by Definition.
func (u *UserTypeDefinition) Walk(walker func(*AttributeDefinition) error) error {
	return walk(u.AttributeDefinition, walker, map[string]bool{u.TypeName: true})
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite recursions by
// keeping track of types that have already been walked.
func walk(at *AttributeDefinition, walker func(*AttributeDefinition) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut *UserTypeDefinition) error {
		if _, ok := seen[ut.TypeName]; ok {
			return nil
		}
		seen[ut.TypeName] = true
		return walk(ut.AttributeDefinition, walker, seen)
	}
	switch actual := at.Type.(type) {
	case Primitive:
		return nil
	case *Array:
		return walk(actual.ElemType, walker, seen)
	case *Hash:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case Object:
		for _, cat := range actual {
			if err := walk(cat, walker, seen); err != nil {
				return err
			}
		}
	case *UserTypeDefinition:
		return walkUt(actual)
	case *MediaTypeDefinition:
		return walkUt(actual.UserTypeDefinition)
	default:
		panic("unknown attribute type") // bug
	}
	return nil
}

// toReflectType converts the DataType to reflect.Type.
func toReflectType(dtype DataType) reflect.Type {
	switch dtype.Kind() {
	case BooleanKind:
		return reflect.TypeOf(true)
	case IntegerKind:
		return reflect.TypeOf(int(0))
	case NumberKind:
		return reflect.TypeOf(float64(0))
	case UUIDKind, StringKind:
		return reflect.TypeOf("")
	case DateTimeKind:
		return reflect.TypeOf(time.Time{})
	case ObjectKind, UserTypeKind, MediaTypeKind:
		return reflect.TypeOf(map[string]interface{}{})
	case ArrayKind:
		return reflect.SliceOf(toReflectType(dtype.ToArray().ElemType.Type))
	case HashKind:
		hash := dtype.ToHash()
		// avoid complication: not allow object as the hash key
		var ktype reflect.Type
		if !hash.KeyType.Type.IsObject() {
			ktype = toReflectType(hash.KeyType.Type)
		} else {
			ktype = reflect.TypeOf([]interface{}{}).Elem()
		}
		return reflect.MapOf(ktype, toReflectType(hash.ElemType.Type))
	default:
		return reflect.TypeOf([]interface{}{}).Elem()
	}
}
