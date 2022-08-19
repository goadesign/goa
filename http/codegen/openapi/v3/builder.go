package openapiv3

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// OpenAPIVersion is the OpenAPI specification version targeted by this package.
const OpenAPIVersion = "3.0.3"

var (
	routeIndexReplacementRegExp = regexp.MustCompile(`\((.*){routeIndex}\)`)
)

const (
	defaultOperationIDFormat = "{service}#{method}(#{routeIndex})"
)

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
		paths    = buildPaths(root.API.HTTP, bodies, root.API)
		security = buildSecurityRequirements(root.API.Requirements)
		tags     = buildTags(root.API)
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
	ver := api.Version
	if ver == "" {
		ver = "1.0" // cannot be empty as per OpenAPI spec
	}
	title := api.Title
	if title == "" {
		title = "Goa API" // cannot be empty as per OpenAPI spec
	}
	info := &Info{
		Title:          title,
		Description:    api.Description,
		TermsOfService: api.TermsOfService,
		Version:        ver,
		Extensions:     openapi.ExtensionsFromExpr(api.Meta),
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
func buildPaths(h *expr.HTTPExpr, bodies map[string]map[string]*EndpointBodies, api *expr.APIExpr) map[string]*PathItem {
	var paths = make(map[string]*PathItem)
	for _, svc := range h.Services {
		if !mustGenerate(svc.Meta) || !mustGenerate(svc.ServiceExpr.Meta) {
			continue
		}

		exts := openapi.ExtensionsFromExpr(svc.Meta)
		sbod := bodies[svc.Name()]

		// endpoints
		for _, e := range svc.HTTPEndpoints {
			if !mustGenerate(e.Meta) || !mustGenerate(e.MethodExpr.Meta) {
				continue
			}

			for _, r := range e.Routes {
				for _, key := range r.FullPaths() {
					// Remove any wildcards that is defined in path as a workaround to
					// https://github.com/OAI/OpenAPI-Specification/issues/291
					key = expr.HTTPWildcardRegex.ReplaceAllString(key, "/{$1}")
					operation := buildOperation(key, r, sbod[e.Name()], api.Random())
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
					if len(exts) > 0 {
						path.Extensions = make(map[string]interface{})
						for k, v := range exts {
							path.Extensions[k] = v
						}
					}
				}
			}
		}

		// file servers
		for _, f := range svc.FileServers {
			if !mustGenerate(f.Meta) || !mustGenerate(f.Service.Meta) {
				continue
			}

			for _, key := range f.RequestPaths {
				operation := buildFileServerOperation(key, f, api)
				path, ok := paths[key]
				if !ok {
					path = new(PathItem)
					paths[key] = path
				}
				path.Get = operation
			}
		}
	}
	return paths
}

// buildOperation builds the OpenAPI Operation object for the given path.
func buildOperation(key string, r *expr.RouteExpr, bodies *EndpointBodies, rand *expr.Random) *Operation {
	e := r.Endpoint
	m := e.MethodExpr
	svc := e.Service

	// OpenAPI summary
	var summary string
	setSummary := func(meta expr.MetaExpr) {
		for n, mdata := range meta {
			if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
				if mdata[0] == "{path}" {
					summary = r.Path
				} else {
					summary = mdata[0]
				}
			}
		}
	}

	{
		summary = fmt.Sprintf("%s %s", e.Name(), svc.Name())
		setSummary(expr.Root.API.Meta)
		setSummary(r.Endpoint.Meta)
		setSummary(m.Meta)
	}

	// OpenAPI operationId
	var operationIDFormat string
	setOperationIDFormat := func(meta expr.MetaExpr) {
		for n, mdata := range meta {
			if (n == "openapi:operationId") && len(mdata) > 0 {
				operationIDFormat = mdata[0]
			}
		}
	}

	{
		operationIDFormat = defaultOperationIDFormat
		setOperationIDFormat(expr.Root.API.Meta)
		setOperationIDFormat(m.Service.Meta)
		setOperationIDFormat(r.Endpoint.Meta)
		setOperationIDFormat(m.Meta)
	}

	// request body
	var requestBody *RequestBodyRef
	if e.Body.Type != expr.Empty {
		ct := "application/json" // TBD: need a way to specify method media type in design...
		if e.MultipartRequest {
			ct = "multipart/form-data"
		}
		mt := &MediaType{Schema: bodies.RequestBody}
		initExamples(mt, e.Body, rand)
		requestBody = &RequestBodyRef{Value: &RequestBody{
			Description: e.Body.Description,
			Required:    e.Body.Type != expr.Empty,
			Content:     map[string]*MediaType{ct: mt},
			Extensions:  openapi.ExtensionsFromExpr(e.Body.Meta),
		}}
	}

	// parameters
	var params []*ParameterRef
	{
		ps := paramsFromPath(e.Params, key, rand)
		ps = append(ps, paramsFromHeadersAndCookies(e, rand)...)
		params = make([]*ParameterRef, len(ps))
		for i, p := range ps {
			params[i] = &ParameterRef{Value: p}
		}
	}

	// responses
	var responses map[string]*ResponseRef
	{
		responses = make(map[string]*ResponseRef, len(e.Responses))
		for _, r := range e.Responses {
			if e.MethodExpr.IsStreaming() {
				// A streaming endpoint allows at most one successful response
				// definition. So it is okay to change the first successful
				// response to a HTTP 101 response for openapi docs.
				if _, ok := responses[strconv.Itoa(expr.StatusSwitchingProtocols)]; !ok {
					b := bodies.ResponseBodies[r.StatusCode]
					delete(bodies.ResponseBodies, r.StatusCode)
					r = r.Dup()
					r.StatusCode = expr.StatusSwitchingProtocols
					bodies.ResponseBodies[r.StatusCode] = b
				}
			}
			resp := responseFromExpr(r, bodies.ResponseBodies, rand)
			responses[strconv.Itoa(r.StatusCode)] = &ResponseRef{Value: resp}
		}
		for _, er := range e.HTTPErrors {
			if er.Description != "" && er.Response.Description == "" {
				er.Response.Description = er.Description
			}
			resp := responseFromExpr(er.Response, bodies.ResponseBodies, rand)
			desc := er.Name
			if resp.Description != nil {
				desc += ": " + *resp.Description
			}
			resp.Description = &desc
			for _, content := range resp.Content {
				content.Example = nil
			}
			responses[strconv.Itoa(er.Response.StatusCode)] = &ResponseRef{Value: resp}
		}
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

	// An endpoint can have multiple routes, so we need to be able to build a unique
	// operationId for each route.
	var routeIndex int
	for i, rt := range e.Routes {
		if rt == r {
			routeIndex = i
			break
		}
	}

	return &Operation{
		Tags:         tagNames,
		Summary:      summary,
		Description:  e.Description(),
		OperationID:  parseOperationIDTemplate(operationIDFormat, svc.Name(), e.Name(), routeIndex),
		Parameters:   params,
		RequestBody:  requestBody,
		Responses:    responses,
		Security:     buildSecurityRequirements(e.Requirements),
		Deprecated:   false,
		ExternalDocs: openapi.DocsFromExpr(m.Docs, m.Meta),
		Extensions:   openapi.ExtensionsFromExpr(m.Meta),
	}
}

// buildOperation builds the OpenAPI Operation object for the given file server.
func buildFileServerOperation(key string, fs *expr.HTTPFileServerExpr, api *expr.APIExpr) *Operation {
	wildcards := expr.ExtractHTTPWildcards(key)
	svc := fs.Service

	// parameters
	var params []*ParameterRef
	{
		if len(wildcards) > 0 {
			pref := ParameterRef{
				Value: &Parameter{
					Name:        wildcards[0],
					Description: "Relative file path",
					In:          "path",
					Required:    true,
				},
			}
			params = []*ParameterRef{&pref}
		}
	}

	// responses
	var responses map[string]*ResponseRef
	{
		desc := "File downloaded"
		rref := ResponseRef{
			Value: &Response{
				Description: &desc,
			},
		}
		responses = map[string]*ResponseRef{
			"200": &rref,
		}
		if len(wildcards) > 0 {
			desc = "File not found"
			responses["404"] = &ResponseRef{
				Value: &Response{
					Description: &desc,
				},
			}
		}
	}

	// OpenAPI summary
	var summary string
	{
		summary = fmt.Sprintf("Download %s", fs.FilePath)
		for n, mdata := range fs.Meta {
			if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
				summary = mdata[0]
			}
		}
	}

	// OpenAPI operationId
	var operationIDFormat string
	setOperationIDFormat := func(meta expr.MetaExpr) {
		for n, mdata := range meta {
			if n == "openapi:operationId" && len(mdata) > 0 {
				operationIDFormat = mdata[0]
			}
		}
	}

	{
		operationIDFormat = defaultOperationIDFormat
		setOperationIDFormat(api.Meta)
		setOperationIDFormat(svc.Meta)
		setOperationIDFormat(fs.Meta)
	}

	// tag names
	var tagNames []string
	{
		tagNames = openapi.TagNamesFromExpr(svc.Meta, fs.Meta)
		if len(tagNames) == 0 {
			// By default tag with service name
			tagNames = []string{svc.Name()}
		}
	}

	return &Operation{
		OperationID:  parseOperationIDTemplate(operationIDFormat, svc.Name(), key, 0),
		Description:  fs.Description,
		Summary:      summary,
		Parameters:   params,
		Responses:    responses,
		Tags:         tagNames,
		Security:     buildSecurityRequirements(api.Requirements),
		Deprecated:   false,
		ExternalDocs: openapi.DocsFromExpr(fs.Docs, fs.Meta),
		Extensions:   openapi.ExtensionsFromExpr(fs.Meta),
	}
}

func parseOperationIDTemplate(template, service, method string, routeIndex int) string {
	// Early return if no replacement is needed for the template.
	if !strings.Contains(template, "{") && routeIndex == 0 {
		return template
	}

	// The template replacer
	repl := strings.NewReplacer(
		"{service}", service,
		"{method}", method,
	)

	operationID := repl.Replace(template)

	if routeIndex == 0 {
		return routeIndexReplacementRegExp.ReplaceAllString(operationID, "")
	}

	// If the routeIndex is greater than 0, we need to add the routeIndex to the operationId.
	if sep := routeIndexReplacementRegExp.FindStringSubmatch(template); sep != nil {
		return routeIndexReplacementRegExp.ReplaceAllString(operationID, fmt.Sprintf("%s%d", sep[1], routeIndex))
	}

	// Fallback in the event that the operationId doesn't contain the routeIndex placeholder.
	return fmt.Sprintf("%s#%d", operationID, routeIndex)
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

			// Get the first URL expression in the host by default.
			// Host expression must have at least one URI (validations would have failed
			// otherwise).
			uExpr := host.URIs[0]
			// attempt to find the first HTTP/HTTPS URL
			for _, ue := range host.URIs {
				s := ue.Scheme()
				if s == "http" || s == "https" {
					uExpr = ue
					break
				}
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
				URL:         string(uExpr),
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
			Scheme:      "bearer",
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

// buildTags builds the OpenAPI Tag object from the API expression.
func buildTags(api *expr.APIExpr) []*openapi.Tag {
	// if a tag with same name is defined in API, Service, and endpoint
	// Meta expressions then the definition in endpoint Meta expression
	// takes highest precedence followed by Service and API.

	m := make(map[string]*openapi.Tag)
	for _, s := range api.HTTP.Services {
		if !mustGenerate(s.Meta) || !mustGenerate(s.ServiceExpr.Meta) {
			continue
		}
		for _, t := range openapi.TagsFromExpr(s.Meta) {
			m[t.Name] = t
		}
		for _, e := range s.HTTPEndpoints {
			if !mustGenerate(e.Meta) || !mustGenerate(e.MethodExpr.Meta) {
				continue
			}
			for _, t := range openapi.TagsFromExpr(e.Meta) {
				m[t.Name] = t
			}
		}
	}

	// sort tag names alphabetically
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tags []*openapi.Tag
	{
		for _, k := range keys {
			tags = append(tags, m[k])
		}

		if len(tags) == 0 {
			// add service name and description to the tags since we tag every
			// operation with service name when no custom tag is defined
			for _, s := range api.HTTP.Services {
				if !mustGenerate(s.Meta) || !mustGenerate(s.ServiceExpr.Meta) {
					continue
				}
				tags = append(tags, &openapi.Tag{Name: s.Name(), Description: s.Description()})
			}
		}
	}
	return tags
}

// mustGenerate returns true if the meta indicates that a OpenAPI specification should be
// generated, false otherwise.
func mustGenerate(meta expr.MetaExpr) bool {
	m, ok := meta.Last("openapi:generate")
	if !ok {
		m, ok = meta.Last("swagger:generate")
	}
	if ok && m == "false" {
		return false
	}
	return true
}
