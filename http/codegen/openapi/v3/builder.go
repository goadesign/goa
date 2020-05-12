package openapiv3

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"goa.design/goa/v3/expr"
)

var uriRegex = regexp.MustCompile("^https?://")

// New returns the OpenAPI v3 specification for the given API.
func New(root *expr.RootExpr) *OpenAPI {
	var (
		info     = buildInfo(root.API)
		servers  = buildServers(root.API.Servers)
		paths    = buildPaths(root.API.HTTP)
		security = buildSecurityRequirements(root.API.Requirements)
	)

	return &OpenAPI{
		OpenAPI:  "3.0.3", // TODO: This is a required string and hardcoded. Need to find some other way to set this.
		Info:     info,
		Paths:    paths,
		Servers:  servers,
		Security: security,
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

func buildSecurityRequirements(servers []*expr.SecurityExpr) []map[string][]string {
	return nil
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
