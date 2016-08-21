package design

type (
	// ServiceExpr contains the global properties for a service expression.
	ServiceExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		*eval.DSLFunc
		// Name of Service
		Name string
		// Title of Service
		Title string
		// Description of Service
		Description string
		// Version is the version of the Service described by this DSL.
		Version string
		// Host is the default Service hostname
		Host string
		// TermsOfService describes or links to the Service terms of service
		TermsOfService string
		// Contact provides the Service users with contact information
		Contact *ContactExpr
		// License describes the Service license
		License *LicenseExpr
		// Docs points to the Service external documentation
		Docs *DocsExpr
		// EndpointGroups lists the endpoint groups exposed by the service.
		EndpointGroups []*EndpointGroupExpr
		// Metadata is a list of key/value pairs
		Metadata *MetadataExpr
	}

	// ContactExpr contains the API contact information.
	ContactExpr struct {
		// Name of the contact person/organization
		Name string `json:"name,omitempty"`
		// Email address of the contact person/organization
		Email string `json:"email,omitempty"`
		// URL pointing to the contact information
		URL string `json:"url,omitempty"`
	}

	// LicenseExpr contains the license information for the API.
	LicenseExpr struct {
		// Name of license used for the API
		Name string `json:"name,omitempty"`
		// URL to the license used for the API
		URL string `json:"url,omitempty"`
	}

	// DocsExpr points to external documentation.
	DocsExpr struct {
		// Description of documentation.
		Description string `json:"description,omitempty"`
		// URL to documentation.
		URL string `json:"url,omitempty"`
	}
)
