package http

import (
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/service"
)

type (
	// CreateAccountBody is type of the "create" request body.
	CreateAccountBody struct {
		Name string `json:"name"`
	}

	// AccountCreateCreated is the type describing the StatusCreated
	// response of the account service create request.
	AccountCreateCreated struct {
		// ID is the account ID.
		ID string `json:"id"`
		// OrgID is the ID of the organization that owns the account.
		OrgID uint `json:"org_id"`
		// Name is the account name
		Name string `json:"name"`
	}

	// AccountResultBody is the type used to marshal AccountResult.
	AccountResultBody struct {
		// Href is the account href.
		Href string `json:"href"`
		// ID is the account ID.
		ID string `json:"id"`
		// OrgID is the ID of the organization that owns the account.
		OrgID uint `json:"org_id"`
		// Name is the account name
		Name string `json:"name"`
	}
)

// NewCreateAccount creates and validates a CreateAccount.
func NewCreateAccount(body *CreateAccountBody, orgID uint) (*service.CreateAccount, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	p := service.CreateAccount{
		Name:  body.Name,
		OrgID: orgID,
	}
	return &p, nil
}

// Validate runs the validations defined in the design.
func (b *CreateAccountBody) Validate() (err error) {
	if b.Name == "" {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}

// NewListAccount creates and validates a CreateAccountPayload.
func NewListAccount(orgID uint, filter *string) (*service.ListAccount, error) {
	return &service.ListAccount{OrgID: orgID, Filter: filter}, nil
}

// NewShowAccountPayload creates and validates a CreateAccountPayload.
func NewShowAccountPayload(orgID uint, id string) (*service.ShowAccountPayload, error) {
	return &service.ShowAccountPayload{OrgID: orgID, ID: id}, nil
}

// NewDeleteAccountPayload creates and validates a CreateAccountPayload.
func NewDeleteAccountPayload(orgID uint, id string) (*service.DeleteAccountPayload, error) {
	return &service.DeleteAccountPayload{OrgID: orgID, ID: id}, nil
}
