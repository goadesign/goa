package codegen

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ClientTypeFiles returns the HTTP transport client types files.
func ClientTypeFiles(genpkg string, data *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(data.Expressions.Services))
	for i, svc := range data.Expressions.Services {
		fw[i] = clientType(genpkg, svc, make(map[string]struct{}), data)
	}
	return fw
}

// clientType return the file containing the type definitions used by the HTTP
// transport for the given service client. seen keeps track of the names of the
// types that have already been generated to prevent duplicate code generation.
//
// Below are the rules governing whether values are pointers or not. Note that
// the rules only applies to values that hold primitive types, values that hold
// slices, maps or objects always use pointers either implicitly - slices and
// maps - or explicitly - objects.
//
//   - The payload struct fields (if a struct) hold pointers when not required
//     and have no default value.
//
//   - Request and response body fields (if the body is a struct) always hold
//     pointers to allow for explicit validation.
//
//   - Request header, path and query string parameter variables hold pointers
//     when not required. Request header, body fields and param variables that
//     have default values are never required (enforced by DSL engine).
//
//   - The result struct fields (if a struct) hold pointers when not required
//     or have a default value (so generated code can set when null).
//
//   - Response header variables hold pointers when not required and have no
//     default value.
func clientType(genpkg string, svc *expr.HTTPServiceExpr, seen map[string]struct{}, services *ServicesData) *codegen.File {
	var (
		path    string
		data    = services.Get(svc.Name())
		svcName = data.Service.PathName
	)
	path = filepath.Join(codegen.Gendir, "http", svcName, "client", "types.go")
	imports := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "unicode/utf8"},
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
		codegen.GoaImport(""),
	}
	header := codegen.Header(svc.Name()+" HTTP client types", "client", imports)

	var (
		initData       []*InitData
		validatedTypes []*TypeData
		seenValidated  = make(map[string]struct{}) // Track validated types to avoid duplicates
		seenInit       = make(map[string]struct{}) // Track init functions to avoid duplicates

		sections = []*codegen.SectionTemplate{header}
	)

	// request body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		if data := adata.Payload.Request.ClientBody; data != nil {
			if _, ok := seen[data.Name]; ok {
				continue
			}
			seen[data.Name] = struct{}{}
			if data.Def != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-request-body",
					Source: httpTemplates.Read(typeDeclT),
					Data:   data,
				})
			}
			if data.Init != nil {
				if _, ok := seenInit[data.Init.Name]; !ok {
					seenInit[data.Init.Name] = struct{}{}
					initData = append(initData, data.Init)
				}
			}
			if data.ValidateDef != "" {
				if _, ok := seenValidated[data.Name]; !ok {
					seenValidated[data.Name] = struct{}{}
					validatedTypes = append(validatedTypes, data)
				}
			}
		}
		if adata.ClientWebSocket != nil {
			if data := adata.ClientWebSocket.Payload; data != nil {
				if _, ok := seen[data.Name]; ok {
					continue
				}
				seen[data.Name] = struct{}{}
				if data.Def != "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-request-body",
						Source: httpTemplates.Read(typeDeclT),
						Data:   data,
					})
				}
				if data.Init != nil {
					initData = append(initData, data.Init)
				}
				if data.ValidateDef != "" {
					if _, ok := seenValidated[data.Name]; !ok {
						seenValidated[data.Name] = struct{}{}
						validatedTypes = append(validatedTypes, data)
					}
				}
			}
		}
	}

	// response body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		for _, resp := range adata.Result.Responses {
			if data := resp.ClientBody; data != nil {
				if _, ok := seen[data.Name]; ok {
					continue
				}
				seen[data.Name] = struct{}{}
				if data.Def != "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-response-body",
						Source: httpTemplates.Read(typeDeclT),
						Data:   data,
					})
				}
				if data.ValidateDef != "" {
					if _, ok := seenValidated[data.Name]; !ok {
						seenValidated[data.Name] = struct{}{}
						validatedTypes = append(validatedTypes, data)
					}
				}
			}
		}
	}

	// error body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		for _, gerr := range adata.Errors {
			for _, herr := range gerr.Errors {
				if data := herr.Response.ClientBody; data != nil {
					if _, ok := seen[data.Name]; ok {
						continue
					}
					seen[data.Name] = struct{}{}
					if data.Def != "" {
						sections = append(sections, &codegen.SectionTemplate{
							Name:   "client-error-body",
							Source: httpTemplates.Read(typeDeclT),
							Data:   data,
						})
					}
					if data.ValidateDef != "" {
						if _, ok := seenValidated[data.Name]; !ok {
							seenValidated[data.Name] = struct{}{}
							validatedTypes = append(validatedTypes, data)
						}
					}
				}
			}
		}
	}

	for _, data := range data.ClientBodyAttributeTypes {
		// Check if this type has already been added to avoid duplicates
		if _, ok := seen[data.Name]; ok {
			continue
		}
		seen[data.Name] = struct{}{}

		if data.Def != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-body-attributes",
				Source: httpTemplates.Read(typeDeclT),
				Data:   data,
			})
		}

		if data.ValidateDef != "" {
			if _, ok := seenValidated[data.Name]; !ok {
				seenValidated[data.Name] = struct{}{}
				validatedTypes = append(validatedTypes, data)
			}
		}
	}

	// body constructors
	for _, init := range initData {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-body-init",
			Source: httpTemplates.Read(clientBodyInitT),
			Data:   init,
		})
	}

	// Track generated init functions to avoid duplicates
	seenInits := make(map[string]struct{})

	for _, adata := range data.Endpoints {
		// response to method result (client)
		for _, resp := range adata.Result.Responses {
			if init := resp.ResultInit; init != nil {
				if _, ok := seenInits[init.Name]; !ok {
					seenInits[init.Name] = struct{}{}
					sections = append(sections, &codegen.SectionTemplate{
						Name:    "client-result-init",
						Source:  httpTemplates.Read(clientTypeInitT),
						Data:    init,
						FuncMap: map[string]any{"fieldCode": fieldCode},
					})
				}
			}
		}

		// error response to method result (client)
		for _, gerr := range adata.Errors {
			for _, herr := range gerr.Errors {
				if init := herr.Response.ResultInit; init != nil {
					if _, ok := seenInits[init.Name]; !ok {
						seenInits[init.Name] = struct{}{}
						sections = append(sections, &codegen.SectionTemplate{
							Name:    "client-error-result-init",
							Source:  httpTemplates.Read(clientTypeInitT),
							Data:    init,
							FuncMap: map[string]any{"fieldCode": fieldCode},
						})
					}
				}
			}
		}
	}

	// body attribute types
	// validate methods
	for _, data := range validatedTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-validate",
			Source: httpTemplates.Read(validateT),
			Data:   data,
		})
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}
