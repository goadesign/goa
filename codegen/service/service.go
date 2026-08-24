// This file renders service declarations and groups declarations with explicit
// package locations into the generated Go package and file where each one is
// written.
package service

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// serviceTypeSectionPhase identifies the stable group containing a service
	// type-file section. Type declarations must precede methods defined on them.
	serviceTypeSectionPhase uint8

	// serviceTypeSection retains the explicit ordering facts for one section in
	// service.go instead of encoding its group in a decorated string key.
	serviceTypeSection struct {
		phase   serviceTypeSectionPhase
		name    string
		section *codegen.SectionTemplate
	}
)

const (
	serviceTypeDefinitionPhase serviceTypeSectionPhase = iota
	serviceErrorImplementationPhase
)

// Files renders every service file described by plans. Each plan must be
// linked so every renderer reads the declarations copied before their names
// were chosen instead of rebuilding service analysis from the expression root.
func Files(plans ...*Plan) ([]*codegen.File, error) {
	var files []*codegen.File
	if len(plans) == 0 {
		return files, nil
	}
	generation := plans[0].generation
	ownedRoots := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		if root, ok := candidate.(*expr.RootExpr); ok {
			ownedRoots[root] = struct{}{}
		}
	}
	for _, plan := range plans[1:] {
		if plan.generation != generation {
			return nil, fmt.Errorf("service plans belong to different generations")
		}
	}
	if len(plans) != len(ownedRoots) {
		return nil, fmt.Errorf("service rendering requires all %d planned roots, got %d", len(ownedRoots), len(plans))
	}
	seenRoots := make(map[*expr.RootExpr]struct{}, len(plans))
	for _, plan := range plans {
		if _, owned := ownedRoots[plan.facts.root]; !owned {
			return nil, rootMembershipError(plan.facts.root)
		}
		if _, exists := seenRoots[plan.facts.root]; exists {
			return nil, fmt.Errorf("service root %p is rendered more than once", plan.facts.root)
		}
		seenRoots[plan.facts.root] = struct{}{}
	}
	analyses := make([]*ServicesData, len(plans))
	for index, plan := range plans {
		analyses[index] = plan.Services()
		for _, facts := range plan.facts.services {
			files = append(files, serviceFiles(plan, facts)...)
		}
	}
	generatedFiles, err := generatedPackageFiles(analyses)
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFiles...)
	for _, plan := range plans {
		for _, facts := range plan.facts.services {
			files = append(files,
				endpointFile(plan, facts),
				clientFile(plan, facts),
			)
			if file := viewsFile(plan, facts); file != nil {
				files = append(files, file)
			}
		}
	}
	conversionFiles, err := externalConversionFiles(plans)
	if err != nil {
		return nil, err
	}
	files = append(files, convertFiles(conversionFiles)...)
	return files, nil
}

// serviceFiles renders the declarations and helpers owned exclusively by one
// service package. Relocated declarations and all union definitions are
// emitted later by generatedPackageFiles.
func serviceFiles(plan *Plan, facts *serviceFacts) []*codegen.File {
	services := plan.Services()
	svc := services.Get(facts.name)
	svcName := svc.PathName
	svcPath := filepath.Join(codegen.Gendir, svcName, "service.go")
	seen := make(map[string]struct{})
	typeSections := make([]serviceTypeSection, 0)
	svcSections := make([]*codegen.SectionTemplate, 0, 10)

	addTypeDefSection := func(name string, section *codegen.SectionTemplate) {
		typeSections = append(typeSections, serviceTypeSection{
			phase:   serviceTypeDefinitionPhase,
			name:    name,
			section: section,
		})
		seen[name] = struct{}{}
	}

	for i, m := range svc.Methods {
		method := facts.orderedMethods[i]
		if m.PayloadLoc == nil && m.PayloadDef != "" {
			if _, ok := seen[m.Payload]; !ok {
				addTypeDefSection(m.Payload, &codegen.SectionTemplate{
					Name:   "service-payload",
					Source: serviceTemplates.Read(payloadT),
					Data:   m,
				})
			}
		}
		if method.streamingPayload != nil && method.streamingPayload.location == nil && m.StreamingPayloadDef != "" {
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
		if method.streamingResult != nil && method.streamingResult.location == nil && m.StreamingResultDef != "" && m.StreamingResult != m.Result {
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
		if et.IsServiceError || et.Loc != nil {
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
			typeSections = append(typeSections, serviceTypeSection{
				phase: serviceErrorImplementationPhase,
				name:  et.Name,
				section: &codegen.SectionTemplate{
					Name:   "service-error",
					Source: serviceTemplates.Read(errorT),
					Data:   et,
				},
			})
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

	header := codegen.Header(facts.name+" service", svc.PkgName, facts.imports.service.specs)
	def := &codegen.SectionTemplate{
		Name:   "service",
		Source: serviceTemplates.Read(serviceT),
		Data:   svc,
		FuncMap: map[string]any{
			"streamInterfaceFor": streamInterfaceFor,
		},
	}

	sort.Slice(typeSections, func(i, j int) bool {
		if typeSections[i].phase != typeSections[j].phase {
			return typeSections[i].phase < typeSections[j].phase
		}
		return typeSections[i].name < typeSections[j].name
	})
	sections := make([]*codegen.SectionTemplate, 0, 2+len(typeSections)+len(svcSections))
	sections = append(sections, header, def)
	for _, record := range typeSections {
		sections = append(sections, record.section)
	}
	sections = append(sections, svcSections...)
	interceptors := interceptorsFiles(plan, facts)
	files := make([]*codegen.File, 1, 1+len(interceptors))
	files[0] = &codegen.File{Path: svcPath, SectionTemplates: sections}
	return append(files, interceptors...)
}

// generatedPackageFiles renders each relocated user type in its configured
// file and one sorted unions.go for every package that owns unions.
func generatedPackageFiles(analyses []*ServicesData) ([]*codegen.File, error) {
	packages, err := aggregateGeneratedPackages(analyses)
	if err != nil {
		return nil, err
	}
	if len(packages) == 0 {
		return nil, nil
	}
	packageOwners := make([]*codegen.GeneratedPackage, 0, len(packages))
	for owner := range packages {
		packageOwners = append(packageOwners, owner)
	}
	slices.SortFunc(packageOwners, func(left, right *codegen.GeneratedPackage) int {
		return strings.Compare(left.ImportPath(), right.ImportPath())
	})

	var files []*codegen.File
	for _, owner := range packageOwners {
		packagePath := owner.ImportPath()
		packageName := strings.ToLower(codegen.Goify(path.Base(packagePath), false))
		generatedPackage := packages[owner]
		typesByFile := make(map[string][]*generatedTypeData)
		for _, generatedType := range generatedPackage.types {
			filePath := filepath.Join(owner.OutputDirectory(), filepath.Base(generatedType.location.FilePath))
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
			var imports []*codegen.ImportSpec
			for _, generatedType := range generatedTypes {
				imports = appendImportSpecs(imports, generatedType.imports)
			}
			sections := []*codegen.SectionTemplate{
				codegen.Header("User types", packageName, imports),
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
			sections := []*codegen.SectionTemplate{
				codegen.Header("Union types", packageName, generatedPackage.unionImports),
			}
			for _, union := range unions {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "service-union-type",
					Source: serviceTemplates.Read(unionTypeT),
					Data:   union,
				})
			}
			files = append(files, &codegen.File{
				Path:             filepath.Join(owner.OutputDirectory(), "unions.go"),
				SectionTemplates: sections,
			})
		}
	}
	return files, nil
}

// aggregateGeneratedPackages selects one render section for each generated
// package declaration across all analyzed roots without changing generation
// state.
func aggregateGeneratedPackages(analyses []*ServicesData) (map[*codegen.GeneratedPackage]*generatedPackageData, error) {
	packages := make(map[*codegen.GeneratedPackage]*generatedPackageData)
	for _, services := range analyses {
		for owner, analyzedPackage := range services.packages {
			generatedPackage, ok := packages[owner]
			if !ok {
				generatedPackage = &generatedPackageData{
					types:  make(map[*codegen.TypeDeclaration]*generatedTypeData),
					unions: make(map[*codegen.UnionDeclaration]*UnionTypeData),
				}
				packages[owner] = generatedPackage
			}
			for declaration, generatedType := range analyzedPackage.types {
				if _, exists := generatedPackage.types[declaration]; exists {
					return nil, fmt.Errorf(
						"generated type declaration %q was assigned to more than one service plan",
						declaration.Name(),
					)
				}
				generatedPackage.types[declaration] = generatedType
			}
			for declaration, union := range analyzedPackage.unions {
				if _, exists := generatedPackage.unions[declaration]; exists {
					return nil, fmt.Errorf(
						"generated union declaration %q was assigned to more than one service plan",
						union.Name,
					)
				}
				generatedPackage.unions[declaration] = union
			}
			generatedPackage.unionImports = appendImportSpecs(generatedPackage.unionImports, analyzedPackage.unionImports)
		}
	}
	return packages, nil
}

// appendImportSpecs merges exact file contributions by complete package path
// and returns them in deterministic path order.
func appendImportSpecs(existing, added []*codegen.ImportSpec) []*codegen.ImportSpec {
	byPath := make(map[string]*codegen.ImportSpec, len(existing)+len(added))
	for _, spec := range existing {
		byPath[spec.Path] = spec
	}
	for _, spec := range added {
		byPath[spec.Path] = spec
	}
	paths := make([]string, 0, len(byPath))
	for importPath := range byPath {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	result := make([]*codegen.ImportSpec, len(paths))
	for index, importPath := range paths {
		result[index] = byPath[importPath]
	}
	return result
}

// streamInterfaceFor builds the data to generate the client and server stream
// interfaces for the given endpoint.
func streamInterfaceFor(typ string, m *MethodData, stream *StreamData) map[string]any {
	return map[string]any{
		"Type":     typ,
		"Endpoint": m.Name,
		"Stream":   stream,
		// If a view is explicitly set (ViewName is not empty) in the Result
		// expression, we can use that view to render the result type instead
		// of iterating through the list of views defined in the result type.
		"IsViewedResult": m.ViewedResult != nil && m.ViewedResult.ViewName == "",
	}
}
