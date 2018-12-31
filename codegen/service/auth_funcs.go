package service

import (
	"os"
	"path"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// AuthFuncsFile returns a file that contains a dummy implementation of the
// authorization functions needed to instantiate the service endpoints.
func AuthFuncsFile(genpkg string, root *expr.RootExpr) *codegen.File {
	var (
		apiPkg   = strings.ToLower(codegen.Goify(root.API.Name, false))
		rootPath = "."
		filepath = "auth.go"
	)
	{
		if _, err := os.Stat(filepath); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
		idx := strings.LastIndex(genpkg, string(os.PathSeparator))
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
	}

	var (
		sections []*codegen.SectionTemplate
		generate bool
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
		header := codegen.Header("", apiPkg, specs)
		sections = []*codegen.SectionTemplate{header}
		for _, s := range root.Services {
			svc := Services.Get(s.Name)
			if len(svc.Schemes) > 0 {
				generate = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "security-authfuncs",
					Source: dummyAuthFuncsT,
					Data:   svc,
				})
			}
		}
	}
	if len(sections) == 0 || !generate {
		return nil
	}

	return &codegen.File{
		Path:             filepath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// data: Data
const dummyAuthFuncsT = `{{ range .Schemes }}
{{ printf "%sAuth implements the authorization logic for service %q for the %q security scheme." .Type $.Name .SchemeName | comment }}
func (s *{{ $.VarName }}srvc) {{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, scheme *security.{{ .Type }}Scheme) (context.Context, error) {
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
