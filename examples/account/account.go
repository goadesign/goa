package basic

import (
	"context"

	"goa.design/goa.v2/examples/account/gen/service"
)

// AccountService implements the account service.
type AccountService struct {
}

// NewAccountService creates a account service.
func NewAccountService() *AccountService {
	return &AccountService{}
}

// Create implements the "create" endpoint.
// Create may return a *app.AccountCreated or a *app.AccountAccepted
func (s *AccountService) Create(ctx context.Context, p *service.CreateAccount) (*service.AccountResult, error) {
	return &service.AccountResult{}, nil
}

// List implements the "list" endpoint.
func (s *AccountService) List(ctx context.Context, filter *service.ListAccount) ([]*service.AccountResult, error) {
	return nil, nil
}

// Show implements the "show" endpoint.
func (s *AccountService) Show(ctx context.Context, p *service.ShowAccountPayload) (*service.AccountResult, error) {
	return &service.AccountResult{}, nil
}

// Delete implements the "delete" endpoint.
func (s *AccountService) Delete(ctx context.Context, p *service.DeleteAccountPayload) error {
	return nil
}
