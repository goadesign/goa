package http

import (
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/services"
)

// CreateAccountBody is type of the "create" request body.
type CreateAccountBody struct {
	Name *string `json:"name,omitempty"`
}

// NewCreateAccountPayload creates and validates a CreateAccountPayload.
func NewCreateAccountPayload(body *CreateAccountBody, orgID int) (*services.CreateAccountPayload, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	p := services.CreateAccountPayload{
		Name:  *body.Name,
		OrgID: orgID,
	}
	return &p, nil
}

// Validate runs the validations defined in the design.
func (b *CreateAccountBody) Validate() (err error) {
	if b.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}

// NewListAccountPayload creates and validates a CreateAccountPayload.
func NewListAccountPayload(filter string) (*services.ListAccountPayload, error) {
	return &services.ListAccountPayload{Filter: filter}, nil
}

// NewShowAccountPayload creates and validates a CreateAccountPayload.
func NewShowAccountPayload(id string) (*services.ShowAccountPayload, error) {
	return &services.ShowAccountPayload{ID: id}, nil
}

// NewDeleteAccountPayload creates and validates a CreateAccountPayload.
func NewDeleteAccountPayload(id string) (*services.DeleteAccountPayload, error) {
	return &services.DeleteAccountPayload{ID: id}, nil
}
