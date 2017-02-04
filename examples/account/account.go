package basic

import (
	"context"

	"goa.design/goa.v2/examples/account/gen/app"
)

type AccountService struct {
}

func NewAccountService() *AccountService {
	return &AccountService{}
}

// Create implements the "create" endpoint.
// Create may return a *app.AccountCreated or a *app.AccountAccepted
func (s *AccountService) Create(ctx context.Context, p *app.CreateAccountPayload) (interface{}, error) {
	return &app.AccountCreated{}, nil
}

// List implements the "list" endpoint.
func (s *AccountService) List(ctx context.Context) ([]*app.Account, error) {
	return nil, nil
}

// Show implements the "show" endpoint.
func (s *AccountService) Show(ctx context.Context, p *app.ShowAccountPayload) (*app.Account, error) {
	return &app.Account{}, nil
}

// Delete implements the "delete" endpoint.
func (s *AccountService) Delete(ctx context.Context, p *app.DeleteAccountPayload) error {
	return nil
}
