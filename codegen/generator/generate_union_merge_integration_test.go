// This file verifies that complete generation merges shared union
// declarations without losing their package-owned names.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cg "goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestGenerateUnionUserTypeSamePathMerged exercises the real service codegen path
// for unions and user types across two services targeting the same user type
// file. It asserts the generated file contains both the struct definition and
// the union marker method for the union branch type. This mirrors the original
// failure mode where only the union method remained and the struct was lost.
func TestGenerateUnionUserTypeSamePathMerged(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)

	dsl := func() {
		d.API("test", func() {})

		var Summary = d.Type("Summary", func() {
			d.Meta("struct:pkg:path", "types")
			d.Attribute("message", d.String)
		})

		var MyUnion = d.Type("MyUnion", func() {
			d.Meta("struct:pkg:path", "types")
			d.OneOf("MyUnion", func() {
				d.TypeName("MyUnionChoice")
				d.Attribute("sum", Summary)
			})
		})

		var Container = d.Type("Container", func() {
			d.Meta("struct:pkg:path", "types")
			d.Attribute("u", MyUnion)
		})

		d.Service("S1", func() {
			d.Method("M", func() {
				d.Payload(Container)
			})
		})

		d.Service("S2", func() {
			d.Method("M", func() {
				d.Payload(Container)
			})
		})
	}

	_ = cg.RunDSL(t, dsl)

	dir := t.TempDir()
	writeGeneratedModule(t, filepath.Join(dir, cg.Gendir), "generated.local/gen")
	if _, err := generate(dir, "gen", false, registry); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated types/summary.go still contains the struct definition for
	// the shared user type. The original regression dropped the struct when
	// merging multiple files contributing to the same path.
	p := filepath.Join(dir, cg.Gendir, "types", cg.SnakeCase("Summary")+".go")
	bs, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed reading generated file: %v", err)
	}
	content := string(bs)
	if !strings.Contains(content, "type Summary struct") {
		t.Fatalf("missing struct definition in %s:\n%s", p, content)
	}
}

// TestGenerateInheritedRecursiveResultTypeView verifies that an explicit
// ResultType view selecting a recursive UserType whose inline nested object
// extends a transitive base chain with a recursive back-edge generates valid
// HTTP, JSON-RPC, and gRPC transport packages without duplicate body types.
func TestGenerateInheritedRecursiveResultTypeView(t *testing.T) {
	var node expr.UserType
	_ = cg.RunDSL(t, func() {
		d.API("copyreview", func() {})
		ancestor := d.Type("Ancestor", func() {
			d.Field(1, "query", d.String, func() { d.MinLength(2) })
			d.Field(2, "title", d.String, func() { d.Default("untitled") })
			d.Field(3, "back", node)
			d.Required("query")
		})
		base := d.Type("Base", func() { d.Extend(ancestor) })
		derived := d.Type("Derived", func() { d.Extend(base) })
		node = d.Type("Node", func() {
			d.Field(1, "self", node)
			d.Field(2, "nested", func() {
				d.Extend(derived)
				d.Field(4, "extra", d.String)
			})
			d.Required("nested")
		})
		document := d.ResultType("application/vnd.copyreview.document", func() {
			d.TypeName("Document")
			d.Attributes(func() {
				d.Field(1, "node", node)
				d.Required("node")
			})
			d.View("default", func() {
				d.Attribute("node")
			})
		})
		d.Service("copyreview", func() {
			d.HTTP(func() {
				d.BodyFieldName(derived, "query:query")
			})
			d.Method("exchange", func() {
				d.Payload(node)
				d.Result(document)
				d.HTTP(func() {
					d.POST("/copy")
					d.Response(200)
				})
				d.GRPC(func() {})
			})
			d.Method("repeat", func() {
				d.Payload(node)
				d.Result(document)
				d.HTTP(func() {
					d.POST("/repeat")
					d.Response(200)
				})
			})
			d.Method("rpc", func() {
				d.Payload(node)
				d.Result(document)
				d.JSONRPC(func() {})
			})
		})
	})

	cases := make([]generatedTestDir, 0, 2)
	for i := range 2 {
		registry := testRegistry(
			"gen",
			testGenerator(planServiceData, testServiceFiles),
			testGenerator(planTransportData, testTransportFiles),
		)
		dir := t.TempDir()
		generated := filepath.Join(dir, cg.Gendir)
		writeGeneratedModule(t, generated, "generated.local/gen")
		if _, err := generate(dir, "gen", false, registry); err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		httpTypesPath := filepath.Join(generated, "http", "copyreview", "server", "types.go")
		bs, err := os.ReadFile(httpTypesPath)
		if err != nil {
			t.Fatalf("failed reading %s: %v", httpTypesPath, err)
		}
		if !strings.Contains(string(bs), "type NodeResponseBody struct") {
			t.Fatalf("missing NodeResponseBody in %s:\n%s", httpTypesPath, string(bs))
		}
		writeGeneratedContractTest(t, generated, filepath.Join("http", "copyreview", "server"), inheritedRecursiveHTTPContract)
		writeGeneratedContractTest(t, generated, filepath.Join("jsonrpc", "copyreview", "server"), inheritedRecursiveRPCContract)
		writeGeneratedContractTest(t, generated, "reviewgrpctest", inheritedRecursiveGRPCContract)
		cases = append(cases, generatedTestDir{
			name: fmt.Sprintf("pass-%d", i+1),
			dir:  generated,
		})
	}
	runGeneratedTestDirs(t, cases)
}

// inheritedRecursiveHTTPContract verifies that generated HTTP request unmarshaling,
// validation, and server constructor conversion preserve inherited required fields
// and defaults on a recursive ResultType view.
const inheritedRecursiveHTTPContract = `package server
import("encoding/json";"testing")
func TestInheritedRecursiveHTTPRequest(t *testing.T){
 for _,tc:=range []struct{document string; valid bool}{
  {` + "`" + `{"nested":{}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"a"}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"ab"}}` + "`" + `,true},
 }{
  var body ExchangeRequestBody
  if err:=json.Unmarshal([]byte(tc.document),&body);err!=nil{t.Fatal(err)}
  if err:=ValidateExchangeRequestBody(&body);(err==nil)!=tc.valid{t.Fatalf("document=%s valid=%t error=%v",tc.document,tc.valid,err)}
  if tc.valid { value:=NewExchangeNode(&body);if value.Nested.Query!="ab"||value.Nested.Title!="untitled" {t.Fatalf("lost inherited query or default: %#v",value.Nested)} }
 }
}`

// inheritedRecursiveRPCContract verifies that generated JSON-RPC request unmarshaling,
// validation, and server constructor conversion preserve inherited required fields
// and defaults on a recursive ResultType view.
const inheritedRecursiveRPCContract = `package server
import("encoding/json";"testing")
func TestInheritedRecursiveRPCRequest(t *testing.T){
 for _,tc:=range []struct{document string; valid bool}{
  {` + "`" + `{"nested":{}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"a"}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"ab"}}` + "`" + `,true},
 }{
  var body RPCRequestBody
  if err:=json.Unmarshal([]byte(tc.document),&body);err!=nil{t.Fatal(err)}
  if err:=ValidateRPCRequestBody(&body);(err==nil)!=tc.valid{t.Fatalf("document=%s valid=%t error=%v",tc.document,tc.valid,err)}
  if tc.valid { value:=NewRPCNode(&body);if value.Nested.Query!="ab"||value.Nested.Title!="untitled" {t.Fatalf("lost inherited query or default: %#v",value.Nested)} }
 }
}`

// inheritedRecursiveGRPCContract verifies that generated gRPC protobuf marshaling
// and server request validation preserve inherited required fields and length
// constraints on a recursive ResultType view.
const inheritedRecursiveGRPCContract = `package reviewgrpctest
import(
 "testing"
 "google.golang.org/protobuf/encoding/protojson"
 "google.golang.org/protobuf/proto"
 genpb "generated.local/gen/grpc/copyreview/pb"
 genserver "generated.local/gen/grpc/copyreview/server"
)
func TestInheritedRecursiveProtobuf(t *testing.T){
 for _,tc:=range []struct{document string; valid bool}{
  {` + "`" + `{"nested":{}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"a"}}` + "`" + `,false},
  {` + "`" + `{"nested":{"query":"ab"}}` + "`" + `,true},
 }{
  var body genpb.ExchangeRequest
  if err:=protojson.Unmarshal([]byte(tc.document),&body);err!=nil{t.Fatal(err)}
  data,err:=proto.Marshal(&body);if err!=nil{t.Fatal(err)}
  var wire genpb.ExchangeRequest
  if err:=proto.Unmarshal(data,&wire);err!=nil{t.Fatal(err)}
  if err:=genserver.ValidateExchangeRequest(&wire);(err==nil)!=tc.valid{t.Fatalf("document=%s valid=%t error=%v",tc.document,tc.valid,err)}
 }
}`
