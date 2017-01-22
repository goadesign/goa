package app

import "context"

// AccountService is the "account" service interface.
type AccountService interface {
	// Create implements the "create" endpoint.
	Create(context.Context, *CreateAccountPayload) (interface{}, error)
	// List implements the "list" endpoint.
	List(context.Context) ([]*Account, error)
	// Show implements the "show" endpoint.
	Show(context.Context, *ShowAccountPayload) (*Account, error)
}
