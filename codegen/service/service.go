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
					Source: payloadT,
					Data:   m,
				})
			}
		}
		if m.StreamingPayloadDef != "" {
			if _, ok := seen[m.StreamingPayload]; !ok {
				addTypeDefSection(payloadPath, m.StreamingPayload, &codegen.SectionTemplate{
					Name:   "service-streamig-payload",
					Source: streamingPayloadT,
					Data:   m,
				})
			}
		}
		if m.ResultDef != "" {
			if _, ok := seen[m.Result]; !ok {
				addTypeDefSection(resultPath, m.Result, &codegen.SectionTemplate{
					Name:   "service-result",
					Source: resultT,
					Data:   m,
				})
			}
		}
	}
	for _, ut := range svc.userTypes {
		if _, ok := seen[ut.VarName]; !ok {
			addTypeDefSection(pathWithDefault(ut.Loc, svcPath), ut.VarName, &codegen.SectionTemplate{
				Name:   "service-user-type",
				Source: userTypeT,
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
					Source: userTypeT,
					Data:   et,
				})
			}
			errorTypes = append(errorTypes, et)
		}
	}

	for _, m := range svc.unionValueMethods {
		addTypeDefSection(pathWithDefault(m.Loc, svcPath), "~"+m.TypeRef+"."+m.Name, &codegen.SectionTemplate{
			Name:   "service-union-value-method",
			Source: unionValueMethodT,
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
			Source:  errorT,
			FuncMap: map[string]interface{}{"errorName": errorName},
			Data:    et,
		})
	}
	for _, er := range svc.errorInits {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "error-init-func",
			Source: errorInitT,
			Data:   er,
		})
	}

	// transform result type functions
	for _, t := range svc.viewedResultTypes {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "viewed-result-type-to-service-result-type",
			Source: typeInitT,
			Data:   t.ResultInit,
		})
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "service-result-type-to-viewed-result-type",
			Source: typeInitT,
			Data:   t.Init,
		})
	}
	var projh []*codegen.TransformFunctionData
	for _, t := range svc.projectedTypes {
		for _, i := range t.TypeInits {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "projected-type-to-service-type",
				Source: typeInitT,
				Data:   i,
			})
		}
		for _, i := range t.Projections {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			svcSections = append(svcSections, &codegen.SectionTemplate{
				Name:   "service-type-to-projected-type",
				Source: typeInitT,
				Data:   i,
			})
		}
	}

	for _, h := range projh {
		svcSections = append(svcSections, &codegen.SectionTemplate{
			Name:   "transform-helpers",
			Source: transformHelperT,
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
		Source:  serviceT,
		Data:    svc,
		FuncMap: map[string]interface{}{"streamInterfaceFor": streamInterfaceFor},
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
func streamInterfaceFor(typ string, m *MethodData, stream *StreamData) map[string]interface{} {
	return map[string]interface{}{
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

// serviceT is the template used to write an service definition.
const serviceT = `
{{ comment .Description }}
type Service interface {
{{- range .Methods }}
	{{ comment .Description }}
	{{- if .ViewedResult }}
		{{- if not .ViewedResult.ViewName }}
			{{ comment "The \"view\" return value must have one of the following views" }}
			{{- range .ViewedResult.Views }}
				{{- if .Description }}
					{{ printf "//	- %q: %s" .Name .Description }}
				{{- else }}
					{{ printf "//	- %q" .Name }}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if .ServerStream }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
	{{- else }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, io.ReadCloser{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}body io.ReadCloser, {{ end }}{{ if .Result }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}{{ end }}err error)
	{{- end }}
{{- end }}
}

{{- if .Schemes }}
// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	{{- range .Schemes }}
	{{ printf "%sAuth implements the authorization logic for the %s security scheme." .Type .Type | comment }}
	{{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, schema *security.{{ .Type }}Scheme) (context.Context, error)
	{{- end }}
}
{{- end }}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = {{ printf "%q" .Name }}

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [{{ len .Methods }}]string{ {{ range .Methods }}{{ printf "%q" .Name }}, {{ end }} }
{{- range .Methods }}
	{{- if .ServerStream }}
		{{ template "stream_interface" (streamInterfaceFor "server" . .ServerStream) }}
		{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
	{{- end }}
{{- end }}

{{- define "stream_interface" }}
{{ printf "%s is the interface a %q endpoint %s stream must satisfy." .Stream.Interface .Endpoint .Type | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendTypeRef }}
		{{ comment .Stream.SendDesc }}
		{{ .Stream.SendName }}({{ .Stream.SendTypeRef }}) error
	{{- end }}
	{{- if .Stream.RecvTypeRef }}
		{{ comment .Stream.RecvDesc }}
		{{ .Stream.RecvName }}() ({{ .Stream.RecvTypeRef }}, error)
	{{- end }}
	{{- if .Stream.MustClose }}
		{{ comment "Close closes the stream." }}
		Close() error
	{{- end }}
	{{- if and .IsViewedResult (eq .Type "server") }}
		{{ comment "SetView sets the view used to render the result before streaming." }}
		SetView(view string)
	{{- end }}
}
{{- end }}
`

const payloadT = `{{ comment .PayloadDesc }}
type {{ .Payload }} {{ .PayloadDef }}
`

const streamingPayloadT = `{{ comment .StreamingPayloadDesc }}
type {{ .StreamingPayload }} {{ .StreamingPayloadDef }}
`

const resultT = `{{ comment .ResultDesc }}
type {{ .Result }} {{ .ResultDef }}
`

const userTypeT = `{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
`

const errorT = `// Error returns an error description.
func (e {{ .Ref }}) Error() string {
	return {{ printf "%q" .Description }}
}

// ErrorName returns {{ printf "%q" .Name }}.
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e {{ .Ref }}) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns {{ printf "%q" .Name }}.
func (e {{ .Ref }}) GoaErrorName() string {
	return {{ errorName . }}
}
`

const unionValueMethodT = `func ({{ .TypeRef }}) {{ .Name }}() {}
`

// input: map[string]{"Type": TypeData, "Error": ErrorData}
const errorInitT = `{{ printf "%s builds a %s from an error." .Name .TypeName |  comment }}
func {{ .Name }}(err error) {{ .TypeRef }} {
	return goa.NewServiceError(err, {{ printf "%q" .ErrName }}, {{ printf "%v" .Timeout }}, {{ printf "%v" .Temporary}}, {{ printf "%v" .Fault}})
}
`

// input: InitData
const typeInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .Ref }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
}
`
