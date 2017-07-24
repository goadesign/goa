package service

import (
	"context"
	"log"

	"goa.design/goa.v2/examples/cellar/gen/storage"

	"github.com/boltdb/bolt"
)

// storage service implementation
type stgsvc struct {
	db     *Bolt
	logger *log.Logger
}

// NewStorage returns the storage service implementation.
func NewStorage(db *bolt.DB, logger *log.Logger) (storage.Service, error) {
	// Setup database
	bolt, err := NewBoltDB(db)
	if err != nil {
		return nil, err
	}

	// Build and return service implementation.
	return &stgsvc{bolt, logger}, nil
}

// Add new bottle and return its ID.
func (s *stgsvc) Add(ctx context.Context, b *storage.Bottle) (string, error) {
	id, err := s.db.NewID("CELLAR")
	if err != nil {
		return "", err // internal error
	}
	sb := storage.StoredBottle{
		ID:          id,
		Name:        b.Name,
		Winery:      b.Winery,
		Vintage:     b.Vintage,
		Composition: b.Composition,
		Description: b.Description,
		Rating:      b.Rating,
	}
	if err = s.db.Save("CELLAR", id, &sb); err != nil {
		return "", err // internal error
	}

	return id, nil
}

// List all stored bottles.
func (s *stgsvc) List(context.Context) (storage.StoredBottleCollection, error) {
	var bottles []*storage.StoredBottle
	if err := s.db.LoadAll("CELLAR", &bottles); err != nil {
		return nil, err // internal error
	}
	return bottles, nil
}

// Show bottle by ID
func (s *stgsvc) Show(ctx context.Context, sp *storage.ShowPayload) (*storage.StoredBottle, error) {
	var b storage.StoredBottle
	if err := s.db.Load("CELLAR", sp.ID, &b); err != nil {
		if err == ErrNotFound {
			return nil, &storage.NotFound{
				Message: err.Error(),
				ID:      sp.ID,
			}
		}
		return nil, err // internal error
	}
	return &b, nil
}

// Remove bottle from cellar.
func (s *stgsvc) Remove(ctx context.Context, rp *storage.RemovePayload) error {
	return s.db.Delete("CELLAR", rp.ID) // internal error if not nil
}
