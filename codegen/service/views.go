package service

import (
	"path/filepath"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type viewedType struct {
	// Name is the type name.
	Name string
	// Views is the view data for all views defined in the type.
	Views []*ViewData
}

// ViewsFile returns the views file for the given service which contains
// logic to render result types using the defined views.
func ViewsFile(_ string, service *expr.ServiceExpr, services *ServicesData) *codegen.File {
	svc := services.Get(service.Name)
	if len(svc.projectedTypes) == 0 {
		return nil
	}

	// Collect union sum-type definitions for the views package.
	//
	// View-projected types cannot import the service package (which already
	// depends on views), therefore unions must be generated in the views package
	// when referenced by projected types.
	unionByHash := make(map[string]*UnionTypeData)
	seenUnions := make(map[string]struct{})
	viewLoc := &codegen.Location{RelImportPath: "views"}
	for _, t := range svc.projectedTypes {
		collectViewUnionTypes(&expr.AttributeExpr{Type: t.Type}, svc.ViewScope, viewLoc, unionByHash, seenUnions)
	}
	unions := make([]*UnionTypeData, 0, len(unionByHash))
	for _, u := range unionByHash {
		unions = append(unions, u)
	}
	sort.Slice(unions, func(i, j int) bool {
		return unions[i].Name < unions[j].Name
	})

	path := filepath.Join(codegen.Gendir, svc.PathName, "views", "view.go")
	imports := []*codegen.ImportSpec{
		codegen.GoaImport(""),
		{Path: "unicode/utf8"},
	}
	if len(unions) > 0 {
		imports = append(imports,
			codegen.SimpleImport("encoding/json"),
			codegen.SimpleImport("fmt"),
		)
	}
	header := codegen.Header(service.Name+" views", "views",
		imports)
	sections := []*codegen.SectionTemplate{header}

	// type definitions
	for _, t := range svc.viewedResultTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "viewed-result-type",
			Source: serviceTemplates.Read(userTypeT),
			Data:   t.UserTypeData,
		})
	}
	for _, t := range svc.projectedTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "projected-type",
			Source: serviceTemplates.Read(userTypeT),
			Data:   t.UserTypeData,
		})
	}
	for _, u := range unions {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "projected-union-type",
			Source: serviceTemplates.Read(unionTypeT),
			Data:   u,
		})
	}

	// generate a map for result types with view name as key and the fields
	// rendered in the view as value.
	var (
		rtdata []*viewedType
		seen   = make(map[string]struct{})
	)
	for _, t := range svc.viewedResultTypes {
		name := t.Views[0].TypeVarName
		if _, ok := seen[name]; !ok {
			rtdata = append(rtdata, &viewedType{Name: name, Views: t.Views})
			seen[name] = struct{}{}
		}
	}
	for _, t := range svc.projectedTypes {
		if len(t.Views) == 0 {
			continue
		}
		name := t.Views[0].TypeVarName
		if _, ok := seen[name]; !ok {
			rtdata = append(rtdata, &viewedType{Name: name, Views: t.Views})
			seen[name] = struct{}{}
		}
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "viewed-type-map",
		Source: serviceTemplates.Read(viewedTypeMapT),
		Data: map[string]any{
			"ViewedTypes": rtdata,
		},
	})

	// validations
	for _, t := range svc.viewedResultTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "validate-viewed-result-type",
			Source: serviceTemplates.Read(validateT),
			Data:   t.Validate,
		})
	}
	for _, t := range svc.projectedTypes {
		for _, v := range t.Validations {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "validate-projected-type",
				Source: serviceTemplates.Read(validateT),
				Data:   v,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}
