package http

import (
	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/services"
)

// createAccountBody is type of the "create" request body.
type createAccountBody struct {
	Name *string `json:"name,omitempty"`
}

// newCreateAccountPayload creates a CreateAccountPayload from the HTTP request
// data.
func newCreateAccountPayload(b interface{}, orgID string) (*services.CreateAccountPayload, error) {
	body := b.(*createAccountBody)
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
func (b *createAccountBody) Validate() (err error) {
	if b.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}

func newListAccountPayload(filter string) (*services.ListAccountPayload, error) {
	return &services.ListAccountPayload{Filter: filter}, nil
}

func newShowAccountPayload(id string) (*services.ShowAccountPayload, error) {
	return &services.ShowAccountPayload{ID: id}, nil
}

func newDeleteAccountPayload(id string) (*services.DeleteAccountPayload, error) {
	return &services.DeleteAccountPayload{ID: id}, nil
}
