package openapiv3

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// AnnotatedSchema makes it possible to annotate a JSON schema so that the
// OpenAPI code generator may provide additional information in descriptions.
// This is used for views and for streaming requests and responses.
type AnnotatedSchema struct {
	*openapi.Schema
	Note string
}

// EndpointBodies describes the request and response HTTP bodies of an endpoint
// using JSON schema. Each body may be described via a reference to a schema
// described in the "Components" section of the OpenAPI document or an actual
// JSON schema data structure. There may also be additional notes attached to
// each body definition to account for cases that are not directly supported in
// OpenAPI such as result types with multiple views or streaming. The possible
// response bodies are indexed by HTTP status.
type EndpointBodies struct {
	RequestBody    *AnnotatedSchema
	ResponseBodies map[int]*AnnotatedSchema
}

var uriRegex = regexp.MustCompile("^https?://")

// New returns the OpenAPI v3 specification for the given API.
// It returns nil if the design does not define HTTP endpoints.
func New(root *expr.RootExpr) *OpenAPI {
	if root == nil || root.API == nil || root.API.HTTP == nil || len(root.API.HTTP.Services) == 0 {
		// No HTTP transport
		return nil
	}

	var (
		info     = buildInfo(root.API)
		comps    = buildComponents(root)
		servers  = buildServers(root.API.Servers)
		paths    = buildPaths(root.API.HTTP)
		security = buildSecurityRequirements(root)
	)

	return &OpenAPI{
		OpenAPI:    "3.0.3", // TODO: This is a required string and hardcoded. Need to find some other way to set this.
		Info:       info,
		Components: comps,
		Paths:      paths,
		Servers:    servers,
		Security:   security,
	}
}

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

func buildComponents(root *expr.RootExpr) *Components {
	schemesRef := make(map[string]*SecuritySchemeRef, len(root.Schemes))
	for _, se := range root.Schemes {
		schemesRef[se.SchemeName] = &SecuritySchemeRef{
			Value: buildSecurityScheme(se),
		}
	}
	return &Components{
		SecuritySchemes: schemesRef,
	}
}

func buildPaths(h *expr.HTTPExpr) map[string]*PathItem {
	return nil
}

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

func buildSecurityRequirements(root *expr.RootExpr) []map[string][]string {
	var srs []map[string][]string
	for _, sch := range root.Schemes {
		sr := make(map[string][]string)
		switch sch.Kind {
		case expr.BasicAuthKind, expr.APIKeyKind:
			sr[sch.SchemeName] = []string{}
		case expr.OAuth2Kind, expr.JWTKind:
			scopes := make([]string, len(sch.Scopes))
			for i, scope := range sch.Scopes {
				scopes[i] = scope.Name
			}
			sr[sch.SchemeName] = scopes
		}
		srs = append(srs, sr)
	}
	return srs
}

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
func defaultURI(h *expr.HostExpr) (uri string) {
	var uExpr expr.URIExpr

	// attempt to find the first HTTP/HTTPS URL
	for _, uExpr = range h.URIs {
		var urlStr = string(uExpr)
		if uriRegex.MatchString(urlStr) {
			uri = urlStr
			break
		}
	}

	// if uri is empty i.e there were no URLs
	// starting with http/https, then pick the first URL
	if uri == "" && len(h.URIs) > 0 {
		uExpr = h.URIs[0]
		uri = string(uExpr)
	}

	vars := expr.AsObject(h.Variables.Type)
	if len(*vars) == 0 {
		return
	}

	// substitute any URI parameters with their
	// default values or first item in their enum
	for _, p := range uExpr.Params() {
		for _, v := range *vars {
			if p == v.Name {
				def := v.Attribute.DefaultValue
				if def == nil {
					def = v.Attribute.Validation.Values[0]
				}
				uri = strings.Replace(uri, fmt.Sprintf("{%s}", p), fmt.Sprintf("%v", def), -1)
			}
		}
	}
	return
}

// buildBodyTypes traverses the design and builds the JSON schemas that
// represent the request and response bodies of each endpoint. The algorithm
// also computes a good unique name for the different types making sure that two
// types that are actually identical share the same name. This is to handle
// properly the data structures created by the code generation algorithms which
// can duplicate types (for example if they are defined inline in the design).
// The result is a map of method details indexed by service name. Each method
// detail is in turn indexed by method name. The details contain JSON schema
// references and the actual JSON schemas are returned in the second result
// value indexed by reference.
func buildBodyTypes(api *expr.APIExpr) (map[string]map[string]*EndpointBodies, map[string]*AnnotatedSchema) {
	bodies := make(map[string]map[string]*EndpointBodies)
	schemas := make(map[string]*AnnotatedSchema)
	for _, s := range api.HTTP.Services {
		errors := make(map[int]*AnnotatedSchema)
		for _, e := range s.HTTPErrors {
			errors[e.Response.StatusCode] = schemafy(e.Response.Body, schemas)
		}
		sbodies := make(map[string]*EndpointBodies, len(s.HTTPEndpoints))
		for _, e := range s.HTTPEndpoints {
			req := schemafy(e.Body, schemas)
			if e.StreamingBody != nil {
				sreq := schemafy(e.StreamingBody, schemas)
				var note string
				if sreq.Schema.Ref != "" {
					note = fmt.Sprintf("Streaming body ref: #/components/schemas/%s", sreq.Schema.Ref)
				} else {
					note = fmt.Sprintf("Streaming body: %s", sreq.Schema.Type)
				}
				req.Note = note
			}
			res := make(map[int]*AnnotatedSchema)
			for c, er := range errors {
				res[c] = er
			}
			for _, resp := range e.Responses {
				res[resp.StatusCode] = schemafy(resp.Body, schemas)
			}
			sbodies[e.Name()] = &EndpointBodies{req, res}
		}
		bodies[s.Name()] = sbodies
	}
	return bodies, schemas
}

func schemafy(attr *expr.AttributeExpr, schemas map[string]*AnnotatedSchema) *AnnotatedSchema {
	s := openapi.NewSchema()
	as := &AnnotatedSchema{Schema: s}
	var note string
	switch t := attr.Type.(type) {
	case expr.Primitive:
		switch t.Kind() {
		case expr.UIntKind, expr.UInt64Kind, expr.UInt32Kind:
			s.Type = openapi.Type("integer")
		case expr.IntKind, expr.Int64Kind:
			s.Type = openapi.Type("integer")
			s.Format = "int64"
		case expr.Int32Kind:
			s.Type = openapi.Type("integer")
			s.Format = "int32"
		case expr.Float32Kind:
			s.Type = openapi.Type("number")
			s.Format = "float"
		case expr.Float64Kind:
			s.Type = openapi.Type("number")
			s.Format = "double"
		case expr.BytesKind, expr.AnyKind:
			s.Type = openapi.Type("string")
			s.Format = "binary"
		default:
			s.Type = openapi.Type(t.Name())
		}
	case *expr.Array:
		s.Type = openapi.Array
		es := schemafy(t.ElemType, schemas)
		s.Items = es.Schema
		if es.Note != "" {
			note = "items: " + es.Note
		}
	case *expr.Object:
		s.Type = openapi.Object
		var itemNotes []string
		for _, nat := range *t {
			prop := schemafy(nat.Attribute, schemas)
			s.Properties[nat.Name] = prop.Schema
			if prop.Note != "" {
				itemNotes = append(itemNotes, nat.Name+": "+prop.Note)
			}
		}
		if len(itemNotes) > 0 {
			note = strings.Join(itemNotes, "\n")
		}
	case *expr.Map:
		s.Type = openapi.Object
		s.AdditionalProperties = true
	case *expr.UserTypeExpr:
		// s.Ref = TypeRefWithPrefix(api, t, prefix)
	case *expr.ResultTypeExpr:
		// Use "default" view by default
		// s.Ref = ResultTypeRefWithPrefix(api, t, expr.DefaultView, prefix)
	default:
		panic(fmt.Sprintf("unknown type %T", t)) // bug
	}
	s.Description = attr.Description
	if note != "" {
		s.Description += "\n" + as.Note
	}
	return as
}
