package testdata

var EndpointInitWithoutRequirementCode = `// NewEndpoints wraps the methods of the "EndpointWithoutRequirement" service
// with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		Unsecure: NewUnsecureEndpoint(s),
	}
}
`

var EndpointInitWithRequirementsCode = `// NewEndpoints wraps the methods of the "EndpointsWithRequirements" service
// with endpoints.
func NewEndpoints(s Service, authBasicFn security.AuthBasicFunc, authJWTFn security.AuthJWTFunc) *Endpoints {
	return &Endpoints{
		SecureWithRequirements:       NewSecureWithRequirementsEndpoint(s, authBasicFn),
		DoublySecureWithRequirements: NewDoublySecureWithRequirementsEndpoint(s, authBasicFn, authJWTFn),
	}
}
`

var EndpointInitWithServiceRequirementsCode = `// NewEndpoints wraps the methods of the "EndpointsWithServiceRequirements"
// service with endpoints.
func NewEndpoints(s Service, authBasicFn security.AuthBasicFunc) *Endpoints {
	return &Endpoints{
		SecureWithRequirements:     NewSecureWithRequirementsEndpoint(s, authBasicFn),
		AlsoSecureWithRequirements: NewAlsoSecureWithRequirementsEndpoint(s, authBasicFn),
	}
}
`

var EndpointInitNoSecurityCode = `// NewEndpoints wraps the methods of the "EndpointNoSecurity" service with
// endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		NoSecurity: NewNoSecurityEndpoint(s),
	}
}
`

var EndpointWithRequiredScopesCode = `// NewSecureWithRequiredScopesEndpoint returns an endpoint function that calls
// the method "SecureWithRequiredScopes" of service
// "EndpointWithRequiredScopes".
func NewSecureWithRequiredScopesEndpoint(s Service, authJWTFn security.AuthJWTFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*SecureWithRequiredScopesPayload)
		var err error
		sc := security.JWTScheme{
			Name:           "jwt",
			Scopes:         []string{"api:read", "api:write", "api:admin"},
			RequiredScopes: []string{"api:read", "api:write"},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authJWTFn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.SecureWithRequiredScopes(ctx, p)
	}
}
`

var EndpointWithAPIKeyOverrideCode = `// NewSecureWithAPIKeyOverrideEndpoint returns an endpoint function that calls
// the method "SecureWithAPIKeyOverride" of service
// "EndpointWithAPIKeyOverride".
func NewSecureWithAPIKeyOverrideEndpoint(s Service, authAPIKeyFn security.AuthAPIKeyFunc) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*SecureWithAPIKeyOverridePayload)
		var err error
		sc := security.APIKeyScheme{
			Name: "api_key",
		}
		var key string
		if p.Key != nil {
			key = *p.Key
		}
		ctx, err = authAPIKeyFn(ctx, key, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.SecureWithAPIKeyOverride(ctx, p)
	}
}
`

var EndpointWithOAuth2Code = `// NewSecureWithOAuth2Endpoint returns an endpoint function that calls the
// method "SecureWithOAuth2" of service "EndpointWithOAuth2".
func NewSecureWithOAuth2Endpoint(s Service, authOAuth2Fn security.AuthOAuth2Func) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*SecureWithOAuth2Payload)
		var err error
		sc := security.OAuth2Scheme{
			Name:           "authCode",
			Scopes:         []string{"api:write", "api:read"},
			RequiredScopes: []string{},
			Flows: []*security.OAuthFlow{
				&security.OAuthFlow{
					Type:             "authorization_code",
					AuthorizationURL: "/authorization",
					TokenURL:         "/token",
					RefreshURL:       "/refresh",
				},
			},
		}
		var token string
		if p.Token != nil {
			token = *p.Token
		}
		ctx, err = authOAuth2Fn(ctx, token, &sc)
		if err != nil {
			return nil, err
		}
		return nil, s.SecureWithOAuth2(ctx, p)
	}
}
`
