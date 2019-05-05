package service

import (
	"os"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// AuthFuncsFile returns a file that contains a dummy implementation of the
// authorization functions needed to instantiate the service endpoints.
func AuthFuncsFile(genpkg string, root *expr.RootExpr) *codegen.File {
	filepath := "auth.go"
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}

	var (
		sections []*codegen.SectionTemplate
		generate bool

		scope = codegen.NewNameScope()
	)
	{
		specs := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "goa.design/goa/v3", Name: "goa"},
			{Path: "goa.design/goa/v3/security"},
		}
		for _, svc := range root.Services {
			sd := Services.Get(svc.Name)
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, codegen.SnakeCase(sd.VarName)),
				Name: scope.Unique(sd.PkgName),
			})
		}

		apiPkg := scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
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
