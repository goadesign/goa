// This file checks that generated gRPC clients validate the fields selected by
// each result view before converting protobuf responses.
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

// TestGenerateGRPCViewedResponseValidation runs unary and streaming client
// decoders against complete and malformed protobuf responses for two views.
func TestGenerateGRPCViewedResponseValidation(t *testing.T) {
	root := codegen.RunDSL(t, grpcViewedResponseValidationDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGRPCViewedResponseValidationTest(t, dir)
	runGeneratedTests(t, dir)
}

// grpcViewedResponseValidationDSL defines required fields for two views. Two
// sibling fields share one nested type, and the tiny view selects only one.
func grpcViewedResponseValidationDSL() {
	d.API("view-validation", func() {})
	nested := d.Type("Nested", func() {
		d.Field(1, "value", d.String)
		d.Required("value")
	})
	item := d.ResultType("application/vnd.view-validation.item", func() {
		d.TypeName("Item")
		d.Attributes(func() {
			d.Field(1, "number", d.Int)
			d.Field(2, "text", d.String)
			d.Field(3, "selected", nested)
			d.Field(4, "omitted", nested)
			d.Required("number", "text", "selected", "omitted")
		})
		d.View("tiny", func() {
			d.Attribute("number")
			d.Attribute("selected")
		})
	})
	d.Service("views", func() {
		d.Method("Show", func() {
			d.Result(item)
			d.GRPC(func() {})
		})
		d.Method("Watch", func() {
			d.StreamingResult(item)
			d.GRPC(func() {})
		})
	})
}

// writeGRPCViewedResponseValidationTest writes tests that call the generated
// unary decoder and streaming Recv method through their public interfaces.
func writeGRPCViewedResponseValidationTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "viewvalidationtest")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create viewed response validation package: %v", err)
	}
	const source = `package viewvalidationtest_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	goa "goa.design/goa/v3/pkg"
	genclient "generated.local/gen/grpc/views/client"
	genpb "generated.local/gen/grpc/views/pb"
	genviews "generated.local/gen/views"
)

func TestUnaryViewedResponseValidation(t *testing.T) {
	zero := int32(0)
	empty := ""
	tests := []struct {
		name    string
		view    string
		message *genpb.ShowResponse
		missing string
		errName string
	}{
		{name: "valid tiny", view: "tiny", message: &genpb.ShowResponse{Number: &zero, Selected: nested("selected")}},
		{name: "malformed tiny", view: "tiny", message: &genpb.ShowResponse{Selected: nested("selected")}, missing: "number"},
		{name: "malformed shared tiny", view: "tiny", message: &genpb.ShowResponse{Number: &zero, Selected: &genpb.Nested{}}, missing: "value"},
		{name: "valid full", view: "default", message: &genpb.ShowResponse{Number: &zero, Text: &empty, Selected: nested("selected"), Omitted: nested("omitted")}},
		{name: "malformed full", view: "default", message: &genpb.ShowResponse{Number: &zero, Selected: nested("selected"), Omitted: nested("omitted")}, missing: "text"},
		{name: "unknown view", view: "unknown", message: &genpb.ShowResponse{Number: &zero, Text: &empty, Selected: nested("selected"), Omitted: nested("omitted")}, errName: "invalid_enum_value"},
	}
	for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
			result, err := genclient.DecodeShowResponse(context.Background(), test.message, metadata.Pairs("goa-view", test.view), nil)
			if test.missing != "" {
				assertMissingField(t, err, test.missing)
				return
			}
			if test.errName != "" {
				assertServiceErrorName(t, err, test.errName)
				return
			}
			if err != nil {
				t.Fatalf("decode %s response: %v", test.view, err)
			}
			if _, ok := result.(*genviews.Item); !ok {
				t.Errorf("unexpected decoded result %T", result)
			}
		})
	}
}

func TestStreamingViewedResponseValidation(t *testing.T) {
	zero := int32(0)
	empty := ""
	tests := []struct {
		name    string
		view    string
		message *genpb.WatchResponse
		missing string
		errName string
	}{
		{name: "valid tiny", view: "tiny", message: &genpb.WatchResponse{Number: &zero, Selected: nested("selected")}},
		{name: "malformed tiny", view: "tiny", message: &genpb.WatchResponse{Selected: nested("selected")}, missing: "number"},
		{name: "malformed shared tiny", view: "tiny", message: &genpb.WatchResponse{Number: &zero, Selected: &genpb.Nested{}}, missing: "value"},
		{name: "valid full", view: "default", message: &genpb.WatchResponse{Number: &zero, Text: &empty, Selected: nested("selected"), Omitted: nested("omitted")}},
		{name: "malformed full", view: "default", message: &genpb.WatchResponse{Number: &zero, Selected: nested("selected"), Omitted: nested("omitted")}, missing: "text"},
		{name: "unknown view", view: "unknown", message: &genpb.WatchResponse{Number: &zero, Text: &empty, Selected: nested("selected"), Omitted: nested("omitted")}, errName: "invalid_enum_value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := &watchClient{
				header:  metadata.Pairs("goa-view", test.view),
				message: test.message,
			}
			decoded, err := genclient.DecodeWatchResponse(context.Background(), transport, nil, nil)
			if err != nil {
				t.Fatalf("build generated stream: %v", err)
			}
			stream := decoded.(interface {
				Recv() (*genviews.Item, error)
			})
			result, err := stream.Recv()
			if test.missing != "" {
				assertMissingField(t, err, test.missing)
				return
			}
			if test.errName != "" {
				assertServiceErrorName(t, err, test.errName)
				return
			}
			if err != nil {
				t.Fatalf("receive %s response: %v", test.view, err)
			}
			if result == nil {
				t.Error("generated stream returned a nil result")
			}
		})
	}
}

func nested(value string) *genpb.Nested {
	return &genpb.Nested{Value: &value}
}

type watchClient struct {
	header  metadata.MD
	message *genpb.WatchResponse
	sent    bool
}

func (c *watchClient) Recv() (*genpb.WatchResponse, error) {
	if c.sent {
		return nil, io.EOF
	}
	c.sent = true
	return c.message, nil
}

func (c *watchClient) Header() (metadata.MD, error) { return c.header, nil }
func (c *watchClient) Trailer() metadata.MD         { return nil }
func (c *watchClient) CloseSend() error              { return nil }
func (c *watchClient) Context() context.Context      { return context.Background() }
func (c *watchClient) SendMsg(any) error             { return nil }
func (c *watchClient) RecvMsg(any) error             { return io.EOF }

var _ genpb.Views_WatchClient = (*watchClient)(nil)
var _ grpc.ClientStream = (*watchClient)(nil)

func assertMissingField(t *testing.T, err error, field string) {
	t.Helper()
	serviceError := assertServiceErrorName(t, err, "missing_field")
	if serviceError.Field == nil || *serviceError.Field != field {
		t.Errorf("expected missing field %q, got %#v", field, serviceError.Field)
	}
}

func assertServiceErrorName(t *testing.T, err error, name string) *goa.ServiceError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %q error", name)
	}
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Fatalf("expected service error, got %T: %v", err, err)
	}
	if serviceError.Name != name {
		t.Errorf("expected %q, got %q: %v", name, serviceError.Name, err)
	}
	return serviceError
}
`
	if err := os.WriteFile(filepath.Join(dir, "view_validation_test.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write viewed response validation test: %v", err)
	}
}
