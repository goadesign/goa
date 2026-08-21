// This file renders gRPC client and server conversion types per service and
// attaches imports to the exact side-specific file that uses them.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ServerTypeFiles returns the server types files containing all the server
// interfaces and types needed to implement gRPC server.
func ServerTypeFiles(services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.Root.API.GRPC.Services))
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = addEndpointImports(typesFile(svc, services, true), services, svc.GRPCEndpoints...)
	}
	return fw
}

// ClientTypeFiles returns the client types files containing all the client
// interfaces and types needed to implement gRPC client.
func ClientTypeFiles(services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.Root.API.GRPC.Services))
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = addEndpointImports(typesFile(svc, services, false), services, svc.GRPCEndpoints...)
	}
	return fw
}

// typesFile returns the file defining the gRPC types for the given service.
// svr indicates whether the file is generated for the server (true) or the
// client (false) package.
func typesFile(svc *expr.GRPCServiceExpr, services *ServicesData, svr bool) *codegen.File {
	var (
		initData []*InitData

		sd = services.Get(svc.Name())
	)
	{
		seen := make(map[string]struct{})
		collect := func(c *ConvertData) {
			if c == nil || c.Init == nil {
				return
			}
			if _, ok := seen[c.Init.Name]; ok {
				return
			}
			seen[c.Init.Name] = struct{}{}
			initData = append(initData, c.Init)
		}
		for _, a := range svc.GRPCEndpoints {
			ed := sd.Endpoint(a.Name())
			if svr {
				collect(ed.Request.ServerConvert)
				if ed.Request.LegacyDecode != nil {
					collect(ed.Request.LegacyDecode.ServerConvert)
				}
				collect(ed.Response.ServerConvert)
				if ed.ServerStream != nil {
					collect(ed.ServerStream.SendConvert)
					collect(ed.ServerStream.RecvConvert)
				}
				for _, e := range ed.Errors {
					collect(e.Response.ServerConvert)
				}
			} else {
				collect(ed.Request.ClientConvert)
				collect(ed.Response.ClientConvert)
				if ed.ClientStream != nil {
					collect(ed.ClientStream.RecvConvert)
					collect(ed.ClientStream.SendConvert)
				}
				for _, e := range ed.Errors {
					collect(e.Response.ClientConvert)
				}
			}
		}
	}

	side := "client"
	skipKind := validateServer
	if svr {
		side = "server"
		skipKind = validateClient
	}

	var (
		fpath    string
		sections []*codegen.SectionTemplate
	)
	{
		svcName := sd.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, side, "types.go")
		imports := []*codegen.ImportSpec{
			{Path: "unicode/utf8"},
			codegen.GoaImport(""),
			services.ServiceImport(svc.Name()),
			services.ViewImport(svc.Name()),
			services.PackageImport(path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		// Add imports if Any type is used
		if usesAnyType(svc.GRPCEndpoints, true) {
			imports = append(imports, &codegen.ImportSpec{Path: "fmt"})
			imports = append(imports, &codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"})
		}
		imports = append(imports, sd.Service.ProtoImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC "+side+" types", side, imports)}
		for _, init := range initData {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   side + "-type-init",
				Source: grpcTemplates.Read(grpcTypeInitT),
				Data:   init,
			})
		}
		for _, data := range sd.validations {
			if data.Kind == skipKind {
				continue
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   side + "-validate",
				Source: grpcTemplates.Read(grpcValidateT),
				Data:   data,
			})
		}
		for _, h := range sd.transformHelpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   side + "-transform-helper",
				Source: grpcTemplates.Read(grpcTransformHelperT),
				Data:   h,
			})
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
