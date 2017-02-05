package basic

import (
	"context"

	"goa.design/goa.v2/examples/account/gen/services"
)

type AccountService struct {
}

func NewAccountService() *AccountService {
	return &AccountService{}
}

// Create implements the "create" endpoint.
// Create may return a *app.AccountCreated or a *app.AccountAccepted
func (s *AccountService) Create(ctx context.Context, p *services.CreateAccountPayload) (interface{}, error) {
	return &services.AccountCreated{}, nil
}

// List implements the "list" endpoint.
func (s *AccountService) List(ctx context.Context) ([]*services.AccountResponse, error) {
	return nil, nil
}

// Show implements the "show" endpoint.
func (s *AccountService) Show(ctx context.Context, p *services.ShowAccountPayload) (*services.AccountResponse, error) {
	return &services.AccountResponse{}, nil
}

// Delete implements the "delete" endpoint.
func (s *AccountService) Delete(ctx context.Context, p *services.DeleteAccountPayload) error {
	return nil
}
