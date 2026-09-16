// This file builds protobuf schema files and compiles each one with the tools
// selected before rendering starts.
package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

const (
	// ProtoVersion is the protocol buffer version used to generate .proto files
	ProtoVersion = "proto3"

	// ProtoPrefix is the prefix added to the proto package name.
	ProtoPrefix = "goagen"
)

// defaultProtocCmd selects protoc when the design does not choose a compiler.
var defaultProtocCmd = []string{expr.DefaultProtoc}

// protoFiles returns the planned protobuf file for every gRPC service.
func protoFiles(services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.servicePlans))
	for i, servicePlan := range services.servicePlans {
		fw[i] = protoFile(servicePlan.expression, services)
	}
	return fw
}

// protoFile returns the protobuf file defining the specified service.
func protoFile(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	data := services.Get(svc.Name())
	svcName := data.Service.PathName
	fileServiceName := svcName
	if planned := services.protobuf[svc]; planned != nil && planned.fileIndex > 1 {
		fileServiceName += strconv.Itoa(planned.fileIndex)
	}
	parts := strings.Split(genpkg, "/")
	var repoName string
	if len(parts) > 1 {
		repoName = parts[len(parts)-2]
	} else {
		repoName = parts[0]
	}
	// Include the repository and service so two services do not write the same
	// file.
	fname := fmt.Sprintf("%s_%s_%s.proto", ProtoPrefix, repoName, fileServiceName)
	path := filepath.Join(codegen.Gendir, "grpc", svcName, pbPkgName, fname)

	sections := make([]*codegen.SectionTemplate, 0, 3+len(data.Messages))
	sections = append(sections,
		// header comments
		&codegen.SectionTemplate{
			Name:   "proto-header",
			Source: grpcTemplates.Read(grpcProtoHeaderT),
			Data: map[string]any{
				"Title": fmt.Sprintf("%s protocol buffer definition", svc.Name()),
			},
		},
		// proto syntax and package
		&codegen.SectionTemplate{
			Name:   "proto-start",
			Source: grpcTemplates.Read(grpcProtoStartT),
			Data: map[string]any{
				"ProtoVersion": ProtoVersion,
				"Pkg":          pkgName(svc, svcName),
				"Imports":      data.ProtoImports,
			},
		},
		// service definition
		&codegen.SectionTemplate{
			Name:   "grpc-service",
			Source: grpcTemplates.Read(grpcServiceT),
			Data:   data,
		},
	)

	// message definition
	for _, m := range data.Messages {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "grpc-message",
			Source: grpcTemplates.Read(grpcMessageT),
			Data:   m,
		})
	}

	tools := services.tools[svc]
	if tools == nil {
		panic(fmt.Sprintf("protobuf tools for service %q were not planned", svc.Name()))
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		FinalizeFunc: func(path string) error {
			return runProtoc(tools, path)
		},
	}
}

// pkgName returns the protobuf package chosen by the design or the service
// name used when the design does not choose one.
func pkgName(svc *expr.GRPCServiceExpr, svcName string) string {
	if svc.ProtoPkg != "" {
		return svc.ProtoPkg
	}
	return codegen.SnakeCase(svcName)
}

// runProtoc compiles one schema with the command fixed during planning.
func runProtoc(tools *protobufToolPlan, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	args := []string{
		path,
		"--proto_path", dir,
		"--go_out", dir,
		"--go-grpc_out", dir,
		"--go_opt=paths=source_relative",
		"--go-grpc_opt=paths=source_relative",
		"--plugin=protoc-gen-go=" + tools.goPlugin,
		"--plugin=protoc-gen-go-grpc=" + tools.goGRPCPlugin,
	}
	for _, include := range tools.includes {
		args = append(args, "-I", include)
	}
	command := tools.command
	cmd := exec.Command(command[0], append(command[1:len(command):len(command)], args...)...)
	cmd.Dir = filepath.Dir(path)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run protoc: %w: %s", err, output)
	}

	return nil
}
