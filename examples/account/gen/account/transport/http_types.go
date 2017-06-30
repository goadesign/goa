// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account HTTP transport types
//
// Command:
// $ goa server goa.design/goa.v2/examples/account/design

package transport

import (
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/account"
)

// CreateServerRequestBody is the type of the account "create" HTTP endpoint
// request body.
type CreateServerRequestBody struct {
	// Name of new account
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
}

// CreateCreatedResponseBody is the type of the account "create" HTTP endpoint
// Created response body.
type CreateCreatedResponseBody struct {
	// ID of account
	ID string `form:"id" json:"id" xml:"id"`
	// ID of organization that owns newly created account
	OrgID uint `form:"org_id" json:"org_id" xml:"org_id"`
	// Name of new account
	Name string `form:"name" json:"name" xml:"name"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
}

// Account type
type Account struct {
	// Href to account
	Href string `form:"href" json:"href" xml:"href"`
	// ID of account
	ID string `form:"id" json:"id" xml:"id"`
	// ID of organization that owns newly created account
	OrgID uint `form:"org_id" json:"org_id" xml:"org_id"`
	// Name of new account
	Name string `form:"name" json:"name" xml:"name"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
}

// NewCreateAccount instantiates and validates the account service create
// endpoint payload.
func NewCreateAccount(body *CreateServerRequestBody, orgID uint) *account.CreateAccount {
	p := account.CreateAccount{
		Name:        *body.Name,
		Description: body.Description,
		OrgID:       orgID,
	}
	return &p
}

// NewListAccount instantiates and validates the account service list endpoint
// payload.
func NewListAccount(filter *string, orgID uint) *account.ListAccount {
	p := account.ListAccount{
		Filter: filter,
		OrgID:  &orgID,
	}
	return &p
}

// NewShowPayload instantiates and validates the account service show endpoint
// payload.
func NewShowPayload(orgID uint, id string) *account.ShowPayload {
	p := account.ShowPayload{
		OrgID: &orgID,
		ID:    &id,
	}
	return &p
}

// NewDeletePayload instantiates and validates the account service delete
// endpoint payload.
func NewDeletePayload(orgID uint, id string) *account.DeletePayload {
	p := account.DeletePayload{
		OrgID: &orgID,
		ID:    &id,
	}
	return &p
}

// CreateServerRequestBody is the type of the account "create" HTTP endpoint
// request body.
func (body *CreateServerRequestBody) Validate() (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("body", "name"))
	}
	return
}
