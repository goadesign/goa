// This file builds Swagger 2.0 operations and schemas from HTTP endpoints. It
// uses the request or response being described to choose each example value.
package openapiv2

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiinternal "goa.design/goa/v3/http/codegen/openapi/internal"
)

// NewV2 returns the OpenAPI v2 specification for the given API.
func NewV2(root *expr.RootExpr, h *expr.HostExpr) (*V2, error) {
	if root == nil {
		return nil, nil
	}
	return NewV2WithValues(
		root,
		h,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
}

// NewV2WithValues returns the OpenAPI v2 specification using values in place
// of matching titles, descriptions, and examples from the evaluated design.
// The generator supplies examples for attributes that have no matching value.
func NewV2WithValues(root *expr.RootExpr, h *expr.HostExpr, generator *expr.ExampleGenerator, values openapi.Values) (*V2, error) {
	if root == nil {
		return nil, nil
	}
	tags := openapi.TagsFromExpr(root.API.Meta, openapi.Version20)
	u, err := url.Parse(defaultURI(h))
	if err != nil {
		// This should never happen because server expression must have been
		// validated. If it does, then we must fix server validation.
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}
	host := u.Host
	if !openapi.MustGenerate(root.API.Servers[0].Meta) || !openapi.MustGenerate(h.Meta) {
		host = ""
	}
	schemas := newSchemaBuilder(values)
	var contact *expr.ContactExpr
	if root.API.Contact != nil {
		contactCopy := *root.API.Contact
		contact = &contactCopy
	}
	var license *expr.LicenseExpr
	if root.API.License != nil {
		licenseCopy := *root.API.License
		license = &licenseCopy
	}

	basePath := root.API.HTTP.Path
	if hasAbsoluteRoutes(root) {
		basePath = ""
	}
	params := paramsFromExpr(nil, root.API.HTTP.Params, basePath, values)
	var paramMap map[string]*Parameter
	if len(params) > 0 {
		paramMap = make(map[string]*Parameter, len(params))
		for _, p := range params {
			paramMap[p.Name] = p
		}
	}
	s := &V2{
		Swagger: "2.0",
		Info: &Info{
			Title:          values.Title(root.API, root.API.Title),
			Description:    values.Description(root.API, root.API.Description),
			TermsOfService: root.API.TermsOfService,
			Contact:        contact,
			License:        license,
			Version:        root.API.Version,
			Extensions:     openapi.ExtensionsFromExpr(root.API.Meta),
		},
		Host:                host,
		BasePath:            basePath,
		Paths:               make(map[string]any),
		Consumes:            slices.Clone(root.API.HTTP.Consumes),
		Produces:            slices.Clone(root.API.HTTP.Produces),
		Parameters:          paramMap,
		Tags:                tags,
		SecurityDefinitions: securitySpecFromExpr(root, values),
		ExternalDocs:        openapi.DocsFromExprWithValues(root.API.Docs, root.API.Meta, values),
	}
	for _, res := range root.API.HTTP.Services {
		if !openapi.MustGenerate(res.Meta) || !openapi.MustGenerate(res.ServiceExpr.Meta) {
			continue
		}
		maps.Copy(s.Paths, openapi.ExtensionsFromExpr(res.Meta))
		for _, fs := range res.FileServers {
			if !openapi.MustGenerate(fs.Meta) || !openapi.MustGenerate(fs.Service.Meta) {
				continue
			}
			buildPathFromFileServer(s, root, fs, schemas, generator, values)
		}
		for _, a := range res.HTTPEndpoints {
			if !openapi.MustGenerate(a.Meta) || !openapi.MustGenerate(a.MethodExpr.Meta) {
				continue
			}
			for _, route := range a.Routes {
				buildPathFromExpr(s, root, h, route, basePath, schemas, generator, values)
			}
		}
	}
	if len(schemas.definitions) > 0 {
		s.Definitions = schemas.definitions
		for _, d := range schemas.definitions {
			// Swagger 2.0 does not support media metadata or schema links.
			d.Media = nil
			d.Links = nil
		}
	}
	// Convert OpenAPI 3.0 references (#/$defs/) to Swagger 2.0 format (#/definitions/)
	convertRefsToV2(s)
	return s, nil
}

// defaultURI returns the first URI defined in the host. It substitutes any URI
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

// addScopeDescription generates and adds required scopes to the scheme's description.
func addScopeDescription(scopes []*expr.ScopeExpr, sd *SecurityDefinition) {
	// Generate scopes to add to description
	lines := make([]string, 0, len(scopes))

	for _, scope := range scopes {
		lines = append(lines, fmt.Sprintf("  * `%s`: %s", scope.Name, scope.Description))
	}
	// Add scope description only if scopes are defined
	if len(lines) > 0 {
		if sd.Description != "" {
			sd.Description += "\n"
		}
		sd.Description += fmt.Sprintf("\n**Security Scopes**:\n%s", strings.Join(lines, "\n"))
	}
}

// securitySpecFromExpr generates the OpenAPI security definitions from the
// security design.
func securitySpecFromExpr(root *expr.RootExpr, values openapi.Values) map[string]*SecurityDefinition {
	sds := make(map[string]*SecurityDefinition)
	for _, svc := range root.API.HTTP.Services {
		if !openapi.MustGenerate(svc.Meta) || !openapi.MustGenerate(svc.ServiceExpr.Meta) {
			continue
		}
		for _, e := range svc.HTTPEndpoints {
			if !openapi.MustGenerate(e.Meta) || !openapi.MustGenerate(e.MethodExpr.Meta) {
				continue
			}
			for _, req := range e.Requirements {
				for _, s := range req.Schemes {
					sd := SecurityDefinition{
						Description: values.Description(s.AuthoredScheme(), s.Description),
						Extensions:  openapi.ExtensionsFromExpr(s.Meta),
					}

					switch s.Kind {
					case expr.BasicAuthKind:
						sd.Type = "basic"
						addScopeDescription(s.Scopes, &sd)
					case expr.APIKeyKind:
						sd.Type = "apiKey"
						sd.In = s.In
						sd.Name = s.Name
						addScopeDescription(s.Scopes, &sd)
					case expr.BearerKind, expr.JWTKind:
						sd.Type = "apiKey"
						// OpenAPI V2 spec does not support HTTP bearer schemes. Hence
						// we add the scheme information to the description.
						addScopeDescription(s.Scopes, &sd)
						sd.In = s.In
						sd.Name = s.Name
					case expr.OAuth2Kind:
						sd.Type = "oauth2"
						if scopesLen := len(s.Scopes); scopesLen > 0 {
							scopes := make(map[string]string, scopesLen)
							for _, scope := range s.Scopes {
								scopes[scope.Name] = scope.Description
							}
							sd.Scopes = scopes
						}
					}
					if len(s.Flows) > 0 {
						switch s.Flows[0].Kind {
						case expr.AuthorizationCodeFlowKind:
							sd.Flow = "accessCode"
						case expr.ImplicitFlowKind:
							sd.Flow = "implicit"
						case expr.PasswordFlowKind:
							sd.Flow = "password"
						case expr.ClientCredentialsFlowKind:
							sd.Flow = "application"
						}
						sd.AuthorizationURL = s.Flows[0].AuthorizationURL
						sd.TokenURL = s.Flows[0].TokenURL
					}
					sds[s.Hash()] = &sd
				}
			}
		}
	}
	return sds
}

// hasAbsoluteRoutes returns true if any endpoint exposed by the API uses an
// absolute route of if the API has file servers. This is needed as OpenAPI does
// not support exceptions to the base path so if the API has any absolute route
// the base path must be "/" and all routes must be absolutes.
func hasAbsoluteRoutes(root *expr.RootExpr) bool {
	hasAbsoluteRoutes := false
	for _, res := range root.API.HTTP.Services {
		if !openapi.MustGenerate(res.Meta) || !openapi.MustGenerate(res.ServiceExpr.Meta) {
			continue
		}
		for _, fs := range res.FileServers {
			if !openapi.MustGenerate(fs.Meta) || !openapi.MustGenerate(fs.Service.Meta) {
				continue
			}
			hasAbsoluteRoutes = true
			break
		}
		for _, a := range res.HTTPEndpoints {
			if !openapi.MustGenerate(a.Meta) || !openapi.MustGenerate(a.MethodExpr.Meta) {
				continue
			}
			for _, ro := range a.Routes {
				if ro.IsAbsolute() {
					hasAbsoluteRoutes = true
					break
				}
			}
			if hasAbsoluteRoutes {
				break
			}
		}
		if hasAbsoluteRoutes {
			break
		}
	}
	return hasAbsoluteRoutes
}

func summaryFromExpr(name string, e *expr.HTTPEndpointExpr, meta expr.MetaExpr) string {
	for n, mdata := range e.Meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			return mdata[0]
		}
	}
	for n, mdata := range e.MethodExpr.Meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			return mdata[0]
		}
	}
	for n, mdata := range e.Service.ServiceExpr.Meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			return mdata[0]
		}
	}
	for n, mdata := range meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			return mdata[0]
		}
	}
	return name
}

func summaryFromMeta(name string, meta expr.MetaExpr) string {
	for n, mdata := range meta {
		if (n == "openapi:summary" || n == "swagger:summary") && len(mdata) > 0 {
			return mdata[0]
		}
	}
	return name
}

func paramsFromExpr(endpoint *expr.HTTPEndpointExpr, params *expr.MappedAttributeExpr, path string, values openapi.Values) []*Parameter {
	if params == nil {
		return nil
	}
	var (
		res       []*Parameter
		wildcards = expr.ExtractHTTPWildcards(path)
	)
	codegen.WalkMappedAttr(params, func(n, pn string, required bool, at *expr.AttributeExpr) error { // nolint: errcheck
		in := "query"
		if slices.Contains(wildcards, n) {
			in = "path"
			required = true
		}
		if endpoint != nil && in != "path" && openapiinternal.IsSecurityParameter(endpoint, in, pn) {
			return nil
		}
		param := paramFor(at, pn, in, required, values)
		res = append(res, param)
		return nil
	})
	return res
}

func paramsFromHeaders(endpoint *expr.HTTPEndpointExpr, values openapi.Values) []*Parameter {
	var params []*Parameter

	expr.WalkMappedAttr(endpoint.Headers, func(name, elem string, att *expr.AttributeExpr) error { // nolint: errcheck
		if openapiinternal.IsSecurityParameter(endpoint, "header", elem) {
			return nil
		}
		required := endpoint.Headers.IsRequiredNoDefault(name)
		params = append(params, paramFor(att, elem, "header", required, values))
		return nil
	})

	return params
}

func paramFor(at *expr.AttributeExpr, name, in string, required bool, values openapi.Values) *Parameter {
	alias := at
	at = resolvedAliasAttribute(at)
	p := &Parameter{
		In:          in,
		Name:        name,
		Default:     openapi.ToStringMap(at.DefaultValue),
		Description: values.Description(alias.AuthoredAttribute(), at.Description),
		Required:    required,
	}
	p.Type, p.Format = openAPITypeFormat(at)
	if expr.IsArray(at.Type) {
		p.Items = itemsFromExpr(expr.AsArray(at.Type).ElemType)
		p.CollectionFormat = "multi"
	}
	p.Extensions = openapi.ExtensionsFromExpr(at.Meta)
	initAttributeValidations(alias, p)
	return p
}

func itemsFromExpr(at *expr.AttributeExpr) *Items {
	itemType, itemFormat := openAPITypeFormat(at)
	items := &Items{
		Type:   itemType,
		Format: itemFormat,
	}
	initAttributeValidations(at, items)
	if expr.IsArray(at.Type) {
		items.Items = itemsFromExpr(expr.AsArray(at.Type).ElemType)
	}
	return items
}

func responseSpecFromExpr(_ *V2, root *expr.RootExpr, r *expr.HTTPResponseExpr, typeNamePrefix string, schemas *schemaBuilder, generator *expr.ExampleGenerator, fallbackDescription string, values openapi.Values) *Response {
	var schema *openapi.Schema
	if mt, ok := r.Body.Type.(*expr.ResultTypeExpr); ok {
		view := expr.DefaultView
		if v, ok := r.Body.Meta.Last(expr.ViewMetaKey); ok {
			view = v
		}
		schema = openapi.NewSchema()
		projection := openapi.ProjectResponseResult(mt, view)
		schema.Ref = schemas.projectedResultTypeRefWithPrefix(root.API, mt, projection.Result, typeNamePrefix, generator)
	} else if r.Body.Type != expr.Empty {
		schema = schemas.attributeTypeSchemaWithPrefix(root.API, r.Body, typeNamePrefix, generator)
	}
	if schema != nil {
		schema.Extensions = openapi.ExtensionsFromExpr(r.Meta)
	}
	headers := headersFromExpr(r.Headers, values)
	desc := values.Description(r, r.Description)
	if desc == "" {
		desc = fallbackDescription
	}
	if desc == "" {
		desc = fmt.Sprintf("%s response.", http.StatusText(r.StatusCode))
	}
	return &Response{
		Description: desc,
		Schema:      schema,
		Headers:     headers,
		Extensions:  openapi.ExtensionsFromExpr(r.Meta),
	}
}

func headersFromExpr(headers *expr.MappedAttributeExpr, values openapi.Values) map[string]*Header {
	if headers == nil {
		return nil
	}
	res := make(map[string]*Header)
	codegen.WalkMappedAttr(headers, func(_, n string, _ bool, at *expr.AttributeExpr) error { // nolint: errcheck
		headerType, headerFormat := openAPITypeFormat(at)
		header := &Header{
			Default:     at.DefaultValue,
			Description: values.Description(at.AuthoredAttribute(), at.Description),
			Type:        headerType,
			Format:      headerFormat,
		}
		if expr.IsArray(at.Type) {
			header.Items = itemsFromExpr(expr.AsArray(at.Type).ElemType)
		}
		initAttributeValidations(at, header)
		res[n] = header
		return nil
	})
	if len(res) == 0 {
		return nil
	}
	return res
}

func openAPITypeFormat(at *expr.AttributeExpr) (string, string) {
	at = resolvedAliasAttribute(at)
	p, ok := at.Type.(expr.Primitive)
	if !ok {
		return at.Type.Name(), ""
	}
	switch p.Kind() {
	case expr.IntKind, expr.UIntKind, expr.Int64Kind, expr.UInt64Kind:
		return "integer", "int64"
	case expr.Int32Kind, expr.UInt32Kind:
		return "integer", "int32"
	case expr.Float32Kind:
		return "number", "float"
	case expr.Float64Kind:
		return "number", "double"
	case expr.BytesKind:
		return "string", "byte"
	case expr.AnyKind:
		return "", ""
	default:
		return p.Name(), ""
	}
}

func resolvedAliasAttribute(at *expr.AttributeExpr) *expr.AttributeExpr {
	if expr.IsAlias(at.Type) {
		return at.Type.(expr.UserType).Attribute()
	}
	return at
}

func initAttributeValidations(at *expr.AttributeExpr, def any) {
	resolved := resolvedAliasAttribute(at)
	if resolved != at {
		initValidations(resolved, def)
	}
	initValidations(at, def)
}

func buildPathFromFileServer(s *V2, root *expr.RootExpr, fs *expr.HTTPFileServerExpr, schemas *schemaBuilder, generator *expr.ExampleGenerator, values openapi.Values) {
	for _, path := range fs.RequestPaths {
		wcs := expr.ExtractHTTPWildcards(path)
		var param []*Parameter
		if len(wcs) > 0 {
			param = []*Parameter{{
				In:          "path",
				Name:        wcs[0],
				Description: "Relative file path",
				Required:    true,
				Type:        "string",
			}}
		}

		responses := map[string]*Response{
			"200": {
				Description: "File downloaded",
				Schema:      &openapi.Schema{Type: openapi.File},
			},
		}
		if len(wcs) > 0 {
			errgen := generator.At(expr.UserTypeExampleIdentity(expr.ErrorResult))
			schema := schemas.typeSchema(root.API, expr.ErrorResult, errgen)
			responses["404"] = &Response{Description: "File not found", Schema: schema}
		}

		operationID := fmt.Sprintf("%s#%s", fs.Service.Name(), path)
		schemes := root.API.Schemes()
		// remove grpc and grpcs from schemes since it is not a valid scheme in
		// openapi.
		for i := len(schemes) - 1; i >= 0; i-- {
			if schemes[i] == "grpc" || schemes[i] == "grpcs" {
				schemes = append(schemes[:i], schemes[i+1:]...)
			}
		}

		tagNames := openapi.TagNamesFromExpr(fs.Meta)
		if len(tagNames) == 0 {
			// By default tag with service name
			tagNames = []string{fs.Service.Name()}
		}

		operation := &Operation{
			Description:  values.Description(fs, fs.Description),
			Summary:      summaryFromMeta(fmt.Sprintf("Download %s", fs.FilePath), fs.Meta),
			ExternalDocs: openapi.DocsFromExprWithValues(fs.Docs, fs.Meta, values),
			OperationID:  operationID,
			Parameters:   param,
			Responses:    responses,
			Schemes:      schemes,
			Tags:         tagNames,
		}

		key := expr.HTTPWildcardRegex.ReplaceAllString(path, "/{$1}")
		if key == "" {
			key = "/"
		}
		var path any
		var ok bool
		if path, ok = s.Paths[key]; !ok {
			path = new(Path)
			s.Paths[key] = path
		}
		p := path.(*Path)
		p.Get = operation
		p.Extensions = openapi.ExtensionsFromExpr(fs.Meta)
	}
}

func buildPathFromExpr(s *V2, root *expr.RootExpr, h *expr.HostExpr, route *expr.RouteExpr, basePath string, schemas *schemaBuilder, generator *expr.ExampleGenerator, values openapi.Values) {
	endpoint := route.Endpoint

	tagNames := openapi.TagNamesFromExpr(endpoint.Meta)
	if len(tagNames) == 0 {
		// By default tag with service name
		tagNames = []string{route.Endpoint.Service.Name()}
	}
	for _, key := range route.FullPaths() {
		// Remove any wildcards that is defined in path as a workaround to
		// https://github.com/OAI/OpenAPI-Specification/issues/291
		key = expr.HTTPWildcardRegex.ReplaceAllString(key, "/{$1}")
		params := paramsFromExpr(endpoint, endpoint.Params, key, values)
		params = append(params, paramsFromHeaders(endpoint, values)...)
		var produces []string

		responses := make(map[string]*Response, len(endpoint.Responses))
		for _, r := range endpoint.Responses {
			responseGenerator := generator.At(expr.ResponseBodyExampleIdentity(endpoint, r))
			if endpoint.UsesWebSocket() {
				// A WebSocket endpoint allows at most one successful response
				// definition. So it is okay to change the first successful
				// response to a HTTP 101 response for OpenAPI docs.
				if _, ok := responses[strconv.Itoa(expr.StatusSwitchingProtocols)]; !ok {
					r = r.Dup()
					r.StatusCode = expr.StatusSwitchingProtocols
				}
			}
			resp := responseSpecFromExpr(s, root, r, endpoint.Service.Name(), schemas, responseGenerator, "", values)
			responses[strconv.Itoa(r.StatusCode)] = resp
			if r.ContentType != "" {
				foundCT := slices.Contains(produces, r.ContentType)
				if !foundCT {
					produces = append(produces, r.ContentType)
				}
			}
		}
		for _, er := range endpoint.HTTPErrors {
			responseGenerator := generator.At(expr.ErrorResponseBodyExampleIdentity(endpoint, er))
			errorDescription := values.Description(er.ErrorExpr, er.Description)
			resp := responseSpecFromExpr(s, root, er.Response, endpoint.Service.Name(), schemas, responseGenerator, errorDescription, values)
			resp.Description = er.Name + ": " + resp.Description
			if example, ok := openapi.ErrorResponseExample(er.ErrorExpr, er.Response.Body, responseGenerator, values); ok {
				resp.Examples = map[string]any{openapi.ResponseContentType(er.Response): example}
			}
			responses[strconv.Itoa(er.Response.StatusCode)] = resp
		}

		var consumes []string
		if endpoint.MultipartRequest {
			consumes = []string{"multipart/form-data"}
		}

		if endpoint.Body.Type != expr.Empty {
			in := "body"
			if endpoint.MultipartRequest {
				in = "formData"
			}
			pp := &Parameter{
				Name:        endpoint.Body.Type.Name(),
				In:          in,
				Description: values.Description(endpoint.Body.AuthoredAttribute(), endpoint.Body.Description),
				Required:    true,
				Schema: schemas.attributeTypeSchemaWithPrefix(
					root.API,
					endpoint.Body,
					codegen.Goify(endpoint.Service.Name(), true),
					generator.At(expr.RequestBodyExampleIdentity(endpoint)),
				),
			}
			params = append(params, pp)
		}

		operationID := fmt.Sprintf("%s#%s", endpoint.Service.Name(), endpoint.Name())
		index := 0
		for i, rt := range endpoint.Routes {
			if rt == route {
				index = i
				break
			}
		}
		if index > 0 {
			operationID = fmt.Sprintf("%s#%d", operationID, index)
		}

		schemes := h.Schemes()
		// remove grpc and grpcs from schemes since it is not a valid scheme in
		// openapi.
		for i := len(schemes) - 1; i >= 0; i-- {
			if schemes[i] == "grpc" || schemes[i] == "grpcs" {
				schemes = append(schemes[:i], schemes[i+1:]...)
			}
		}

		// replace HTTP with WebSocket schemes for WebSocket endpoints
		if endpoint.UsesWebSocket() {
			for i := len(schemes) - 1; i >= 0; i-- {
				if schemes[i] == "http" {
					news := append([]string{"ws"}, schemes[i+1:]...)
					schemes = append(schemes[:i], news...)
				}
				if schemes[i] == "https" {
					news := append([]string{"wss"}, schemes[i+1:]...)
					schemes = append(schemes[:i], news...)
				}
			}
		}

		description := values.Description(endpoint.MethodExpr, endpoint.Description())

		var requirements SecurityRequirements
		if len(endpoint.Requirements) > 0 {
			requirements = make(SecurityRequirements, len(endpoint.Requirements))
		}
		for i, req := range endpoint.Requirements {
			requirement := make(map[string][]string)
			for _, s := range req.Schemes {
				requirement[s.Hash()] = nil
				switch s.Kind {
				case expr.OAuth2Kind:
					if len(req.Scopes) > 0 {
						requirement[s.Hash()] = req.Scopes
					}
				case expr.BasicAuthKind, expr.APIKeyKind, expr.BearerKind, expr.JWTKind:
					lines := make([]string, 0, len(req.Scopes))
					for _, scope := range req.Scopes {
						lines = append(lines, fmt.Sprintf("  * `%s`", scope))
					}
					// List scopes only if they are defined
					if len(lines) > 0 {
						if description != "" {
							description += "\n"
						}
						description += fmt.Sprintf("\n**Required security scopes for %s**:\n%s", s.SchemeName, strings.Join(lines, "\n"))
					}
				}
			}
			requirements[i] = requirement
		}
		if expr.HasNoSecurity(endpoint.MethodExpr.Requirements) {
			requirements = SecurityRequirements{}
		}
		_, deprecated := endpoint.MethodExpr.Meta.Last("openapi:deprecated")
		operation := &Operation{
			Tags:         tagNames,
			Description:  description,
			Summary:      summaryFromExpr(endpoint.Name()+" "+endpoint.Service.Name(), endpoint, root.API.Meta),
			ExternalDocs: openapi.DocsFromExprWithValues(endpoint.MethodExpr.Docs, endpoint.MethodExpr.Meta, values),
			OperationID:  operationID,
			Parameters:   params,
			Consumes:     consumes,
			Produces:     produces,
			Responses:    responses,
			Schemes:      schemes,
			Deprecated:   deprecated,
			Extensions:   openapi.ExtensionsFromMethod(endpoint.MethodExpr),
			Security:     requirements,
		}

		if key == "" {
			key = "/"
		}
		bp := expr.HTTPWildcardRegex.ReplaceAllStringFunc(
			basePath,
			func(w string) string {
				return fmt.Sprintf("/{%s}", w[2:])
			},
		)
		if bp != "/" {
			key = strings.TrimPrefix(key, bp)
		}
		var path any
		var ok bool
		if path, ok = s.Paths[key]; !ok {
			path = new(Path)
			s.Paths[key] = path
		}
		p := path.(*Path)
		switch route.Method {
		case "GET":
			p.Get = operation
		case "PUT":
			p.Put = operation
		case "POST":
			p.Post = operation
		case "DELETE":
			p.Delete = operation
		case "OPTIONS":
			p.Options = operation
		case "HEAD":
			p.Head = operation
		case "PATCH":
			p.Patch = operation
		}
		p.Extensions = openapi.ExtensionsFromExpr(route.Endpoint.Meta)
	}
}

func initEnumValidation(def any, values []any) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Enum = values
	case *Header:
		actual.Enum = values
	case *Items:
		actual.Enum = values
	}
}

func initFormatValidation(def any, format string) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Format = format
	case *Header:
		actual.Format = format
	case *Items:
		actual.Format = format
	}
}

func initPatternValidation(def any, pattern string) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Pattern = pattern
	case *Header:
		actual.Pattern = pattern
	case *Items:
		actual.Pattern = pattern
	}
}

func initExclusiveMinimumValidation(def any, exclMin *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Minimum = exclMin
		actual.ExclusiveMinimum = true
	case *Header:
		actual.Minimum = exclMin
		actual.ExclusiveMinimum = true
	case *Items:
		actual.Minimum = exclMin
		actual.ExclusiveMinimum = true
	}
}

func initMinimumValidation(def any, minimum *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Minimum = minimum
		actual.ExclusiveMinimum = false
	case *Header:
		actual.Minimum = minimum
		actual.ExclusiveMinimum = false
	case *Items:
		actual.Minimum = minimum
		actual.ExclusiveMinimum = false
	}
}

func initExclusiveMaximumValidation(def any, exclMax *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Maximum = exclMax
		actual.ExclusiveMaximum = true
	case *Header:
		actual.Maximum = exclMax
		actual.ExclusiveMaximum = true
	case *Items:
		actual.Maximum = exclMax
		actual.ExclusiveMaximum = true
	}
}

func initMaximumValidation(def any, maximum *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Maximum = maximum
		actual.ExclusiveMaximum = false
	case *Header:
		actual.Maximum = maximum
		actual.ExclusiveMaximum = false
	case *Items:
		actual.Maximum = maximum
		actual.ExclusiveMaximum = false
	}
}

func initMinLengthValidation(def any, isArray bool, minLength *int) {
	switch actual := def.(type) {
	case *Parameter:
		if isArray {
			actual.MinItems = minLength
		} else {
			actual.MinLength = minLength
		}
	case *Header:
		actual.MinLength = minLength
	case *Items:
		actual.MinLength = minLength
	}
}

func initMaxLengthValidation(def any, isArray bool, maxLength *int) {
	switch actual := def.(type) {
	case *Parameter:
		if isArray {
			actual.MaxItems = maxLength
		} else {
			actual.MaxLength = maxLength
		}
	case *Header:
		actual.MaxLength = maxLength
	case *Items:
		actual.MaxLength = maxLength
	}
}

func initValidations(attr *expr.AttributeExpr, def any) {
	val := attr.Validation
	if val == nil {
		return
	}
	initEnumValidation(def, val.Values)
	initFormatValidation(def, string(val.Format))
	initPatternValidation(def, val.Pattern)
	if val.ExclusiveMinimum != nil {
		initExclusiveMinimumValidation(def, val.ExclusiveMinimum)
	}
	if val.Minimum != nil {
		initMinimumValidation(def, val.Minimum)
	}
	if val.ExclusiveMaximum != nil {
		initExclusiveMaximumValidation(def, val.ExclusiveMaximum)
	}
	if val.Maximum != nil {
		initMaximumValidation(def, val.Maximum)
	}
	if val.MinLength != nil {
		initMinLengthValidation(def, expr.IsArray(attr.Type), val.MinLength)
	}
	if val.MaxLength != nil {
		initMaxLengthValidation(def, expr.IsArray(attr.Type), val.MaxLength)
	}
}

// convertRefsToV2 converts all OpenAPI 3.0 references (#/$defs/) to Swagger 2.0
// format (#/definitions/) in all schemas throughout the V2 specification.
func convertRefsToV2(s *V2) {
	// Convert references in definitions
	for _, def := range s.Definitions {
		convertSchemaRefs(def)
	}
	// Convert references in paths
	for _, pathVal := range s.Paths {
		if path, ok := pathVal.(*Path); ok {
			convertPathRefs(path)
		}
	}
	// Convert references in parameters
	for _, param := range s.Parameters {
		if param.Schema != nil {
			convertSchemaRefs(param.Schema)
		}
	}
}

// convertPathRefs converts references in all operations of a path.
func convertPathRefs(path *Path) {
	ops := []*Operation{path.Get, path.Put, path.Post, path.Delete, path.Options, path.Head, path.Patch}
	for _, op := range ops {
		if op == nil {
			continue
		}
		// Convert references in parameters
		for _, param := range op.Parameters {
			if param.Schema != nil {
				convertSchemaRefs(param.Schema)
			}
		}
		// Convert references in responses
		for _, resp := range op.Responses {
			if resp.Schema != nil {
				convertSchemaRefs(resp.Schema)
			}
		}
	}
}

// convertSchemaRefs recursively converts all #/$defs/ references to #/definitions/
// in a schema and all nested schemas.
func convertSchemaRefs(schema *openapi.Schema) {
	if schema == nil {
		return
	}
	// Convert the reference itself
	if strings.HasPrefix(schema.Ref, "#/$defs/") {
		schema.Ref = strings.Replace(schema.Ref, "#/$defs/", "#/definitions/", 1)
	}
	// Convert references in items
	if schema.Items != nil {
		convertSchemaRefs(schema.Items)
	}
	// Convert references in properties
	for _, prop := range schema.Properties {
		convertSchemaRefs(prop)
	}
	// Convert references in anyOf
	for _, anyOfSchema := range schema.AnyOf {
		convertSchemaRefs(anyOfSchema)
	}
	// Convert references in additionalProperties when it's a schema
	if apSchema, ok := schema.AdditionalProperties.(*openapi.Schema); ok {
		convertSchemaRefs(apSchema)
	}
	// Convert references in $defs (though these shouldn't exist in Swagger 2.0)
	for _, def := range schema.Defs {
		convertSchemaRefs(def)
	}
}
