package app

import "context"

// AccountService is the "account" service interface.
type (
	AccountService interface {
		// Create implements the "create" endpoint.
		// The implementation should return an instance of
		// *CreateAccountCreated or of *CreateAccountAccepted.
		Create(context.Context, *CreateAccountPayload) (interface{}, error)
		// List implements the "list" endpoint.
		List(context.Context) ([]*Account, error)
		// Show implements the "show" endpoint.
		Show(context.Context, *ShowAccountPayload) (*Account, error)
		// Delete implements the "delete" endpoint.
		Delete(context.Context, *DeleteAccountPayload) error
	}

	// Account type
	Account struct {
		// Href to account
		Href string
		// ID of account
		ID string
		// ID of organization that owns newly created account
		OrgID string
		// Name of new account
		Name string
	}

	// AccountCreated is the type that describes the "create" endpoint HTTP
	// response with status code 201.
	AccountCreated struct {
		// Href is the value of the Location header
		Href string
		// Body describes the response body.
		Body *Account
	}

	// AccountAccepted is the type that describes the "create" endpoint HTTP
	// response with status code 202.
	AccountAccepted struct {
		// Href is the value of the Location header
		Href string
	}

	CreateAccountPayload struct {
		OrgID string
		Name  string
	}

	ShowAccountPayload struct {
		ID string
	}

	DeleteAccountPayload struct {
		ID string
	}
)
