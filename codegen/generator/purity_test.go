package generator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpcdata "goa.design/goa/v3/grpc/codegen/testdata"
	httpdata "goa.design/goa/v3/http/codegen/testdata"
	jsonrpcdata "goa.design/goa/v3/jsonrpc/codegen/testdata"
)

type (
	// attrState captures the mutable state of a design attribute expression:
	// the identity of its type, the user type naming, the object shape and
	// deep copies of the meta and validation expressions. Two snapshots of
	// the same attribute are equal if and only if no reachable state was
	// rewritten in between.
	attrState struct {
		Type         uintptr
		Primitive    expr.Kind
		TypeName     string
		UID          string
		Identifier   string
		Views        []string
		UTAttr       uintptr
		Fields       []string
		Description  string
		Meta         expr.MetaExpr
		Validation   *expr.ValidationExpr
		DefaultValue any
	}

	// visitKey identifies a visited pointer during the design walk. The type
	// disambiguates a struct from its first field which share the address.
	visitKey struct {
		ptr uintptr
		typ reflect.Type
	}
)

// TestGeneratorsTreatDesignAsReadOnly is the design purity invariant: once
// eval finalization ran and codegen.NormalizeRoot applied the only sanctioned
// post-finalization rewrite, running every generator ("gen" and "example")
// must leave the design expression tree bit for bit unchanged. The fixtures
// cover alias chains, result views, websocket streaming, SSE with anonymous
// object payloads and results (the NormalizeRoot wrapping case), mixed
// HTTP+JSON-RPC transports and gRPC unions and streaming.
//
// Process global state is deliberately out of the snapshot:
//   - expr.GeneratedResultTypes is appended to by expr.Dup when generators
//     duplicate generated result types; it is a separate eval root, not part
//     of the design tree (known purity hole, documented in expr/dup.go).
//   - the example randomizer seen-value cache lives on the API expression
//     example generator and is legitimately filled by example and OpenAPI
//     generation; the walk skips it.
func TestGeneratorsTreatDesignAsReadOnly(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"alias-chains", httpdata.AliasTypeDSL},
		{"result-views", httpdata.ResultBodyMultipleViewsDSL},
		{"websocket-bidirectional", httpdata.BidirectionalStreamingDSL},
		{"sse-anonymous-object", httpdata.SSEObjectDSL},
		{"jsonrpc-mixed-transport", jsonrpcdata.JSONRPCKitchenSinkDSL},
		{"grpc-union-streaming", grpcdata.ClientStreamingRPCWithUnionPayloadDSL},
		{"grpc-streaming-views", grpcdata.ServerStreamingResultWithViewsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			codegen.NormalizeRoot(root)
			before := snapshotDesign(root)

			for _, cmd := range []string{"gen", "example"} {
				genfuncs, err := Generators(cmd)
				require.NoError(t, err)
				for _, gen := range genfuncs {
					_, err := gen("gen", []eval.Root{root})
					require.NoError(t, err)
				}
			}

			after := snapshotDesign(root)
			assert.Len(t, after, len(before), "attributes appeared in or disappeared from the design")
			for att, b := range before {
				a, ok := after[att]
				if !assert.True(t, ok, "attribute %q (%p) disappeared from the design", b.Description, att) {
					continue
				}
				assert.Equal(t, b, a, "attribute %q (%p) was mutated by a generator", b.Description, att)
			}
		})
	}
}

// snapshotDesign walks every expression reachable from the root via exported
// fields and captures the state of each attribute expression encountered.
func snapshotDesign(root *expr.RootExpr) map[*expr.AttributeExpr]attrState {
	atts := make(map[*expr.AttributeExpr]attrState)
	visited := make(map[visitKey]struct{})
	exampleGenType := reflect.TypeOf((*expr.ExampleGenerator)(nil))
	var walk func(v reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			key := visitKey{ptr: v.Pointer(), typ: v.Type()}
			if _, ok := visited[key]; ok {
				return
			}
			visited[key] = struct{}{}
			if v.Type() == exampleGenType {
				// The example generator carries the randomizer seen-value
				// cache which generation legitimately fills; it is not part
				// of the design.
				return
			}
			if v.CanInterface() {
				if att, ok := v.Interface().(*expr.AttributeExpr); ok {
					atts[att] = snapshotAttribute(att)
				}
			}
			walk(v.Elem())
		case reflect.Interface:
			if v.IsNil() {
				return
			}
			walk(v.Elem())
		case reflect.Struct:
			for i := range v.NumField() {
				if v.Type().Field(i).PkgPath != "" {
					continue // unexported
				}
				walk(v.Field(i))
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				walk(v.Index(i))
			}
		case reflect.Map:
			iter := v.MapRange()
			for iter.Next() {
				walk(iter.Key())
				walk(iter.Value())
			}
		}
	}
	walk(reflect.ValueOf(root))
	return atts
}

// snapshotAttribute captures the mutable state of att. Meta and validation
// are deep copied so in-place writes are detected; the type is captured by
// identity together with the user type name, attribute and shape so renames,
// attribute swaps and field changes are detected too.
func snapshotAttribute(att *expr.AttributeExpr) attrState {
	s := attrState{
		Description:  att.Description,
		DefaultValue: att.DefaultValue,
	}
	if att.Meta != nil {
		s.Meta = att.Meta.Dup()
	}
	if att.Validation != nil {
		s.Validation = att.Validation.Dup()
	}
	switch dt := att.Type.(type) {
	case nil:
	case expr.Primitive:
		s.Primitive = dt.Kind()
	case expr.UserType:
		s.Type = reflect.ValueOf(att.Type).Pointer()
		s.TypeName = dt.Name()
		s.UID = dt.ID()
		s.UTAttr = reflect.ValueOf(dt.Attribute()).Pointer()
		if rt, ok := dt.(*expr.ResultTypeExpr); ok {
			s.Identifier = rt.Identifier
			for _, v := range rt.Views {
				s.Views = append(s.Views, v.Name)
			}
		}
	default:
		s.Type = reflect.ValueOf(att.Type).Pointer()
	}
	if obj := expr.AsObject(att.Type); obj != nil {
		for _, nat := range *obj {
			s.Fields = append(s.Fields, nat.Name)
		}
	}
	return s
}
