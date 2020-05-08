package openapiv3

import (
	"context"
	"fmt"

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
		// TODO: Implement me
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
