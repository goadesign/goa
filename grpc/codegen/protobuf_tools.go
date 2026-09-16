// This file chooses the protobuf compiler and Go plugins before any generated
// file is built.
package codegen

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// protobufToolPlan stores the compiler, plugins, and include paths for one
	// service.
	protobufToolPlan struct {
		command      []string
		includes     []string
		goPlugin     string
		goGRPCPlugin string
	}

	// protobufToolResolver lets tests provide fixed program paths and versions
	// without changing PATH or running real programs.
	protobufToolResolver struct {
		resolve func(string) (string, error)
		version func(string) (string, error)
	}
)

const (
	protocGenGoName        = "protoc-gen-go"
	protocGenGoGRPCName    = "protoc-gen-go-grpc"
	protocGenGoVersion     = "protoc-gen-go v1.36.12"
	protocGenGoGRPCVersion = "protoc-gen-go-grpc 1.6.2"
)

// planProtobufTools copies the compiler settings for every gRPC service and
// uses one checked pair of Go plugins for one call to NewPlans.
func planProtobufTools(inputs []PlanInput, resolver protobufToolResolver) (map[*expr.GRPCServiceExpr]*protobufToolPlan, error) {
	goPlugin, err := resolveProtobufPlugin(resolver, protocGenGoName, protocGenGoVersion)
	if err != nil {
		return nil, err
	}
	goGRPCPlugin, err := resolveProtobufPlugin(resolver, protocGenGoGRPCName, protocGenGoGRPCVersion)
	if err != nil {
		return nil, err
	}

	plans := make(map[*expr.GRPCServiceExpr]*protobufToolPlan)
	compilers := make(map[string]string)
	for _, input := range inputs {
		for _, service := range input.Root.API.GRPC.Services {
			command := protobufCompilerCommand(input.Root, service)
			if len(command) == 0 {
				return nil, fmt.Errorf(`Meta("protoc:cmd"): must be given arguments`)
			}
			if plugin := replacedGoPlugin(command[1:]); plugin != "" {
				return nil, fmt.Errorf(`Meta("protoc:cmd") cannot replace required plugin %q`, plugin)
			}
			compiler := compilers[command[0]]
			if compiler == "" {
				compiler, err = resolver.resolve(command[0])
				if err != nil {
					return nil, fmt.Errorf("resolve protobuf compiler %q: %w", command[0], err)
				}
				compilers[command[0]] = compiler
			}
			command[0] = compiler
			includes := append([]string{}, service.ServiceExpr.Meta["protoc:include"]...)
			includes = append(includes, input.Root.API.Meta["protoc:include"]...)
			plans[service] = &protobufToolPlan{
				command:      command,
				includes:     includes,
				goPlugin:     goPlugin,
				goGRPCPlugin: goGRPCPlugin,
			}
		}
	}
	return plans, nil
}

// systemProtobufTools uses the programs available to Goa.
func systemProtobufTools() protobufToolResolver {
	return protobufToolResolver{
		resolve: resolveProtobufExecutable,
		version: protobufExecutableVersion,
	}
}

// protobufCompilerCommand copies the command selected by the service or API.
func protobufCompilerCommand(root *expr.RootExpr, service *expr.GRPCServiceExpr) []string {
	command := defaultProtocCmd
	if configured, ok := root.API.Meta["protoc:cmd"]; ok {
		command = configured
	}
	if configured, ok := service.ServiceExpr.Meta["protoc:cmd"]; ok {
		command = configured
	}
	return append([]string{}, command...)
}

// resolveProtobufPlugin finds one required plugin and checks its version.
func resolveProtobufPlugin(resolver protobufToolResolver, name, wantVersion string) (string, error) {
	path, err := resolver.resolve(name)
	if err != nil {
		return "", fmt.Errorf("resolve protobuf plugin %s: %w", name, err)
	}
	version, err := resolver.version(path)
	if err != nil {
		return "", fmt.Errorf("read protobuf plugin %s version: %w", name, err)
	}
	if !protobufPluginVersionMatches(name, version, wantVersion) {
		return "", fmt.Errorf("protobuf plugin %s reports version %s, want %s", name, version, wantVersion)
	}
	return path, nil
}

// protobufPluginVersionMatches accepts the program name printed on Unix and
// the same name with the executable suffix printed on Windows.
func protobufPluginVersionMatches(name, version, wantVersion string) bool {
	windowsVersion := name + ".exe" + strings.TrimPrefix(wantVersion, name)
	return version == wantVersion || version == windowsVersion
}

// resolveProtobufExecutable returns an absolute path for one executable.
func resolveProtobufExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("make executable path absolute: %w", err)
	}
	return path, nil
}

// protobufExecutableVersion returns the single version line printed by an
// executable.
func protobufExecutableVersion(path string) (string, error) {
	output, err := exec.Command(path, "--version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run %s --version: %w: %s", path, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// replacedGoPlugin returns a required plugin name when the flags try to
// replace it.
func replacedGoPlugin(arguments []string) string {
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		var plugin string
		switch {
		case argument == "--plugin" && index+1 < len(arguments):
			index++
			plugin = arguments[index]
		case strings.HasPrefix(argument, "--plugin="):
			plugin = strings.TrimPrefix(argument, "--plugin=")
		default:
			continue
		}
		var name string
		if configuredName, _, ok := strings.Cut(plugin, "="); ok {
			name = configuredName
		} else {
			name = filepath.Base(plugin)
		}
		if name == protocGenGoName || name == protocGenGoGRPCName {
			return name
		}
	}
	return ""
}
