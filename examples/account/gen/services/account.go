package services

import (
	"fmt"

	"golang.org/x/net/context"
)

// Account is the "account" service interface.
type (
	Account interface {
		// Create implements the "create" endpoint.
		// The implementation should return an instance of
		// *AccountCreated or of *AccountAccepted.
		// Possible errors are BadRequest, NameAlreadyToken
		Create(context.Context, *CreateAccountPayload) (interface{}, error)
		// List implements the "list" endpoint.
		List(context.Context, *ListAccountPayload) ([]*AccountBody, error)
		// Show implements the "show" endpoint.
		Show(context.Context, *ShowAccountPayload) (*AccountBody, error)
		// Delete implements the "delete" endpoint.
		Delete(context.Context, *DeleteAccountPayload) error
	}

	// AccountCreated is the type that describes the "create" endpoint HTTP
	// response with status code 201.
	AccountCreated struct {
		// Href is the value of the Location header
		Href string
		// Body describes the response body.
		Body *AccountBody
	}

	// AccountAccepted is the type that describes the "create" endpoint HTTP
	// response with status code 202.
	AccountAccepted struct {
		// Href is the value of the Location header
		Href string
	}

	// NameAlreadyTaken is the error returned when the account name is
	// already taken.
	NameAlreadyTaken struct {
		// Message of error
		Message string
	}

	// ListFilter defines an optional list filter.
	ListFilter struct {
		// Filter is the account name prefix filter.
		Filter *string
	}

	// AccountBody type
	AccountBody struct {
		// Href to account
		Href string
		// ID of account
		ID string
		// ID of organization that owns newly created account
		OrgID int
		// Name of new account
		Name string
	}

	CreateAccountPayload struct {
		OrgID int
		Name  string
	}

	ListAccountPayload struct {
		Filter string
	}

	ShowAccountPayload struct {
		ID string
	}

	DeleteAccountPayload struct {
		ID string
	}
)

func (nat *NameAlreadyTaken) Error() string {
	return fmt.Sprintf("Message: %v", nat.Message)
}
