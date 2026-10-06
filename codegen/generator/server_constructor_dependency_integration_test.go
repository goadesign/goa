// These tests compile server constructors and their generated examples from
// the same dependency plan. HTTP and JSON-RPC must retain the supplied value,
// and factory names must remain consistent when another declaration collides.
package generator

import (
	"cmp"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type constructorDependencyOrder string

// ComparePackageName makes factory collisions independent of declaration order.
func (o constructorDependencyOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return cmp.Compare(string(o), string(other.(constructorDependencyOrder)))
}

func TestServerConstructorDependenciesCompileWithExamples(t *testing.T) {
	for _, test := range []struct {
		mode           string
		beforeExamples bool
	}{
		{mode: "http"},
		{mode: "jsonrpc"},
		{mode: "mixed"},
		{mode: "files"},
		{mode: "multipart"},
		{mode: "websocket"},
		{mode: "sse"},
		{mode: "http", beforeExamples: true},
		{mode: "jsonrpc", beforeExamples: true},
		{mode: "mixed", beforeExamples: true},
	} {
		mode := test.mode
		t.Run(mode, func(t *testing.T) {
			root := codegen.RunDSL(t, func() {
				dsl.API("records", func() {})
				dsl.Service("records", func() {
					if mode == "files" {
						dsl.Files("/assets/{*path}", "assets")
						return
					}
					if mode == "multipart" {
						dsl.Method("upload", func() {
							dsl.Payload(dsl.String)
							dsl.HTTP(func() {
								dsl.POST("/upload")
								dsl.MultipartRequest()
							})
						})
						return
					}
					if mode == "websocket" || mode == "sse" {
						dsl.Method("watch", func() {
							dsl.StreamingResult(dsl.String)
							dsl.HTTP(func() {
								dsl.GET("/watch")
								if mode == "sse" {
									dsl.ServerSentEvents()
								}
							})
						})
						return
					}
					dsl.Method("read", func() {
						dsl.Result(dsl.String)
						if mode != "http" {
							dsl.JSONRPC(func() {})
						} else {
							dsl.HTTP(func() { dsl.GET("/records") })
						}
					})
					if mode == "mixed" {
						dsl.Method("native", func() {
							dsl.Result(dsl.String)
							dsl.HTTP(func() { dsl.GET("/native") })
						})
					}
				})
			})
			var checks []*codegen.File
			declare := func(plan *Plan) error {
				dependencyType, err := codegen.PlanGoType(&expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"struct:field:type": {"*http.Client", "net/http", "http"}},
				}, codegen.GoTypePlanOptions{Owner: "generated.local"})
				if err != nil {
					return err
				}
				locationType, err := codegen.PlanGoType(&expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"struct:field:type": {"*url.URL", "net/url", "url"}},
				}, codegen.GoTypePlanOptions{Owner: "generated.local"})
				if err != nil {
					return err
				}
				if transport, exists := plan.HTTP(root); exists {
					service := root.API.HTTP.Services[0]
					if _, err := transport.DeclareServerConstructorDependency(service, "location", locationType, "NewRecords", constructorDependencyOrder("http location")); err != nil {
						return err
					}
					if _, err := transport.DeclareServerConstructorDependency(service, "client", dependencyType, "NewRecords", constructorDependencyOrder("http")); err != nil {
						return err
					}
				}
				if transport, exists := plan.JSONRPC(root); exists {
					service := root.API.JSONRPC.Services[0]
					if _, err := transport.DeclareServerConstructorDependency(service, "location", locationType, "NewRecords", constructorDependencyOrder("jsonrpc location")); err != nil {
						return err
					}
					if _, err := transport.DeclareServerConstructorDependency(service, "client", dependencyType, "NewRecords", constructorDependencyOrder("jsonrpc")); err != nil {
						return err
					}
				}
				return nil
			}
			planners := []func(*Plan) error{planExampleData, declare}
			if test.beforeExamples {
				planners = []func(*Plan) error{planTransportData, declare, planExampleData}
			}
			plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planners...)
			files, err := testServiceFiles(plan)
			require.NoError(t, err)
			transports, err := testTransportFiles(plan)
			require.NoError(t, err)
			files = append(files, transports...)
			examples, err := assembleExampleFilesForTest(plan)
			require.NoError(t, err)
			files = append(files, examples...)
			for _, dialect := range []string{"http", "jsonrpc"} {
				if mode != "mixed" && dialect != mode {
					continue
				}
				check := &codegen.File{
					Path: path.Join("gen", dialect, "records", "server", "dependency_test.go"),
					SectionTemplates: []*codegen.SectionTemplate{{
						Name:   "dependency-value-test",
						Source: constructorDependencyCheck,
						Data:   struct{ HTTP, Mixed bool }{HTTP: dialect == "http", Mixed: mode == "mixed"},
					}},
				}
				checks = append(checks, check)
			}
			files = append(files, checks...)
			files, err = mergeFilesByPath(files)
			require.NoError(t, err)
			directory := t.TempDir()
			writeGeneratedModule(t, directory, "generated.local")
			for _, file := range files {
				_, err := file.Render(directory)
				require.NoError(t, err)
			}
			runGeneratedTests(t, directory)
		})
	}
}

const constructorDependencyCheck = `package server
import (
    "context"
    "net/http"
    "net/url"
    "reflect"
    "testing"

    goahttp "goa.design/goa/v3/http"
    genrecords "generated.local/gen/records"
)
type fixtureService struct{}
func (s *fixtureService) Read(context.Context) (string, error) { return "accepted", nil }
{{ if .Mixed }}
func (s *fixtureService) Native(context.Context) (string, error) { return "native", nil }
{{ end }}
func TestConstructorRetainsDependency(t *testing.T) {
    client := &http.Client{}
    location := &url.URL{Scheme: "https", Host: "example.com"}
    server := New(genrecords.NewEndpoints(&fixtureService{}), goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder, nil{{ if .HTTP }}, nil{{ end }}, client, location)
    signature := reflect.TypeOf(New)
    expected := {{ if .HTTP }}8{{ else }}7{{ end }}
    if signature.IsVariadic() || signature.NumIn() != expected {
        t.Error("dependency became an optional constructor argument")
    }
    if server.location != location {
        t.Error("constructor changed the supplied location")
    }
    if server.client != client {
        t.Error("constructor changed the supplied dependency")
    }
}
`
