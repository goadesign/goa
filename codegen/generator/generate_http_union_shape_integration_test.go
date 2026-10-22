// This file verifies that HTTP generation keeps request- and response-shaped
// unions separate when their nested branch types have different Go names.
package generator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

type (
	generatedModuleTemplate struct {
		goMod string
		goSum []byte
	}

	generatedTestDir struct {
		name string
		dir  string
	}
)

const generatedModulePlaceholder = "module __GENERATED_MODULE_PATH__"

const generatedModuleSeed = `package seed

import (
	_ "github.com/getkin/kin-openapi/openapi3"
	_ "github.com/gorilla/websocket"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
	_ "go.opentelemetry.io/otel"
	_ "goa.design/clue/debug"
	_ "goa.design/clue/health"
	_ "goa.design/clue/interceptors"
	_ "goa.design/clue/log"
	_ "goa.design/goa/v3/grpc"
	_ "goa.design/goa/v3/grpc/pb"
	_ "goa.design/goa/v3/http"
	_ "goa.design/goa/v3/jsonrpc"
	_ "goa.design/goa/v3/pkg"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
	_ "google.golang.org/grpc"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/metadata"
	_ "google.golang.org/grpc/reflection"
	_ "google.golang.org/grpc/status"
	_ "google.golang.org/grpc/test/bufconn"
	_ "google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/reflect/protoreflect"
	_ "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "gopkg.in/yaml.v3"
)
`

var (
	parallelGeneratedTests sync.Map
	cachedGoaModuleDir     = sync.OnceValues(func() (string, error) {
		return filepath.Abs(filepath.Join("..", ".."))
	})
	cachedGeneratedModule = sync.OnceValues(func() (generatedModuleTemplate, error) {
		goaRoot, err := cachedGoaModuleDir()
		if err != nil {
			return generatedModuleTemplate{}, err
		}
		tmpDir, err := os.MkdirTemp("", "goa-generated-mod-template-*")
		if err != nil {
			return generatedModuleTemplate{}, err
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		initialGoMod := generatedModulePlaceholder + "\n\ngo 1.25.0\n\n" +
			"require (\n\tgoa.design/goa/v3 v3.0.0\n\tgoa.design/clue v1.2.6\n)\n\n" +
			"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(initialGoMod), 0o600); err != nil {
			return generatedModuleTemplate{}, err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "seed.go"), []byte(generatedModuleSeed), 0o600); err != nil {
			return generatedModuleTemplate{}, err
		}
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return generatedModuleTemplate{}, contextError("go mod tidy template", err, out)
		}
		goModBytes, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if err != nil {
			return generatedModuleTemplate{}, err
		}
		goSumBytes, err := os.ReadFile(filepath.Join(tmpDir, "go.sum"))
		if err != nil {
			return generatedModuleTemplate{}, err
		}
		return generatedModuleTemplate{
			goMod: string(goModBytes),
			goSum: goSumBytes,
		}, nil
	})
)

func contextError(prefix string, err error, out []byte) error {
	return &execError{prefix: prefix, err: err, output: string(out)}
}

type execError struct {
	prefix string
	err    error
	output string
}

func (e *execError) Error() string {
	return e.prefix + ": " + e.err.Error() + "\n" + e.output
}

func parallelGeneratedTest(t *testing.T) {
	t.Helper()
	if _, loaded := parallelGeneratedTests.LoadOrStore(t, struct{}{}); !loaded {
		t.Parallel()
	}
}

func runGeneratedTestDirs(t *testing.T, cases []generatedTestDir) {
	t.Helper()
	parallelGeneratedTest(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runGeneratedTests(t, tc.dir)
		})
	}
}

func TestGenerateHTTPUnionUsedByRequestAndResponseCompiles(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)

	dsl := func() {
		d.API("test", func() {})

		siteSet := d.Type("SiteSet", func() {
			d.Attribute("site_ids", d.ArrayOf(d.String))
			d.Required("site_ids")
		})
		allSites := d.Type("AllSites", func() {
			d.Attribute("include_current", d.Boolean)
			d.Required("include_current")
		})
		setup := d.Type("Setup", func() {
			d.OneOf("scope", func() {
				d.Attribute("site_set", siteSet)
				d.Attribute("all_sites", allSites)
			})
			d.Required("scope")
		})

		d.Service("front", func() {
			d.Method("configure", func() {
				d.Payload(setup)
				d.Result(setup)
				d.HTTP(func() {
					d.POST("/configure")
					d.Response(200)
				})
			})
			d.Method("reconfigure", func() {
				d.Payload(setup)
				d.Result(d.String)
				d.HTTP(func() {
					d.POST("/reconfigure")
					d.Response(200)
				})
			})
		})
	}

	_ = codegen.RunDSL(t, dsl)
	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "generated.local/gen")
	if _, err := generate(dir, "gen", false, registry); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	assertGeneratedUnionDeclarations(t, genDir)
	runGeneratedTests(t, genDir)
}

// assertGeneratedUnionDeclarations proves identical request derivations reuse
// one union while the differently shaped response receives another.
func assertGeneratedUnionDeclarations(t *testing.T, genDir string) {
	t.Helper()
	path := filepath.Join(genDir, "http", "front", "server", "types.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated server types: %v", err)
	}
	code := string(content)
	if strings.Count(code, "type ScopeRequestBody struct {") != 1 ||
		strings.Count(code, "type ScopeResponseBody struct {") != 1 {
		t.Fatalf("expected one request union and one response union:\n%s", code)
	}
	if strings.Contains(code, "type Scope2 struct {") {
		t.Fatalf("union declarations received a numeric suffix:\n%s", code)
	}
	if strings.Count(code, "type SiteSetRequestBody struct {") != 1 ||
		strings.Count(code, "type SiteSetResponseBody struct {") != 1 {
		t.Fatalf("expected one request branch and one response branch declaration:\n%s", code)
	}
	if !strings.Contains(code, "Scope *ScopeRequestBody `") ||
		!strings.Contains(code, "Scope ScopeResponseBody `") {
		t.Fatalf("request and response bodies do not use their role names:\n%s", code)
	}
	if strings.Count(code, "\tsiteSet  *SiteSetRequestBody\n") != 1 ||
		strings.Count(code, "\tsiteSet  *SiteSetResponseBody\n") != 1 {
		t.Fatalf("unions do not reference their request and response branches:\n%s", code)
	}
}

// writeGeneratedModule creates a temporary module that resolves this Goa
// checkout explicitly instead of downloading a released generator runtime.
func writeGeneratedModule(t *testing.T, dir, modulePath string) {
	t.Helper()
	tmpl, err := cachedGeneratedModule()
	if err != nil {
		t.Fatalf("prepare generated module template: %v", err)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create generated module directory: %v", err)
	}
	module := strings.Replace(tmpl.goMod, generatedModulePlaceholder, "module "+modulePath, 1)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o600); err != nil {
		t.Fatalf("write generated go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), tmpl.goSum, 0o600); err != nil {
		t.Fatalf("write generated go.sum: %v", err)
	}
}

// runGeneratedTests compiles every rendered package in the isolated module.
func runGeneratedTests(t *testing.T, dir string) {
	t.Helper()
	parallelGeneratedTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile generated packages: %v\n%s", err, output)
	}
}
