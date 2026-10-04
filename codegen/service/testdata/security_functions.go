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
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		SecureWithRequirements:       NewSecureWithRequirementsEndpoint(s, a.BasicAuth),
		DoublySecureWithRequirements: NewDoublySecureWithRequirementsEndpoint(s, a.BasicAuth, a.JWTAuth),
	}
}
`

var EndpointInitWithServiceRequirementsCode = `// NewEndpoints wraps the methods of the "EndpointsWithServiceRequirements"
// service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	// Casting service to Auther interface
	a := s.(Auther)
	return &Endpoints{
		SecureWithRequirements:     NewSecureWithRequirementsEndpoint(s, a.BasicAuth),
		AlsoSecureWithRequirements: NewAlsoSecureWithRequirementsEndpoint(s, a.BasicAuth),
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
