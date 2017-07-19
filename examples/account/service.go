package service

import (
	"context"
	"log"
	"strconv"

	"github.com/boltdb/bolt"

	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/examples/account/gen/account/http/server"
)

// service state
type service struct {
	db     *Bolt
	logger *log.Logger
}

// New returns the account service implementation.
func New(db *bolt.DB, logger *log.Logger) (account.Service, error) {
	// Setup database
	bolt, err := NewBoltDB(db)
	if err != nil {
		return nil, err
	}

	// Build and return service implementation.
	return &service{bolt, logger}, nil
}

// Create new account
func (s *service) Create(ctx context.Context, ca *account.CreatePayload) (*account.Account, error) {
	accountID, err := s.db.NewID("ACCOUNTS")
	if err != nil {
		return nil, err
	}

	a := account.Account{
		Href:        server.ShowAccountPath(ca.OrgID, accountID),
		ID:          accountID,
		OrgID:       ca.OrgID,
		Name:        ca.Name,
		Description: ca.Description,
	}

	if err = s.db.Save("ACCOUNTS", dbID(ca.OrgID, accountID), &a); err != nil {
		return nil, err
	}

	return &a, nil
}

// List all accounts
func (s *service) List(context.Context, *account.ListPayload) ([]*account.Account, error) {
	var accounts []*account.Account
	if err := s.db.LoadAll("ACCOUNTS", &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// Show account by ID
func (s *service) Show(ctx context.Context, sp *account.ShowPayload) (*account.Account, error) {
	var a account.Account
	if err := s.db.Load("ACCOUNTS", dbID(sp.OrgID, sp.ID), &a); err != nil {
		if err == ErrNotFound {
			return nil, &account.NotFound{
				Message: err.Error(),
				ID:      sp.ID,
				OrgID:   sp.OrgID,
			}
		}
		return nil, err
	}
	return &a, nil
}

// Delete account by ID
func (s *service) Delete(ctx context.Context, dp *account.DeletePayload) error {
	return s.db.Delete("ACCOUNTS", dbID(dp.OrgID, dp.ID))
}

// dbID builds a database ID given an org and account IDs.
func dbID(orgID uint, id string) string {
	return strconv.FormatUint(uint64(orgID), 10) + ":" + id
}
