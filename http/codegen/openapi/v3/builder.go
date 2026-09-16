// This file builds OpenAPI 3 operations from HTTP endpoints. It uses the
// request or response being described to choose each example value.
package openapiv3

import (
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// OpenAPIVersion is the OpenAPI specification version of the documents
// generated for openapi.Version30.
const OpenAPIVersion = "3.0.3"

// OpenAPIVersion32 is the OpenAPI specification version of the documents
// generated for openapi.Version32.
const OpenAPIVersion32 = "3.2.0"

var (
	routeIndexReplacementRegExp = regexp.MustCompile(`\((.*){routeIndex}\)`)
)

const (
	defaultOperationIDFormat = "{service}#{method}(#{routeIndex})"
)

// New returns the OpenAPI specification conforming to the given version
// (openapi.Version30 or openapi.Version32) for the given API. It returns nil if
// the design does not define HTTP endpoints.
func New(root *expr.RootExpr, ver openapi.Version) *OpenAPI {
	if root == nil || root.API == nil {
		return nil
	}
	return NewWithValues(
		root,
		ver,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
}

// NewWithValues returns an OpenAPI specification using values in place of
// matching titles, descriptions, and examples from the evaluated design.
// The generator supplies examples for attributes that have no matching value.
func NewWithValues(root *expr.RootExpr, ver openapi.Version, generator *expr.ExampleGenerator, values openapi.Values) *OpenAPI {
	if root == nil || root.API == nil || root.API.HTTP == nil || len(root.API.HTTP.Services) == 0 {
		// No HTTP transport
		return nil
	}

	specVersion := OpenAPIVersion
	if ver == openapi.Version32 {
		specVersion = OpenAPIVersion32
	}

	var (
		bodies, types = buildBodyTypes(root.API, root.Types, root.ResultTypes, ver, generator, values)

		info     = buildInfo(root.API, ver, values)
		comps    = buildComponents(root, types, values)
		servers  = buildServers(root.API.Servers, ver, values)
		paths    = buildPaths(root.API.HTTP, bodies, root.API, ver, generator, values)
		security = buildSecurityRequirements(root.API.Requirements)
		tags     = buildTags(root.API, ver, values)
	)

	return &OpenAPI{
		OpenAPI:    specVersion,
		Info:       info,
		Components: comps,
		Paths:      paths,
		Servers:    servers,
		Security:   security,
		Tags:       tags,
	}
}

// buildInfo builds the OpenAPI Info object.
func buildInfo(api *expr.APIExpr, ver openapi.Version, values openapi.Values) *Info {
	title := values.Title(api, api.Title)
	if title == "" {
		title = "Goa API" // cannot be empty as per OpenAPI spec
	}
	info := &Info{
		Title:          title,
		Description:    values.Description(api, api.Description),
		TermsOfService: api.TermsOfService,
		Version:        api.Version,
		Extensions:     openapi.ExtensionsFromExpr(api.Meta),
	}
	if ver == openapi.Version32 {
		if s, ok := api.Meta.Last("openapi:info:summary"); ok {
			info.Summary = s
		}
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
func buildComponents(root *expr.RootExpr, types map[string]*openapi.Schema, values openapi.Values) *Components {
	var schemesRef map[string]*SecuritySchemeRef
	{
		schemesRef = make(map[string]*SecuritySchemeRef)
		for _, s := range root.API.HTTP.Services {
			if !openapi.MustGenerate(s.Meta) || !openapi.MustGenerate(s.ServiceExpr.Meta) {
				continue
			}
			for _, e := range s.HTTPEndpoints {
				if !openapi.MustGenerate(e.Meta) || !openapi.MustGenerate(e.MethodExpr.Meta) {
					continue
				}
				for _, r := range e.Requirements {
					for _, sch := range r.Schemes {
						schemesRef[sch.Hash()] = &SecuritySchemeRef{
							Value: buildSecurityScheme(sch, values),
						}
					}
				}
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
func buildPaths(h *expr.HTTPExpr, bodies map[string]map[string]*EndpointBodies, api *expr.APIExpr, ver openapi.Version, generator *expr.ExampleGenerator, values openapi.Values) map[string]*PathItem {
	var paths = make(map[string]*PathItem)
	for _, svc := range h.Services {
		if !openapi.MustGenerate(svc.Meta) || !openapi.MustGenerate(svc.ServiceExpr.Meta) {
			continue
		}
		exts := openapi.ExtensionsFromExpr(svc.Meta)
		sbod := bodies[svc.Name()]

		// endpoints
		for _, e := range svc.HTTPEndpoints {
			if !openapi.MustGenerate(e.Meta) || !openapi.MustGenerate(e.MethodExpr.Meta) {
				continue
			}
			for _, r := range e.Routes {
				for _, key := range r.FullPaths() {
					// Remove any wildcards that is defined in path as a workaround to
					// https://github.com/OAI/OpenAPI-Specification/issues/291
					key = expr.HTTPWildcardRegex.ReplaceAllString(key, "/{$1}")
					operation := buildOperation(key, r, sbod[e.Name()], generator, api.Meta, ver, values)
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
						path.Extensions = make(map[string]any)
						maps.Copy(path.Extensions, exts)
					}
				}
			}
		}

		// file servers
		for _, f := range svc.FileServers {
			if !openapi.MustGenerate(f.Meta) || !openapi.MustGenerate(f.Service.Meta) {
				continue
			}

			for _, key := range f.RequestPaths {
				// Replace wildcards in the path to OpenAPI path parameter form
				// e.g. "/ui/{*filepath}" -> "/ui/{filepath}"
				key = expr.HTTPWildcardRegex.ReplaceAllString(key, "/{$1}")
				operation := buildFileServerOperation(key, f, api, values)
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
func buildOperation(key string, r *expr.RouteExpr, bodies *EndpointBodies, rand *expr.ExampleGenerator, meta expr.MetaExpr, ver openapi.Version, values openapi.Values) *Operation {
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

	summary = fmt.Sprintf("%s %s", e.Name(), svc.Name())
	setSummary(meta)
	setSummary(svc.ServiceExpr.Meta)
	setSummary(e.Meta)
	setSummary(m.Meta)

	// OpenAPI operationId
	var operationIDFormat string
	setOperationIDFormat := func(meta expr.MetaExpr) {
		for n, mdata := range meta {
			if (n == "openapi:operationId") && len(mdata) > 0 {
				operationIDFormat = mdata[0]
			}
		}
	}

	operationIDFormat = defaultOperationIDFormat
	setOperationIDFormat(meta)
	setOperationIDFormat(m.Service.Meta)
	setOperationIDFormat(e.Meta)
	setOperationIDFormat(m.Meta)

	// request body
	var requestBody *RequestBodyRef
	if e.Body.Type != expr.Empty {
		ct := "application/json" // TBD: need a way to specify method media type in design...
		if e.MultipartRequest {
			ct = "multipart/form-data"
		}
		mt := &MediaType{Schema: bodies.RequestBody}
		initExamples(mt, e.Body, rand.At(expr.RequestBodyExampleIdentity(e)), values)
		requestBody = &RequestBodyRef{Value: &RequestBody{
			Description: requestBodyDescription(e, values),
			Required:    e.Body.Type != expr.Empty,
			Content:     map[string]*MediaType{ct: mt},
			Extensions:  openapi.ExtensionsFromExpr(e.Body.Meta),
		}}
	}

	// parameters
	var params []*ParameterRef
	{
		ps := paramsFromPath(e, key, rand, values)
		ps = append(ps, paramsFromHeadersAndCookies(e, rand, values)...)
		if e.MapQueryParams != nil {
			name := *e.MapQueryParams
			if name == "" {
				name = "payload"
			}
			ps = append(ps, &Parameter{
				Name:        name,
				Description: "Query parameters",
				In:          "query",
				Required:    name == "payload" || e.MethodExpr.Payload.IsRequired(name),
				Schema: &openapi.Schema{
					Type:                 "object",
					AdditionalProperties: true,
				},
				Style: "deepObject",
			})
		}
		params = make([]*ParameterRef, len(ps))
		for i, p := range ps {
			params[i] = &ParameterRef{Value: p}
		}
	}

	// responses
	responses := make(map[string]*ResponseRef, len(e.Responses))
	responseBodyIndexes := make(map[int]int)
	for i, r := range e.Responses {
		var resultCT string
		switch {
		case e.UsesWebSocket():
			// A WebSocket endpoint allows at most one successful response
			// definition. So it is okay to change the first successful
			// response to a HTTP 101 response for OpenAPI docs.
			if _, ok := responses[strconv.Itoa(expr.StatusSwitchingProtocols)]; !ok {
				b := bodies.ResponseBodies[r.StatusCode]
				delete(bodies.ResponseBodies, r.StatusCode)
				r = r.Dup()
				r.StatusCode = expr.StatusSwitchingProtocols
				bodies.ResponseBodies[r.StatusCode] = b
			}
		case e.UsesSSE():
			resultCT = openapi.ResponseContentType(r)
			r = r.Dup()
			r.ContentType = "text/event-stream"
		}
		var body *openapi.Schema
		if r.Body.Type != expr.Empty {
			bodyIndex := responseBodyIndexes[r.StatusCode]
			body = bodies.ResponseBodies[r.StatusCode][bodyIndex]
			responseBodyIndexes[r.StatusCode]++
		}
		owner := expr.MethodResultExampleIdentity(m)
		bodyOwner := expr.ResponseBodyExampleIdentity(e, e.Responses[i])
		resp := responseFromExpr(r, body, rand, m.Result, owner, bodyOwner, "", values)
		if ver == openapi.Version32 && e.UsesSSE() {
			setSSEContent(resp, bodies, resultCT, m.HasMixedResults())
		}
		responses[strconv.Itoa(r.StatusCode)] = &ResponseRef{Value: resp}
	}
	for _, er := range e.HTTPErrors {
		var body *openapi.Schema
		if er.Response.Body.Type != expr.Empty {
			bodyIndex := responseBodyIndexes[er.Response.StatusCode]
			body = bodies.ResponseBodies[er.Response.StatusCode][bodyIndex]
			responseBodyIndexes[er.Response.StatusCode]++
		}
		owner := expr.MethodErrorExampleIdentity(m, er.ErrorExpr)
		bodyOwner := expr.ErrorResponseBodyExampleIdentity(e, er)
		errorDescription := values.Description(er.ErrorExpr, er.Description)
		resp := responseFromExpr(er.Response, body, rand, er.AttributeExpr, owner, bodyOwner, errorDescription, values)
		desc := er.Name
		if resp.Description != nil {
			desc += ": " + *resp.Description
		}
		resp.Description = &desc
		if example, ok := openapi.ErrorResponseExample(er.ErrorExpr, er.Response.Body, rand.At(bodyOwner), values); ok {
			for _, content := range resp.Content {
				content.Example = example
			}
		}
		responses[strconv.Itoa(er.Response.StatusCode)] = &ResponseRef{Value: resp}
	}

	// tag names
	var tagNames []string
	tagNames = openapi.TagNamesFromExpr(e.Meta)
	if len(tagNames) == 0 {
		// By default tag with service name
		tagNames = []string{e.Service.Name()}
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

	// An endpoint may be marked as deprecated. if the openapi:deprecated tag is present, we populate it to true
	_, deprecated := e.Meta.Last("openapi:deprecated")
	security := buildSecurityRequirements(e.Requirements)
	if expr.HasNoSecurity(e.MethodExpr.Requirements) {
		security = SecurityRequirements{}
	}
	return &Operation{
		Tags:         tagNames,
		Summary:      summary,
		Description:  values.Description(e.MethodExpr, e.Description()),
		OperationID:  parseOperationIDTemplate(operationIDFormat, svc.Name(), e.Name(), routeIndex),
		Parameters:   params,
		RequestBody:  requestBody,
		Responses:    responses,
		Security:     security,
		Deprecated:   deprecated,
		ExternalDocs: openapi.DocsFromExprWithValues(m.Docs, m.Meta, values),
		Extensions:   openapi.ExtensionsFromMethod(m),
	}
}

// requestBodyDescription returns the request body description authored in the
// HTTP body, payload, or referenced type. It uses a deterministic default for
// computed bodies so generated OpenAPI requestBody objects are always
// self-describing.
func requestBodyDescription(e *expr.HTTPEndpointExpr, values openapi.Values) string {
	if description := values.Description(e.Body.AuthoredAttribute(), e.Body.Description); description != "" {
		return description
	}
	if ut, ok := e.Body.Type.(expr.UserType); ok {
		if desc := values.Description(ut.Attribute().AuthoredAttribute(), ut.Attribute().Description); desc != "" {
			return desc
		}
	}
	if e.MethodExpr.Payload != nil {
		if desc := values.Description(e.MethodExpr.Payload.AuthoredAttribute(), e.MethodExpr.Payload.Description); desc != "" {
			return desc
		}
	}
	return defaultRequestBodyDescription(e)
}

// defaultRequestBodyDescription returns the conventional description for
// request bodies computed from payload fields after parameters and headers have
// been projected out.
func defaultRequestBodyDescription(e *expr.HTTPEndpointExpr) string {
	return fmt.Sprintf("Request body for %s.", e.Name())
}

// buildFileServerOperation builds the OpenAPI Operation object for the given file server.
func buildFileServerOperation(key string, fs *expr.HTTPFileServerExpr, api *expr.APIExpr, values openapi.Values) *Operation {
	wildcards := expr.ExtractHTTPWildcards(key)
	svc := fs.Service

	// parameters
	var params []*ParameterRef
	if len(wildcards) > 0 {
		pref := ParameterRef{
			Value: &Parameter{
				// Use the literal wildcard (including leading '*') as name to match path if needed
				// Note: HTTPWildcardRegex already strips '*' in ExtractHTTPWildcards; however
				// the path key has been normalized to "/{name}" so the correct parameter name
				// is the bare wildcard identifier.
				Name:        wildcards[0],
				Description: "Relative file path",
				In:          "path",
				Required:    true,
				Schema: &openapi.Schema{ // string schema makes validators happy
					Type: openapi.String,
				},
			},
		}
		params = []*ParameterRef{&pref}
	}

	// responses
	var responses map[string]*ResponseRef
	{
		desc200 := "File downloaded"
		rref := ResponseRef{
			Value: &Response{
				Description: &desc200,
			},
		}
		responses = map[string]*ResponseRef{
			"200": &rref,
		}
		if len(wildcards) > 0 {
			desc404 := "File not found"
			responses["404"] = &ResponseRef{
				Value: &Response{
					Description: &desc404,
				},
			}
		}
	}

	// OpenAPI summary
	var summary string
	summary = fmt.Sprintf("Download %s", fs.FilePath)
	for n, mdata := range fs.Meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			summary = mdata[0]
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

	operationIDFormat = defaultOperationIDFormat
	setOperationIDFormat(api.Meta)
	setOperationIDFormat(svc.Meta)
	setOperationIDFormat(fs.Meta)

	// tag names
	var tagNames []string
	tagNames = openapi.TagNamesFromExpr(fs.Meta)
	if len(tagNames) == 0 {
		// By default tag with service name
		tagNames = []string{svc.Name()}
	}

	return &Operation{
		OperationID:  parseOperationIDTemplate(operationIDFormat, svc.Name(), key, 0),
		Description:  values.Description(fs, fs.Description),
		Summary:      summary,
		Parameters:   params,
		Responses:    responses,
		Tags:         tagNames,
		Security:     buildSecurityRequirements(api.Requirements),
		Deprecated:   false,
		ExternalDocs: openapi.DocsFromExprWithValues(fs.Docs, fs.Meta, values),
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
func buildServers(servers []*expr.ServerExpr, ver openapi.Version, values openapi.Values) []*Server {
	var svrs []*Server
	for _, svr := range servers {
		if !openapi.MustGenerate(svr.Meta) {
			continue
		}
		var server *Server
		for _, host := range svr.Hosts {
			if !openapi.MustGenerate(host.Meta) {
				continue
			}

			serverVariable := make(map[string]*ServerVariable)

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
				defaultValue := v.Attribute.DefaultValue
				var validationValues []any

				if v.Attribute.Validation != nil && len(v.Attribute.Validation.Values) > 0 {
					validationValues = append([]any(nil), v.Attribute.Validation.Values...)
					if defaultValue == nil {
						defaultValue = v.Attribute.Validation.Values[0]
					}
				}

				if defaultValue != nil {
					serverVariable[v.Name] = &ServerVariable{
						Enum:        validationValues,
						Default:     defaultValue,
						Description: values.Description(v.Attribute.AuthoredAttribute(), v.Attribute.Description),
					}
				}
			}

			server = &Server{
				URL:         string(uExpr),
				Description: values.Description(svr, svr.Description),
				Variables:   serverVariable,
			}
			if ver == openapi.Version32 {
				server.Name = serverName(svr, host)
			}
			svrs = append(svrs, server)
		}
	}
	return svrs
}

// serverName computes the OpenAPI 3.2 name of the Server object built for the
// given design server and host. The host name qualifies the server name when
// the server defines several hosts so generated names are unique.
func serverName(svr *expr.ServerExpr, host *expr.HostExpr) string {
	if len(svr.Hosts) == 1 {
		return svr.Name
	}
	return svr.Name + "-" + host.Name
}

// buildSecurityRequirements builds the OpenAPI security requirements for the
// given security expressions.
func buildSecurityRequirements(reqs []*expr.SecurityExpr) SecurityRequirements {
	if expr.HasNoSecurity(reqs) {
		return SecurityRequirements{}
	}
	if len(reqs) == 0 {
		return nil
	}
	srs := make(SecurityRequirements, len(reqs))
	for i, req := range reqs {
		sr := make(map[string][]string, len(req.Schemes))
		for _, sch := range req.Schemes {
			scopes := make([]string, 0)
			if sch.Kind == expr.OAuth2Kind && len(req.Scopes) > 0 {
				scopes = req.Scopes
			}
			sr[sch.Hash()] = scopes
		}
		srs[i] = sr
	}
	return srs
}

// buildSecurityScheme builds the OpenAPI SecurityScheme object from the
// top-level security scheme definition.
func buildSecurityScheme(se *expr.SchemeExpr, values openapi.Values) *SecurityScheme {
	description := values.Description(se.AuthoredScheme(), se.Description)
	var scheme *SecurityScheme
	switch se.Kind {
	case expr.BasicAuthKind:
		scheme = &SecurityScheme{
			Type:        "http",
			Scheme:      "basic",
			Description: description,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	case expr.APIKeyKind:
		scheme = &SecurityScheme{
			Type:        "apiKey",
			Description: description,
			In:          se.In,
			Name:        se.Name,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	case expr.BearerKind, expr.JWTKind:
		bearerFormat := se.BearerFormat
		if bearerFormat == "" && se.Kind == expr.JWTKind {
			bearerFormat = "JWT"
		}
		scheme = &SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: bearerFormat,
			Description:  description,
			Extensions:   openapi.ExtensionsFromExpr(se.Meta),
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
			Description: description,
			Flows:       &flows,
			Extensions:  openapi.ExtensionsFromExpr(se.Meta),
		}
	}
	return scheme
}

// buildTags builds the OpenAPI Tag object from the API expression.
func buildTags(api *expr.APIExpr, ver openapi.Version, values openapi.Values) []*openapi.Tag {
	m := make(map[string]*openapi.Tag)
	for _, t := range openapi.TagsFromExpr(api.Meta, ver) {
		m[t.Name] = t
	}
	for _, s := range api.HTTP.Services {
		if !openapi.MustGenerate(s.Meta) || !openapi.MustGenerate(s.ServiceExpr.Meta) {
			continue
		}
		for _, t := range openapi.TagsFromExpr(s.Meta, ver) {
			m[t.Name] = t
		}
	}

	// sort tag names alphabetically
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tags := make([]*openapi.Tag, 0, len(keys))
	for _, k := range keys {
		tags = append(tags, m[k])
	}

	if len(tags) == 0 {
		// add service name and description to the tags since we tag every
		// operation with service name when no custom tag is defined
		for _, s := range api.HTTP.Services {
			if !openapi.MustGenerate(s.Meta) || !openapi.MustGenerate(s.ServiceExpr.Meta) {
				continue
			}
			tags = append(tags, &openapi.Tag{
				Name:        s.Name(),
				Description: values.Description(s.ServiceExpr, s.Description()),
			})
		}
	}
	return tags
}
