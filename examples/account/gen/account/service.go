// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account service
//
// Command:
// $ goa server goa.design/goa.v2/examples/account/design

package account

import "context"

// Manage accounts
type Service interface {
	// Create new account
	Create(context.Context, *CreateAccount) (*Account, error)
	// List all accounts
	List(context.Context, *ListAccount) ([]*Account, error)
	// Show account by ID
	Show(context.Context, *ShowPayload) (*Account, error)
	// Delete account by IF
	Delete(context.Context, *DeletePayload) error
}

// CreateAccount is the payload type of the account service create method.
type CreateAccount struct {
	// ID of organization that owns newly created account
	OrgID uint
	// Name of new account
	Name string
	// Description of new account
	Description *string
}

// Account is the result type of the account service create method.
type Account struct {
	// Href to account
	Href string
	// ID of account
	ID string
	// ID of organization that owns newly created account
	OrgID uint
	// Name of new account
	Name string
	// Description of new account
	Description *string
}

// ListAccount is the payload type of the account service list method.
type ListAccount struct {
	// ID of organization that owns newly created account
	OrgID *uint
	// Filter is the account name prefix filter
	Filter *string
}

// ShowPayload is the payload type of the account service show method.
type ShowPayload struct {
	// ID of organization that owns  account
	OrgID uint
	// ID of account to show
	ID string
}

// DeletePayload is the payload type of the account service delete method.
type DeletePayload struct {
	// ID of organization that owns  account
	OrgID uint
	// ID of account to show
	ID string
}

// NameAlreadyTaken is the type returned when creating an account fails because
// its name is already taken
type NameAlreadyTaken struct {
	// Message of error
	Message string
}

// NotFound is the type returned when attempting to show or delete an account
// that does not exist.
type NotFound struct {
	// Message of error
	Message string
	// ID of missing account owner organization
	OrgID uint
	// ID of missing account
	ID string
}

// NameAlreadyTaken is the type returned when creating an account fails because
// its name is already taken
func (e *NameAlreadyTaken) Error() string {
	return "NameAlreadyTaken"
}

// NotFound is the type returned when attempting to show or delete an account
// that does not exist.
func (e *NotFound) Error() string {
	return "NotFound"
}
