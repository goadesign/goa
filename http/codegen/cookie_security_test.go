package codegen

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

func TestCookieAPIKeySecurity(t *testing.T) {
	t.Run("endpoint requirement uses cookie transport", func(t *testing.T) {
		root := expr.RunDSL(t, cookieAPIKeySecurityDSL)
		endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]
		require.Len(t, endpoint.Requirements, 1)
		require.Len(t, endpoint.Requirements[0].Schemes, 1)

		scheme := endpoint.Requirements[0].Schemes[0]
		require.Equal(t, "cookie", scheme.In)
		require.Equal(t, "__Host-ak_session", scheme.Name)

		headers := expr.AsObject(endpoint.Headers.Type)
		require.Zero(t, len(*headers), "cookie-backed api key must not synthesize an Authorization header")
	})

	t.Run("openapi uses cookie security scheme", func(t *testing.T) {
		root := expr.RunDSL(t, cookieAPIKeySecurityDSL)
		openapi.Definitions = make(map[string]*openapi.Schema)

		v2Files, err := openapiv2.Files(root, openapi.DefaultPath20)
		require.NoError(t, err)
		v2JSON := renderOpenAPIJSON(t, v2Files)
		var swagger openapi2.T
		require.NoError(t, swagger.UnmarshalJSON(v2JSON))
		require.Len(t, swagger.SecurityDefinitions, 1)
		require.Len(t, swagger.Paths, 1)
		require.Contains(t, swagger.Paths, "/auth/profile")
		require.NotNil(t, swagger.Paths["/auth/profile"].Get.Security)
		require.Len(t, *swagger.Paths["/auth/profile"].Get.Security, 1)
		for name, def := range swagger.SecurityDefinitions {
			require.Equal(t, "apiKey", def.Type, name)
			require.Equal(t, "cookie", def.In, name)
			require.Equal(t, "__Host-ak_session", def.Name, name)
			require.Contains(t, (*swagger.Paths["/auth/profile"].Get.Security)[0], name)
		}

		openapi.Definitions = make(map[string]*openapi.Schema)
		v3JSON := renderOpenAPIJSON(t, openapiv3.Files(root, openapi.Version30, openapi.DefaultPath30))
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(v3JSON)
		require.NoError(t, err)
		require.NoError(t, doc.Validate(context.Background()))
		require.Len(t, doc.Components.SecuritySchemes, 1)
		require.NotNil(t, doc.Paths.Find("/auth/profile"))
		require.NotNil(t, doc.Paths.Find("/auth/profile").Get.Security)
		require.Len(t, *doc.Paths.Find("/auth/profile").Get.Security, 1)
		for name, ref := range doc.Components.SecuritySchemes {
			require.NotNil(t, ref.Value, name)
			require.Equal(t, "apiKey", ref.Value.Type, name)
			require.Equal(t, "cookie", ref.Value.In, name)
			require.Equal(t, "__Host-ak_session", ref.Value.Name, name)
			require.Contains(t, (*doc.Paths.Find("/auth/profile").Get.Security)[0], name)
		}
	})

	t.Run("http codegen does not duplicate cookie-backed auth fields", func(t *testing.T) {
		root := expr.RunDSL(t, cookieAPIKeySecurityDSL)
		services := CreateHTTPServices(root)

		serverTypes := typesFile("gen", root.API.HTTP.Services[0], true, services)
		var serverTypesBuf bytes.Buffer
		for _, section := range serverTypes.SectionTemplates[1:] {
			require.NoError(t, section.Write(&serverTypesBuf))
		}
		serverTypesCode := codegen.FormatTestCode(t, "package foo\n"+serverTypesBuf.String())
		require.Contains(t, serverTypesCode, "func NewProfilePayload(browserSession string)")
		require.NotContains(t, serverTypesCode, "browserSession *string, browserSession *string")
		require.NotContains(t, serverTypesCode, "browserSession string, browserSession string")

		serverFiles := ServerFiles("", services)
		require.Len(t, serverFiles, 2)
		serverDecode := codegen.SectionCode(t, serverFiles[1].SectionTemplates[2])
		require.Contains(t, serverDecode, `r.Cookie("__Host-ak_session")`)
		require.NotContains(t, serverDecode, "Authorization")
		require.NotContains(t, serverDecode, "browserSession *string, browserSession *string")
		require.NotContains(t, serverDecode, "browserSession string, browserSession string")

		clientFiles := ClientFiles("", services)
		require.Len(t, clientFiles, 2)
		clientEncode := codegen.SectionCode(t, clientFiles[1].SectionTemplates[2])
		require.Contains(t, clientEncode, `req.AddCookie(&http.Cookie{`)
		require.Contains(t, clientEncode, `Name:  "__Host-ak_session"`)
		require.NotContains(t, clientEncode, "Authorization")
	})
}

func renderOpenAPIJSON(t *testing.T, files []*codegen.File) []byte {
	t.Helper()

	for _, f := range files {
		if filepath.Ext(f.Path) != ".json" {
			continue
		}
		require.Len(t, f.SectionTemplates, 1)
		section := f.SectionTemplates[0]
		require.NotEmpty(t, section.Source)
		require.NotNil(t, section.Data)

		var buf bytes.Buffer
		tmpl := template.Must(template.New("openapi").Funcs(section.FuncMap).Parse(section.Source))
		require.NoError(t, tmpl.Execute(&buf, section.Data))
		return buf.Bytes()
	}

	t.Fatalf("no JSON OpenAPI file generated")
	return nil
}

var cookieAPIKeySecurityDSL = func() {
	var browserSessionCookie = dsl.APIKeySecurity("browser_session_cookie", func() {
		dsl.Description("Browser session cookie")
	})

	dsl.Service("cookieSecurity", func() {
		dsl.Method("profile", func() {
			dsl.Security(browserSessionCookie)
			dsl.Payload(func() {
				dsl.APIKey("browser_session_cookie", "browser_session", dsl.String, func() {
					dsl.Description("Opaque browser session cookie")
				})
				dsl.Required("browser_session")
			})
			dsl.Result(dsl.Empty)
			dsl.HTTP(func() {
				dsl.GET("/auth/profile")
				dsl.Cookie("browser_session:__Host-ak_session")
				dsl.Response(dsl.StatusOK)
			})
		})
	})
}
