package service

import (
	"fmt"

	"context"
)

// Account is the "account" service interface.
type (
	Account interface {
		// Create implements the "create" endpoint.
		// Possible errors are BadRequest, NameAlreadyToken
		Create(context.Context, *CreateAccount) (*AccountResult, error)
		// List implements the "list" endpoint.
		List(context.Context, *ListAccount) ([]*AccountResult, error)
		// Show implements the "show" endpoint.
		Show(context.Context, *ShowAccountPayload) (*AccountResult, error)
		// Delete implements the "delete" endpoint.
		Delete(context.Context, *DeleteAccountPayload) error
	}

	// AccountResult type
	AccountResult struct {
		// Href to account
		Href string
		// ID of account
		ID string
		// ID of organization that owns account
		OrgID uint
		// Name of account
		Name string
		// Description of account
		Description *string
	}

	// NameAlreadyTaken is the error returned when the account name is
	// already taken.
	NameAlreadyTaken struct {
		// Message of error
		Message string
	}

	// ListAccount defines an optional list filter.
	ListAccount struct {
		// ID of organization that owns account
		OrgID uint
		// Filter is the account name prefix filter.
		Filter *string
	}

	// CreateAccount is the account creation payload.
	CreateAccount struct {
		OrgID       uint
		Name        string
		Description string
	}

	ShowAccountPayload struct {
		// ID of organization that owns account
		OrgID uint
		// ID of account to show
		ID string
	}

	DeleteAccountPayload struct {
		// ID of organization that owns account
		OrgID uint
		// ID of account to delete
		ID string
	}
)

func (nat *NameAlreadyTaken) Error() string {
	return fmt.Sprintf("Message: %v", nat.Message)
}
