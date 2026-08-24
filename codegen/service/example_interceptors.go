// This file renders starter interceptor implementations that depend only on
// the service package and interceptor metadata, not service type packages.
package service

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

type (
	// exampleInterceptorData contains the generated type and constructor names
	// plus service metadata rendered by one starter interceptor implementation.
	exampleInterceptorData struct {
		// ServiceName is the design service name described by the comments.
		ServiceName string
		// ServicePkg is the generated service package qualifier.
		ServicePkg string
		// StructDeclaration is the starter interceptor implementation type.
		StructDeclaration *codegen.NameDeclaration
		// ConstructorDeclaration creates StructDeclaration.
		ConstructorDeclaration *codegen.NameDeclaration
		// Interceptors contains the interceptor methods implemented by the type.
		Interceptors []*InterceptorData
	}
)

// ExampleInterceptorsFiles returns starter server and client interceptor files
// for every service copied into plan.
func ExampleInterceptorsFiles(plan *Plan) []*codegen.File {
	var fw []*codegen.File
	for _, facts := range plan.facts.services {
		if f := exampleInterceptorsFile(plan, facts); f != nil {
			fw = append(fw, f...)
		}
	}
	return fw
}

// exampleInterceptorsFile renders starter interceptors from one service copied
// into plan.
func exampleInterceptorsFile(plan *Plan, facts *serviceFacts) []*codegen.File {
	if len(facts.serverInterceptors) == 0 && len(facts.clientInterceptors) == 0 {
		return nil
	}
	genpkg := plan.generation.GenPkg()
	services := plan.Services()
	sdata := services.Get(facts.name)
	servicePath := path.Join(genpkg, sdata.PathName)
	servicePkg := services.aliases.name(path.Join(path.Dir(genpkg), "interceptors"), servicePath)

	var files []*codegen.File

	// Generate the server interceptor starter when the service uses one.
	if len(sdata.ServerInterceptors) > 0 {
		data := &exampleInterceptorData{
			ServiceName:            sdata.Name,
			ServicePkg:             servicePkg,
			StructDeclaration:      facts.exampleServerStruct,
			ConstructorDeclaration: facts.exampleServerConstructor,
			Interceptors:           sdata.ServerInterceptors,
		}
		serverPath := filepath.Join("interceptors", sdata.PathName+"_server.go")
		files = append(files, &codegen.File{
			Path: serverPath,
			SectionTemplates: []*codegen.SectionTemplate{
				codegen.Header(fmt.Sprintf("%s example server interceptors", sdata.Name), "interceptors", facts.imports.exampleServerInterceptors.specs),
				{
					Name:   "example-server-interceptor",
					Source: serviceTemplates.Read(exampleServerInterceptorT),
					Data:   data,
				},
			},
			SkipExist: true,
		})
	}

	// Generate the client interceptor starter when the service uses one.
	if len(sdata.ClientInterceptors) > 0 {
		data := &exampleInterceptorData{
			ServiceName:            sdata.Name,
			ServicePkg:             servicePkg,
			StructDeclaration:      facts.exampleClientStruct,
			ConstructorDeclaration: facts.exampleClientConstructor,
			Interceptors:           sdata.ClientInterceptors,
		}
		clientPath := filepath.Join("interceptors", sdata.PathName+"_client.go")
		files = append(files, &codegen.File{
			Path: clientPath,
			SectionTemplates: []*codegen.SectionTemplate{
				codegen.Header(fmt.Sprintf("%s example client interceptors", sdata.Name), "interceptors", facts.imports.exampleClientInterceptors.specs),
				{
					Name:   "example-client-interceptor",
					Source: serviceTemplates.Read(exampleClientInterceptorT),
					Data:   data,
				},
			},
			SkipExist: true,
		})
	}

	return files
}
