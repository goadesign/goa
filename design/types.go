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
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/goadesign/goa/dslengine"
)

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
		GenerateExample(r *RandomGenerator) interface{}
	}

	// DataStructure is the interface implemented by all data structure types.
	// That is attribute definitions, user types and media types.
	DataStructure interface {
		// Definition returns the data structure definition.
		Definition() *AttributeDefinition
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
	case String:
		return "string"
	case DateTime:
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
	case Boolean, Integer, Number, String:
		ok = true
	}
	return
}

// IsCompatible returns true if val is compatible with p.
func (p Primitive) IsCompatible(val interface{}) (ok bool) {
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
	}
	return false
}

var anyPrimitive = []Primitive{Boolean, Integer, Number, DateTime}

// GenerateExample returns an instance of the given data type.
func (p Primitive) GenerateExample(r *RandomGenerator) interface{} {
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
	case Any:
		// to not make it too complicated, pick one of the primitive types
		return anyPrimitive[r.Int()%len(anyPrimitive)].GenerateExample(r)
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
		if !a.ElemType.Type.IsCompatible(v.Index(i).Interface()) {
			return false
		}
	}
	return true
}

// GenerateExample produces a random array value.
func (a *Array) GenerateExample(r *RandomGenerator) interface{} {
	count := r.Int()%3 + 1
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = a.ElemType.Type.GenerateExample(r)
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
func (o Object) GenerateExample(r *RandomGenerator) interface{} {
	// ensure fixed ordering
	keys := make([]string, 0, len(o))
	for n := range o {
		keys = append(keys, n)
	}
	sort.Strings(keys)

	res := make(map[string]interface{})
	for _, n := range keys {
		att := o[n]
		res[n] = att.Type.GenerateExample(r)
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
		if !h.KeyType.Type.IsCompatible(key.Interface()) || !h.ElemType.Type.IsCompatible(v.MapIndex(key).Interface()) {
			return false
		}
	}
	return true
}

// GenerateExample returns a random hash value.
func (h *Hash) GenerateExample(r *RandomGenerator) interface{} {
	count := r.Int()%3 + 1
	pair := map[interface{}]interface{}{}
	for i := 0; i < count; i++ {
		pair[h.KeyType.Type.GenerateExample(r)] = h.ElemType.Type.GenerateExample(r)
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
func (u *UserTypeDefinition) IsPrimitive() bool { return u.Type.IsPrimitive() }

// HasAttributes calls the HasAttributes on the user type underlying data type.
func (u *UserTypeDefinition) HasAttributes() bool { return u.Type.HasAttributes() }

// IsObject calls IsObject on the user type underlying data type.
func (u *UserTypeDefinition) IsObject() bool { return u.Type.IsObject() }

// IsArray calls IsArray on the user type underlying data type.
func (u *UserTypeDefinition) IsArray() bool { return u.Type.IsArray() }

// IsHash calls IsHash on the user type underlying data type.
func (u *UserTypeDefinition) IsHash() bool { return u.Type.IsHash() }

// ToObject calls ToObject on the user type underlying data type.
func (u *UserTypeDefinition) ToObject() Object { return u.Type.ToObject() }

// ToArray calls ToArray on the user type underlying data type.
func (u *UserTypeDefinition) ToArray() *Array { return u.Type.ToArray() }

// ToHash calls ToHash on the user type underlying data type.
func (u *UserTypeDefinition) ToHash() *Hash { return u.Type.ToHash() }

// CanHaveDefault calls CanHaveDefault on the user type underlying data type.
func (u *UserTypeDefinition) CanHaveDefault() bool { return u.Type.CanHaveDefault() }

// IsCompatible returns true if val is compatible with p.
func (u *UserTypeDefinition) IsCompatible(val interface{}) bool {
	return u.Type.IsCompatible(val)
}

// Finalize merges base type attributes.
func (u *UserTypeDefinition) Finalize() {
	if u.Reference != nil {
		if bat := u.AttributeDefinition; bat != nil {
			u.AttributeDefinition.Inherit(bat)
		}
	}

	u.finalizeExample(nil)
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

// IsBuiltIn returns true if the media type is implemented via a goa struct.
func (m *MediaTypeDefinition) IsBuiltIn() bool {
	return m == ErrorMedia
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

// Project creates a MediaTypeDefinition derived from the given definition that matches the given
// view.
func (m *MediaTypeDefinition) Project(view string) (p *MediaTypeDefinition, links *UserTypeDefinition, err error) {
	if _, ok := m.Views[view]; !ok {
		return nil, nil, fmt.Errorf("unknown view %#v", view)
	}
	if m.IsArray() {
		return m.projectCollection(view)
	}
	if m.Type.ToObject() == nil {
		return m, nil, nil
	}
	return m.projectSingle(view)
}

func (m *MediaTypeDefinition) projectSingle(view string) (p *MediaTypeDefinition, links *UserTypeDefinition, err error) {
	v := m.Views[view]
	canonical := CanonicalIdentifier(m.Identifier)
	typeName := m.TypeName
	if view != "default" {
		typeName += strings.Title(view)
		canonical += "; view=" + view
	}
	var ok bool
	if p, ok = GeneratedMediaTypes[canonical]; ok {
		mLinks := GeneratedMediaTypes[canonical+"; links"]
		if mLinks != nil {
			links = mLinks.UserTypeDefinition
		}
		return
	}

	// Compute validations - view may not have all attributes
	viewObj := v.Type.ToObject()
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
	p = &MediaTypeDefinition{
		Identifier: m.Identifier,
		UserTypeDefinition: &UserTypeDefinition{
			TypeName: typeName,
			AttributeDefinition: &AttributeDefinition{
				Type:       Dup(v.Type),
				Validation: val,
			},
		},
	}
	GeneratedMediaTypes[canonical] = p
	projectedObj := p.Type.ToObject()
	mtObj := m.Type.ToObject()
	for n := range viewObj {
		if n == "links" {
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
				vl, _, err := mtAtt.Type.(*MediaTypeDefinition).Project(linkView)
				if err != nil {
					return nil, nil, err
				}
				linkObj[n] = &AttributeDefinition{Type: vl}
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
			GeneratedMediaTypes[canonical+"; links"] = &MediaTypeDefinition{UserTypeDefinition: links}
		} else {
			if at := mtObj[n]; at != nil {
				if at.View != "" {
					m, ok := at.Type.(*MediaTypeDefinition)
					if !ok {
						return nil, nil, fmt.Errorf("View specified on non media type attribute %#v", n)
					}
					pr, _, err := m.Project(at.View)
					if err != nil {
						return nil, nil, fmt.Errorf("view %#v on field %#v cannot be computed: %s", at.View, n, err)
					}
					at.Type = pr
				}
				projectedObj[n] = at
			}
		}
	}
	return
}

func (m *MediaTypeDefinition) projectCollection(view string) (p *MediaTypeDefinition, links *UserTypeDefinition, err error) {
	e := m.ToArray().ElemType.Type.(*MediaTypeDefinition) // validation checked this cast would work
	pe, le, err2 := e.Project(view)
	if err2 != nil {
		return nil, nil, fmt.Errorf("collection element: %s", err2)
	}
	p = &MediaTypeDefinition{
		Identifier: m.Identifier,
		UserTypeDefinition: &UserTypeDefinition{
			AttributeDefinition: &AttributeDefinition{
				Type: &Array{ElemType: &AttributeDefinition{Type: pe}},
			},
			TypeName: pe.TypeName + "Collection",
		},
	}
	if !dslengine.Execute(p.DSL(), p) {
		return nil, nil, dslengine.Errors
	}
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
	return
}

// DataStructure implementation

// Definition returns the underlying attribute definition.
// Note that this function is "inherited" by both UserTypeDefinition and
// MediaTypeDefinition.
func (a *AttributeDefinition) Definition() *AttributeDefinition {
	return a
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
	case StringKind:
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
