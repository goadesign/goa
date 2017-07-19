// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account HTTP client types
//
// Command:
// $ goa gen goa.design/goa.v2/examples/account/design

package client

import (
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
)

// CreateServerRequestBody is the type of the account create HTTP endpoint
// request body.
type CreateServerRequestBody struct {
	// Name of new account
	Name string `form:"name" json:"name" xml:"name"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
}

// CreateCreatedResponseBody is the type of the account create HTTP endpoint
// response body.
type CreateCreatedResponseBody struct {
	// ID of account
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// ID of organization that owns newly created account
	OrgID *uint `form:"org_id,omitempty" json:"org_id,omitempty" xml:"org_id,omitempty"`
	// Name of new account
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Status of account
	Status *string `form:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`
}

// ShowResponseBody is the type of the account show HTTP endpoint response body.
type ShowResponseBody struct {
	// Href to account
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// ID of account
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// ID of organization that owns newly created account
	OrgID *uint `form:"org_id,omitempty" json:"org_id,omitempty" xml:"org_id,omitempty"`
	// Name of new account
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Status of account
	Status *string `form:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`
}

// CreateNameAlreadyTakenResponseBody is the type of the account "create" HTTP
// endpoint name_already_taken error response body.
type CreateNameAlreadyTakenResponseBody struct {
	// Message of error
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
}

// ShowNotFoundResponseBody is the type of the account "show" HTTP endpoint
// not_found error response body.
type ShowNotFoundResponseBody struct {
	// Message of error
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// ID of missing account owner organization
	OrgID *uint `form:"org_id,omitempty" json:"org_id,omitempty" xml:"org_id,omitempty"`
	// ID of missing account
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
}

// AccountResponseBody is used to define fields on response body types.
type AccountResponseBody struct {
	// Href to account
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// ID of account
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// ID of organization that owns newly created account
	OrgID *uint `form:"org_id,omitempty" json:"org_id,omitempty" xml:"org_id,omitempty"`
	// Name of new account
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Status of account
	Status *string `form:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`
}

// NewCreateServerRequestBody builds the account service create endpoint
// request body from a payload.
func NewCreateServerRequestBody(p *account.CreatePayload) *CreateServerRequestBody {
	body := &CreateServerRequestBody{
		Name:        p.Name,
		Description: p.Description,
	}

	return body
}

// NewCreateAccountAccepted builds a account service create endpoint Accepted
// result.
func NewCreateAccountAccepted(href string) *account.Account {
	return &account.Account{
		Href: href,
	}
}

// NewCreateAccountCreated builds a account service create endpoint Created
// result.
func NewCreateAccountCreated(body *CreateCreatedResponseBody, href string) *account.Account {
	v := &account.Account{
		ID:          *body.ID,
		OrgID:       *body.OrgID,
		Name:        *body.Name,
		Description: body.Description,
		Status:      body.Status,
	}
	if body.Description == nil {
		tmp := "An active account"
		v.Description = &tmp
	}
	v.Href = href

	return v
}

// NewCreateNameAlreadyTaken builds a account service create endpoint
// name_already_taken error.
func NewCreateNameAlreadyTaken(body *CreateNameAlreadyTakenResponseBody) *account.NameAlreadyTaken {
	v := &account.NameAlreadyTaken{
		Message: *body.Message,
	}

	return v
}

// NewListAccountOK builds a account service list endpoint OK result.
func NewListAccountOK(body []*AccountResponseBody) []*account.Account {
	v := make([]*account.Account, len(body))
	for i, val := range body {
		v[i] = &account.Account{
			Href:        *val.Href,
			ID:          *val.ID,
			OrgID:       *val.OrgID,
			Name:        *val.Name,
			Description: val.Description,
			Status:      val.Status,
		}
		if val.Description == nil {
			tmp := "An active account"
			v[i].Description = &tmp
		}
	}

	return v
}

// NewShowAccountOK builds a account service show endpoint OK result.
func NewShowAccountOK(body *ShowResponseBody) *account.Account {
	v := &account.Account{
		Href:        *body.Href,
		ID:          *body.ID,
		OrgID:       *body.OrgID,
		Name:        *body.Name,
		Description: body.Description,
		Status:      body.Status,
	}
	if body.Description == nil {
		tmp := "An active account"
		v.Description = &tmp
	}

	return v
}

// NewShowNotFound builds a account service show endpoint not_found error.
func NewShowNotFound(body *ShowNotFoundResponseBody) *account.NotFound {
	v := &account.NotFound{
		Message: *body.Message,
		OrgID:   *body.OrgID,
		ID:      *body.ID,
	}

	return v
}

// AccountResponseBody is used to define fields on response body types.
func (body *AccountResponseBody) Validate() (err error) {
	if body.Href == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("href", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.OrgID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("org_id", "body"))
	}
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.Status != nil {
		if !(*body.Status == "provisioning" || *body.Status == "ready" || *body.Status == "deprovisioning") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.status", *body.Status, []interface{}{"provisioning", "ready", "deprovisioning"}))
		}
	}
	return
}
