package app

import goa "goa.design/goa.v2"

// createAccountBody is type of the "create" request body.
type createAccountBody struct {
	OrgID *string `json:"org_id,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// newCreateAccountPayload creates a CreateAccountPayload from the HTTP request data.
func newCreateAccountPayload(body *createAccountBody) (*CreateAccountPayload, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	p := CreateAccountPayload{
		Name:  *body.Name,
		OrgID: *body.OrgID,
	}
	return &p, nil
}

// Validate runs the validations defined in the design.
func (b *createAccountBody) Validate() (err error) {
	if b.OrgID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("org_id", "body"))
	}
	if b.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}

func newShowAccountPayload(id string) (*ShowAccountPayload, error) {
	return &ShowAccountPayload{ID: id}, nil
}

func newDeleteAccountPayload(id string) (*DeleteAccountPayload, error) {
	return &DeleteAccountPayload{ID: id}, nil
}
