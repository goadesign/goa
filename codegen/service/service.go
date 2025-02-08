package service

import (
	"fmt"
	"path/filepath"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Files returns the generated files for the given service as well as a map
// indexing user type names by custom path as defined by the "struct:pkg:path"
// metadata. The map is built over each invocation of Files to avoid duplicate
// type definitions.
func Files(genpkg string, service *expr.ServiceExpr, userTypePkgs map[string][]string) []*codegen.File {
	svc := Services.Get(service.Name)
	svc.initUserTypeImports(genpkg)
	svcName := svc.PathName
	svcPath := filepath.Join(codegen.Gendir, svcName, "service.go")
	seen := make(map[string]struct{})
	typeDefSections := make(map[string]map[string]*codegen.SectionTemplate)
	typesByPath := make(map[string][]string)
	var svcSections []*codegen.SectionTemplate

	addTypeDefSection := func(path, name string, section *codegen.SectionTemplate) {
		if typeDefSections[path] == nil {
			typeDefSections[path] = make(map[string]*codegen.SectionTemplate)
		}
		typeDefSections[path][name] = section
		typesByPath[path] = append(typesByPath[path], name)
		seen[name] = struct{}{}
	}

	for _, m := range svc.Methods {
		payloadPath := pathWithDefault(m.PayloadLoc, svcPath)
		resultPath := pathWithDefault(m.ResultLoc, svcPath)
		if m.PayloadDef != "" {
			if _, ok := seen[m.Payload]; !ok {
				addTypeDefSection(payloadPath, m.Payload, &codegen.SectionTemplate{
					Name:   "service-payload",
					Source: readTemplate("payload"),
					Data:   m,
				})
			}
		}
		if m.StreamingPayloadDef != "" {
			if _, ok := seen[m.StreamingPayload]; !ok {
				addTypeDefSection(payloadPath, m.StreamingPayload, &codegen.SectionTemplate{
					Name:   "service-streaming-payload",
					Source: readTemplate("streaming_payload"),
					Data:   m,
				})
			}
		}
		if m.ResultDef != "" {
			if _, ok := seen[m.Result]; !ok {
				addTypeDefSection(resultPath, m.Result, &codegen.SectionTemplate{
					Name:   "service-result",
					Source: readTemplate("result"),
					Data:   m,
				})
			}
		}
	}
	for _, ut := range svc.userTypes {
		if _, ok := seen[ut.VarName]; !ok {
			addTypeDefSection(pathWithDefault(ut.Loc, svcPath), ut.VarName, &codegen.SectionTemplate{
				Name:   "service-user-type",
				Source: readTemplate("user_type"),
				Data:   ut,
			})
		}
	}

	var errorTypes []*UserTypeData
	seenErrs := make(map[string]struct{})
	for _, et := range svc.errorTypes {
		if et.Type == expr.ErrorResult {
			continue
		}
		if _, ok := seenErrs[et.Name]; !ok {
			seenErrs[et.Name] = struct{}{}
			if _, ok := seen[et.Name]; !ok {
				addTypeDefSection(pathWithDefault(et.Loc, svcPath), et.Name, &codegen.SectionTemplate{
					Name:   "error-user-type",
					Source: readTemplate("user_type"),
					Data:   et,
				})
			}
			errorTypes = append(errorTypes, et)
		}
	}

	for _, m := range svc.unionValueMethods {
		addTypeDefSection(pathWithDefault(m.Loc, svcPath), "~"+m.TypeRef+"."+m.Name, &codegen.SectionTemplate{
			Name:   "service-union-value-method",
			Source: readTemplate("union_value_method"),
			Data:   m,
		})
	}

	for _, et := range errorTypes {
		// Don't override the section created for the error type
		// declaration, make sure the key does not clash with existing
		// type names, make it generated last.
		key := "|" + et.Name
		addTypeDefSection(pathWithDefault(et.Loc, svcPath), key, &codegen.SectionTemplate{
			Name:    "service-error",
			Source:  readTemplate("error"),
			FuncMap: map[string]any{"errorName": errorName},
			Data:    et,
		})
	}
	for _, er := range svc.errorInits {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "error-init-func",
			Source: readTemplate("error_init"),
			Data:   er,
		})
	}

	// transform result type functions
	for _, t := range svc.viewedResultTypes {
		svcSections = append(svcSections,
			&codegen.SectionTemplate{Name: "viewed-result-type-to-service-result-type", Source: readTemplate("type_init"), Data: t.ResultInit},
			&codegen.SectionTemplate{Name: "service-result-type-to-viewed-result-type", Source: readTemplate("type_init"), Data: t.Init})
	}
	var projh []*codegen.TransformFunctionData
	for _, t := range svc.projectedTypes {
		for _, i := range t.TypeInits {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "projected-type-to-service-type",
				Source: readTemplate("type_init"),
				Data:   i,
			})
		}
		for _, i := range t.Projections {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "service-type-to-projected-type",
				Source: readTemplate("type_init"),
				Data:   i,
			})
		}
	}

	for _, h := range projh {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "transform-helpers",
			Source: readTemplate("transform_helper"),
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
	imports = append(imports, svc.UserTypeImports...)
	header := codegen.Header(service.Name+" service", svc.PkgName, imports)
	def := &codegen.SectionTemplate{
		Name:    "service",
		Source:  readTemplate("service"),
		Data:    svc,
		FuncMap: map[string]any{"streamInterfaceFor": streamInterfaceFor},
	}

	// service.go
	var sections []*codegen.SectionTemplate
	{
		sections = []*codegen.SectionTemplate{header, def}
		names := make([]string, len(typeDefSections[svcPath]))
		i := 0
		for n := range typeDefSections[svcPath] {
			names[i] = n
			i++
		}
		sort.Strings(names)
		for _, n := range names {
			sections = append(sections, typeDefSections[svcPath][n])
		}
		sections = append(sections, svcSections...)
	}
	files := []*codegen.File{{Path: svcPath, SectionTemplates: sections}}

	// service and client interceptors
	files = append(files, InterceptorsFiles(genpkg, service)...)

	// user types
	paths := make([]string, len(typeDefSections))
	i := 0
	for p := range typesByPath {
		paths[i] = p
		i++
	}
	sort.Strings(paths)
	for _, p := range paths {
		if p == svcPath {
			continue
		}
		var secs []*codegen.SectionTemplate
		ts := typesByPath[p]
		sort.Strings(ts)
		for _, name := range ts {
			hasName := false
			for _, n := range userTypePkgs[p] {
				if hasName = n == name; hasName {
					break
				}
			}
			if hasName {
				continue
			}
			userTypePkgs[p] = append(userTypePkgs[p], name)
			secs = append(secs, typeDefSections[p][name])
		}
		if len(secs) == 0 {
			continue
		}
		fullRelPath := filepath.Join(codegen.Gendir, p)
		dir, _ := filepath.Split(fullRelPath)
		h := codegen.Header("User types", codegen.Goify(filepath.Base(dir), false), nil)
		sections := append([]*codegen.SectionTemplate{h}, secs...)
		files = append(files, &codegen.File{Path: fullRelPath, SectionTemplates: sections})
	}

	return files
}

// AddServiceDataMetaTypeImports Adds all imports defined by struct:field:type from the service expr and the service data
func AddServiceDataMetaTypeImports(header *codegen.SectionTemplate, serviceE *expr.ServiceExpr) {
	codegen.AddServiceMetaTypeImports(header, serviceE)
	svc := Services.Get(serviceE.Name)
	for _, ut := range svc.userTypes {
		codegen.AddImport(header, codegen.GetMetaTypeImports(ut.Type.Attribute())...)
	}
	for _, et := range svc.errorTypes {
		codegen.AddImport(header, codegen.GetMetaTypeImports(et.Type.Attribute())...)
	}
	for _, t := range svc.viewedResultTypes {
		codegen.AddImport(header, codegen.GetMetaTypeImports(t.Type.Attribute())...)
	}
	for _, t := range svc.projectedTypes {
		codegen.AddImport(header, codegen.GetMetaTypeImports(t.Type.Attribute())...)
	}
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

func pathWithDefault(loc *codegen.Location, def string) string {
	if loc == nil {
		return def
	}
	return loc.FilePath
}
