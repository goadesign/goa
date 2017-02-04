package services

import "context"

// Account is the "account" service interface.
type (
	Account interface {
		// Create implements the "create" endpoint.
		// The implementation should return an instance of
		// *AccountCreated or of *AccountAccepted.
		Create(context.Context, *CreateAccountPayload) (interface{}, error)
		// List implements the "list" endpoint.
		List(context.Context) ([]*Account, error)
		// Show implements the "show" endpoint.
		Show(context.Context, *ShowAccountPayload) (*Account, error)
		// Delete implements the "delete" endpoint.
		Delete(context.Context, *DeleteAccountPayload) error
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

	// AccountBody type
	AccountBody struct {
		// Href to account
		Href string
		// ID of account
		ID string
		// ID of organization that owns newly created account
		OrgID string
		// Name of new account
		Name string
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
