// This file checks that recursive collection designs generate working gRPC code.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGenerateRecursiveGRPCResult(t *testing.T) {
	root := codegen.RunDSL(t, grpcRecursiveCollectionsDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range append(files, transport...) {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGRPCRecursiveCollectionsTest(t, dir)
	runGeneratedTests(t, dir)
}

// grpcRecursiveCollectionsDSL keeps the unconstrained result from #2515 and
// adds request and response types with validation and nested collection wrappers.
func grpcRecursiveCollectionsDSL() {
	category := d.ResultType("application/vnd.category", func() {
		d.TypeName("CategoryResult")
		d.Attributes(func() {
			d.Field(1, "id", d.Int)
			d.Field(3, "children_category", d.ArrayOf("CategoryResult"))
			d.Field(4, "name", d.String)
		})
	})
	node := d.Type("Node", func() {
		d.Field(1, "children", d.ArrayOf("Node"))
		d.Field(2, "by_name", d.MapOf(d.String, "Node"))
		d.Field(3, "groups", d.MapOf(d.String, d.ArrayOf("Node")))
		d.Field(4, "matrix", d.ArrayOf(d.ArrayOf("Node")))
		d.Field(5, "branches", d.ArrayOf("Branch"))
		d.Field(6, "name", d.String, func() {
			d.Pattern("^[a-z]+$")
		})
		d.Field(7, "count", d.Int)
		d.Field(8, "enabled", d.Boolean, func() {
			d.Default(true)
		})
		d.Required("name")
	})
	d.Type("Branch", func() {
		d.Field(1, "nodes", d.MapOf(d.String, node))
	})
	d.Service("categories", func() {
		d.Method("list", func() {
			d.Result(category)
			d.GRPC(func() {})
		})
		d.Method("exchange", func() {
			d.Payload(node)
			d.Result(node)
			d.GRPC(func() {})
		})
	})
}

// writeGRPCRecursiveCollectionsTest exercises all four generated conversion
// directions and verifies that validators inspect values below recursive fields.
func writeGRPCRecursiveCollectionsTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "recursivetest")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	const source = `package recursivetest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	gencategories "generated.local/gen/categories"
	genclient "generated.local/gen/grpc/categories/client"
	genpb "generated.local/gen/grpc/categories/pb"
	genserver "generated.local/gen/grpc/categories/server"
)

func TestRecursiveResultRoundTrip(t *testing.T) {
	id, name := 42, "category"
	result := &gencategories.CategoryResult{ID: &id, Name: &name, ChildrenCategory: []*gencategories.CategoryResult{
		{ID: &id, Name: &name, ChildrenCategory: []*gencategories.CategoryResult{{ID: &id, Name: &name}}},
	}}
	viewed := gencategories.NewViewedCategoryResult(result, "default")
	headers, trailers := metadata.MD{}, metadata.MD{}
	message, err := genserver.EncodeListResponse(context.Background(), viewed, &headers, &trailers)
	require.NoError(t, err)
	wire, ok := message.(proto.Message)
	require.True(t, ok)
	data, err := proto.Marshal(wire)
	require.NoError(t, err)
	received := wire.ProtoReflect().New().Interface()
	require.NoError(t, proto.Unmarshal(data, received))
	converted, err := genclient.DecodeListResponse(context.Background(), received, headers, trailers)
	require.NoError(t, err)
	require.Equal(t, result, converted)
}

func TestRecursiveCollectionsRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		value *gencategories.Node
	}{
		{
			name: "array",
			value: &gencategories.Node{Name: "root", Children: []*gencategories.Node{
				{Name: "child", Children: []*gencategories.Node{{Name: "leaf"}}},
			}},
		},
		{
			name: "map",
			value: &gencategories.Node{Name: "root", ByName: map[string]*gencategories.Node{
				"child": {Name: "child", ByName: map[string]*gencategories.Node{"leaf": {Name: "leaf"}}},
			}},
		},
		{
			name: "map of arrays",
			value: &gencategories.Node{Name: "root", Groups: map[string][]*gencategories.Node{
				"group": {{Name: "child", Groups: map[string][]*gencategories.Node{"group": {{Name: "leaf"}}}}},
			}},
		},
		{
			name: "array of arrays",
			value: &gencategories.Node{Name: "root", Matrix: [][]*gencategories.Node{
				{{Name: "child", Matrix: [][]*gencategories.Node{{{Name: "leaf"}}}}},
			}},
		},
		{
			name: "mutual recursion",
			value: &gencategories.Node{Name: "root", Branches: []*gencategories.Branch{
				{Nodes: map[string]*gencategories.Node{"child": {Name: "child", Branches: []*gencategories.Branch{
					{Nodes: map[string]*gencategories.Node{"leaf": {Name: "leaf"}}},
				}}}},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := genclient.NewProtoExchangeRequest(test.value)
			data, err := proto.Marshal(request)
			require.NoError(t, err)
			received := new(genpb.ExchangeRequest)
			require.NoError(t, proto.Unmarshal(data, received))
			require.NoError(t, genserver.ValidateExchangeRequest(received))
			payload := genserver.NewExchangePayload(received)
			require.Equal(t, test.value, payload)

			response := genserver.NewProtoExchangeResponse(payload)
			data, err = proto.Marshal(response)
			require.NoError(t, err)
			result := new(genpb.ExchangeResponse)
			require.NoError(t, proto.Unmarshal(data, result))
			require.NoError(t, genclient.ValidateExchangeResponse(result))
			require.Equal(t, test.value, genclient.NewExchangeResult(result))
		})
	}
}

func TestRecursiveCollectionsValidateDescendants(t *testing.T) {
	valid := "valid"
	invalid := "INVALID"
	for _, name := range []string{"array", "map", "map of arrays", "array of arrays", "mutual recursion"} {
		t.Run(name, func(t *testing.T) {
			leaf := &gencategories.Node{Name: valid}
			middle := &gencategories.Node{Name: valid}
			root := &gencategories.Node{Name: valid}
			for parent, child := range map[*gencategories.Node]*gencategories.Node{root: middle, middle: leaf} {
				switch name {
				case "array":
					parent.Children = []*gencategories.Node{child}
				case "map":
					parent.ByName = map[string]*gencategories.Node{"child": child}
				case "map of arrays":
					parent.Groups = map[string][]*gencategories.Node{"group": {child}}
				case "array of arrays":
					parent.Matrix = [][]*gencategories.Node{{child}}
				case "mutual recursion":
					parent.Branches = []*gencategories.Branch{{Nodes: map[string]*gencategories.Node{"child": child}}}
				}
			}
			require.NoError(t, genserver.ValidateExchangeRequest(genclient.NewProtoExchangeRequest(root)))
			leaf.Name = invalid
			require.Error(t, genserver.ValidateExchangeRequest(genclient.NewProtoExchangeRequest(root)))
			require.Error(t, genclient.ValidateExchangeResponse(genserver.NewProtoExchangeResponse(root)))
		})
	}
}

func TestRecursiveCollectionPresenceAndDefaults(t *testing.T) {
	name, zero, disabled := "valid", int32(0), false
	message := &genpb.ExchangeRequest{Name: &name, Children: []*genpb.Node{
		{Name: &name},
		{Name: &name, Count: &zero, Enabled: &disabled},
	}}
	require.NoError(t, genserver.ValidateExchangeRequest(message))
	payload := genserver.NewExchangePayload(message)
	require.Nil(t, payload.Children[0].Count)
	require.True(t, payload.Children[0].Enabled)
	require.NotNil(t, payload.Children[1].Count)
	require.Zero(t, *payload.Children[1].Count)
	require.False(t, payload.Children[1].Enabled)
	response := genserver.NewProtoExchangeResponse(payload)
	require.Equal(t, payload, genclient.NewExchangeResult(response))
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "recursive_test.go"), []byte(source), 0o600))
}
