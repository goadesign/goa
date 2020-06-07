package openapiv3

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

type (
	// EndpointBodies describes the request and response HTTP bodies of an endpoint
	// using JSON schema. Each body may be described via a reference to a schema
	// described in the "Components" section of the OpenAPI document or an actual
	// JSON schema data structure. There may also be additional notes attached to
	// each body definition to account for cases that are not directly supported in
	// OpenAPI such as streaming. The possible response bodies are indexed by HTTP
	// status, there may be more than one when the result type defined multiple
	// views.
	EndpointBodies struct {
		RequestBody    *openapi.Schema
		ResponseBodies map[int][]*openapi.Schema
	}

	// schemafier is an internal data structure used to keep the state required to
	// create JSON schemas for all the request and response body types.
	schemafier struct {
		// type schemas indexed by ref
		schemas map[string]*openapi.Schema
		// type names indexed by hashes
		hashes map[uint64]string
		rand   *expr.Random
	}
)

// newSchemafier initializes a schemafier.
func newSchemafier(rand *expr.Random) *schemafier {
	return &schemafier{
		schemas: make(map[string]*openapi.Schema),
		hashes:  make(map[uint64]string),
		rand:    rand,
	}
}

// buildBodyTypes traverses the design and builds the JSON schemas that
// represent the request and response bodies of each endpoint. The algorithm
// also computes a good unique name for the different types making sure that two
// types that are actually identical share the same name. This is to handle
// properly the data structures created by the code generation algorithms which
// can duplicate types (for example if they are defined inline in the design).
// The result is a map of method details indexed by service name. Each method
// detail is in turn indexed by method name. The details contain JSON schema
// references and the actual JSON schemas are returned in the second result
// value indexed by reference name.
func buildBodyTypes(api *expr.APIExpr) (map[string]map[string]*EndpointBodies, map[string]*openapi.Schema) {
	bodies := make(map[string]map[string]*EndpointBodies)
	sf := newSchemafier(api.Random())
	for _, s := range api.HTTP.Services {
		errors := make(map[int]*openapi.Schema)
		for _, e := range s.HTTPErrors {
			errors[e.Response.StatusCode] = sf.schemafy(e.Response.Body)
		}
		sbodies := make(map[string]*EndpointBodies, len(s.HTTPEndpoints))
		for _, e := range s.HTTPEndpoints {
			req := sf.schemafy(e.Body)
			if e.StreamingBody != nil {
				sreq := sf.schemafy(e.StreamingBody)
				var note string
				if sreq.Ref != "" {
					note = sreq.Ref
				} else {
					note = string(sreq.Type)
				}
				if req.Description != "" {
					req.Description += "\n"
				}
				req.Description += fmt.Sprintf("Streaming body: %s", note)
			}
			res := make(map[int][]*openapi.Schema)
			for c, er := range errors {
				res[c] = []*openapi.Schema{er}
			}
			for _, resp := range e.Responses {
				var view string
				if vs, ok := resp.Body.Meta["view"]; ok {
					view = vs[0]
				}
				body := resp.Body
				if view != "" {
					// Static view
					rt, err := expr.Project(body.Type.(*expr.ResultTypeExpr), view)
					if err != nil { // bug
						panic(fmt.Sprintf("failed to project %q to view %q", body.Type.Name(), view))
					}
					body.Type = rt
				}
				js := sf.schemafy(body)
				if rt, ok := resp.Body.Type.(*expr.ResultTypeExpr); ok {
					if view == "" && rt.HasMultipleViews() {
						// Dynamic views
						if len(js.Description) > 0 {
							js.Description += "\n"
						}
						js.Description += sf.viewsNote(rt)
					}
				}
				res[resp.StatusCode] = append(res[resp.StatusCode], js)
			}
			sbodies[e.Name()] = &EndpointBodies{req, res}
		}
		bodies[s.Name()] = sbodies
	}
	return bodies, sf.schemas
}

func (sf *schemafier) schemafy(attr *expr.AttributeExpr) *openapi.Schema {
	s := openapi.NewSchema()
	var note string

	// Initialize type and format
	switch t := attr.Type.(type) {
	case expr.Primitive:
		switch t.Kind() {
		case expr.UIntKind, expr.UInt64Kind, expr.UInt32Kind:
			s.Type = openapi.Type("integer")
		case expr.IntKind, expr.Int64Kind:
			s.Type = openapi.Type("integer")
			s.Format = "int64"
		case expr.Int32Kind:
			s.Type = openapi.Type("integer")
			s.Format = "int32"
		case expr.Float32Kind:
			s.Type = openapi.Type("number")
			s.Format = "float"
		case expr.Float64Kind:
			s.Type = openapi.Type("number")
			s.Format = "double"
		case expr.BytesKind, expr.AnyKind:
			s.Type = openapi.Type("string")
			s.Format = "binary"
		default:
			s.Type = openapi.Type(t.Name())
		}
	case *expr.Array:
		s.Type = openapi.Array
		s.Items = sf.schemafy(t.ElemType)
	case *expr.Object:
		s.Type = openapi.Object
		var itemNotes []string
		for _, nat := range *t {
			s.Properties[nat.Name] = sf.schemafy(nat.Attribute)
		}
		if len(itemNotes) > 0 {
			note = strings.Join(itemNotes, "\n")
		}
	case *expr.Map:
		s.Type = openapi.Object
		s.AdditionalProperties = true
	case expr.UserType:
		h := hashAttribute(attr, fnv.New64())
		if ref, ok := sf.hashes[h]; ok {
			s.Ref = ref
		} else {
			s.Ref = sf.uniquify(fmt.Sprintf("#/components/schemas/%s", codegen.Goify(t.Name(), true)))
			sf.hashes[h] = s.Ref
			sf.schemas[s.Ref] = s
		}
	default:
		panic(fmt.Sprintf("unknown type %T", t)) // bug
	}
	s.Description = attr.Description
	if note != "" {
		s.Description += "\n" + note
	}

	// Default value, example, extensions
	s.DefaultValue = toStringMap(attr.DefaultValue)
	s.Example = attr.Example(sf.rand)
	s.Extensions = openapi.ExtensionsFromExpr(attr.Meta)

	// Validations
	val := attr.Validation
	if val == nil {
		return s
	}
	s.Enum = val.Values
	s.Format = string(val.Format)
	s.Pattern = val.Pattern
	if val.Minimum != nil {
		s.Minimum = val.Minimum
	}
	if val.Maximum != nil {
		s.Maximum = val.Maximum
	}
	if val.MinLength != nil {
		if _, ok := attr.Type.(*expr.Array); ok {
			s.MinItems = val.MinLength
		} else {
			s.MinLength = val.MinLength
		}
	}
	if val.MaxLength != nil {
		if _, ok := attr.Type.(*expr.Array); ok {
			s.MaxItems = val.MaxLength
		} else {
			s.MaxLength = val.MaxLength
		}
	}
	s.Required = val.Required

	return s
}

// uniquify returns n if n is not a known type name. Otherwise uniquify appends
// the smallest integer greater than 1 to n so the result is not a known type
// name.
func (sf *schemafier) uniquify(n string) string {
	exists := func(n string) bool {
		_, ok := sf.schemas[n]
		return ok
	}
	i := 1
	for exists(n) {
		i++
		n = strings.TrimRight(n, "0123456789") + strconv.Itoa(i)
	}
	return n
}

// viewsNote returns a human friendly description of the different possible
// response body shapes for the different views supported by the attribute type
// which must be a ResultType.
func (sf *schemafier) viewsNote(rt *expr.ResultTypeExpr) string {
	var alts []string
	for _, v := range rt.Views {
		if v.Name != expr.DefaultView {
			pr, err := expr.Project(rt, v.Name)
			if err != nil {
				panic(fmt.Sprintf("failed to project %q with view %q", rt.Name(), v.Name)) // bug, DSL should have performed validations
			}
			js := sf.schemafy(&expr.AttributeExpr{Type: pr})
			alts = append(alts, js.Ref)
		}
	}
	oneof := ""
	last := ""
	if len(alts) > 1 {
		oneof = "one of "
		last = " or " + alts[len(alts)-1]
		alts = alts[:len(alts)-1]
	}
	return "Response body may alternatively be " + oneof + strings.Join(alts, ", ") + last
}

// toStringMap converts map[interface{}]interface{} to a map[string]interface{}
// when possible.
func toStringMap(val interface{}) interface{} {
	switch actual := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range actual {
			m[toString(k)] = toStringMap(v)
		}
		return m
	case []interface{}:
		mapSlice := make([]interface{}, len(actual))
		for i, e := range actual {
			mapSlice[i] = toStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// toString returns the string representation of the given type.
func toString(val interface{}) string {
	switch actual := val.(type) {
	case string:
		return actual
	case int:
		return strconv.Itoa(actual)
	case float64:
		return strconv.FormatFloat(actual, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(actual)
	default:
		panic("unexpected key type")
	}
}

// hashAttribute is helper function that computes a unique hash for the given
// attribute type. The algorithm returns the same value for two attributes whose
// types are structurally equivalent unless they are result types with different
// identifiers. Structurally identical means same primitive types, arrays with
// structurally equivalent element types, maps with structurally equivalent key
// and value types or object with identical attribute names and structurally
// equivalent types and identical set of required attributes.
func hashAttribute(att *expr.AttributeExpr, h hash.Hash64) uint64 {
	t := att.Type
	switch t.Kind() {

	case expr.ObjectKind:
		o := expr.AsObject(t)
		keys := make([]string, len(*o))
		for i, m := range *o {
			keys[i] = m.Name
		}
		sort.Strings(keys)
		var res uint64
		for _, k := range keys {
			kh := hashString(k, h)
			vh := hashAttribute(o.Attribute(k), h)
			res = res ^ orderedHash(kh, vh, h)
		}
		// Objects with a different set of required attributes should produce
		// different hashes.
		if att.Validation != nil {
			for _, req := range att.Validation.Required {
				rh := hashString(req, h)
				res = res ^ rh
			}
		}
		return res

	case expr.ArrayKind:
		kh := hashString("[]", h)
		vh := hashAttribute(expr.AsArray(t).ElemType, h)
		return orderedHash(kh, vh, h)

	case expr.MapKind:
		m := expr.AsMap(t)
		kh := hashAttribute(m.KeyType, h)
		vh := hashAttribute(m.ElemType, h)
		return orderedHash(kh, vh, h)

	case expr.UserTypeKind:
		return hashAttribute(t.(expr.UserType).Attribute(), h)

	case expr.ResultTypeKind:
		// The identifier specified in the design for result types should drive
		// the computation of the hash.
		rt := t.(*expr.ResultTypeExpr)
		res := hashString(rt.Identifier, h)
		view := rt.AttributeExpr.Meta["view"]
		if len(view) > 0 {
			return orderedHash(res, hashString(view[0], h), h)
		}
		return res

	default: // Primitives or Any
		return hashString(t.Name(), h)
	}
}

func hashString(s string, h hash.Hash64) uint64 {
	h.Reset()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err) // should not fail
	}
	return h.Sum64()
}

func orderedHash(a, b uint64, h hash.Hash64) uint64 {
	h.Reset()
	if err := binary.Write(h, binary.LittleEndian, a); err != nil {
		panic(err) // should not fail
	}
	if err := binary.Write(h, binary.LittleEndian, b); err != nil {
		panic(err) // should not fail
	}
	return h.Sum64()
}
