package main

import (
	"context"

	"goa.design/goa.v2/examples/account/gen/account"
)

type service struct {
}

// NewAccountService returns the account service implementation.
func NewAccountService() account.Service {
	return &service{}
}

// Create new account
func (s *service) Create(context.Context, *account.CreateAccount) (*account.Account, error) {
	return nil, nil
}

// List all accounts
func (s *service) List(context.Context, *account.ListAccount) ([]*account.Account, error) {
	return nil, nil
}

// Show account by ID
func (s *service) Show(context.Context, *account.ShowPayload) (*account.Account, error) {
	return nil, nil
}

// Delete account by IF
func (s *service) Delete(context.Context, *account.DeletePayload) error {
	return nil
}
