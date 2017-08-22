package cellar

import (
	"context"
	"log"

	"goa.design/goa/examples/cellar/gen/storage"

	"github.com/boltdb/bolt"
)

// storage service example implementation.
// The example methods log the requests and return zero values.
type storagesvc struct {
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
	return &storagesvc{bolt, logger}, nil
}

// List all stored bottles
func (s *storagesvc) List(ctx context.Context) (storage.StoredBottleCollection, error) {
	var bottles []*storage.StoredBottle
	if err := s.db.LoadAll("CELLAR", &bottles); err != nil {
		return nil, err // internal error
	}
	return bottles, nil
}

// Show bottle by ID
func (s *storagesvc) Show(ctx context.Context, p *storage.ShowPayload) (*storage.StoredBottle, error) {
	var b storage.StoredBottle
	if err := s.db.Load("CELLAR", p.ID, &b); err != nil {
		if err == ErrNotFound {
			return nil, &storage.NotFound{
				Message: err.Error(),
				ID:      p.ID,
			}
		}
		return nil, err // internal error
	}
	return &b, nil
}

// Add new bottle and return its ID.
func (s *storagesvc) Add(ctx context.Context, p *storage.Bottle) (string, error) {
	id, err := s.db.NewID("CELLAR")
	if err != nil {
		return "", err // internal error
	}
	sb := storage.StoredBottle{
		ID:          id,
		Name:        p.Name,
		Winery:      p.Winery,
		Vintage:     p.Vintage,
		Composition: p.Composition,
		Description: p.Description,
		Rating:      p.Rating,
	}
	if err = s.db.Save("CELLAR", id, &sb); err != nil {
		return "", err // internal error
	}

	return id, nil
}

// Remove bottle from storage
func (s *storagesvc) Remove(ctx context.Context, p *storage.RemovePayload) error {
	return s.db.Delete("CELLAR", p.ID) // internal error if not nil
}
