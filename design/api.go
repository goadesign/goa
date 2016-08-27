package design

type (
	// APIExpr contains the global properties for a API expression.
	APIExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		*eval.DSLFunc
		// Name of API
		Name string
		// Title of API
		Title string
		// Description of API
		Description string
		// Version is the version of the API described by this DSL.
		Version string
		// Host is the default API hostname
		Host string
		// TermsOfAPI describes or links to the API terms of API
		TermsOfAPI string
		// Contact provides the API users with contact information
		Contact *ContactExpr
		// License describes the API license
		License *LicenseExpr
		// Docs points to the API external documentation
		Docs *DocsExpr
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
