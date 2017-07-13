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

// CreateServerRequestBody is the type of the account create HTTP endpoint
// request body.
type CreateServerRequestBody struct {
	// Name of new account
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Description of new account
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
}

// CreateCreatedResponseBody is the type of the account create HTTP endpoint
// response body.
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

// ShowResponseBody is the type of the account show HTTP endpoint response body.
type ShowResponseBody struct {
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

// CreateNameAlreadyTakenResponseBody is the type of the account "create" HTTP
// endpoint name_already_taken Conflict error response body.
type CreateNameAlreadyTakenResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
}

// ShowNotFoundResponseBody is the type of the account "show" HTTP endpoint
// not_found Not Found error response body.
type ShowNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// ID of missing account owner organization
	OrgID uint `form:"org_id" json:"org_id" xml:"org_id"`
	// ID of missing account
	ID string `form:"id" json:"id" xml:"id"`
}

// Account type
type AccountResponseBody struct {
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

// NewCreateServerRequestBody builds the account service create endpoint
// request body from a payload.
func NewCreateServerRequestBody(p *account.CreatePayload) *CreateServerRequestBody {
	body := &CreateServerRequestBody{
		Name:        &p.Name,
		Description: p.Description,
	}

	return body
}

// NewCreateCreatedResponseBody builds the account service create endpoint
// response body from a result.
func NewCreateCreatedResponseBody(res *account.Account) *CreateCreatedResponseBody {
	body := &CreateCreatedResponseBody{
		ID:          res.ID,
		OrgID:       res.OrgID,
		Name:        res.Name,
		Description: res.Description,
	}

	return body
}

// NewAccountResponseBody builds the account service list endpoint response
// body from a result.
func NewAccountResponseBody(res []*account.Account) []*AccountResponseBody {
	body := make([]*AccountResponseBody, len(res))
	for i, val := range res {
		body[i] = &AccountResponseBody{
			Href:        val.Href,
			ID:          val.ID,
			OrgID:       val.OrgID,
			Name:        val.Name,
			Description: val.Description,
		}
	}

	return body
}

// NewShowResponseBody builds the account service show endpoint response body
// from a result.
func NewShowResponseBody(res *account.Account) *ShowResponseBody {
	body := &ShowResponseBody{
		Href:        res.Href,
		ID:          res.ID,
		OrgID:       res.OrgID,
		Name:        res.Name,
		Description: res.Description,
	}

	return body
}

// NewCreateNameAlreadyTakenResponseBody builds the account service create
// endpoint response body from a result.
func NewCreateNameAlreadyTakenResponseBody(res *account.NameAlreadyTaken) *CreateNameAlreadyTakenResponseBody {
	body := &CreateNameAlreadyTakenResponseBody{
		Message: res.Message,
	}

	return body
}

// NewShowNotFoundResponseBody builds the account service show endpoint
// response body from a result.
func NewShowNotFoundResponseBody(res *account.NotFound) *ShowNotFoundResponseBody {
	body := &ShowNotFoundResponseBody{
		Message: res.Message,
		OrgID:   res.OrgID,
		ID:      res.ID,
	}

	return body
}

// NewCreateCreatePayload builds a account service create endpoint payload.
func NewCreateCreatePayload(body *CreateServerRequestBody, orgID uint) *account.CreatePayload {
	v := &account.CreatePayload{
		Name:        *body.Name,
		Description: body.Description,
	}
	v.OrgID = orgID

	return v
}

// NewCreateAccount builds a account service create endpoint result.
func NewCreateAccount(body *CreateCreatedResponseBody, href string) *account.Account {
	v := &account.Account{
		ID:          body.ID,
		OrgID:       body.OrgID,
		Name:        body.Name,
		Description: body.Description,
	}
	if body.Description == nil {
		tmp := "An active account"
		v.Description = &tmp
	}
	v.Href = href

	return v
}

// NewListListPayload builds a account service list endpoint payload.
func NewListListPayload(orgID uint, filter *string) *account.ListPayload {
	return &account.ListPayload{
		OrgID:  &orgID,
		Filter: filter,
	}
}

// NewListAccount builds a account service list endpoint result.
func NewListAccount(body []*AccountResponseBody) []*account.Account {
	v := make([]*account.Account, len(body))
	for i, val := range body {
		v[i] = &account.Account{
			Href:        val.Href,
			ID:          val.ID,
			OrgID:       val.OrgID,
			Name:        val.Name,
			Description: val.Description,
		}
		if val.Description == nil {
			tmp := "An active account"
			v[i].Description = &tmp
		}
	}

	return v
}

// NewShowShowPayload builds a account service show endpoint payload.
func NewShowShowPayload(orgID uint, id string) *account.ShowPayload {
	return &account.ShowPayload{
		OrgID: orgID,
		ID:    id,
	}
}

// NewShowAccount builds a account service show endpoint result.
func NewShowAccount(body *ShowResponseBody) *account.Account {
	v := &account.Account{
		Href:        body.Href,
		ID:          body.ID,
		OrgID:       body.OrgID,
		Name:        body.Name,
		Description: body.Description,
	}
	if body.Description == nil {
		tmp := "An active account"
		v.Description = &tmp
	}

	return v
}

// NewDeleteDeletePayload builds a account service delete endpoint payload.
func NewDeleteDeletePayload(orgID uint, id string) *account.DeletePayload {
	return &account.DeletePayload{
		OrgID: orgID,
		ID:    id,
	}
}

// CreateServerRequestBody is the type of the account create HTTP endpoint
// request body.
func (body *CreateServerRequestBody) Validate() (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}
