// This file checks that gRPC planning fixes the protobuf commands before any
// generated file is rendered or compiled.
package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestNewPlansRetainsProtobufCommands checks that later design and PATH changes
// cannot replace the compiler, plugins, or include paths chosen by NewPlans.
func TestNewPlansRetainsProtobufCommands(t *testing.T) {
	root := grpcPlanRoots(t, "Calc")[0]
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	recordPath := filepath.Join(t.TempDir(), "compiler-arguments")
	compilerPath, err := filepath.Abs(os.Args[0])
	require.NoError(t, err)
	root.API.Meta = make(expr.MetaExpr)
	root.API.GRPC.Services[0].ServiceExpr.Meta = make(expr.MetaExpr)
	root.API.Meta["protoc:cmd"] = []string{
		compilerPath,
		"-test.run=TestProtobufCompilerProcess",
		"--",
		recordPath,
	}
	root.API.Meta["protoc:include"] = []string{"api-before"}
	root.API.GRPC.Services[0].ServiceExpr.Meta["protoc:include"] = []string{"service-before"}

	resolver := protobufToolResolver{
		resolve: func(name string) (string, error) {
			switch name {
			case compilerPath:
				return compilerPath, nil
			case protocGenGoName:
				return "/planned/protoc-gen-go", nil
			case protocGenGoGRPCName:
				return "/planned/protoc-gen-go-grpc", nil
			default:
				t.Fatalf("unexpected executable lookup %q", name)
				return "", nil
			}
		},
		version: func(path string) (string, error) {
			switch path {
			case "/planned/protoc-gen-go":
				return protocGenGoVersion, nil
			case "/planned/protoc-gen-go-grpc":
				return protocGenGoGRPCVersion, nil
			default:
				t.Fatalf("unexpected version check %q", path)
				return "", nil
			}
		},
	}
	plans, err := newPlans(generation, resolver, PlanInput{Root: root, Service: services[0]})
	require.NoError(t, err)

	root.API.Meta["protoc:cmd"] = []string{"compiler-after"}
	root.API.Meta["protoc:include"] = []string{"api-after"}
	root.API.GRPC.Services[0].ServiceExpr.Meta["protoc:include"] = []string{"service-after"}
	t.Setenv("PATH", t.TempDir())
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())

	grpcService := root.API.GRPC.Services[0]
	renderData := newServicesData(services[0].Services(), plans[0])
	renderData.GRPCServices[grpcService.Name()] = &ServiceData{
		Service: services[0].Services().Get(grpcService.Name()),
	}
	files := protoFiles(renderData)
	require.Len(t, files, 1)
	t.Setenv("GO_WANT_PROTOBUF_COMPILER_PROCESS", "1")
	require.NoError(t, files[0].FinalizeFunc(filepath.Join(t.TempDir(), "service.proto")))
	encoded, err := os.ReadFile(recordPath)
	require.NoError(t, err)
	arguments := strings.Split(string(encoded), "\n")
	require.Contains(t, arguments, "--plugin=protoc-gen-go=/planned/protoc-gen-go")
	require.Contains(t, arguments, "--plugin=protoc-gen-go-grpc=/planned/protoc-gen-go-grpc")
	require.Contains(t, arguments, "service-before")
	require.Contains(t, arguments, "api-before")
	require.NotContains(t, arguments, "service-after")
	require.NotContains(t, arguments, "api-after")
}

// TestNewPlansChecksProtobufPluginVersions checks both required plugin
// versions before planning succeeds.
func TestNewPlansChecksProtobufPluginVersions(t *testing.T) {
	tests := []struct {
		name        string
		plugin      string
		gotVersion  string
		wantVersion string
	}{
		{"Go plugin", protocGenGoName, "protoc-gen-go v1.36.11", protocGenGoVersion},
		{"gRPC plugin", protocGenGoGRPCName, "protoc-gen-go-grpc 1.6.1", protocGenGoGRPCVersion},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := grpcPlanRoots(t, "Calc")[0]
			generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
			resolver := fixedProtobufToolResolver()
			resolver.version = func(path string) (string, error) {
				if filepath.Base(path) == test.plugin {
					return test.gotVersion, nil
				}
				if filepath.Base(path) == protocGenGoName {
					return protocGenGoVersion, nil
				}
				return protocGenGoGRPCVersion, nil
			}
			_, err := newPlans(generation, resolver, PlanInput{Root: root, Service: services[0]})
			require.EqualError(t, err, "protobuf plugin "+test.plugin+" reports version "+
				test.gotVersion+", want "+test.wantVersion)
		})
	}
}

// TestResolveProtobufPluginAcceptsWindowsExecutableName checks the version text
// printed when Windows adds its executable suffix to the plugin name.
func TestResolveProtobufPluginAcceptsWindowsExecutableName(t *testing.T) {
	resolver := fixedProtobufToolResolver()
	resolver.version = func(string) (string, error) {
		return "protoc-gen-go.exe v1.36.12", nil
	}

	plugin, err := resolveProtobufPlugin(resolver, protocGenGoName, protocGenGoVersion)

	require.NoError(t, err)
	require.Equal(t, "/tools/protoc-gen-go", plugin)
}

// TestNewPlansRejectsGoPluginOverrides checks every protoc flag form that
// could replace either required Go plugin.
func TestNewPlansRejectsGoPluginOverrides(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"Go plugin equals", []string{"--plugin=protoc-gen-go=/other/go"}},
		{"gRPC plugin equals", []string{"--plugin=protoc-gen-go-grpc=/other/grpc"}},
		{"Go plugin separate", []string{"--plugin", "protoc-gen-go=/other/go"}},
		{"gRPC plugin separate", []string{"--plugin", "protoc-gen-go-grpc=/other/grpc"}},
		{"Go plugin path", []string{"--plugin=/other/protoc-gen-go"}},
		{"gRPC plugin path", []string{"--plugin=/other/protoc-gen-go-grpc"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := grpcPlanRoots(t, "Calc")[0]
			generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
			root.API.Meta = make(expr.MetaExpr)
			root.API.Meta["protoc:cmd"] = append([]string{"protoc"}, test.args...)
			_, err := newPlans(
				generation,
				fixedProtobufToolResolver(),
				PlanInput{Root: root, Service: services[0]},
			)
			require.ErrorContains(t, err, `Meta("protoc:cmd") cannot replace`)
		})
	}
}

// TestProtobufCompilerProcess records the compiler arguments for its parent
// test and exits without compiling the schema.
func TestProtobufCompilerProcess(t *testing.T) {
	if os.Getenv("GO_WANT_PROTOBUF_COMPILER_PROCESS") != "1" {
		return
	}
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || len(os.Args) <= separator+1 {
		os.Exit(2)
	}
	recordPath := os.Args[separator+1]
	arguments := strings.Join(os.Args[separator+2:], "\n")
	if err := os.WriteFile(recordPath, []byte(arguments), 0o600); err != nil {
		os.Exit(3)
	}
	os.Exit(0)
}

// fixedProtobufToolResolver returns stable paths and required versions.
func fixedProtobufToolResolver() protobufToolResolver {
	return protobufToolResolver{
		resolve: func(name string) (string, error) {
			return "/tools/" + filepath.Base(name), nil
		},
		version: func(path string) (string, error) {
			if filepath.Base(path) == protocGenGoName {
				return protocGenGoVersion, nil
			}
			return protocGenGoGRPCVersion, nil
		},
	}
}

// protoc compiles a schema directly for tests that inspect protobuf output.
func protoc(command []string, path string) error {
	if len(command) == 0 {
		return fmt.Errorf("protobuf compiler command is empty")
	}
	resolver := systemProtobufTools()
	goPlugin, err := resolveProtobufPlugin(resolver, protocGenGoName, protocGenGoVersion)
	if err != nil {
		return err
	}
	goGRPCPlugin, err := resolveProtobufPlugin(resolver, protocGenGoGRPCName, protocGenGoGRPCVersion)
	if err != nil {
		return err
	}
	compiler, err := resolver.resolve(command[0])
	if err != nil {
		return fmt.Errorf("resolve protobuf compiler %q: %w", command[0], err)
	}
	return runProtoc(&protobufToolPlan{
		command:      append([]string{compiler}, command[1:]...),
		goPlugin:     goPlugin,
		goGRPCPlugin: goGRPCPlugin,
	}, path)
}
