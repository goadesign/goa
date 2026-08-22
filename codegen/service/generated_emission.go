// This file attaches linked service render data to the exact relocated type
// and union emission records selected by batch planning before names freeze.
package service

import (
	"fmt"

	"goa.design/goa/v3/codegen"
)

// generatedPackage returns the retained render data for one exact generated
// package, creating the render container without changing package ownership.
func (d *ServicesData) generatedPackage(importPath string) *generatedPackageData {
	owner := d.generation.Package(importPath)
	if generatedPackage, ok := d.packages[owner]; ok {
		return generatedPackage
	}
	generatedPackage := &generatedPackageData{
		types:  make(map[*codegen.TypeDeclaration]*generatedTypeData),
		unions: make(map[*codegen.UnionDeclaration]*UnionTypeData),
	}
	d.packages[owner] = generatedPackage
	return generatedPackage
}

// registerPackageData attaches linked template data to the exact type and
// union emission records selected by the batch planner before freeze.
func (d *ServicesData) registerPackageData() {
	for _, emission := range d.facts.generatedTypes {
		section, errorSection := generatedTypeSections(emission)
		generatedPackage := d.generatedPackage(emission.declaration.PackagePath())
		generatedPackage.types[emission.declaration] = &generatedTypeData{
			declaration: emission.declaration,
			location:    emission.location,
			imports:     emission.service.generatedTypeImports[emission.declaration].specs,
			section:     section,
			error:       errorSection,
		}
	}
	for _, emission := range d.facts.generatedUnions {
		generatedPackage := d.generatedPackage(emission.union.declaration.PackagePath())
		generatedPackage.unions[emission.union.declaration] = emission.union.data
		generatedPackage.unionImports = appendImportSpecs(
			generatedPackage.unionImports,
			emission.union.imports.specs,
		)
	}
}

// generatedTypeSections formats the selected template family from linked data
// retained on the owning service and method facts.
func generatedTypeSections(emission *generatedTypeEmissionFacts) (*codegen.SectionTemplate, *codegen.SectionTemplate) {
	if emission.method != nil {
		var methodData *MethodData
		for index, method := range emission.service.orderedMethods {
			if method == emission.method {
				methodData = emission.service.data.Methods[index]
				break
			}
		}
		switch emission.kind {
		case generatedPayloadEmission:
			return &codegen.SectionTemplate{Name: "service-payload", Source: serviceTemplates.Read(payloadT), Data: methodData}, nil
		case generatedStreamingPayloadEmission:
			return &codegen.SectionTemplate{Name: "service-streaming-payload", Source: serviceTemplates.Read(streamingPayloadT), Data: methodData}, nil
		case generatedResultEmission:
			return &codegen.SectionTemplate{Name: "service-result", Source: serviceTemplates.Read(resultT), Data: methodData}, nil
		case generatedStreamingResultEmission:
			return &codegen.SectionTemplate{
				Name:   "service-streaming-result",
				Source: serviceTemplates.Read(resultT),
				Data: map[string]any{
					"Result":     methodData.StreamingResult,
					"ResultDef":  methodData.StreamingResultDef,
					"ResultDesc": methodData.StreamingResultDesc,
				},
			}, nil
		}
	}
	data := generatedUserTypeData(emission)
	name := "service-user-type"
	if emission.kind == generatedErrorTypeEmission {
		name = "error-user-type"
	}
	section := &codegen.SectionTemplate{Name: name, Source: serviceTemplates.Read(userTypeT), Data: data}
	if !emission.error {
		return section, nil
	}
	return section, &codegen.SectionTemplate{Name: "service-error", Source: serviceTemplates.Read(errorT), Data: data}
}

// generatedUserTypeData returns the linked record for one retained authored
// type declaration.
func generatedUserTypeData(emission *generatedTypeEmissionFacts) *UserTypeData {
	candidates := emission.service.data.userTypes
	if emission.kind == generatedErrorTypeEmission {
		candidates = emission.service.data.errorTypes
	}
	for _, candidate := range candidates {
		if candidate.Declaration == emission.declaration {
			return candidate
		}
	}
	panic(fmt.Sprintf("generated type %q has no linked render data", emission.declaration.Name()))
}

// copyGeneratedLocation retains location metadata before callers can mutate
// the expression graph after planning.
func copyGeneratedLocation(location *codegen.Location) *codegen.Location {
	if location == nil {
		return nil
	}
	copy := *location
	return &copy
}

// sameGeneratedLocation compares the explicit package and file selected for a
// generated declaration.
func sameGeneratedLocation(left, right *codegen.Location) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.RelImportPath == right.RelImportPath && left.FilePath == right.FilePath
}
