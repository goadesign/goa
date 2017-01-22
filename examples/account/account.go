package basic

import (
	"context"

	"goa.design/goa.v2/examples/account/gen/app"
)

type AccountService struct {
}

func NewAccountService() app.AccountService {

}

// Create implements the "create" endpoint.
// Create may return a *app.Account or a string (href field of Account)
func (s *AccountService) Create(ctx context.Context, p *app.CreateAccountPayload) (interface{}, error) {
	return &app.Account{}, nil
}

// List implements the "list" endpoint.
func (s *AccountService) List(ctx context.Context) ([]*app.Account, error) {
	return nil, nil
}

// Show implements the "show" endpoint.
func (s *AccountService) Show(ctx context.Context, p *app.ShowAccountPayload) (*app.Account, error) {
	return &app.Account{}, nil
}
