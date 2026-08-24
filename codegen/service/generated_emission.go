// This file attaches service template data to types and Goa OneOf unions that
// are written outside their service package. The generation command selects
// each output package and declaration before source is written.
package service

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
)

// generatedPackage returns the template data collected for one generated Go
// package. It creates an empty container when this is the package's first type
// or union.
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

// registerPackageData associates each selected type and union declaration with
// the template section and imports written to its generated package.
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

// generatedTypeSections returns the template sections for one selected payload,
// result, error, or user type.
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

// generatedUserTypeData returns the template data for one authored type
// declaration selected for a generated package.
func generatedUserTypeData(emission *generatedTypeEmissionFacts) *UserTypeData {
	candidates := emission.service.data.userTypes
	if emission.kind == generatedErrorTypeEmission {
		candidates = emission.service.data.errorTypes
	}
	for _, candidate := range candidates {
		if candidate.Declaration == emission.declaration {
			data := *candidate
			if data.Description == "" {
				data.Description = generatedUserTypeDescription(emission)
			}
			return &data
		}
	}
	panic(fmt.Sprintf("generated type %q has no linked render data", emission.declaration.Name()))
}

// generatedUserTypeDescription explains where an authored type is used. A
// nested type has no method role of its own.
func generatedUserTypeDescription(emission *generatedTypeEmissionFacts) string {
	name := emission.declaration.Name()
	if len(emission.uses) == 0 {
		return fmt.Sprintf("%s is a named type defined in the service design.", name)
	}
	if len(emission.uses) == 1 {
		use := emission.uses[0]
		return fmt.Sprintf(
			"%s is the %s type of the %s service %s method.",
			name,
			generatedTypeRoleNames(use.roles),
			use.service,
			use.method,
		)
	}
	var description strings.Builder
	fmt.Fprintf(&description, "%s is used by these service methods:", name)
	for _, use := range emission.uses {
		fmt.Fprintf(
			&description,
			"\n- %s %s: %s",
			use.service,
			use.method,
			generatedTypeRoleNames(use.roles),
		)
	}
	return description.String()
}

// generatedTypeRoleNames joins the method fields that use one authored type.
func generatedTypeRoleNames(roles generatedTypeMethodRoles) string {
	names := make([]string, 0, 4)
	for _, role := range []struct {
		value generatedTypeMethodRoles
		name  string
	}{
		{generatedPayloadRole, "payload"},
		{generatedStreamingPayloadRole, "streaming payload"},
		{generatedResultRole, "result"},
		{generatedStreamingResultRole, "streaming result"},
	} {
		if roles&role.value != 0 {
			names = append(names, role.name)
		}
	}
	if len(names) == 1 {
		return names[0]
	}
	if len(names) == 2 {
		return names[0] + " and " + names[1]
	}
	return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
}

// copyGeneratedLocation copies the requested package and file so later changes
// to the design expression cannot change where the type is written.
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
