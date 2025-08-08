package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	goa "goa.design/goa/v3/pkg"
)

const (
	// ProtoVersion is the protocol buffer version used to generate .proto files
	ProtoVersion = "proto3"

	// ProtoPrefix is the prefix added to the proto package name.
	ProtoPrefix = "goagen"
)

// ProtoFiles returns the protobuf file for every gRPC service.
func ProtoFiles(genpkg string, services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.Root.API.GRPC.Services))
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = protoFile(genpkg, svc, services)
	}
	return fw
}

// protoFile returns the protobuf file defining the specified service.
func protoFile(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	parts := strings.Split(genpkg, "/")
	var repoName string
	if len(parts) > 1 {
		repoName = parts[len(parts)-2]
	} else {
		repoName = parts[0]
	}
	// the filename is used by protoc to set the namespace so try to make it unique
	fname := fmt.Sprintf("%s_%s_%s.proto", ProtoPrefix, repoName, svcName)
	path := filepath.Join(codegen.Gendir, "grpc", svcName, pbPkgName, fname)

	sections := []*codegen.SectionTemplate{
		// header comments
		{
			Name:   "proto-header",
			Source: grpcTemplates.Read(grpcProtoHeaderT),
			Data: map[string]any{
				"Title":       fmt.Sprintf("%s protocol buffer definition", svc.Name()),
				"ToolVersion": goa.Version(),
			},
		},
		// proto syntax and package
		{
			Name:   "proto-start",
			Source: grpcTemplates.Read(grpcProtoStartT),
			Data: map[string]any{
				"ProtoVersion": ProtoVersion,
				"Pkg":          pkgName(svc, svcName),
				"Imports":      data.ProtoImports,
			},
		},
		// service definition
		{
			Name:   "grpc-service",
			Source: grpcTemplates.Read(grpcServiceT),
			Data:   data,
		},
	}

	// message definition
	for _, m := range data.Messages {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "grpc-message",
			Source: grpcTemplates.Read(grpcMessageT),
			Data:   m,
		})
	}

	runProtoc := func(path string) error {
		includes := svc.ServiceExpr.Meta["protoc:include"]
		includes = append(includes, services.Root.API.Meta["protoc:include"]...)

		cmd := defaultProtocCmd
		if c, ok := services.Root.API.Meta["protoc:cmd"]; ok {
			cmd = c
		}
		if c, ok := svc.ServiceExpr.Meta["protoc:cmd"]; ok {
			cmd = c
		}
		if len(cmd) == 0 {
			return fmt.Errorf(`Meta("protoc:cmd"): must be given arguments`)
		}

		return protoc(cmd, path, includes)
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		FinalizeFunc:     runProtoc,
	}
}

func pkgName(svc *expr.GRPCServiceExpr, svcName string) string {
	if svc.ProtoPkg != "" {
		return svc.ProtoPkg
	}
	return codegen.SnakeCase(svcName)
}

var defaultProtocCmd = []string{expr.DefaultProtoc}

func protoc(protocCmd []string, path string, includes []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// Build args for protoc
	args := []string{
		path,
		"--proto_path", dir,
		"--go_out", dir,
		"--go-grpc_out", dir,
		"--go_opt=paths=source_relative",
		"--go-grpc_opt=paths=source_relative",
	}
	for _, include := range includes {
		args = append(args, "-I", include)
	}

	// Resolve plugins robustly across platforms
	pluginArgs, env := resolveProtocPlugins()
	args = append(args, pluginArgs...)

	cmd := exec.Command(protocCmd[0], append(protocCmd[1:len(protocCmd):len(protocCmd)], args...)...)
	cmd.Dir = filepath.Dir(path)
	if len(env) > 0 {
		cmd.Env = env
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run protoc: %w: %s", err, output)
	}

	return nil
}

// resolveProtocPlugins attempts to locate protoc plugins and augment PATH for cross-platform reliability.
// It returns additional protoc arguments (e.g., --plugin=...) and an environment slice that includes an augmented PATH.
func resolveProtocPlugins() (extraArgs []string, env []string) {
	currentEnv := os.Environ()
	pathEnv := os.Getenv("PATH")

	// Gather potential bin directories
	bins := make([]string, 0, 4)
	if gobin := goEnv("GOBIN"); gobin != "" {
		bins = append(bins, gobin)
	}
	if gopath := goEnv("GOPATH"); gopath != "" {
		// GOPATH may be a list; take all
		sep := ":"
		if runtime.GOOS == "windows" {
			sep = ";"
		}
		for _, p := range strings.Split(gopath, sep) {
			if p == "" {
				continue
			}
			bins = append(bins, filepath.Join(p, "bin"))
		}
	}
	// Also add common locations
	bins = append(bins, filepath.Join(os.Getenv("HOME"), "go", "bin"))

	// Augment PATH
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	augmented := strings.Join(append([]string{pathEnv}, bins...), sep)

	// Compose new env with augmented PATH
	seenPath := false
	for i, e := range currentEnv {
		if strings.HasPrefix(e, "PATH=") || strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			currentEnv[i] = "PATH=" + augmented
			seenPath = true
			break
		}
	}
	if !seenPath {
		currentEnv = append(currentEnv, "PATH="+augmented)
	}

	// Try to resolve plugin paths explicitly and pass --plugin flags when available.
	plugins := []struct{
		name   string
		flag   string
	}{
		{"protoc-gen-go", "protoc-gen-go"},
		{"protoc-gen-go-grpc", "protoc-gen-go-grpc"},
	}

	for _, pl := range plugins {
		path := lookPathCrossPlatform(pl.name, bins)
		if path != "" {
			extraArgs = append(extraArgs, "--plugin="+pl.flag+"="+path)
		}
	}

	return extraArgs, currentEnv
}

// goEnv runs `go env KEY` and returns the trimmed output or empty string on error.
func goEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// lookPathCrossPlatform tries exec.LookPath first, then searches in provided bins with OS extension handling.
func lookPathCrossPlatform(base string, bins []string) string {
	// First try the environment PATH
	if p, err := exec.LookPath(base); err == nil {
		return p
	}

	exts := []string{""}
	if runtime.GOOS == "windows" {
		exts = []string{".exe", ".bat", ".cmd", ""}
	}
	for _, b := range bins {
		for _, ext := range exts {
			candidate := filepath.Join(b, base+ext)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return candidate
			}
		}
	}
	return ""
}
