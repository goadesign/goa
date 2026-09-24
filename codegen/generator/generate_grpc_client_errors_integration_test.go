// This file generates a catalog client and runs it against a local gRPC server.
// The tests check which errors reach callers after encoding, RPC, and decoding.
package generator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestGenerateGRPCClientErrorIdentity checks error types through generated
// endpoints. The generated tests themselves run with the race detector.
func TestGenerateGRPCClientErrorIdentity(t *testing.T) {
	root := codegen.RunDSL(t, grpcClientErrorsDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transport...)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	fixtures, err := filepath.Glob("testdata/grpc_client_errors/*_test.go")
	require.NoError(t, err)
	require.NotEmpty(t, fixtures)
	target := filepath.Join(dir, "clienterrors")
	require.NoError(t, os.MkdirAll(target, 0o750))
	for _, fixture := range fixtures {
		source, err := os.ReadFile(fixture)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(target, filepath.Base(fixture)), source, 0o600))
	}

	// Bound compilation and the local server tests together. Passing -race to
	// this child command instruments the generated client, not just its generator.
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "-race", "-count=1", "-v", "./...")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	t.Logf("generated client tests:\n%s", output)
	require.NoError(t, err)
}

// grpcClientErrorsDSL gives ordinary and idempotent methods the same result so
// the fixture can compare validation without changing the requested operation.
func grpcClientErrorsDSL() {
	d.API("client-errors", func() {})
	selection := d.Type("Selection", func() {
		d.Field(1, "key", d.String)
		d.Required("key")
	})
	entry := d.Type("Entry", func() {
		d.Field(1, "state", d.String, func() {
			d.Enum("open", "closed")
		})
		d.Field(2, "title", d.String, func() {
			d.MinLength(1)
		})
		d.Required("state")
	})
	denied := d.Type("Denied", func() {
		d.Field(1, "reason", d.String)
		d.Required("reason")
	})
	metadataEntry := d.Type("MetadataEntry", func() {
		d.Field(1, "state", d.String)
		d.Field(2, "count", d.Int)
		d.Field(3, "flags", d.ArrayOf(d.Boolean))
		d.Required("state", "count", "flags")
	})
	viewed := d.ResultType("application/vnd.catalog.item", func() {
		d.TypeName("Item")
		d.Attributes(func() {
			d.Field(1, "state", d.String)
			d.Field(2, "title", d.String)
			d.Required("state", "title")
		})
		d.View("compact", func() {
			d.Attribute("state")
		})
	})
	d.Service("Catalog", func() {
		for _, method := range []struct {
			name       string
			idempotent bool
			errors     bool
		}{
			{"Read", false, false},
			{"ReadErrors", false, true},
			{"Retry", true, false},
			{"RetryErrors", true, true},
		} {
			d.Method(method.name, func() {
				d.Payload(selection)
				d.Result(entry)
				if method.idempotent {
					d.Idempotent()
				}
				if method.errors {
					d.Error("denied", denied)
					d.Error("busy", d.ErrorResult, func() {
						d.Temporary()
					})
					d.Error("invalid_enum_value", d.ErrorResult, func() {
						d.Temporary()
					})
				}
				d.GRPC(func() {
					if method.errors {
						d.Response("denied", d.CodePermissionDenied)
						d.Response("busy", d.CodeUnavailable)
						d.Response("invalid_enum_value", d.CodeUnavailable)
					}
				})
			})
		}
		for _, name := range []string{"Metadata", "RetryMetadata"} {
			d.Method(name, func() {
				d.Result(metadataEntry)
				if name == "RetryMetadata" {
					d.Idempotent()
					d.Error("missing_field", d.ErrorResult, func() {
						d.Temporary()
					})
					d.Error("invalid_field_type", d.ErrorResult, func() {
						d.Temporary()
					})
				}
				d.GRPC(func() {
					d.Response(func() {
						d.Headers(func() {
							d.Attribute("count")
						})
						d.Trailers(func() {
							d.Attribute("flags")
						})
					})
					if name == "RetryMetadata" {
						d.Response("missing_field", d.CodeUnavailable)
						d.Response("invalid_field_type", d.CodeUnavailable)
					}
				})
			})
		}
		d.Method("View", func() {
			d.Result(viewed)
			d.GRPC(func() {})
		})
		d.Method("Notify", func() {
			d.Payload(selection)
			d.GRPC(func() {})
		})
		d.Method("Watch", func() {
			d.Payload(selection)
			d.StreamingResult(entry)
			d.Error("denied", denied)
			d.GRPC(func() {
				d.Response("denied", d.CodePermissionDenied)
			})
		})
		d.Method("Upload", func() {
			d.Payload(selection)
			d.StreamingPayload(entry)
			d.Result(entry)
			d.Error("denied", denied)
			d.GRPC(func() {
				d.Response("denied", d.CodePermissionDenied)
			})
		})
		d.Method("Exchange", func() {
			d.Payload(selection)
			d.StreamingPayload(entry)
			d.StreamingResult(entry)
			d.Error("denied", denied)
			d.GRPC(func() {
				d.Response("denied", d.CodePermissionDenied)
			})
		})
		d.Method("Collect", func() {
			d.StreamingPayload(entry)
			d.Error("denied", denied)
			d.GRPC(func() {
				d.Response("denied", d.CodePermissionDenied)
			})
		})
	})
}
