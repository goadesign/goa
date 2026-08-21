// This file renders service declarations and aggregates relocated declarations
// into the exact generated Go packages and files that own them.
package service

import (
	"fmt"
	"path/filepath"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Files returns every service-local file from analyses and emits each relocated
// user type and union once across the complete generation.
func Files(genpkg string, analyses []*ServicesData) []*codegen.File {
	var files []*codegen.File
	for _, services := range analyses {
		for _, service := range services.Root.Services {
			files = append(files, serviceFiles(genpkg, service, services)...)
		}
	}
	return append(files, generatedPackageFiles(genpkg, analyses)...)
}

// serviceFiles renders the declarations and helpers owned exclusively by one
// service package. Relocated declarations and all union definitions are
// emitted later by generatedPackageFiles.
func serviceFiles(genpkg string, service *expr.ServiceExpr, services *ServicesData) []*codegen.File {
	svc := services.Get(service.Name)
	svcName := svc.PathName
	svcPath := filepath.Join(codegen.Gendir, svcName, "service.go")
	seen := make(map[string]struct{})
	typeDefSections := make(map[string]*codegen.SectionTemplate)
	svcSections := make([]*codegen.SectionTemplate, 0, 10)

	addTypeDefSection := func(name string, section *codegen.SectionTemplate) {
		typeDefSections[name] = section
		seen[name] = struct{}{}
	}

	for i, m := range svc.Methods {
		method := service.Methods[i]
		if m.PayloadLoc == nil && m.PayloadDef != "" {
			if _, ok := seen[m.Payload]; !ok {
				addTypeDefSection(m.Payload, &codegen.SectionTemplate{
					Name:   "service-payload",
					Source: serviceTemplates.Read(payloadT),
					Data:   m,
				})
			}
		}
		if method.StreamingPayload != nil && codegen.UserTypeLocation(method.StreamingPayload.Type) == nil && m.StreamingPayloadDef != "" {
			if _, ok := seen[m.StreamingPayload]; !ok {
				addTypeDefSection(m.StreamingPayload, &codegen.SectionTemplate{
					Name:   "service-streaming-payload",
					Source: serviceTemplates.Read(streamingPayloadT),
					Data:   m,
				})
			}
		}
		if m.ResultLoc == nil && m.ResultDef != "" {
			if _, ok := seen[m.Result]; !ok {
				addTypeDefSection(m.Result, &codegen.SectionTemplate{
					Name:   "service-result",
					Source: serviceTemplates.Read(resultT),
					Data:   m,
				})
			}
		}
		// Generate streaming result type if different from result
		if method.StreamingResult != nil && codegen.UserTypeLocation(method.StreamingResult.Type) == nil && m.StreamingResultDef != "" && m.StreamingResult != m.Result {
			if _, ok := seen[m.StreamingResult]; !ok {
				addTypeDefSection(m.StreamingResult, &codegen.SectionTemplate{
					Name:   "service-streaming-result",
					Source: serviceTemplates.Read(resultT),
					Data: map[string]any{
						"Result":     m.StreamingResult,
						"ResultDef":  m.StreamingResultDef,
						"ResultDesc": m.StreamingResultDesc,
					},
				})
			}
		}
	}
	for _, ut := range svc.userTypes {
		if ut.Loc == nil {
			if _, ok := seen[ut.VarName]; !ok {
				addTypeDefSection(ut.VarName, &codegen.SectionTemplate{
					Name:   "service-user-type",
					Source: serviceTemplates.Read(userTypeT),
					Data:   ut,
				})
			}
		}
	}

	seenErrs := make(map[string]struct{})
	for _, et := range svc.errorTypes {
		if et.Type == expr.ErrorResult || et.Loc != nil {
			continue
		}
		if _, ok := seenErrs[et.Name]; !ok {
			seenErrs[et.Name] = struct{}{}
			if _, ok := seen[et.Name]; !ok {
				addTypeDefSection(et.Name, &codegen.SectionTemplate{
					Name:   "error-user-type",
					Source: serviceTemplates.Read(userTypeT),
					Data:   et,
				})
			}
			typeDefSections["|"+et.Name] = &codegen.SectionTemplate{
				Name:    "service-error",
				Source:  serviceTemplates.Read(errorT),
				FuncMap: map[string]any{"errorName": errorName},
				Data:    et,
			}
		}
	}
	for _, er := range svc.errorInits {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "error-init-func",
			Source: serviceTemplates.Read(errorInitT),
			Data:   er,
		})
	}

	// transform result type functions
	for _, t := range svc.viewedResultTypes {
		svcSections = append(svcSections,
			&codegen.SectionTemplate{Name: "viewed-result-type-to-service-result-type", Source: serviceTemplates.Read(typeInitT), Data: t.ResultInit},
			&codegen.SectionTemplate{Name: "service-result-type-to-viewed-result-type", Source: serviceTemplates.Read(typeInitT), Data: t.Init})
	}
	var projh []*codegen.TransformFunctionData
	for _, t := range svc.projectedTypes {
		for _, i := range t.TypeInits {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "projected-type-to-service-type",
				Source: serviceTemplates.Read(typeInitT),
				Data:   i,
			})
		}
		for _, i := range t.Projections {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "service-type-to-projected-type",
				Source: serviceTemplates.Read(typeInitT),
				Data:   i,
			})
		}
	}

	for _, h := range projh {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "transform-helpers",
			Source: serviceTemplates.Read(transformHelperT),
			Data:   h,
		})
	}

	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("context"),
		codegen.SimpleImport("io"),
		codegen.GoaImport(""),
		codegen.GoaImport("security"),
		codegen.NewImport(svc.ViewsPkg, genpkg+"/"+svcName+"/views"),
	}
	outputPackage := genpkg + "/" + svcName
	attributes := serviceReferenceAttributes(service)
	attributes = append(attributes, normalizedMethodDefinitions(service)...)
	for _, userType := range svc.userTypes {
		if userType.Loc == nil {
			attributes = append(attributes, userType.Type.Attribute())
		}
	}
	for _, errorType := range svc.errorTypes {
		if errorType.Loc == nil {
			attributes = append(attributes, errorType.Type.Attribute())
		}
	}
	imports = append(imports, services.AttributeImports(outputPackage, attributes...)...)
	header := codegen.Header(service.Name+" service", svc.PkgName, imports)
	def := &codegen.SectionTemplate{
		Name:   "service",
		Source: serviceTemplates.Read(serviceT),
		Data:   svc,
		FuncMap: map[string]any{
			"hasJSONRPCStreaming": hasJSONRPCStreaming,
			"isJSONRPCWebSocket":  hasJSONRPCWebSocket,
			"streamInterfaceFor":  streamInterfaceFor,
			"dedupeByResult":      dedupeByResult,
		},
	}

	names := make([]string, 0, len(typeDefSections))
	for name := range typeDefSections {
		names = append(names, name)
	}
	sort.Strings(names)
	sections := make([]*codegen.SectionTemplate, 0, 2+len(names)+len(svcSections))
	sections = append(sections, header, def)
	for _, name := range names {
		sections = append(sections, typeDefSections[name])
	}
	sections = append(sections, svcSections...)
	files := []*codegen.File{{Path: svcPath, SectionTemplates: sections}}
	return append(files, InterceptorsFiles(genpkg, service, services)...)
}

// normalizedMethodDefinitions returns the underlying object definitions
// emitted for semantic method wrappers in service.go. Their nested references
// contribute imports even though endpoint and client files stop at the wrapper
// declaration.
func normalizedMethodDefinitions(service *expr.ServiceExpr) []*expr.AttributeExpr {
	var definitions []*expr.AttributeExpr
	for _, method := range service.Methods {
		definitions = append(definitions, normalizedMethodDefinitionsFor(method)...)
	}
	return definitions
}

// normalizedMethodDefinitionsFor returns the raw object definitions emitted
// for one method so service.go imports their nested references.
func normalizedMethodDefinitionsFor(method *expr.MethodExpr) []*expr.AttributeExpr {
	var definitions []*expr.AttributeExpr
	definitions = appendNormalizedMethodDefinition(definitions, method, method.Payload, "Payload")
	definitions = appendNormalizedMethodDefinition(definitions, method, method.StreamingPayload, "StreamingPayload")
	definitions = appendNormalizedMethodDefinition(definitions, method, method.Result, "Result")
	if method.HasMixedResults() {
		definitions = appendNormalizedMethodDefinition(definitions, method, method.StreamingResult, "StreamingResult")
	}
	return definitions
}

// appendNormalizedMethodDefinition appends the underlying object only when
// attribute is the semantic wrapper created for the requested method role.
func appendNormalizedMethodDefinition(definitions []*expr.AttributeExpr, method *expr.MethodExpr, attribute *expr.AttributeExpr, suffix string) []*expr.AttributeExpr {
	if attribute == nil {
		return definitions
	}
	userType, ok := attribute.Type.(expr.UserType)
	if !ok || userType.ID() != normalizedMethodTypeID(method.Service, method, suffix) {
		return definitions
	}
	return append(definitions, userType.Attribute())
}

// generatedPackageFiles renders each relocated user type in its configured
// file and one sorted unions.go for every package that owns unions.
func generatedPackageFiles(genpkg string, analyses []*ServicesData) []*codegen.File {
	packages := aggregateGeneratedPackages(analyses)
	if len(packages) == 0 {
		return nil
	}
	aliases := analyses[0].aliases
	packagePaths := make([]string, 0, len(packages))
	for packagePath := range packages {
		packagePaths = append(packagePaths, packagePath)
	}
	sort.Strings(packagePaths)

	var files []*codegen.File
	for _, packagePath := range packagePaths {
		generatedPackage := packages[packagePath]
		typesByFile := make(map[string][]*generatedTypeData)
		for _, generatedType := range generatedPackage.types {
			filePath := filepath.Join(codegen.Gendir, generatedType.location.FilePath)
			typesByFile[filePath] = append(typesByFile[filePath], generatedType)
		}
		filePaths := make([]string, 0, len(typesByFile))
		for filePath := range typesByFile {
			filePaths = append(filePaths, filePath)
		}
		sort.Strings(filePaths)
		for _, filePath := range filePaths {
			generatedTypes := typesByFile[filePath]
			sort.Slice(generatedTypes, func(i, j int) bool {
				return generatedTypes[i].declaration.Name() < generatedTypes[j].declaration.Name()
			})
			imports := []*codegen.ImportSpec{
				codegen.SimpleImport("fmt"),
				codegen.GoaImport(""),
			}
			collector := newImportCollector(aliases, genpkg, packagePath)
			for _, generatedType := range generatedTypes {
				collector.collect(generatedType.userType.Attribute())
			}
			imports = append(imports, collector.imports()...)
			sections := []*codegen.SectionTemplate{
				codegen.Header("User types", generatedPackage.packageName, imports),
			}
			for _, generatedType := range generatedTypes {
				sections = append(sections, generatedType.section)
				if generatedType.error != nil {
					sections = append(sections, generatedType.error)
				}
			}
			files = append(files, &codegen.File{Path: filePath, SectionTemplates: sections})
		}

		if len(generatedPackage.unions) > 0 {
			unions := make([]*UnionTypeData, 0, len(generatedPackage.unions))
			for _, union := range generatedPackage.unions {
				unions = append(unions, union)
			}
			sort.Slice(unions, func(i, j int) bool {
				return unions[i].Name < unions[j].Name
			})
			imports := []*codegen.ImportSpec{
				codegen.SimpleImport("bytes"),
				codegen.SimpleImport("encoding/json"),
				codegen.SimpleImport("fmt"),
				codegen.GoaImport(""),
			}
			collector := newImportCollector(aliases, genpkg, packagePath)
			for _, union := range unions {
				for _, field := range union.Fields {
					collector.collect(field.reference)
					collector.collect(field.definition)
				}
			}
			imports = append(imports, collector.imports()...)
			sections := []*codegen.SectionTemplate{
				codegen.Header("Union types", generatedPackage.packageName, imports),
			}
			for _, union := range unions {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "service-union-type",
					Source: serviceTemplates.Read(unionTypeT),
					Data:   union,
				})
			}
			files = append(files, &codegen.File{
				Path:             filepath.Join(generatedPackage.outputPath, "unions.go"),
				SectionTemplates: sections,
			})
		}
	}
	return files
}

// aggregateGeneratedPackages selects one render section per canonical package
// declaration across all analyzed roots without mutating generation state.
func aggregateGeneratedPackages(analyses []*ServicesData) map[string]*generatedPackageData {
	packages := make(map[string]*generatedPackageData)
	for _, services := range analyses {
		for packagePath, analyzedPackage := range services.packages {
			generatedPackage, ok := packages[packagePath]
			if !ok {
				generatedPackage = &generatedPackageData{
					outputPath:  analyzedPackage.outputPath,
					packageName: analyzedPackage.packageName,
					types:       make(map[*codegen.TypeDeclaration]*generatedTypeData),
					unions:      make(map[codegen.UnionTypeID]*UnionTypeData),
				}
				packages[packagePath] = generatedPackage
			}
			for declaration, generatedType := range analyzedPackage.types {
				if _, exists := generatedPackage.types[declaration]; !exists {
					generatedPackage.types[declaration] = generatedType
				}
			}
			for identity, union := range analyzedPackage.unions {
				if _, exists := generatedPackage.unions[identity]; !exists {
					generatedPackage.unions[identity] = union
				}
			}
		}
	}
	return packages
}

// dedupeByResult returns a slice of methods where only a single representative
// per unique ResultRef is kept (first occurrence wins). Methods without a
// ResultRef are ignored.
func dedupeByResult(ms []*MethodData) []*MethodData {
	seen := make(map[string]struct{})
	out := make([]*MethodData, 0, len(ms))
	for _, m := range ms {
		key := m.Result
		if key == "" {
			key = m.StreamingResult
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, m)
	}
	return out
}

func errorName(et *UserTypeData) string {
	obj := expr.AsObject(et.Type)
	if obj != nil {
		for _, att := range *obj {
			if _, ok := att.Attribute.Meta["struct:error:name"]; ok {
				return fmt.Sprintf("e.%s", codegen.GoifyAtt(att.Attribute, att.Name, true))
			}
		}
	}
	// if error type is a custom user type and used by at most one error, then
	// error Finalize should have added "struct:error:name" to the user type
	// attribute's meta.
	if v, ok := et.Type.Attribute().Meta["struct:error:name"]; ok {
		return fmt.Sprintf("%q", v[0])
	}
	return fmt.Sprintf("%q", et.Name)
}

// hasJSONRPCStreaming returns true if the service has a JSON-RPC streaming
// endpoint (WebSocket or SSE).
func hasJSONRPCStreaming(sd *Data) bool {
	for _, m := range sd.Methods {
		if m.IsJSONRPC && m.ServerStream != nil {
			return true
		}
	}
	return false
}

// hasJSONRPCWebSocket returns true if the service has a JSON-RPC streaming
// endpoint that uses the WebSocket transport.
func hasJSONRPCWebSocket(sd *Data) bool {
	for _, m := range sd.Methods {
		if m.IsJSONRPCWebSocket {
			return true
		}
	}
	return false
}

// streamInterfaceFor builds the data to generate the client and server stream
// interfaces for the given endpoint.
func streamInterfaceFor(typ string, m *MethodData, stream *StreamData) map[string]any {
	return map[string]any{
		"Type":               typ,
		"Endpoint":           m.Name,
		"Stream":             stream,
		"MethodVarName":      m.VarName,
		"IsJSONRPC":          m.IsJSONRPC,
		"IsJSONRPCSSE":       m.IsJSONRPCSSE && typ == "server",
		"IsJSONRPCWebSocket": m.IsJSONRPCWebSocket,
		// If a view is explicitly set (ViewName is not empty) in the Result
		// expression, we can use that view to render the result type instead
		// of iterating through the list of views defined in the result type.
		"IsViewedResult": m.ViewedResult != nil && m.ViewedResult.ViewName == "",
	}
}
