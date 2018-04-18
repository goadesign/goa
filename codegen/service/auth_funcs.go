package service

import (
	"os"
	"path"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// AuthFuncsFile returns a file that contains a dummy implementation of the
// authorization functions needed to instantiate the service endpoints.
func AuthFuncsFile(genpkg string, root *design.RootExpr) *codegen.File {
	var (
		apiPkg   = strings.ToLower(codegen.Goify(root.API.Name, false))
		rootPath = "."
	)
	{
		idx := strings.LastIndex(genpkg, string(os.PathSeparator))
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
	}

	var (
		sections []*codegen.SectionTemplate
	)
	{
		specs := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/security"},
			{Path: rootPath, Name: apiPkg},
		}
		for _, svc := range root.Services {
			pkgName := Services.Get(svc.Name).PkgName
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, codegen.SnakeCase(svc.Name)),
				Name: pkgName,
			})
		}
		header := codegen.Header(root.API.Name+" authentication logic.", apiPkg, specs)
		sections = []*codegen.SectionTemplate{header}
		for _, s := range root.Services {
			svc := Services.Get(s.Name)
			if len(svc.Schemes) > 0 {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "security-authfuncs",
					Source: dummyAuthFuncsT,
					Data:   svc,
				})
			}
		}
	}
	if len(sections) == 0 {
		return nil
	}

	return &codegen.File{
		Path:             "auth.go",
		SectionTemplates: sections,
	}
}

// data: AuthFuncsData
const dummyAuthFuncsT = `{{ range .Schemes }}
{{ printf "%s%sAuth implements the authorization logic for service %q for the %q security scheme." $.StructName .Type $.Name .SchemeName | comment }}
func {{ $.StructName }}{{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, s *security.{{ .Type }}Scheme) (context.Context, error) {
	//
	// TBD: add authorization logic.
	//
	// In case of authorization failure this function should return
	// one of the generated error structs, e.g.:
	//
	//    return ctx, myservice.MakeUnauthorizedError("invalid token")
	//
	// Alternatively this function may return an instance of
	// goa.ServiceError with a Name field value that matches one of
	// the design error names, e.g:
	//
	//    return ctx, goa.PermanentError("unauthorized", "invalid token")
	//
	return ctx, fmt.Errorf("not implemented")
}
{{- end }}
`
