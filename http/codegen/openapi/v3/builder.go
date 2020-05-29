package openapiv3

import (
	"fmt"
	"net/url"
	"regexp"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// OpenAPIVersion is the OpenAPI specification version targeted by this package.
const OpenAPIVersion = "3.0.3"

var uriRegex = regexp.MustCompile("^https?://")

// New returns the OpenAPI v3 specification for the given API.
// It returns nil if the design does not define HTTP endpoints.
func New(root *expr.RootExpr) *OpenAPI {
	if root == nil || root.API == nil || root.API.HTTP == nil || len(root.API.HTTP.Services) == 0 {
		// No HTTP transport
		return nil
	}

	var (
		bodies, types = buildBodyTypes(root.API)

		info     = buildInfo(root.API)
		comps    = buildComponents(root, types)
		servers  = buildServers(root.API.Servers)
		paths    = buildPaths(root.API.HTTP, bodies)
		security = buildSecurityRequirements(root.API.Requirements)
		tags     = openapi.TagsFromExpr(root.API.Meta)
	)

	return &OpenAPI{
		OpenAPI:    OpenAPIVersion,
		Info:       info,
		Components: comps,
		Paths:      paths,
		Servers:    servers,
		Security:   security,
		Tags:       tags,
	}
}

// buildInfo builds the OpenAPI Info object.
func buildInfo(api *expr.APIExpr) *Info {
	info := &Info{
		Title:          api.Title,
		Description:    api.Description,
		TermsOfService: api.TermsOfService,
		Version:        api.Version,
	}
	if c := api.Contact; c != nil {
		info.Contact = &Contact{
			Name:  c.Name,
			Email: c.Email,
			URL:   c.URL,
		}
	}
	if l := api.License; l != nil {
		info.License = &License{
			Name: l.Name,
			URL:  l.URL,
		}
	}
	return info
}

// buildComponents builds the OpenAPI Components object.
func buildComponents(root *expr.RootExpr, types map[string]*openapi.Schema) *Components {
	var schemesRef map[string]*SecuritySchemeRef
	{
		var schemes []*expr.SchemeExpr
		for _, s := range root.API.HTTP.Services {
			for _, e := range s.HTTPEndpoints {
				for _, r := range e.Requirements {
					schemes = append(schemes, r.Schemes...)
				}
			}
		}
		schemesRef = make(map[string]*SecuritySchemeRef, len(schemes))
		for _, se := range schemes {
			schemesRef[se.Hash()] = &SecuritySchemeRef{
				Value: buildSecurityScheme(se),
			}
		}
	}
	return &Components{
		SecuritySchemes: schemesRef,
		Schemas:         types,
	}
}

// buildPaths builds the OpenAPI Paths map with key as the HTTP path string and
// the value as the corresponding PathItem object.
func buildPaths(h *expr.HTTPExpr, bodies map[string]map[string]*EndpointBodies) map[string]*PathItem {
	var paths = make(map[string]*PathItem)
	for _, svc := range h.Services {
		sbod := bodies[svc.Name()]
		for _, e := range svc.HTTPEndpoints {
			for _, r := range e.Routes {
				for _, key := range r.FullPaths() {
						operation := buildOperation(key, r, sbod[e.Name()])
						path, ok := paths[key]
						if !ok {
							path = new(PathItem)
							paths[key] = path
						}
					switch r.Method {
					case "GET":
						path.Get = operation
					case "PUT":
						path.Put = operation
					case "POST":
						path.Post = operation
					case "DELETE":
						path.Delete = operation
					case "OPTIONS":
						path.Options = operation
					case "HEAD":
						path.Head = operation
					case "PATCH":
						path.Patch = operation
					}
					path.Extensions = openapi.ExtensionsFromExpr(r.Endpoint.Meta)
				}
			}
		}
	}
	return paths
}

// buildOperation builds the OpenAPI Operation object for the given path.
func buildOperation(key string, r *expr.RouteExpr, bodies *EndpointBodies) *Operation {
	e := r.Endpoint
	m := e.MethodExpr
	svc := e.Service

	// operation ID
	var opID string
	{
		opID = fmt.Sprintf("%s#%s", svc.Name(), e.Name())
		// An endpoint can have multiple routes. If there are multiple routes for
		// the endpoint suffix the operation ID with the route index.
		index := 0
		for i, rt := range r.Endpoint.Routes {
			if rt == r {
				index = i
				break
			}
		}
		if index > 0 {
			opID = fmt.Sprintf("%s#%d", opID, index)
		}
	}

	// swagger summary
	var summary string
	{
		summary = fmt.Sprintf("%s %s", e.Name(), svc.Name())
		for n, mdata := range r.Endpoint.Meta {
			if n == "swagger:summary" && len(mdata) > 0 {
				summary = mdata[0]
			}
		}
		for n, mdata := range m.Meta {
			if n == "swagger:summary" && len(mdata) > 0 {
				summary = mdata[0]
			}
		}
	}

	// body
	var body RequestBody
	{
		hb := r.Endpoint.Body
		body.Description = hb.Description
		body.Required = hb.Type != expr.Empty
		body.Extensions = openapi.ExtensionsFromExpr(hb.Meta)
	}

	// tag names
	var tagNames []string
	{
		tagNames = openapi.TagNamesFromExpr(svc.Meta, e.Meta)
		if len(tagNames) == 0 {
			// By default tag with service name
			tagNames = []string{r.Endpoint.Service.Name()}
		}
	}

	return &Operation{
		Tags:         tagNames,
		Summary:      summary,
		Description:  e.Description(),
		OperationID:  opID,
		RequestBody:  &RequestBodyRef{Value: &body},
		Security:     buildSecurityRequirements(e.Requirements),
		Deprecated:   false,
		ExternalDocs: openapi.DocsFromExpr(m.Docs, m.Meta),
		Extensions:   openapi.ExtensionsFromExpr(m.Meta),
	}
}

// buildServers builds the OpenAPI Server objects from the given server
// expressions.
func buildServers(servers []*expr.ServerExpr) []*Server {
	var svrs []*Server
	for _, svr := range servers {
		var server *Server
		for _, host := range svr.Hosts {
			var (
				serverVariable   = make(map[string]*ServerVariable)
				defaultValue     interface{}
				validationValues []interface{}
			)

			// retrieve host URL
			u, err := url.Parse(defaultURI(host))
			if err != nil {
				// bug: should be validated by DSL
				panic("invalid host " + host.Name)
			}

			// retrieve host variables
			vars := expr.AsObject(host.Variables.Type)
			for _, v := range *vars {
				defaultValue = v.Attribute.DefaultValue

				if v.Attribute.Validation != nil && len(v.Attribute.Validation.Values) > 0 {
					validationValues = append(validationValues, v.Attribute.Validation.Values...)
					if defaultValue == nil {
						defaultValue = v.Attribute.Validation.Values[0]
					}
				}

				if defaultValue != nil {
					serverVariable[v.Name] = &ServerVariable{
						Enum:        validationValues,
						Default:     defaultValue,
						Description: host.Variables.Description,
					}
				}
			}

			server = &Server{
				URL:         u.Host,
				Description: svr.Description,
				Variables:   serverVariable,
			}
			svrs = append(svrs, server)
		}
	}
	return svrs
}

// buildSecurityRequirements builds the OpenAPI security requirements for the
// given security expressions.
func buildSecurityRequirements(reqs []*expr.SecurityExpr) []map[string][]string {
	srs := make([]map[string][]string, len(reqs))
	for i, req := range reqs {
		sr := make(map[string][]string, len(req.Schemes))
		for _, sch := range req.Schemes {
			switch sch.Kind {
			case expr.BasicAuthKind, expr.APIKeyKind:
				sr[sch.Hash()] = []string{}
			case expr.OAuth2Kind, expr.JWTKind:
				scopes := make([]string, len(sch.Scopes))
				for i, scope := range sch.Scopes {
					scopes[i] = scope.Name
				}
				sr[sch.Hash()] = scopes
			}
		}
		srs[i] = sr
	}
	return srs
}

// buildSecurityScheme builds the OpenAPI SecurityScheme object from the
// top-level security scheme definition.
func buildSecurityScheme(se *expr.SchemeExpr) *SecurityScheme {
	var scheme *SecurityScheme
	switch se.Kind {
	case expr.BasicAuthKind:
		scheme = &SecurityScheme{
			Type:        "http",
			Scheme:      "basic",
			Description: se.Description,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	case expr.APIKeyKind:
		scheme = &SecurityScheme{
			Type:        "apiKey",
			Description: se.Description,
			In:          se.In,
			Name:        se.Name,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	case expr.JWTKind:
		scheme = &SecurityScheme{
			Type:        "http",
			Scheme:      "Bearer",
			Description: se.Description,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	case expr.OAuth2Kind:
		scopes := make(map[string]string, len(se.Scopes))
		for _, scope := range se.Scopes {
			scopes[scope.Name] = scope.Description
		}
		var flows OAuthFlows
		for _, f := range se.Flows {
			switch f.Kind {
			case expr.AuthorizationCodeFlowKind:
				flows.AuthorizationCode = &OAuthFlow{
					AuthorizationURL: f.AuthorizationURL,
					TokenURL:         f.TokenURL,
					RefreshURL:       f.RefreshURL,
					Scopes:           scopes,
				}
			case expr.ClientCredentialsFlowKind:
				flows.ClientCredentials = &OAuthFlow{
					TokenURL:   f.TokenURL,
					RefreshURL: f.RefreshURL,
					Scopes:     scopes,
				}
			case expr.ImplicitFlowKind:
				flows.Implicit = &OAuthFlow{
					AuthorizationURL: f.AuthorizationURL,
					RefreshURL:       f.RefreshURL,
					Scopes:           scopes,
				}
			case expr.PasswordFlowKind:
				flows.Password = &OAuthFlow{
					TokenURL:   f.TokenURL,
					RefreshURL: f.RefreshURL,
					Scopes:     scopes,
				}
			}
		}
		scheme = &SecurityScheme{
			Type:        "oauth2",
			Description: se.Description,
			Flows:       &flows,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	}
	return scheme
}

// defaultURI returns the first HTTP URI defined in the host. It substitutes any URI
// parameters with their default values or the first item in their enum.
func defaultURI(h *expr.HostExpr) string {
	// Get the first URL expression in the host by default.
	// Host expression must have at least one URI (validations would have failed
	// otherwise).
	uExpr := h.URIs[0]
	// attempt to find the first HTTP/HTTPS URL
	for _, ue := range h.URIs {
		s := ue.Scheme()
		if s == "http" || s == "https" {
			uExpr = ue
			break
		}
	}
	uri, err := h.URIString(uExpr)
	if err != nil {
		panic(err) // should never hit this!
	}
	return uri
}
