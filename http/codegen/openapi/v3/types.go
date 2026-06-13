package openapiv3

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/gohugoio/hashstructure"

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
		// SSEItemSchema describes a single event streamed by a server-sent
		// events endpoint. It is only computed for OpenAPI 3.2 documents
		// which render it as the itemSchema of the text/event-stream media
		// type.
		SSEItemSchema *openapi.Schema
	}

	// schemafier is an internal data structure used to keep the state required to
	// create JSON schemas for all the request and response body types.
	schemafier struct {
		// type schemas indexed by ref
		schemas map[string]*openapi.Schema
		// type names indexed by hashes
		hashes map[uint64][]string
		rand   *expr.ExampleGenerator
		// nameAliases generates named component schemas for primitive alias
		// types instead of inlining them. Only set when the schemas map feeds
		// the document components (OpenAPI 3.2 documents): schemafiers whose
		// schemas are discarded (parameters, headers) must keep inlining
		// aliases as their references would dangle.
		nameAliases bool
	}
)

// derived returns a schemafier drawing example values from a stream derived
// from the given identity, sharing all other state. See
// expr.ExampleGenerator.Derived.
func (sf *schemafier) derived(id string) *schemafier {
	c := *sf
	c.rand = sf.rand.Derived(id)
	return &c
}

// rebased returns a schemafier whose example value stream is anchored to the
// given absolute design identity, sharing all other state. See
// expr.ExampleGenerator.Rebased.
func (sf *schemafier) rebased(id string) *schemafier {
	c := *sf
	c.rand = sf.rand.Rebased(id)
	return &c
}

// bodyExampleID returns the absolute design identity anchoring the example
// streams of an endpoint request or response body. Anonymous body types
// (inline arrays, maps and primitives) have no type identity of their own so
// their examples anchor on the endpoint that owns them.
func bodyExampleID(svc, endpoint, role string) string {
	return svc + "." + endpoint + "." + role
}

// fieldOf returns a schemafier whose example value stream is anchored to the
// identity of the named field of the given parent attribute, sharing all
// other state. See expr.ExampleGenerator.Field.
func (sf *schemafier) fieldOf(parent *expr.AttributeExpr, name string) *schemafier {
	c := *sf
	c.rand = sf.rand.Field(parent, name)
	return &c
}

// newSchemafier initializes a schemafier.
func newSchemafier(rand *expr.ExampleGenerator) *schemafier {
	return &schemafier{
		schemas: make(map[string]*openapi.Schema),
		hashes:  make(map[uint64][]string),
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
// value indexed by type name.
//
// NOTE: entries are nil when the corresponding type is Empty.
func buildBodyTypes(api *expr.APIExpr, types []expr.UserType, resultTypes []*expr.ResultTypeExpr, ver openapi.Version) (map[string]map[string]*EndpointBodies, map[string]*openapi.Schema) {
	bodies := make(map[string]map[string]*EndpointBodies)
	sf := newSchemafier(api.ExampleGenerator)
	sf.nameAliases = ver == openapi.Version32
	services := openAPIGeneratedServices(api)

	// Generates the types referenced from the endpoints.
	for _, t := range types {
		if !mustGenerateType(t.Attribute().Meta, services) {
			continue
		}
		sf.schemafy(&expr.AttributeExpr{Type: t})
	}
	for _, t := range resultTypes {
		if !mustGenerateType(t.Attribute().Meta, services) {
			continue
		}
		sf.schemafy(&expr.AttributeExpr{Type: t})
	}

	for _, s := range api.HTTP.Services {
		if !openapi.MustGenerate(s.Meta) || !openapi.MustGenerate(s.ServiceExpr.Meta) {
			continue
		}

		sbodies := make(map[string]*EndpointBodies, len(s.HTTPEndpoints))
		for _, e := range s.HTTPEndpoints {
			if !openapi.MustGenerate(e.Meta) || !openapi.MustGenerate(e.MethodExpr.Meta) {
				continue
			}

			reqBody := e.Body
			if reqBody.Type != expr.Empty && reqBody.Description == "" {
				if _, ok := reqBody.Type.(expr.UserType); ok {
					reqBody = expr.DupAtt(reqBody)
					reqBody.Description = defaultRequestBodyDescription(e)
				}
			}
			req := sf.rebased(bodyExampleID(s.Name(), e.Name(), "request")).schemafy(reqBody)
			if e.StreamingBody != nil {
				sreq := sf.schemafy(e.StreamingBody)
				var note string
				if sreq.Ref != "" {
					note = sreq.Ref
				} else {
					note = string(sreq.Type)
				}
				if req == nil {
					req = sreq
					if req.Description != "" {
						req.Description += "\n"
					}
					req.Description += "Streaming body."
				} else {
					if req.Description != "" {
						req.Description += "\n"
					}
					req.Description += fmt.Sprintf("Streaming body: %s", note)
				}
			}
			res := make(map[int][]*openapi.Schema)
			resps := e.Responses
			for _, er := range e.HTTPErrors {
				resps = append(resps, er.Response)
			}
			for i, resp := range resps {
				id := bodyExampleID(s.Name(), e.Name(), "response."+strconv.Itoa(resp.StatusCode)+"."+strconv.Itoa(i))
				js := sf.rebased(id).schemafy(staticViewBody(resp))
				res[resp.StatusCode] = append(res[resp.StatusCode], js)
			}
			eb := &EndpointBodies{RequestBody: req, ResponseBodies: res}
			if ver == openapi.Version32 && e.UsesSSE() {
				eb.SSEItemSchema = sf.buildSSEItemSchema(e)
			}
			sbodies[e.Name()] = eb
		}
		bodies[s.Name()] = sbodies
	}
	return bodies, sf.schemas
}

// buildSSEItemSchema returns the JSON schema describing a single event
// streamed by the given server-sent events endpoint as defined by the OpenAPI
// 3.2 sequential media types. The schema is an object whose properties mirror
// the SSE event fields mapped by the design: data is always present, event,
// id and retry only when the design maps them. String and bytes data is
// written raw on the wire while other types are JSON-encoded, which the data
// property reflects using the JSON schema contentMediaType and contentSchema
// keywords.
func (sf *schemafier) buildSSEItemSchema(e *expr.HTTPEndpointExpr) *openapi.Schema {
	sse := e.SSE
	sr := e.MethodExpr.StreamingResult
	data := sr
	dsf := sf
	if sse.DataField != "" {
		data = expr.AsObject(sr.Type).Attribute(sse.DataField)
		dsf = sf.fieldOf(sr, sse.DataField)
	}
	var dataSchema *openapi.Schema
	switch data.Type {
	case expr.String, expr.Bytes:
		dataSchema = dsf.schemafy(data)
	default:
		dataSchema = &openapi.Schema{
			Type:             openapi.String,
			ContentMediaType: "application/json",
			ContentSchema:    dsf.schemafy(data),
		}
	}
	props := map[string]*openapi.Schema{"data": dataSchema}
	if sse.EventField != "" {
		props["event"] = sf.fieldOf(sr, sse.EventField).schemafy(expr.AsObject(sr.Type).Attribute(sse.EventField))
	}
	if sse.IDField != "" {
		props["id"] = sf.fieldOf(sr, sse.IDField).schemafy(expr.AsObject(sr.Type).Attribute(sse.IDField))
	}
	if sse.RetryField != "" {
		props["retry"] = sf.fieldOf(sr, sse.RetryField).schemafy(expr.AsObject(sr.Type).Attribute(sse.RetryField))
	}
	return &openapi.Schema{
		Type:       openapi.Object,
		Properties: props,
		Required:   []string{"data"},
	}
}

// staticViewBody returns the response body attribute used to compute the
// OpenAPI schema and examples. When the design pins the response to a single
// view the result type is projected onto a detached copy of the body: the
// design expression tree is read-only for the generators.
func staticViewBody(resp *expr.HTTPResponseExpr) *expr.AttributeExpr {
	view, ok := resp.Body.Meta.Last(expr.ViewMetaKey)
	if !ok || view == "" {
		return resp.Body
	}
	body := expr.DupAtt(resp.Body)
	rt, err := expr.Project(body.Type.(*expr.ResultTypeExpr), view)
	if err != nil {
		panic(fmt.Sprintf("failed to project %q to view %q", body.Type.Name(), view)) // bug
	}
	body.Type = rt
	return body
}

func (sf *schemafier) schemafy(attr *expr.AttributeExpr, noref ...bool) *openapi.Schema {
	if attr.Type == expr.Empty {
		return nil
	}

	s := openapi.NewSchema()
	var note string

	// Initialize type and format
	switch t := attr.Type.(type) {
	case expr.Primitive:
		switch t.Kind() {
		case expr.IntKind, expr.UIntKind, expr.Int64Kind, expr.UInt64Kind:
			// Use int64 format for IntKind and UIntKind because the OpenAPI
			// generator produced int32 by default.
			s.Type = openapi.Type("integer")
			s.Format = "int64"
		case expr.Int32Kind, expr.UInt32Kind:
			s.Type = openapi.Type("integer")
			s.Format = "int32"
		case expr.Float32Kind:
			s.Type = openapi.Type("number")
			s.Format = "float"
		case expr.Float64Kind:
			s.Type = openapi.Type("number")
			s.Format = "double"
		case expr.BytesKind:
			if bases := attr.Bases; len(bases) > 0 {
				for _, b := range bases {
					// Union type
					val := sf.schemafy(&expr.AttributeExpr{Type: b}, false)
					s.AnyOf = append(s.AnyOf, val)
				}
			} else {
				s.Type = openapi.Type("string")
				s.Format = "binary"
			}
		case expr.AnyKind:
			// A schema without a type matches any data type.
			// See https://swagger.io/docs/specification/data-models/data-types/#any.
			s.Type = openapi.Type("")
		default:
			s.Type = openapi.Type(t.Name())
		}
	case *expr.Array:
		s.Type = openapi.Array
		s.Items = sf.derived("0").schemafy(t.ElemType)
	case *expr.Object:
		s.Type = openapi.Object
		var itemNotes []string
		for _, nat := range *t {
			if !openapi.MustGenerate(nat.Attribute.Meta) {
				continue
			}
			s.Properties[nat.Name] = sf.derived(nat.Name).schemafy(nat.Attribute)
		}
		if len(itemNotes) > 0 {
			note = strings.Join(itemNotes, "\n")
		}
	case *expr.Map:
		s.Type = openapi.Object
		if t.ElemType.Type == expr.Any {
			// Use free-form objects when elements are of type "Any", otherwise, use full schema
			// See https://swagger.io/docs/specification/data-models/dictionaries/.
			s.AdditionalProperties = true
		} else {
			s.AdditionalProperties = sf.derived("val0").schemafy(t.ElemType)
		}
	case *expr.Union:
		// Represent unions as an object with discriminator and value fields.
		// The field names are configurable via Meta tags (defaults: "type" and "value").
		typeKey := t.GetTypeKey()
		valueKey := t.GetValueKey()

		s.Type = openapi.Object
		if s.Properties == nil {
			s.Properties = make(map[string]*openapi.Schema)
		}
		typeSchema := &openapi.Schema{Type: "string"}
		typeSchema.Enum = make([]any, len(t.Values))
		for i, val := range t.Values {
			typeSchema.Enum[i] = val.Name
		}
		valueSchema := &openapi.Schema{}
		for _, val := range t.Values {
			valueSchema.AnyOf = append(valueSchema.AnyOf, sf.derived(val.Name).schemafy(val.Attribute))
		}
		s.Properties[typeKey] = typeSchema
		s.Properties[valueKey] = valueSchema
		s.Required = append(s.Required, typeKey, valueKey)
	case expr.UserType:
		if expr.IsAlias(t) && !sf.nameAliases {
			return sf.rebased(t.ID()).schemafy(t.Attribute())
		}
		h := sf.hashAttribute(attr, fnv.New64())

		var metaName string
		if n, ok := t.Attribute().Meta["openapi:typename"]; ok {
			metaName = codegen.Goify(n[0], true)
		}
		metaRef := toRef(metaName)

		// If it is named, it refers to the same structure and name.
		// If it is not named, it refers to the same structure.
		refs, ok := sf.hashes[h]
		if len(noref) == 0 && ok {
			for _, ref := range refs {
				if ref == metaRef || metaName == "" {
					s.Ref = ref
					sf.ensureSchemaDescription(ref, t, attr)
					return s
				}
			}
		}

		// There is no type to refer to, generate a new one.
		name := t.Name()
		if metaName != "" {
			name = metaName
		} else if n, ok := t.Attribute().Meta["name:original"]; ok {
			name = n[0]
		}

		typeName := sf.uniquify(codegen.Goify(name, true))
		s.Ref = toRef(typeName)
		sf.hashes[h] = append(sf.hashes[h], s.Ref)
		schema := sf.rebased(t.ID()).schemafy(t.Attribute(), true)
		if schema.Description == "" {
			schema.Description = userTypeDescription(t, attr)
		}
		sf.schemas[typeName] = schema
		return s // All other schema properties are set in the reference
	default:
		panic(fmt.Sprintf("unknown type %T", t)) // bug
	}
	s.Description = attr.Description
	if note != "" {
		s.Description += "\n" + note
	}

	// Default value, example, extensions
	s.DefaultValue = toStringMap(attr.DefaultValue)
	s.Example = openapi.Example(attr, sf.rand)
	s.Extensions = openapi.ExtensionsFromExpr(attr.Meta)

	// Validations
	if ap := openapi.AdditionalPropertiesFromExpr(attr.Meta); ap != nil {
		s.AdditionalProperties = ap
	}
	val := attr.Validation
	if val == nil {
		return s
	}
	s.Enum = val.Values
	if val.Format != "" {
		s.Format = string(val.Format)
	}
	s.Pattern = val.Pattern
	if val.ExclusiveMinimum != nil {
		s.ExclusiveMinimum = val.ExclusiveMinimum
	}
	if val.Minimum != nil {
		s.Minimum = val.Minimum
	}
	if val.ExclusiveMaximum != nil {
		s.ExclusiveMaximum = val.ExclusiveMaximum
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
	for _, v := range val.Required {
		if a := attr.Find(v); a != nil {
			if !openapi.MustGenerate(a.Meta) {
				continue
			}
		}
		s.Required = append(s.Required, v)
	}

	return s
}

// ensureSchemaDescription updates an existing component schema with the type or
// reference attribute description if the component was first created without
// one. This preserves user type descriptions when structurally equivalent types
// are reused under a component reference.
func (sf *schemafier) ensureSchemaDescription(ref string, t expr.UserType, attr *expr.AttributeExpr) {
	const prefix = "#/components/schemas/"
	typeName := strings.TrimPrefix(ref, prefix)
	if typeName == ref {
		return
	}
	schema := sf.schemas[typeName]
	if schema == nil || schema.Description != "" {
		return
	}
	schema.Description = userTypeDescription(t, attr)
}

// userTypeDescription returns the canonical description for a user type schema,
// falling back to the description of the attribute that introduced the type.
func userTypeDescription(t expr.UserType, attr *expr.AttributeExpr) string {
	if desc := t.Attribute().Description; desc != "" {
		return desc
	}
	return attr.Description
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

// toRef creates a relative JSON Schema reference from a type name that points
// to the corresponding definition in the OpenAPI "components" field.
func toRef(n string) string {
	return fmt.Sprintf("#/components/schemas/%s", n)
}

// toStringMap converts map[any]any to a map[string]any
// when possible.
func toStringMap(val any) any {
	switch actual := val.(type) {
	case map[any]any:
		m := make(map[string]any)
		for k, v := range actual {
			m[toString(k)] = toStringMap(v)
		}
		return m
	case []any:
		mapSlice := make([]any, len(actual))
		for i, e := range actual {
			mapSlice[i] = toStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// toString returns the string representation of the given type.
func toString(val any) string {
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
// equivalent types and identical set of validation rules.
func (*schemafier) hashAttribute(att *expr.AttributeExpr, h hash.Hash64) uint64 {
	return *hashAttribute(att, h, make(map[string]*uint64))
}

func hashAttribute(att *expr.AttributeExpr, h hash.Hash64, seen map[string]*uint64) *uint64 {
	t := att.Type
	if h, ok := seen[t.Hash()]; ok {
		return h
	}
	var res *uint64
	{
		var tmp uint64
		res = &tmp
	}
	seen[t.Hash()] = res

	hv := hashValidation(att.Validation, h)
	switch t.Kind() {
	case expr.ObjectKind:
		o := expr.AsObject(t)
		for _, m := range *o {
			if !openapi.MustGenerate(m.Attribute.Meta) {
				continue
			}
			kh := hashString(m.Name, h)
			vh := hashAttribute(m.Attribute, h, seen)
			*res ^= orderedHash(kh, *vh, h)
		}
		if hv != 0 {
			*res = orderedHash(*res, hv, h)
		}

	case expr.ArrayKind:
		kh := hashString("[]", h)
		vh := hashAttribute(expr.AsArray(t).ElemType, h, seen)
		*res = orderedHash(kh, *vh, h)
		if hv != 0 {
			*res = orderedHash(*res, hv, h)
		}

	case expr.MapKind:
		m := expr.AsMap(t)
		kh := hashAttribute(m.KeyType, h, seen)
		vh := hashAttribute(m.ElemType, h, seen)
		*res = orderedHash(*kh, *vh, h)
		if hv != 0 {
			*res = orderedHash(*res, hv, h)
		}

	case expr.UserTypeKind:
		*res = *hashAttribute(t.(expr.UserType).Attribute(), h, seen)

	case expr.ResultTypeKind:
		// The identifier specified in the design for result types should drive
		// the computation of the hash.
		rt := t.(*expr.ResultTypeExpr)
		*res = hashString(rt.Identifier, h)
		if view, ok := rt.Meta.Last(expr.ViewMetaKey); ok {
			*res = orderedHash(*res, hashString(view, h), h)
		}

	default: // Primitives or Any
		*res = hashString(t.Name(), h)
		if hv != 0 {
			*res = orderedHash(*res, hv, h)
		}
	}

	return res
}

func hashValidation(val *expr.ValidationExpr, h hash.Hash64) uint64 {
	// Note: we can't use hashstructure for attributes because it doesn't
	// handle recursive structures.
	if val == nil {
		return 0
	}

	res, err := hashstructure.Hash(val, &hashstructure.HashOptions{
		Hasher:          h,
		ZeroNil:         false,
		IgnoreZeroValue: true,
		SlicesAsSets:    true,
	})
	if err != nil {
		// should really never happen (OOM maybe)
		return 0
	}
	return res
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

func openAPIGeneratedServices(api *expr.APIExpr) map[string]struct{} {
	services := make(map[string]struct{}, len(api.HTTP.Services))
	for _, s := range api.HTTP.Services {
		if !openapi.MustGenerate(s.Meta) || !openapi.MustGenerate(s.ServiceExpr.Meta) {
			continue
		}
		services[s.Name()] = struct{}{}
	}
	return services
}

func mustGenerateType(meta expr.MetaExpr, services map[string]struct{}) bool {
	if !openapi.MustGenerate(meta) {
		return false
	}
	if svcs, ok := meta["type:generate:force"]; ok {
		if len(svcs) == 0 {
			return true
		}
		for _, svc := range svcs {
			if _, ok := services[svc]; ok {
				return true
			}
		}
	}
	return false
}
