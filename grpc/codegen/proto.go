package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// ProtoFiles returns a *.proto file for each gRPC service.
func ProtoFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.API.GRPC.Services))
	for i, svc := range root.API.GRPC.Services {
		fw[i] = protoFile(genpkg, svc)
	}
	return fw
}

func protoFile(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	data := GRPCServices.Get(svc.Name())
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
			Source: protoHeaderT,
			Data: map[string]any{
				"Title":       fmt.Sprintf("%s protocol buffer definition", svc.Name()),
				"ToolVersion": goa.Version(),
			},
		},
		// proto syntax and package
		{
			Name:   "proto-start",
			Source: protoStartT,
			Data: map[string]any{
				"ProtoVersion": ProtoVersion,
				"Pkg":          pkgName(svc, svcName),
				"Imports":      data.ProtoImports,
			},
		},
		// service definition
		{
			Name:   "grpc-service",
			Source: serviceT,
			Data:   data,
		},
	}

	// message definition
	for _, m := range data.Messages {
		sections = append(sections, &codegen.SectionTemplate{Name: "grpc-message", Source: messageT, Data: m})
	}

	runProtoc := func(path string) error {
		includes := svc.ServiceExpr.Meta["protoc:include"]
		includes = append(includes, expr.Root.API.Meta["protoc:include"]...)
		return protoc(path, includes)
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

func protoc(path string, includes []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

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
	cmd := exec.Command("protoc", args...)
	cmd.Dir = filepath.Dir(path)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run protoc: %s: %s", err, output)
	}

	return nil
}

const (
	protoHeaderT = `{{ if .Title -}}
// Code generated with goa {{ .ToolVersion }}, DO NOT EDIT.
//
// {{ .Title }}
//
// Command:
{{ comment commandLine }}
{{- end }}
`

	protoStartT = `
syntax = {{ printf "%q" .ProtoVersion }};

package {{ .Pkg }};

option go_package = "/{{ .Pkg }}pb";
{{- range .Imports }}
import "{{ . }}";
{{- end }}
`

	// input: ServiceData
	serviceT = `
{{ .Description | comment }}
service {{ .Name }} {
	{{- range .Endpoints }}
	{{ if .Method.Description }}{{ .Method.Description | comment }}{{ end }}
	{{- $serverStream := or (eq .Method.StreamKind 3) (eq .Method.StreamKind 4) }}
	{{- $clientStream := or (eq .Method.StreamKind 2) (eq .Method.StreamKind 4) }}
	rpc {{ .Method.VarName }} ({{ if $clientStream }}stream {{ end }}{{ .Request.Message.VarName }}) returns ({{ if $serverStream }}stream {{ end }}{{ .Response.Message.VarName }});
	{{- end }}
}
`

	// input: service.UserTypeData
	messageT = `{{ comment .Description }}
message {{ .VarName }}{{ .Def }}
`
)
