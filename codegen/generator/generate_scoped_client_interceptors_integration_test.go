package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestGeneratedScopedClientInterceptors compiles complete service, transport,
// and example output, including a server that hosts only the second service.
func TestGeneratedScopedClientInterceptors(t *testing.T) {
	for _, transport := range []string{"grpc", "http", "jsonrpc"} {
		scopes := []string{"method", "API", "service", "none", "server"}
		if transport == "grpc" {
			scopes = append(scopes, "HTTP method")
		}
		for _, scope := range scopes {
			t.Run(transport+"/"+scope, func(t *testing.T) {
				root := codegen.RunDSL(t, func() {
					trace := dsl.Interceptor("Trace")
					dsl.API("scoped interceptors", func() {
						if scope == "API" {
							dsl.ClientInterceptor(trace)
						}
						dsl.Server("all", func() {
							dsl.Services("events", "quiet")
							dsl.Host("local", func() {
								if transport == "grpc" {
									dsl.URI("grpc://localhost:8080")
								} else {
									dsl.URI("http://localhost:8080")
								}
								if scope == "HTTP method" {
									dsl.URI("http://localhost:8081")
								}
							})
						})
						dsl.Server("quiet", func() {
							dsl.Services("quiet")
							dsl.Host("local", func() {
								if transport == "grpc" {
									dsl.URI("grpc://localhost:8082")
									if scope == "HTTP method" {
										dsl.URI("http://localhost:8083")
									}
								} else {
									dsl.URI("http://localhost:8082")
								}
							})
						})
					})
					dsl.Service("events", func() {
						if scope == "service" {
							dsl.ClientInterceptor(trace)
						}
						if scope == "server" {
							dsl.ServerInterceptor(trace)
						}
						dsl.Method("read", func() {
							if scope == "method" || scope == "HTTP method" {
								dsl.ClientInterceptor(trace)
							}
							dsl.Result(dsl.String)
							if scope == "HTTP method" {
								dsl.HTTP(func() {
									dsl.GET("/read")
								})
							} else {
								scopedInterceptorTransport(transport, "/read")
							}
						})
						dsl.Method("plain", func() {
							dsl.Result(dsl.String)
							scopedInterceptorTransport(transport, "/plain")
						})
						dsl.Method("watch", func() {
							if scope == "method" {
								dsl.ClientInterceptor(trace)
							}
							dsl.StreamingResult(dsl.String)
							if transport == "jsonrpc" {
								dsl.JSONRPC(func() {
									dsl.ServerSentEvents()
								})
							} else {
								scopedInterceptorTransport(transport, "/watch")
							}
						})
					})
					dsl.Service("quiet", func() {
						dsl.Method("read", func() {
							dsl.Result(dsl.String)
							scopedInterceptorTransport(transport, "/quiet")
							if scope == "HTTP method" {
								dsl.HTTP(func() {
									dsl.GET("/quiet")
								})
							}
						})
					})
				})
				plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
				files, err := testServiceFiles(plan)
				require.NoError(t, err)
				transports, err := testTransportFiles(plan)
				require.NoError(t, err)
				files = append(files, transports...)
				examples, err := assembleExampleFilesForTest(plan)
				require.NoError(t, err)
				files = append(files, examples...)
				files, err = mergeFilesByPath(files)
				require.NoError(t, err)

				dir := t.TempDir()
				writeGeneratedModule(t, dir, "generated.local")
				for _, file := range files {
					_, err := file.Render(dir)
					require.NoError(t, err)
				}
				for _, server := range []string{"all", "quiet"} {
					want := scope == "API" || (server == "all" && (scope == "method" || scope == "service" || scope == "HTTP method"))
					parser, err := os.ReadFile(filepath.Join(dir, "gen", transport, "cli", server, "cli.go"))
					require.NoError(t, err)
					require.Equal(t, want, strings.Contains(string(parser), ".ClientInterceptors"))
					example, err := os.ReadFile(filepath.Join(dir, "cmd", server+"-cli", transport+".go"))
					require.NoError(t, err)
					require.Equal(t, want, strings.Contains(string(example), `"generated.local/interceptors"`))
				}
				if scope == "method" || scope == "HTTP method" {
					require.NoError(t, os.WriteFile(filepath.Join(dir, "client_scope_test.go"), []byte(clientInterceptorScopeTest), 0o600))
				}
				runGeneratedTests(t, dir)
			})
		}
	}
}

// scopedInterceptorTransport exposes each fixture method through the transport
// under test, so the same service interface exercises all three import planners.
func scopedInterceptorTransport(transport, route string) {
	switch transport {
	case "http":
		dsl.HTTP(func() {
			dsl.GET(route)
		})
	case "jsonrpc":
		dsl.JSONRPC(func() {})
	case "grpc":
		dsl.GRPC(func() {})
	}
}

const clientInterceptorScopeTest = `package scopedinterceptors_test

import (
	"context"
	"testing"

	genevents "generated.local/gen/events"
	goa "goa.design/goa/v3/pkg"
)

type counter struct {
	calls int
}

func (c *counter) Trace(ctx context.Context, info genevents.TraceInfo, next goa.Endpoint) (any, error) {
	c.calls++
	return next(ctx, info.RawPayload())
}

func TestClientMethodScope(t *testing.T) {
	interceptors := &counter{}
	endpoint := func(context.Context, any) (any, error) {
		return "ok", nil
	}
	client := genevents.NewClient(endpoint, endpoint, endpoint, interceptors)
	if _, err := client.Read(context.Background()); err != nil {
		t.Fatal(err)
	}
	if interceptors.calls != 1 {
		t.Fatalf("intercepted read invoked Trace %d times", interceptors.calls)
	}
	if _, err := client.Plain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if interceptors.calls != 1 {
		t.Fatalf("unintercepted plain invoked Trace: calls=%d", interceptors.calls)
	}
}
`
