// This file renders starter interceptor implementations that depend only on
// the service package and interceptor metadata, not service type packages.
package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

type (
	// exampleInterceptorData contains the canonical declarations and service
	// metadata rendered by one starter interceptor implementation.
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
// for every service retained by plan.
func ExampleInterceptorsFiles(plan *Plan) []*codegen.File {
	var fw []*codegen.File
	for _, facts := range plan.facts.services {
		if f := exampleInterceptorsFile(plan, facts); f != nil {
			fw = append(fw, f...)
		}
	}
	return fw
}

// exampleInterceptorsFile renders starter interceptors from one retained
// service.
func exampleInterceptorsFile(plan *Plan, facts *serviceFacts) []*codegen.File {
	genpkg := plan.generation.GenPkg()
	services := plan.Services()
	sdata := services.Get(facts.name)
	servicePath := path.Join(genpkg, sdata.PathName)
	servicePkg := services.aliases.name(servicePath)

	var files []*codegen.File

	// Generate server interceptor if needed and file doesn't exist
	if len(sdata.ServerInterceptors) > 0 {
		data := &exampleInterceptorData{
			ServiceName:            sdata.Name,
			ServicePkg:             servicePkg,
			StructDeclaration:      facts.exampleServerStruct,
			ConstructorDeclaration: facts.exampleServerConstructor,
			Interceptors:           sdata.ServerInterceptors,
		}
		serverPath := filepath.Join("interceptors", sdata.PathName+"_server.go")
		if _, err := os.Stat(serverPath); os.IsNotExist(err) {
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
			})
		}
	}

	// Generate client interceptor if needed and file doesn't exist
	if len(sdata.ClientInterceptors) > 0 {
		data := &exampleInterceptorData{
			ServiceName:            sdata.Name,
			ServicePkg:             servicePkg,
			StructDeclaration:      facts.exampleClientStruct,
			ConstructorDeclaration: facts.exampleClientConstructor,
			Interceptors:           sdata.ClientInterceptors,
		}
		clientPath := filepath.Join("interceptors", sdata.PathName+"_client.go")
		if _, err := os.Stat(clientPath); os.IsNotExist(err) {
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
			})
		}
	}

	return files
}
