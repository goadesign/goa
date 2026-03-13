package openapiv3

import (
	"encoding/json"
	"hash/fnv"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/openapi/v3/testdata/dsls"
	"goa.design/goa/v3/http/codegen/testdata"
)

// describes a type for comparison in tests.
type typ struct {
	Type                 string
	Format               string
	Props                []attr
	SkipProps            bool
	AdditionalProperties *additionalPropsType // nil means no additionalProperties check
}

// additionalPropsType describes additionalProperties for testing
type additionalPropsType struct {
	Type  string               // "string", "array", "object", "" (for reference)
	Items *additionalPropsType // for array items
	Ref   string               // for references like "#/components/schemas/MapData"
}

type attr struct {
	Name string
	Val  typ
}

// types mapped by response code.
type rt map[int]typ

// helpers
var (
	tempty  typ
	tstring = typ{Type: "string"}
	tuuid   = typ{Type: "string", Format: "uuid"}
	tbinary = typ{Type: "string", Format: "binary"}
	tint    = typ{Type: "integer"}
	tarray  = typ{Type: "array"}
)

func tobj(attrs ...any) typ {
	res := typ{Type: "object"}
	if len(attrs) == 0 {
		res.SkipProps = true
	}
	for i := 0; i < len(attrs); i += 2 {
		res.Props = append(res.Props, attr{Name: attrs[i].(string), Val: attrs[i+1].(typ)})
	}
	return res
}

func tmap() typ {
	return typ{Type: "object", Props: []attr{{Name: "map", Val: typ{Type: "object"}}}}
}

func (tt typ) Prop(n string) (typ, bool) {
	for _, att := range tt.Props {
		if att.Name == n {
			return att.Val, true
		}
	}
	return tempty, false
}

func TestBuildBodyTypes(t *testing.T) {
	const svcName = "test service"

	cases := []struct {
		Name string
		DSL  func()

		ExpectedType          typ
		ExpectedFormat        string
		ExpectedResponseTypes rt
		ExpectedExtraTypes    map[string]typ
	}{{
		Name: "string_body",
		DSL:  dsls.StringBodyDSL(svcName, "string_body"),

		ExpectedType:          tstring,
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "alias_string_body",
		DSL:  dsls.AliasStringBodyDSL(svcName, "alias_string_body"),

		ExpectedType:          tuuid,
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "object_body",
		DSL:  dsls.ObjectBodyDSL(svcName, "object_body"),

		ExpectedType:          tobj("name", tstring, "age", tint),
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "map_body",
		DSL:  dsls.MapBodyDSL(svcName, "map_body"),

		ExpectedType:          tmap(),
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "streaming_string_body",
		DSL:  dsls.RequestStreamingStringBody(svcName, "streaming_string_body"),

		ExpectedType:          tstring,
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "streaming_object_body",
		DSL:  dsls.RequestStreamingObjectBody(svcName, "streaming_object_body"),

		ExpectedType:          tobj("name", tstring, "age", tint),
		ExpectedResponseTypes: rt{204: tempty},
	}, {
		Name: "string_response_body",
		DSL:  dsls.StringResponseBodyDSL(svcName, "string_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tstring},
	}, {
		Name: "object_response_body",
		DSL:  dsls.ObjectResponseBodyDSL(svcName, "object_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tobj("name", tstring, "age", tint, "misc", tempty)},
	}, {
		Name: "multi_cookie_response_body",
		DSL:  dsls.MultiCookieResponseBodyDSL(svcName, "multi_cookie_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tobj("name", tstring)},
	}, {
		Name: "string_streaming_response_body",
		DSL:  dsls.StringStreamingResponseBodyDSL(svcName, "string_streaming_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tstring},
	}, {
		Name: "object_streaming_response_body",
		DSL:  dsls.ObjectResponseBodyDSL(svcName, "object_streaming_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tobj("name", tstring, "age", tint, "misc", tempty)},
	}, {
		Name: "string_error_response",
		DSL:  dsls.StringErrorResponseBodyDSL(svcName, "string_error_response"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty, 400: tstring},
	}, {
		Name: "object_error_response",
		DSL:  dsls.ObjectErrorResponseBodyDSL(svcName, "object_error_response"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty, 400: tobj("name", tstring, "age", tint)},
	}, {
		Name: "forced_type",
		DSL:  dsls.ForcedTypeDSL(svcName, "forced_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedExtraTypes:    map[string]typ{"Forced": tobj("foo", tstring)},
	}, {
		Name: "forced_result_type",
		DSL:  dsls.ForcedResultTypeDSL(svcName, "forced_result_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedExtraTypes:    map[string]typ{"Forced": tobj("foo", tstring)},
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)

			bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes)

			svc, ok := bodies[svcName]
			if !ok {
				t.Errorf("bodies does not contain details for service %q", svcName)
				return
			}
			met, ok := svc[c.Name]
			if !ok {
				t.Errorf("bodies does not contain details for method %q", c.Name)
				return
			}
			requestBody := met.RequestBody
			for s, r := range met.ResponseBodies {
				if len(r) != 1 {
					t.Errorf("got %d response bodies for %d, expected 1", len(r), s)
					return
				}
			}

			matchesSchema(t, "request", requestBody, types, c.ExpectedType)
			if len(c.ExpectedResponseTypes) != len(met.ResponseBodies) {
				t.Errorf("got %d response body(ies), expected %d", len(met.ResponseBodies), len(c.ExpectedResponseTypes))
				return
			}
			for s, r := range c.ExpectedResponseTypes {
				if len(met.ResponseBodies[s]) != 1 {
					t.Errorf("got %d response bodies for code %d, expected 1", len(met.ResponseBodies[s]), s)
					return
				}
				matchesSchema(t, "response", met.ResponseBodies[s][0], types, r)
			}
			for name, forced := range c.ExpectedExtraTypes {
				got, ok := types[name]
				if !ok {
					t.Errorf("missing forced type %q", name)
					continue
				}
				matchesSchema(t, "extra type", got, types, forced)
			}
		})
	}
}

func TestBuildBodyTypesUnionIncludesDiscriminatorOneOf(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		textResult := dsl.Type("TextResult", func() {
			dsl.Attribute("text", dsl.String)
			dsl.Required("text")
		})
		jsonResult := dsl.Type("JSONResult", func() {
			dsl.Attribute("message", dsl.String)
			dsl.Required("message")
		})
		dsl.Service("union-service", func() {
			dsl.Method("show", func() {
				dsl.Payload(dsl.OneOf(textResult, jsonResult))
				dsl.Result(dsl.OneOf(textResult, jsonResult))
				dsl.HTTP(func() {
					dsl.POST("/")
				})
			})
		})
	})

	bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes)
	methodBodies := bodies["union-service"]["show"]

	requestSchema := methodBodies.RequestBody
	if requestSchema != nil && requestSchema.Ref != "" {
		requestSchema = types[nameFromRef(requestSchema.Ref)]
	}
	if requestSchema == nil {
		t.Fatal("expected request body schema")
	}
	matchesUnionSchemaExtensions(t, requestSchema, types)

	responseSchema := methodBodies.ResponseBodies[200][0]
	if responseSchema != nil && responseSchema.Ref != "" {
		responseSchema = types[nameFromRef(responseSchema.Ref)]
	}
	if responseSchema == nil {
		t.Fatal("expected response body schema")
	}
	matchesUnionSchemaExtensions(t, responseSchema, types)
}

func matchesUnionSchemaExtensions(t *testing.T, schema *openapi.Schema, types map[string]*openapi.Schema) {
	t.Helper()

	if schema.Extensions == nil {
		t.Fatal("expected union schema extensions")
	}
	discriminator, ok := schema.Extensions["discriminator"].(map[string]any)
	if !ok {
		t.Fatalf("expected discriminator extension, got %T", schema.Extensions["discriminator"])
	}
	if discriminator["propertyName"] != "type" {
		t.Fatalf("expected discriminator propertyName %q, got %#v", "type", discriminator["propertyName"])
	}
	mapping, ok := discriminator["mapping"].(map[string]any)
	if !ok {
		t.Fatalf("expected discriminator mapping, got %T", discriminator["mapping"])
	}

	rawOneOf, ok := schema.Extensions["oneOf"].([]any)
	if !ok {
		t.Fatalf("expected oneOf extension, got %T", schema.Extensions["oneOf"])
	}
	if len(rawOneOf) != 2 {
		t.Fatalf("expected 2 oneOf branches, got %d", len(rawOneOf))
	}
	for _, raw := range rawOneOf {
		branch, ok := raw.(*openapi.Schema)
		if !ok {
			t.Fatalf("expected oneOf branch schema, got %T", raw)
		}
		if branch.Ref == "" {
			t.Fatal("expected oneOf branch schema reference")
		}
		refName := nameFromRef(branch.Ref)
		refSchema, ok := types[refName]
		if !ok {
			t.Fatalf("expected referenced branch schema %q", refName)
		}
		if refSchema.Type != openapi.Object {
			t.Fatalf("expected branch object type, got %q", refSchema.Type)
		}
		if len(refSchema.Required) != 2 || refSchema.Required[0] != "type" || refSchema.Required[1] != "value" {
			t.Fatalf("expected required fields [type value], got %#v", refSchema.Required)
		}
		typeSchema, ok := refSchema.Properties["type"]
		if !ok {
			t.Fatal("expected type discriminator property")
		}
		if len(typeSchema.Enum) != 1 {
			t.Fatalf("expected singleton discriminator enum, got %#v", typeSchema.Enum)
		}
		discriminatorValue, ok := typeSchema.Enum[0].(string)
		if !ok {
			t.Fatalf("expected string discriminator enum, got %#v", typeSchema.Enum[0])
		}
		if mappedRef, ok := mapping[discriminatorValue].(string); !ok || mappedRef != branch.Ref {
			t.Fatalf("expected discriminator mapping for %q to be %q, got %#v", discriminatorValue, branch.Ref, mapping[discriminatorValue])
		}
		if _, ok := refSchema.Properties["value"]; !ok {
			t.Fatal("expected value property")
		}
	}
}

func matchesSchema(t *testing.T, ctx string, s *openapi.Schema, types map[string]*openapi.Schema, tt typ) {
	matchesSchemaWithPrefix(t, ctx, s, types, tt, "")
}
func matchesSchemaWithPrefix(t *testing.T, ctx string, s *openapi.Schema, types map[string]*openapi.Schema, tt typ, prefix string) {
	if s == nil {
		if tt.Type != "" {
			t.Errorf("%s: %sgot type Empty, expected %q", ctx, prefix, tt.Type)
		}
		return
	}
	if s.Ref != "" {
		var ok bool
		s, ok = types[nameFromRef(s.Ref)]
		if !ok {
			t.Errorf("could not find type for ref %q", s.Ref)
			return
		}
	}
	if tt.Type != string(s.Type) {
		t.Errorf("%s: %sgot type %q, expected %q", ctx, prefix, s.Type, tt.Type)
	}
	if tt.Format != "" {
		if s.Format != tt.Format {
			t.Errorf("%s: %sgot format %q, expected %q", ctx, prefix, s.Format, tt.Format)
		}
	}
	if tt.Type == "object" {
		if tt.SkipProps {
			return
		}
		for n, v := range s.Properties {
			p, ok := tt.Prop(n)
			if !ok {
				t.Errorf("%s: %sgot unexpected field %q", ctx, prefix, n)
				continue
			}
			matchesSchemaWithPrefix(t, ctx, v, types, p, n+": ")
		}

		// Check additionalProperties
		if tt.AdditionalProperties != nil {
			validateAdditionalProperties(t, ctx, s.AdditionalProperties, types, tt.AdditionalProperties, prefix)
		}
	}
}

func TestMapTypes(t *testing.T) {
	svcName := "test-service"

	testCases := []struct {
		Name     string
		DSL      func()
		Expected typ
	}{
		{
			Name: "map_int_array_string",
			DSL:  dsls.MapIntKeyBodyDSL(svcName, "map_int_array_string"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "intmap", Val: typ{
					Type: "object",
					AdditionalProperties: &additionalPropsType{
						Type:  "array",
						Items: &additionalPropsType{Type: "string"},
					},
				}}},
			},
		},
		{
			Name: "map_int_array_object",
			DSL:  dsls.MapIntKeyObjectBodyDSL(svcName, "map_int_array_object"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "intmap", Val: typ{
					Type: "object",
					AdditionalProperties: &additionalPropsType{
						Type:  "array",
						Items: &additionalPropsType{Ref: "#/components/schemas/MapData"},
					},
				}}},
			},
		},
		{
			Name: "map_int_string",
			DSL:  dsls.MapIntKeyStringBodyDSL(svcName, "map_int_string"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "intmap", Val: typ{
					Type:                 "object",
					AdditionalProperties: &additionalPropsType{Type: "string"},
				}}},
			},
		},
		{
			Name: "map_int_object_direct",
			DSL:  dsls.MapIntKeyObjectDirectBodyDSL(svcName, "map_int_object_direct"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "intmap", Val: typ{
					Type:                 "object",
					AdditionalProperties: &additionalPropsType{Ref: "#/components/schemas/MapData"},
				}}},
			},
		},
		{
			Name: "map_string_int",
			DSL:  dsls.MapStringKeyIntBodyDSL(svcName, "map_string_int"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "stringmap", Val: typ{
					Type:                 "object",
					AdditionalProperties: &additionalPropsType{Type: "integer"},
				}}},
			},
		},
		{
			Name: "map_string_object_direct",
			DSL:  dsls.MapStringKeyObjectDirectBodyDSL(svcName, "map_string_object_direct"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "stringmap", Val: typ{
					Type:                 "object",
					AdditionalProperties: &additionalPropsType{Ref: "#/components/schemas/MapData"},
				}}},
			},
		},
		{
			Name: "map_string_array_object",
			DSL:  dsls.MapStringKeyArrayObjectBodyDSL(svcName, "map_string_array_object"),
			Expected: typ{
				Type: "object",
				Props: []attr{{Name: "stringmap", Val: typ{
					Type: "object",
					AdditionalProperties: &additionalPropsType{
						Type:  "array",
						Items: &additionalPropsType{Ref: "#/components/schemas/MapData"},
					},
				}}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Build the OpenAPI spec
			root := codegen.RunDSL(t, tc.DSL)
			bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes)

			// Find the service and method
			svcBodies, ok := bodies[svcName]
			if !ok {
				t.Fatalf("Could not find service %s in bodies", svcName)
			}

			methodBody, ok := svcBodies[tc.Name]
			if !ok {
				t.Fatalf("Could not find method %s in service bodies", tc.Name)
			}

			// Get the request body schema
			requestBodyRef := methodBody.RequestBody.Ref
			if requestBodyRef == "" {
				t.Fatal("Expected request body reference")
			}

			requestBodyTypeName := nameFromRef(requestBodyRef)
			requestBodySchema, ok := types[requestBodyTypeName]
			if !ok {
				t.Fatalf("Could not find request body type %s", requestBodyTypeName)
			}

			// Validate the schema
			matchesSchema(t, tc.Name, requestBodySchema, types, tc.Expected)
		})
	}
}

func validateAdditionalProperties(t *testing.T, ctx string, addProps any, types map[string]*openapi.Schema, expected *additionalPropsType, prefix string) {
	if addProps == nil {
		t.Errorf("%s: %sexpected additionalProperties to be set", ctx, prefix)
		return
	}

	// Check if additionalProperties is a schema
	schema, ok := addProps.(*openapi.Schema)
	if !ok {
		t.Errorf("%s: %sexpected additionalProperties to be schema, got %T", ctx, prefix, addProps)
		return
	}

	validateAdditionalPropsSchema(t, ctx, schema, types, expected, prefix+"additionalProperties: ")
}

func validateAdditionalPropsSchema(t *testing.T, ctx string, schema *openapi.Schema, types map[string]*openapi.Schema, expected *additionalPropsType, prefix string) {
	// Handle reference case
	if expected.Ref != "" {
		if schema.Ref == "" {
			t.Errorf("%s: %sexpected reference to %s, but got inline schema", ctx, prefix, expected.Ref)
			return
		}
		if schema.Ref != expected.Ref {
			t.Errorf("%s: %sexpected reference %s, got %s", ctx, prefix, expected.Ref, schema.Ref)
		}
		return
	}

	// Resolve reference if present
	if schema.Ref != "" {
		typeName := nameFromRef(schema.Ref)
		resolvedSchema, ok := types[typeName]
		if !ok {
			t.Errorf("%s: %scould not resolve reference %s", ctx, prefix, schema.Ref)
			return
		}
		schema = resolvedSchema
	}

	// Check type
	if string(schema.Type) != expected.Type {
		t.Errorf("%s: %sexpected type %s, got %s", ctx, prefix, expected.Type, schema.Type)
	}

	// Check array items if expected
	if expected.Items != nil {
		if schema.Items == nil {
			t.Errorf("%s: %sexpected array items to be set", ctx, prefix)
		} else {
			validateAdditionalPropsSchema(t, ctx, schema.Items, types, expected.Items, prefix+"items: ")
		}
	}
}

func TestTypesOnlyDifferByEnum(t *testing.T) {
	root := codegen.RunDSL(t, dsls.StringEnumBodyDSL())

	bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes)

	svc1, ok := bodies["svc_enum_1"]
	if !ok {
		t.Errorf("bodies does not contain details for service %q", "svc_enum_1")
		return
	}
	svc2, ok := bodies["svc_enum_2"]
	if !ok {
		t.Errorf("bodies does not contain details for service %q", "svc_enum_2")
		return
	}

	svc1MethodRB := svc1["method_enum"].RequestBody.Ref
	svc2MethodRB := svc2["method_enum"].RequestBody.Ref

	if svc1MethodRB == svc2MethodRB {
		t.Errorf("expected different refs, got %q", svc1MethodRB)

		name := nameFromRef(svc1MethodRB)
		derefed := types[name]
		jsoned, _ := json.Marshal(derefed)
		t.Errorf("shared referenced type (%s) was: %v", name, string(jsoned))
		return
	}
}

func TestHashAttribute(t *testing.T) {
	type (
		testAttr struct {
			name string
			att  *expr.AttributeExpr
		}

		hashBehavior int

		testGroup struct {
			name     string
			attrs    []testAttr
			behavior hashBehavior
		}
	)

	const (
		uniqueHashes hashBehavior = iota
		identicalHashes
	)

	var (
		metaNotGenerate = expr.MetaExpr{"openapi:generate": []string{"false"}}
		metaEmpty       = expr.MetaExpr{}
	)

	cases := []testGroup{
		{
			name:     "Primitive types",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "bool", att: &expr.AttributeExpr{Type: expr.Boolean}},
				{name: "int", att: &expr.AttributeExpr{Type: expr.Int}},
				{name: "int32", att: &expr.AttributeExpr{Type: expr.Int32}},
				{name: "int64", att: &expr.AttributeExpr{Type: expr.Int64}},
				{name: "uint", att: &expr.AttributeExpr{Type: expr.UInt}},
				{name: "uint32", att: &expr.AttributeExpr{Type: expr.UInt32}},
				{name: "uint64", att: &expr.AttributeExpr{Type: expr.UInt64}},
				{name: "float32", att: &expr.AttributeExpr{Type: expr.Float32}},
				{name: "float64", att: &expr.AttributeExpr{Type: expr.Float64}},
				{name: "string", att: &expr.AttributeExpr{Type: expr.String}},
				{name: "bytes", att: &expr.AttributeExpr{Type: expr.Bytes}},
				{name: "any", att: &expr.AttributeExpr{Type: expr.Any}},
			},
		}, {
			name:     "Collection types",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "array-bool", att: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Boolean}}}},
				{name: "array-int", att: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}}},
				{name: "map-str-int", att: &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.Int}}}},
				{name: "map-str-str", att: &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.String}}}},
			},
		}, {
			name:     "Objects with validation rules",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "no-validation", att: newObj("foo", false)},
				{name: "required-validation", att: newObj("foo", true)},
				{name: "pattern-validation", att: &expr.AttributeExpr{
					Type: expr.String,
					Validation: &expr.ValidationExpr{
						Pattern: "^[a-z]+$",
					},
				}},
				{name: "enum-validation", att: &expr.AttributeExpr{
					Type: expr.String,
					Validation: &expr.ValidationExpr{
						Values: []any{"foo", "bar"},
					},
				}},
			},
		}, {
			name:     "Result types with different views",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "no-view", att: newRT("id", newObj("foo", true))},
				{name: "default-view", att: newRTWithView("id", newObj("foo", true), "default")},
				{name: "tiny-view", att: newRTWithView("id", newObj("foo", true), "tiny")},
			},
		}, {
			name:     "Objects with openapi:generate:false metadata",
			behavior: identicalHashes,
			attrs: []testAttr{
				{name: "obj-with-skipped-field", att: newObj2Meta("foo", "bar", expr.String, expr.String, metaEmpty, metaNotGenerate)},
				{name: "obj-without-skipped-field", att: newObj("foo", false)},
			},
		}, {
			name:     "Complex map types",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "map-int-array", att: &expr.AttributeExpr{Type: &expr.Map{
					KeyType:  &expr.AttributeExpr{Type: expr.Int},
					ElemType: &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
				}}},
				{name: "map-array-int", att: &expr.AttributeExpr{Type: &expr.Map{
					KeyType:  &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.String}}},
					ElemType: &expr.AttributeExpr{Type: expr.Int},
				}}},
			},
		}, {
			name:     "Nested user types",
			behavior: uniqueHashes,
			attrs: []testAttr{
				{name: "single-nest", att: newUserType("foo", newObj("bar", false))},
				{name: "double-nest", att: newUserType("foo", newUserType("bar", newObj("baz", false)))},
			},
		}, {
			name:     "Recursive types",
			behavior: identicalHashes,
			attrs: []testAttr{
				{name: "recursive-1", att: newRecursiveType("foo")},
				{name: "recursive-2", att: newRecursiveType("foo")},
			},
		},
	}

	h := fnv.New64()
	sf := newSchemafier(expr.NewRandom("test"))

	for _, group := range cases {
		t.Run(group.name, func(t *testing.T) {
			seen := make(map[uint64][]string)

			// Collect all hashes in this group
			for _, attr := range group.attrs {
				hash := sf.hashAttribute(attr.att, h)
				seen[hash] = append(seen[hash], attr.name)
			}

			switch group.behavior {
			case uniqueHashes:
				// Verify all hashes are different
				for hash, names := range seen {
					if len(names) > 1 {
						t.Errorf("expected unique hashes but got collision between %v (hash: %d)",
							names, hash)
					}
				}
			case identicalHashes:
				// Verify all hashes are the same
				if len(seen) > 1 {
					t.Errorf("expected identical hashes but got different ones: %v", seen)
				}
			}
		})
	}
}

func newObj(n string, req bool) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type:       &expr.Object{{Name: n, Attribute: &expr.AttributeExpr{Type: expr.String}}},
		Validation: &expr.ValidationExpr{},
	}
	if req {
		attr.Validation.Required = []string{n}
	}
	return attr
}

func newObj2Meta(n, o string, t, u expr.DataType, l, m expr.MetaExpr, reqs ...string) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: n, Attribute: &expr.AttributeExpr{Type: t, Meta: l}},
			{Name: o, Attribute: &expr.AttributeExpr{Type: u, Meta: m}},
		},
		Validation: &expr.ValidationExpr{},
	}
	attr.Validation.Required = append(attr.Validation.Required, reqs...)
	return attr
}

func newRT(id string, att *expr.AttributeExpr) *expr.AttributeExpr {
	return &expr.AttributeExpr{
		Type: &expr.ResultTypeExpr{
			Identifier: id,
			UserTypeExpr: &expr.UserTypeExpr{
				AttributeExpr: att,
			},
		},
	}
}

// Helper function for result types with views
func newRTWithView(id string, att *expr.AttributeExpr, view string) *expr.AttributeExpr {
	rt := newRT(id, att)
	rt.Type.(*expr.ResultTypeExpr).Meta = expr.MetaExpr{
		expr.ViewMetaKey: []string{view},
	}
	return rt
}

// Helper function for user types
func newUserType(name string, att *expr.AttributeExpr) *expr.AttributeExpr {
	return &expr.AttributeExpr{
		Type: &expr.UserTypeExpr{
			AttributeExpr: att,
			TypeName:      name,
		},
	}
}

// Helper function for recursive types
func newRecursiveType(name string) *expr.AttributeExpr {
	// Create a user type that references itself
	ut := &expr.UserTypeExpr{
		TypeName: name,
	}
	att := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{
				Name: "self",
				Attribute: &expr.AttributeExpr{
					Type: ut,
				},
			},
		},
	}
	ut.AttributeExpr = att
	return &expr.AttributeExpr{Type: ut}
}

type testExampler struct {
	example  any
	examples map[string]*ExampleRef
}

func (t *testExampler) setExample(v any) {
	t.example = v
}

func (t *testExampler) setExamples(v map[string]*ExampleRef) {
	t.examples = v
}

func TestInitExamplesCanonicalizesNestedUnionExamples(t *testing.T) {
	root := codegen.RunDSL(t, testdata.NestedConstructorUnionHTTPDSL)
	attr := root.Services[0].Methods[0].Payload
	exampler := &testExampler{}

	initExamples(exampler, attr, root.API.ExampleGenerator)

	raw, ok := exampler.example.(map[string]any)
	if !ok {
		t.Fatalf("expected object example, got %T", exampler.example)
	}
	choice, ok := raw["choice"].(map[string]any)
	if !ok {
		t.Fatalf("expected canonicalized union example, got %#v", raw["choice"])
	}
	if choice["type"] != "TextPayload" {
		t.Fatalf("expected nested union type %q, got %#v", "TextPayload", choice["type"])
	}
	if _, ok := choice["value"].(map[string]any); !ok {
		t.Fatalf("expected nested union value object, got %#v", choice["value"])
	}
}

func TestInitExamplesUsesMatchingObjectBranchForUserExample(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ConstructorUnionUserExampleSecondBranchHTTPDSL)
	attr := root.Services[0].Methods[0].Payload.Find("choice")
	exampler := &testExampler{}

	initExamples(exampler, attr, root.API.ExampleGenerator)

	choice, ok := exampler.example.(map[string]any)
	if !ok {
		t.Fatalf("expected canonicalized union example, got %T", exampler.example)
	}
	if choice["type"] != "JSONPayload" {
		t.Fatalf("expected union type %q, got %#v", "JSONPayload", choice["type"])
	}
	value, ok := choice["value"].(map[string]any)
	if !ok {
		t.Fatalf("expected union value object, got %#v", choice["value"])
	}
	if value["message"] != "hello" {
		t.Fatalf("expected user example message %q, got %#v", "hello", value["message"])
	}
}

func TestInitExamplesPropagatesUnionFieldUserExampleToPayloadExample(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ConstructorUnionUserExampleSecondBranchHTTPDSL)
	attr := root.API.HTTP.Services[0].HTTPEndpoints[0].Body
	exampler := &testExampler{}
	choice := attr.Find("choice")
	if choice == nil {
		t.Fatalf("expected choice attribute on HTTP body")
	}
	if len(choice.ExtractUserExamples()) == 0 {
		t.Fatalf("expected direct body choice user examples")
	}
	if ut, ok := attr.Type.(expr.UserType); ok {
		utChoice := ut.Attribute().Find("choice")
		if utChoice == nil {
			t.Fatalf("expected choice attribute on body user type")
		}
		if len(utChoice.ExtractUserExamples()) == 0 {
			t.Fatalf("expected user type choice user examples")
		}
	}

	initExamples(exampler, attr, root.API.ExampleGenerator)

	raw, ok := exampler.example.(map[string]any)
	if !ok {
		t.Fatalf("expected payload object example, got %T", exampler.example)
	}
	choiceExample, ok := raw["choice"].(map[string]any)
	if !ok {
		t.Fatalf("expected canonicalized union example, got %#v", raw["choice"])
	}
	if choiceExample["type"] != "JSONPayload" {
		t.Fatalf("expected payload union type %q, got %#v", "JSONPayload", choiceExample["type"])
	}
	value, ok := choiceExample["value"].(map[string]any)
	if !ok {
		t.Fatalf("expected union value object, got %#v", choiceExample["value"])
	}
	if value["message"] != "hello" {
		t.Fatalf("expected user example message %q, got %#v", "hello", value["message"])
	}
}

func TestSchemafyUsesUnionFieldUserExample(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ConstructorUnionUserExampleSecondBranchHTTPDSL)
	attr := root.API.HTTP.Services[0].HTTPEndpoints[0].Body.Find("choice")
	if attr == nil {
		t.Fatalf("expected choice attribute on HTTP body")
	}
	sf := newSchemafier(root.API.ExampleGenerator)

	schema := sf.schemafy(attr)
	example, ok := schema.Example.(map[string]any)
	if !ok {
		t.Fatalf("expected schema example map, got %T", schema.Example)
	}
	if example["type"] != "JSONPayload" {
		t.Fatalf("expected schema union type %q, got %#v", "JSONPayload", example["type"])
	}
	value, ok := example["value"].(map[string]any)
	if !ok {
		t.Fatalf("expected union value object, got %#v", example["value"])
	}
	if value["message"] != "hello" {
		t.Fatalf("expected schema user example message %q, got %#v", "hello", value["message"])
	}
}

func TestInitExamplesCanonicalizesNestedTopLevelUnionExamples(t *testing.T) {
	root := codegen.RunDSL(t, testdata.NestedTopLevelConstructorUnionHTTPDSL)
	attr := root.Services[0].Methods[0].Payload
	exampler := &testExampler{}

	initExamples(exampler, attr, root.API.ExampleGenerator)

	raw, ok := exampler.example.(map[string]any)
	if !ok {
		t.Fatalf("expected top-level union example, got %T", exampler.example)
	}
	if raw["type"] != "OuterA" {
		t.Fatalf("expected outer union type %q, got %#v", "OuterA", raw["type"])
	}
	value, ok := raw["value"].(map[string]any)
	if !ok {
		t.Fatalf("expected outer union value object, got %#v", raw["value"])
	}
	choice, ok := value["choice"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested union choice object, got %#v", value["choice"])
	}
	if choice["type"] == nil || choice["value"] == nil {
		t.Fatalf("expected canonical nested union example, got %#v", choice)
	}
}

func TestInitExamplesCanonicalizesNestedTopLevelUnionExamplesCustomKeys(t *testing.T) {
	root := codegen.RunDSL(t, testdata.NestedTopLevelConstructorUnionCustomKeysHTTPDSL)
	attr := root.Services[0].Methods[0].Payload
	exampler := &testExampler{}

	initExamples(exampler, attr, root.API.ExampleGenerator)

	raw, ok := exampler.example.(map[string]any)
	if !ok {
		t.Fatalf("expected top-level union example, got %T", exampler.example)
	}
	if raw["kind"] != "OuterA" {
		t.Fatalf("expected outer union type %q, got %#v", "OuterA", raw["kind"])
	}
	value, ok := raw["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected outer union value object, got %#v", raw["data"])
	}
	choice, ok := value["choice"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested union choice object, got %#v", value["choice"])
	}
	if choice["kind"] == nil || choice["data"] == nil {
		t.Fatalf("expected canonical nested union example with custom keys, got %#v", choice)
	}
}

func TestBuildBodyTypesRecursiveConstructorUnion(t *testing.T) {
	root := codegen.RunDSL(t, testdata.RecursiveConstructorUnionHTTPDSL)

	bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes)
	methodBodies := bodies["RecursiveConstructorUnion"]["Show"]
	if methodBodies == nil || methodBodies.RequestBody == nil {
		t.Fatalf("expected request body schema for recursive constructor union")
	}
	if len(types) == 0 {
		t.Fatalf("expected generated schemas for recursive constructor union")
	}
}

func TestHTTPBodyPreservesUnionFieldUserExamples(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ConstructorUnionUserExampleSecondBranchHTTPDSL)
	body := root.API.HTTP.Services[0].HTTPEndpoints[0].Body
	choice := body.Find("choice")
	if choice == nil {
		t.Fatalf("expected choice attribute on HTTP body")
	}
	examples := choice.ExtractUserExamples()
	if len(examples) == 0 {
		t.Fatalf("expected union field user examples to be preserved on HTTP body")
	}
	value, ok := examples[len(examples)-1].Value.(map[string]any)
	if !ok {
		t.Fatalf("expected user example map, got %T", examples[len(examples)-1].Value)
	}
	if value["message"] != "hello" {
		t.Fatalf("expected preserved user example message %q, got %#v", "hello", value["message"])
	}
}

func TestEnsureUnionBranchSchemaAvoidsExistingComponentCollision(t *testing.T) {
	sf := newSchemafier(expr.NewRandom("collision"))
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{
				Name: "TextPayload",
				Attribute: &expr.AttributeExpr{
					Type: &expr.UserTypeExpr{
						TypeName: "TextPayload",
						AttributeExpr: &expr.AttributeExpr{
							Type: &expr.Object{
								&expr.NamedAttributeExpr{
									Name:      "text",
									Attribute: &expr.AttributeExpr{Type: expr.String},
								},
							},
						},
					},
				},
			},
		},
	}
	key := sf.unionBranchSchemaKey(union, union.Values[0])
	name := deterministicUnionBranchSchemaName(union, union.Values[0], key)
	existing := openapi.NewSchema()
	existing.Description = "existing"
	sf.schemas[name] = existing

	ref := sf.ensureUnionBranchSchema(union, union.Values[0])
	if nameFromRef(ref) == name {
		t.Fatalf("expected wrapper schema name to avoid existing component collision")
	}
	if sf.schemas[name] != existing {
		t.Fatalf("expected existing component schema to remain unchanged")
	}
}

func TestDeterministicUnionBranchSchemaNameStableAcrossOrder(t *testing.T) {
	newNamedObject := func(typeName, fieldName string, fieldType expr.DataType) *expr.NamedAttributeExpr {
		return &expr.NamedAttributeExpr{
			Name: typeName,
			Attribute: &expr.AttributeExpr{
				Type: &expr.UserTypeExpr{
					TypeName: typeName,
					AttributeExpr: &expr.AttributeExpr{
						Type: &expr.Object{
							&expr.NamedAttributeExpr{
								Name: fieldName,
								Attribute: &expr.AttributeExpr{
									Type: fieldType,
								},
							},
						},
					},
				},
			},
		}
	}

	union1 := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			newNamedObject("A", "text", expr.String),
			newNamedObject("B", "id", expr.Int),
		},
	}
	union2 := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			newNamedObject("B", "id", expr.Int),
			newNamedObject("A", "text", expr.String),
		},
	}

	key1A := (&schemafier{}).unionBranchSchemaKey(union1, union1.Values[0])
	key1B := (&schemafier{}).unionBranchSchemaKey(union1, union1.Values[1])
	key2B := (&schemafier{}).unionBranchSchemaKey(union2, union2.Values[0])
	key2A := (&schemafier{}).unionBranchSchemaKey(union2, union2.Values[1])

	name1A := deterministicUnionBranchSchemaName(union1, union1.Values[0], key1A)
	name1B := deterministicUnionBranchSchemaName(union1, union1.Values[1], key1B)
	name2B := deterministicUnionBranchSchemaName(union2, union2.Values[0], key2B)
	name2A := deterministicUnionBranchSchemaName(union2, union2.Values[1], key2A)

	if name1A != name2A {
		t.Fatalf("expected stable name for A branch, got %q and %q", name1A, name2A)
	}
	if name1B != name2B {
		t.Fatalf("expected stable name for B branch, got %q and %q", name1B, name2B)
	}
}

// nameFromRef does the reverse of toRef: it returns the type name from its
// JSON Schema reference.
func nameFromRef(ref string) string {
	elems := strings.Split(ref, "/")
	return elems[len(elems)-1]
}
