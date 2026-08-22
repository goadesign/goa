// This file renders HTTP request and response types per service and transport
// side, using imports attached to that exact generated type file.
package codegen

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// serverTypeFiles builds the server request and response types read by Plan.Link.
func serverTypeFiles(data *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(data.Expressions.Services))
	for i, svc := range data.Expressions.Services {
		fw[i] = addEndpointImports(typesFile(svc, true, data), data, svc.HTTPEndpoints...)
	}
	return fw
}

// clientTypeFiles builds the client request and response types read by Plan.Link.
func clientTypeFiles(data *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(data.Expressions.Services))
	for i, svc := range data.Expressions.Services {
		fw[i] = addEndpointImports(typesFile(svc, false, data), data, svc.HTTPEndpoints...)
	}
	return fw
}

// typesFile returns the file containing the type definitions used by the HTTP
// transport for the given service server (svr is true) or client.
//
// Below are the rules governing whether values are pointers or not. Note that
// the rules only apply to values that hold primitive types, values that hold
// slices, maps or objects always use pointers either implicitly - slices and
// maps - or explicitly - objects.
//
//   - The payload struct fields (if a struct) hold pointers when not required
//     and have no default value.
//
//   - Request body fields (if the body is a struct) always hold pointers to
//     allow for explicit validation. The same applies to response body fields
//     in the client code.
//
//   - Request header, path and query string parameter variables hold pointers
//     when not required. Request header, body fields and param variables that
//     have default values are never required (enforced by DSL engine).
//
//   - The result struct fields (if a struct) hold pointers when not required
//     or have a default value (so generated code can set when null).
//
//   - Response body fields (if the body is a struct) and header variables hold
//     pointers when not required and have no default value.
func typesFile(svc *expr.HTTPServiceExpr, svr bool, services *ServicesData) *codegen.File {
	var (
		data    = services.Get(svc.Name())
		svcName = data.Service.PathName

		side = "server"

		requestBodySection  = "request-body-type-decl"
		wsPayloadSection    = "request-stream-payload-type-decl"
		responseBodySection = "response-server-body"
		errorBodySection    = "error-body-type-decl"
		attributeSection    = "server-body-attributes"
		unionSection        = "server-union-type"
		bodyInitSection     = "server-body-init"
		validateSection     = "server-validate"
		bodyInitT           = serverBodyInitT
	)
	if !svr {
		side = "client"
		requestBodySection = "client-request-body"
		wsPayloadSection = "client-request-body"
		responseBodySection = "client-response-body"
		errorBodySection = "client-error-body"
		attributeSection = "client-body-attributes"
		unionSection = "client-union-type"
		bodyInitSection = "client-body-init"
		validateSection = "client-validate"
		bodyInitT = clientBodyInitT
	}
	unionTypes := data.wireTypes(svr).unionTypes()
	path := filepath.Join(codegen.Gendir, services.dir(), svcName, side, "types.go")
	imports := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "unicode/utf8"},
		services.ServiceImport(svc.Name()),
	}
	if serviceHasViewedResult(data, nil) {
		imports = append(imports, services.ViewImport(svc.Name()))
	}
	if len(unionTypes) > 0 {
		imports = append(imports, &codegen.ImportSpec{Path: "bytes"})
	}
	imports = append(imports, codegen.GoaImport(""))
	header := codegen.Header(svc.Name()+" "+services.label()+" "+side+" types", side, imports)

	var (
		initData       []*InitData
		validatedTypes []*TypeData

		sections = []*codegen.SectionTemplate{header}

		// seen tracks the canonical package records already emitted. Type names
		// and references are outputs of these records, never declaration
		// identity.
		seen          = make(map[*wireTypeRecord]struct{})
		seenInits     = make(map[string]struct{})
		seenValidated = make(map[*wireTypeRecord]struct{})
	)
	// addDecl emits the type declaration section if the type has a
	// definition.
	addDecl := func(name string, td *TypeData) {
		if td.declaration == nil || td.Def == "" {
			return
		}
		if _, ok := seen[td.declaration]; ok {
			return
		}
		seen[td.declaration] = struct{}{}
		declaration := td.declaration.data
		if declaration != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   name,
				Source: httpTemplates.Read(typeDeclT),
				Data:   declaration,
			})
		}
	}
	// addValidated records each package-owned validation helper once.
	addValidated := func(td *TypeData) {
		if td.declaration == nil || td.ValidateDef == "" {
			return
		}
		if _, ok := seenValidated[td.declaration]; ok {
			return
		}
		seenValidated[td.declaration] = struct{}{}
		validatedTypes = append(validatedTypes, td.declaration.data)
	}

	// request body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		var body, wsPayload *TypeData
		if svr {
			body = adata.Payload.Request.ServerBody
			if adata.ServerWebSocket != nil && !adata.IsJSONRPC {
				wsPayload = adata.ServerWebSocket.Payload
			}
		} else {
			body = adata.Payload.Request.ClientBody
			if adata.ClientWebSocket != nil {
				wsPayload = adata.ClientWebSocket.Payload
			}
		}
		for i, td := range []*TypeData{body, wsPayload} {
			if td == nil {
				continue
			}
			name := requestBodySection
			if i == 1 {
				name = wsPayloadSection
			}
			addDecl(name, td)
			// Server-side request constructors are emitted per endpoint
			// below ("server-payload-init"), only client bodies contribute
			// body constructors here.
			if !svr && td.Init != nil {
				if _, ok := seenInits[td.Init.Name]; !ok {
					seenInits[td.Init.Name] = struct{}{}
					initData = append(initData, td.Init)
				}
			}
			addValidated(td)
		}
	}

	// response body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		for _, resp := range adata.Result.Responses {
			bodies := resp.ServerBody
			if !svr {
				bodies = nil
				if len(resp.ViewedRepresentations) > 0 {
					for _, representation := range resp.ViewedRepresentations {
						bodies = append(bodies, representation.ClientBody)
					}
				} else if resp.ClientBody != nil {
					bodies = []*TypeData{resp.ClientBody}
				}
			}
			for _, td := range bodies {
				if td == nil {
					continue
				}
				addDecl(responseBodySection, td)
				if td.Init != nil {
					if _, ok := seenInits[td.Init.Name]; !ok {
						seenInits[td.Init.Name] = struct{}{}
						initData = append(initData, td.Init)
					}
				}
				addValidated(td)
			}
		}
	}

	// error body types
	for _, a := range svc.HTTPEndpoints {
		adata := data.Endpoint(a.Name())
		for _, gerr := range adata.Errors {
			for _, herr := range gerr.Errors {
				bodies := herr.Response.ServerBody
				if !svr {
					bodies = nil
					if herr.Response.ClientBody != nil {
						bodies = []*TypeData{herr.Response.ClientBody}
					}
				}
				for _, td := range bodies {
					addDecl(errorBodySection, td)
					if td.Init != nil {
						if _, ok := seenInits[td.Init.Name]; !ok {
							seenInits[td.Init.Name] = struct{}{}
							initData = append(initData, td.Init)
						}
					}
					addValidated(td)
				}
			}
		}
	}

	// body attribute types
	atts := data.ServerBodyAttributeTypes
	if !svr {
		atts = data.ClientBodyAttributeTypes
	}
	for _, td := range atts {
		addDecl(attributeSection, td)
		addValidated(td)
	}

	// union sum types
	for _, u := range unionTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   unionSection,
			Source: httpTemplates.Read(unionTypeT),
			Data:   u,
		})
	}

	// body constructors
	for _, init := range initData {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   bodyInitSection,
			Source: httpTemplates.Read(bodyInitT),
			Data:   init,
		})
	}

	if svr {
		// request to method payload constructors
		for _, adata := range data.Endpoints {
			if init := adata.Payload.Request.PayloadInit; init != nil {
				sections = append(sections, payloadInitSection(init))
			}
			if IsWebSocketEndpoint(adata) && adata.ServerWebSocket.Payload != nil {
				if init := adata.ServerWebSocket.Payload.Init; init != nil {
					sections = append(sections, payloadInitSection(init))
				}
			}
		}
	} else {
		// response and error response to method result constructors
		seenResultInits := make(map[string]struct{})
		for _, adata := range data.Endpoints {
			for _, resp := range adata.Result.Responses {
				inits := []*InitData{resp.ResultInit}
				for _, representation := range resp.ViewedRepresentations {
					inits = append(inits, representation.ResultInit)
				}
				for _, init := range inits {
					if init != nil {
						if _, ok := seenResultInits[init.Name]; !ok {
							seenResultInits[init.Name] = struct{}{}
							sections = append(sections, resultInitSection("client-result-init", init))
						}
					}
				}
			}
			for _, gerr := range adata.Errors {
				for _, herr := range gerr.Errors {
					if init := herr.Response.ResultInit; init != nil {
						if _, ok := seenResultInits[init.Name]; !ok {
							seenResultInits[init.Name] = struct{}{}
							sections = append(sections, resultInitSection("client-error-result-init", init))
						}
					}
				}
			}
		}
	}

	// validate methods
	for _, td := range validatedTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   validateSection,
			Source: httpTemplates.Read(validateT),
			Data:   td,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// payloadInitSection returns the section rendering the server-side constructor
// that builds a method payload from request elements.
func payloadInitSection(init *InitData) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:    "server-payload-init",
		Source:  httpTemplates.Read(serverTypeInitT),
		Data:    init,
		FuncMap: map[string]any{"fieldCode": fieldCode},
	}
}

// resultInitSection returns the section rendering the client-side constructor
// that builds a method result or error from response elements.
func resultInitSection(name string, init *InitData) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:    name,
		Source:  httpTemplates.Read(clientTypeInitT),
		Data:    init,
		FuncMap: map[string]any{"fieldCode": fieldCode},
	}
}

// fieldCode returns the code to initialize the return struct fields. It is
// used only in templates.
func fieldCode(init *InitData, typ string) string {
	varn := "res"
	if init.ReturnTypeAttribute == "" {
		varn = "v"
	}
	args := init.ServerArgs
	if typ == "client" {
		args = init.ClientArgs
	}
	initArgs := make([]*codegen.InitArgData, len(args))
	for i, arg := range args {
		initArgs[i] = &codegen.InitArgData{
			Name:         arg.VarName,
			Pointer:      arg.Pointer,
			Type:         arg.Type,
			FieldName:    arg.FieldName,
			FieldPointer: arg.FieldPointer,
			FieldType:    arg.FieldType,
		}
	}
	// We can ignore the transform helpers as there won't be any generated
	// because the headers and params cannot be user types.
	c, _, err := codegen.InitStructFields(initArgs, varn, "", init.ReturnTypePkg)
	if err != nil {
		panic(err) // bug
	}
	return c
}
