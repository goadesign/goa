package openapiv3

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"goa.design/goa/v3/expr"
)

type (
	// V3 represents an instance of OpenAPI v3 swagger object.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md
	V3 struct {
		OpenAPI  string                        `json:"openapi" yaml:"openapi"`
		Info     *openapi3.Info                `json:"info" yaml:"info"`
		Servers  openapi3.Servers              `json:"servers,omitempty" yaml:"servers,omitempty"`
		Paths    openapi3.Paths                `json:"paths" yaml:"paths"`
		Security openapi3.SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	}
)

var uriRegex = regexp.MustCompile("^https?://")

// New returns the OpenAPI v3 specification for the given API.
func New(root *expr.RootExpr) (*V3, error) {
	var (
		err error

		ctx = context.Background()
	)

	var info *openapi3.Info
	{
		if info, err = buildInfo(ctx, root.API); err != nil {
			return nil, err
		}
	}

	var servers openapi3.Servers
	{
		if len(root.API.Servers) > 0 {
			if servers, err = buildServers(ctx, root.API.Servers); err != nil {
				return nil, err
			}
		}
	}

	var paths openapi3.Paths
	{
		if paths, err = buildPaths(ctx, root.API.HTTP); err != nil {
			return nil, err
		}
	}

	var security openapi3.SecurityRequirements
	{
		if security, err = buildSecurityRequirements(ctx, root.API.Requirements); err != nil {
			return nil, err
		}
	}

	return &V3{
		OpenAPI:  "3.0.3", // TODO: This is a required string and hardcoded. Need to find some other way to set this.
		Info:     info,
		Paths:    paths,
		Servers:  servers,
		Security: security,
	}, nil
}

func buildInfo(ctx context.Context, api *expr.APIExpr) (*openapi3.Info, error) {
	var info *openapi3.Info
	{
		info = &openapi3.Info{
			Title:          api.Title,
			Description:    api.Description,
			TermsOfService: api.TermsOfService,
			Version:        api.Version,
		}
		if c := api.Contact; c != nil {
			info.Contact = &openapi3.Contact{
				Name:  c.Name,
				Email: c.Email,
				URL:   c.URL,
			}
		}
		if l := api.License; l != nil {
			info.License = &openapi3.License{
				Name: l.Name,
				URL:  l.URL,
			}
		}
	}
	if err := info.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to build Info: %s", err)
	}
	return info, nil
}

func buildPaths(ctx context.Context, h *expr.HTTPExpr) (openapi3.Paths, error) {
	var paths openapi3.Paths
	{
		// TODO: Implement me
	}
	if err := paths.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to build Paths: %s", err)
	}
	return paths, nil
}

func buildServers(ctx context.Context, servers []*expr.ServerExpr) (openapi3.Servers, error) {
	var svrs openapi3.Servers
	{
		for _, svr := range servers {
			var server *openapi3.Server
			for _, host := range svr.Hosts {
				var (
					serverVariable   map[string]*openapi3.ServerVariable
					defaultValue     interface{}
					validationValues []interface{}
				)

				u, err := url.Parse(defaultURI(host))
				if err != nil {
					return nil, err
				}

				defaultValue, validationValues = paramsFromHostVariables(host.Variables)
				if defaultValue != nil {
					serverVariable = map[string]*openapi3.ServerVariable{
						host.Name: {
							Enum:        validationValues,
							Default:     defaultValue,
							Description: host.Variables.Description,
						},
					}
				}

				server = &openapi3.Server{
					URL:         u.Host,
					Description: svr.Description,
					Variables:   serverVariable,
				}
				svrs = append(svrs, server)
			}
		}
	}
	if err := svrs.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to build Servers: %s", err)
	}
	return svrs, nil
}

func buildSecurityRequirements(ctx context.Context, servers []*expr.SecurityExpr) (openapi3.SecurityRequirements, error) {
	var reqs openapi3.SecurityRequirements
	{
		// TODO: Implement me
	}
	if err := reqs.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to build SecurityRequirements: %s", err)
	}
	return reqs, nil
}

// paramsFromHostVariables returns
// - defaultValue. If empty, it substitues first value from ValidatonValues
// - validatonValues of the host if it exists
func paramsFromHostVariables(hostVariables *expr.AttributeExpr) (defaultValue interface{}, validationValues []interface{}) {
	vars := expr.AsObject(hostVariables.Type)
	for _, v := range *vars {
		defaultValue = v.Attribute.DefaultValue
		if v.Attribute.Validation != nil && len(v.Attribute.Validation.Values) > 0 {
			validationValues = append(validationValues, v.Attribute.Validation.Values...)

			if defaultValue == nil {
				defaultValue = v.Attribute.Validation.Values[0]
			}
		}
	}

	return
}

func defaultURI(h *expr.HostExpr) (uri string) {
	var uExpr expr.URIExpr

	// attempt to find the first HTTP/HTTPS URL
	for _, uExpr = range h.URIs {
		var urlStr = string(uExpr)
		if uriRegex.Match([]byte(urlStr)) {
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
