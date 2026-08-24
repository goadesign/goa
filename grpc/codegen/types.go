// This file renders gRPC client and server conversion types per service and
// attaches imports to the exact side-specific file that uses them.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

// serverTypeFiles returns the planned conversion types used by gRPC servers.
func serverTypeFiles(services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.servicePlans))
	for i, servicePlan := range services.servicePlans {
		fw[i] = addEndpointImports(typesFile(servicePlan, services, true), services, servicePlan)
	}
	return fw
}

// clientTypeFiles returns the planned conversion types used by gRPC clients.
func clientTypeFiles(services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.servicePlans))
	for i, servicePlan := range services.servicePlans {
		fw[i] = addEndpointImports(typesFile(servicePlan, services, false), services, servicePlan)
	}
	return fw
}

// typesFile returns the file defining the gRPC types for the given service.
// svr indicates whether the file is generated for the server (true) or the
// client (false) package.
func typesFile(servicePlan *grpcServicePlan, services *ServicesData, svr bool) *codegen.File {
	svc := servicePlan.expression
	var (
		initData []*InitData

		sd = services.Get(svc.Name())
	)
	{
		seen := make(map[*codegen.NameDeclaration]struct{})
		collect := func(c *ConvertData) {
			if c == nil || c.Init == nil {
				return
			}
			if _, ok := seen[c.Init.Declaration]; ok {
				return
			}
			seen[c.Init.Declaration] = struct{}{}
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
				for _, conversion := range ed.Response.ServerConverts {
					collect(conversion.Convert)
				}
				if ed.ServerStream != nil {
					collect(ed.ServerStream.SendConvert)
					for _, conversion := range ed.ServerStream.SendConverts {
						collect(conversion.Convert)
					}
					collect(ed.ServerStream.RecvConvert)
				}
				for _, e := range ed.Errors {
					collect(e.Response.ServerConvert)
				}
			} else {
				collect(ed.Request.ClientConvert)
				collect(ed.Response.ClientConvert)
				for _, conversion := range ed.Response.ClientConverts {
					collect(conversion.Convert)
				}
				if ed.ClientStream != nil {
					collect(ed.ClientStream.RecvConvert)
					for _, conversion := range ed.ClientStream.RecvConverts {
						collect(conversion.Convert)
					}
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
		outputPackage := path.Join(services.GenPkg(), "grpc", svcName, side)
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, side, "types.go")
		imports := []*codegen.ImportSpec{
			{Path: "unicode/utf8"},
			codegen.GoaImport(""),
			services.ServiceImport(outputPackage, svc.Name()),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		if serviceHasViewedResult(sd) {
			imports = append(imports, services.ViewImport(outputPackage, svc.Name()))
		}
		// Add imports if Any type is used
		if servicePlan.usesAnyInErrors {
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
		helpers := sd.clientTransformHelpers
		if svr {
			helpers = sd.serverTransformHelpers
		}
		for _, h := range helpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   side + "-transform-helper",
				Source: grpcTemplates.Read(grpcTransformHelperT),
				Data:   h,
			})
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
