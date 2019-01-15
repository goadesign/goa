package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
	svcName := codegen.SnakeCase(svc.Name())
	path := filepath.Join(codegen.Gendir, "grpc", svcName, pbPkgName, svcName+".proto")
	data := GRPCServices.Get(svc.Name())

	title := fmt.Sprintf("%s protocol buffer definition", svc.Name())
	sections := []*codegen.SectionTemplate{
		Header(title, svc.Name(), []*codegen.ImportSpec{}),
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

const (
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
