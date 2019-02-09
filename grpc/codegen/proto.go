package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/pkg"
)

const (
	// ProtoVersion is the protocol buffer version used to generate .proto files
	ProtoVersion = "proto3"
)

// ProtoFiles returns a *.proto file for each gRPC service.
func ProtoFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.API.GRPC.Services))
	for i, svc := range root.API.GRPC.Services {
		fw[i] = protoFile(genpkg, root.API, svc)
	}
	return fw
}

func protoFile(genpkg string, api *expr.APIExpr, svc *expr.GRPCServiceExpr) *codegen.File {
	svcName := codegen.SnakeCase(svc.Name())
	path := filepath.Join(codegen.Gendir, "grpc", svcName, pbPkgName, svcName+".proto")
	data := GRPCServices.Get(svc.Name())

	title := fmt.Sprintf("%s protocol buffer definition", svc.Name())
	sections := []*codegen.SectionTemplate{
		header(title, api.Name, svc.Name(), []*codegen.ImportSpec{}),
		&codegen.SectionTemplate{
			Name:   "grpc-service",
			Source: serviceT,
			Data:   data,
		},
	}

	for _, m := range data.Messages {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "grpc-message",
			Source: messageT,
			Data:   m,
		})
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		FinalizeFunc:     protoc,
	}
}

func protoc(path string) error {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0777)

	args := []string{"--go_out=plugins=grpc:.", path, "--proto_path", dir}
	cmd := exec.Command("protoc", args...)
	cmd.Dir = filepath.Dir(path)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run protoc: %s: %s", err, output)
	}

	return nil
}

// header returns a proto source file header section template.
func header(title, pack, gopkg string, imports []*codegen.ImportSpec) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:   "source-header",
		Source: headerT,
		Data: map[string]interface{}{
			"Title":        title,
			"ToolVersion":  pkg.Version(),
			"ProtoVersion": ProtoVersion,
			"Pkg":          codegen.SnakeCase(codegen.Goify(pack, false)),
			"GoPkg":        codegen.SnakeCase(codegen.Goify(gopkg, false)),
			"Imports":      imports,
		},
	}
}

const (
	headerT = `{{ if .Title -}}
// Code generated with goa {{ .ToolVersion }}, DO NOT EDIT.
//
// {{ .Title }}
//
// Command:
{{ comment commandLine }}
{{- end }}

syntax = {{ printf "%q" .ProtoVersion }};

package {{ .Pkg }};

option go_package = "{{ .GoPkg }}pb";

{{ range .Imports }}
import {{ .Code }};
{{ end }}`

	// input: ServiceData
	serviceT = `{{ .Description | comment }}
service {{ .Name }} {
	{{- range .Endpoints }}
	{{ if .Method.Description }}{{ .Method.Description | comment }}{{ end }}
	{{- $serverStream := or (eq .Method.StreamKind 3) (eq .Method.StreamKind 4) }}
	{{- $clientStream := or (eq .Method.StreamKind 2) (eq .Method.StreamKind 4) }}
	rpc {{ .Method.VarName }} ({{ if $clientStream }}stream {{ end }}{{ .Request.Message.Name }}) returns ({{ if $serverStream }}stream {{ end }}{{ .Response.Message.Name }});
	{{- end }}
}
`

	// input: service.UserTypeData
	messageT = `{{ comment .Description }}
message {{ .VarName }}{{ .Def }}
`
)
