package openapiv3

import (
	"hash/fnv"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/openapi/v3/testdata/dsls"
)

// describes a type for comparison in tests.
type typ struct {
	Type      string
	Format    string
	Props     []attr
	SkipProps bool
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
	tint    = typ{Type: "integer"}
	tarray  = typ{Type: "array"}
)

func tobj(attrs ...interface{}) typ {
	res := typ{Type: "object"}
	if len(attrs) == 0 {
		res.SkipProps = true
	}
	for i := 0; i < len(attrs); i += 2 {
		res.Props = append(res.Props, attr{Name: attrs[i].(string), Val: attrs[i+1].(typ)})
	}
	return res
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
		ExpectedResponseTypes: rt{200: tobj("name", tstring, "age", tint)},
	}, {
		Name: "string_streaming_response_body",
		DSL:  dsls.StringStreamingResponseBodyDSL(svcName, "string_streaming_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tstring},
	}, {
		Name: "object_streaming_response_body",
		DSL:  dsls.ObjectResponseBodyDSL(svcName, "object_streaming_response_body"),

		ExpectedType:          tempty,
		ExpectedResponseTypes: rt{200: tobj("name", tstring, "age", tint)},
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
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			api := codegen.RunDSL(t, c.DSL).API

			bodies, types := buildBodyTypes(api)

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
	}
}

func TestHashAttribute(t *testing.T) {
	var (
		h1 = uint64(12943244719806607708)
		h2 = uint64(7733915756259492975)
		h3 = uint64(7729867354446285276)
		h4 = uint64(12938215553621425391)
		h5 = uint64(590638987843676710)
	)
	cases := []struct {
		Name string
		att  *expr.AttributeExpr
		h    uint64
	}{
		{"bool", &expr.AttributeExpr{Type: expr.Boolean}, 1200285950329868815},
		{"int", &expr.AttributeExpr{Type: expr.Int}, 15618947606512183472},
		{"int32", &expr.AttributeExpr{Type: expr.Int32}, 9710406772214674507},
		{"int64", &expr.AttributeExpr{Type: expr.Int64}, 9710410070749559206},
		{"uint", &expr.AttributeExpr{Type: expr.UInt}, 9334303408231097877},
		{"uint32", &expr.AttributeExpr{Type: expr.UInt32}, 14693036559411812390},
		{"uint64", &expr.AttributeExpr{Type: expr.UInt64}, 14693033260876927695},
		{"float32", &expr.AttributeExpr{Type: expr.Float32}, 3496747786213902106},
		{"float64", &expr.AttributeExpr{Type: expr.Float64}, 3496753283772043155},
		{"string", &expr.AttributeExpr{Type: expr.String}, 11035750783128163470},
		{"bytes", &expr.AttributeExpr{Type: expr.Bytes}, 9376284137219620846},
		{"any", &expr.AttributeExpr{Type: expr.Any}, 15626582615256966821},
		{"array-bool", &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Boolean}}}, 11710318443436489022},
		{"array-int", &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 16304700464423429033},
		{"map-str-int", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 957614225485715479},
		{"map-str-str", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: &expr.AttributeExpr{Type: expr.String}}}, 10408036596908747853},
		{"map-int-str", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.Int}, ElemType: &expr.AttributeExpr{Type: expr.String}}}, 16377853221392883275},
		{"map-int-int", &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.Int}, ElemType: &expr.AttributeExpr{Type: expr.Int}}}, 3290208366554661977},
		{"obj-str-req", newObj("foo", expr.String, true), 2958992150570065940},
		{"obj-str-noreq", newObj("foo", expr.String, false), 17427721879237743911},
		{"obj-int-req", newObj("foo", expr.Int, true), 8915021286725901502},
		{"obj-int-noreq", newObj("foo", expr.Int, false), 11777831908257753485},
		{"obj-other", newObj("bar", expr.Int, false), 12868551315046025641},
		{"obj-str-str-noreq", newObj2("foo", "bar", expr.String, expr.String), h1},
		{"obj-str-str-req1", newObj2("foo", "bar", expr.String, expr.String, "foo"), h2},
		{"obj-str-str-req2", newObj2("foo", "bar", expr.String, expr.String, "bar"), h3},
		{"obj-str-str-req3", newObj2("foo", "bar", expr.String, expr.String, "foo", "bar"), h4},
		{"obj-int-str-noreq", newObj2("foo", "bar", expr.Int, expr.String), 16228531529443692022},
		{"obj1-str-str-noreq", newObj2("bar", "foo", expr.String, expr.String), h1},
		{"obj1-str-str-req1", newObj2("bar", "foo", expr.String, expr.String, "foo"), h2},
		{"obj1-str-str-req2", newObj2("bar", "foo", expr.String, expr.String, "bar"), h3},
		{"obj1-str-str-req3", newObj2("bar", "foo", expr.String, expr.String, "bar", "foo"), h4},
		{"result", newRT("id", newObj("foo", expr.String, true)), h5},
		{"result-diff", newRT("id2", newObj("foo", expr.String, true)), 15618941009442414240},
		{"result-same", newRT("id", newObj("foo", expr.Int, true)), h5},
	}
	h := fnv.New64()
	sf := newSchemafier(expr.NewRandom("test"))
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := sf.hashAttribute(c.att, h)
			if got != c.h {
				t.Errorf("got %v, expected %v", got, c.h)
			}
		})
	}
}

func newObj(n string, t expr.DataType, req bool) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type:       &expr.Object{{n, &expr.AttributeExpr{Type: t}}},
		Validation: &expr.ValidationExpr{},
	}
	if req {
		attr.Validation.Required = []string{n}
	}
	return attr
}

func newObj2(n, o string, t, u expr.DataType, reqs ...string) *expr.AttributeExpr {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			{n, &expr.AttributeExpr{Type: t}},
			{o, &expr.AttributeExpr{Type: u}},
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

// nameFromRef does the reverse of toRef: it returns the type name from its
// JSON Schema reference.
func nameFromRef(ref string) string {
	elems := strings.Split(ref, "/")
	return elems[len(elems)-1]
}
