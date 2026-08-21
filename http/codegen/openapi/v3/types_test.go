// This file verifies OpenAPI v3 schema construction and stable example
// generation for primitive, collection, object, and transport body types.
package openapiv3

import (
	"encoding/json"
	"hash/fnv"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/openapi/v3/testdata/dsls"
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
		ExpectedAbsentTypes   []string
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
		Name: "unused_type",
		DSL:  dsls.UnusedTypeDSL(svcName, "unused_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedAbsentTypes:   []string{"Unused"},
	}, {
		Name: "forced_type",
		DSL:  dsls.ForcedTypeDSL(svcName, "forced_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedExtraTypes:    map[string]typ{"Forced": tobj("foo", tstring)},
	}, {
		Name: "service_scoped_forced_type",
		DSL:  dsls.ServiceScopedForcedTypeDSL(svcName, "service_scoped_forced_type", svcName),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedExtraTypes:    map[string]typ{"Forced": tobj("foo", tstring)},
	}, {
		Name: "other_service_scoped_forced_type",
		DSL:  dsls.ServiceScopedForcedTypeDSL(svcName, "other_service_scoped_forced_type", "other"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedAbsentTypes:   []string{"Forced"},
	}, {
		Name: "openapi_suppressed_forced_type",
		DSL:  dsls.OpenAPISuppressedForcedTypeDSL(svcName, "openapi_suppressed_forced_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedAbsentTypes:   []string{"Forced"},
	}, {
		Name: "forced_result_type",
		DSL:  dsls.ForcedResultTypeDSL(svcName, "forced_result_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedExtraTypes:    map[string]typ{"Forced": tobj("foo", tstring)},
	}, {
		Name: "openapi_suppressed_forced_result_type",
		DSL:  dsls.OpenAPISuppressedForcedResultTypeDSL(svcName, "openapi_suppressed_forced_result_type"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{204: tempty},
		ExpectedAbsentTypes:   []string{"Forced"},
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)

			bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes, openapi.Version30, expr.NewExampleGenerator(root.API.RandomizerFactory))

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
			for _, name := range c.ExpectedAbsentTypes {
				if _, ok := types[name]; ok {
					t.Errorf("unexpected type %q", name)
				}
			}
		})
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
			bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes, openapi.Version30, expr.NewExampleGenerator(root.API.RandomizerFactory))

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

	bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes, openapi.Version30, expr.NewExampleGenerator(root.API.RandomizerFactory))

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

func TestBuildBodyTypesPreservesPrimitiveAliasComponents(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		unauthorized := dsl.Type("Unauthorized", dsl.String, func() {
			dsl.Description("Authentication is required.")
		})
		forbidden := dsl.Type("Forbidden", dsl.String, func() {
			dsl.Description("Access is forbidden.")
		})

		dsl.Service("auth", func() {
			dsl.Method("show", func() {
				dsl.Error("unauthorized", unauthorized)
				dsl.Error("forbidden", forbidden)
				dsl.HTTP(func() {
					dsl.GET("/")
					dsl.Response("unauthorized", dsl.StatusUnauthorized)
					dsl.Response("forbidden", dsl.StatusForbidden)
				})
			})
		})
	})

	bodies, types := buildBodyTypes(root.API, root.Types, root.ResultTypes, openapi.Version32, expr.NewExampleGenerator(root.API.RandomizerFactory))
	tests := []struct {
		name        string
		status      int
		description string
	}{
		{name: "Unauthorized", status: 401, description: "Authentication is required."},
		{name: "Forbidden", status: 403, description: "Access is forbidden."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			responses := bodies["auth"]["show"].ResponseBodies[test.status]
			if len(responses) != 1 {
				t.Errorf("got %d response schemas, expected 1", len(responses))
				return
			}
			wantRef := "#/components/schemas/" + test.name
			if responses[0].Ref != wantRef {
				t.Errorf("got response ref %q, expected %q", responses[0].Ref, wantRef)
			}
			schema, ok := types[test.name]
			if !ok {
				t.Errorf("missing component schema %q", test.name)
				return
			}
			if schema.Description != test.description {
				t.Errorf("got description %q, expected %q", schema.Description, test.description)
			}
		})
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
	sf := newSchemafier(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")))

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

// nameFromRef does the reverse of toRef: it returns the type name from its
// JSON Schema reference.
func nameFromRef(ref string) string {
	elems := strings.Split(ref, "/")
	return elems[len(elems)-1]
}
