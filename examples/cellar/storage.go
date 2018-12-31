package cellar

import (
	"context"
	"fmt"
	"log"
	"strings"

	storage "goa.design/goa/examples/cellar/gen/storage"

	"github.com/boltdb/bolt"
)

// storage service example implementation.
// The example methods log the requests and return zero values.
type storageSvc struct {
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
	return &storageSvc{bolt, logger}, nil
}

// List all stored bottles
func (s *storageSvc) List(ctx context.Context) (res storage.StoredBottleCollection, err error) {
	if err = s.db.LoadAll("CELLAR", &res); err != nil {
		return nil, err // internal error
	}
	return res, nil
}

// Show bottle by ID
func (s *storageSvc) Show(ctx context.Context, p *storage.ShowPayload) (res *storage.StoredBottle, view string, err error) {
	if p.View != nil {
		view = *p.View
	} else {
		view = "default"
	}
	if err = s.db.Load("CELLAR", p.ID, &res); err != nil {
		if err == ErrNotFound {
			return nil, view, &storage.NotFound{
				Message: err.Error(),
				ID:      p.ID,
			}
		}
		return nil, view, err // internal error
	}
	return res, view, nil
}

// Add new bottle and return its ID.
func (s *storageSvc) Add(ctx context.Context, p *storage.Bottle) (res string, err error) {
	res, err = s.db.NewID("CELLAR")
	if err != nil {
		return "", err // internal error
	}
	sb := storage.StoredBottle{
		ID:          res,
		Name:        p.Name,
		Winery:      p.Winery,
		Vintage:     p.Vintage,
		Composition: p.Composition,
		Description: p.Description,
		Rating:      p.Rating,
	}
	if err = s.db.Save("CELLAR", res, &sb); err != nil {
		return "", err // internal error
	}
	return res, nil
}

// Remove bottle from storage
func (s *storageSvc) Remove(ctx context.Context, p *storage.RemovePayload) (err error) {
	return s.db.Delete("CELLAR", p.ID) // internal error if not nil
}

// Rate bottles by IDs
func (s *storageSvc) Rate(ctx context.Context, p map[uint32][]string) (err error) {
	for rating, ids := range p {
		for _, id := range ids {
			var b storage.StoredBottle
			if err = s.db.Load("CELLAR", id, &b); err != nil {
				if err == ErrNotFound {
					continue
				}
			}
			sb := storage.StoredBottle{
				ID:          id,
				Name:        b.Name,
				Winery:      b.Winery,
				Vintage:     b.Vintage,
				Composition: b.Composition,
				Description: b.Description,
				Rating:      &rating,
			}
			if err = s.db.Save("CELLAR", id, &sb); err != nil {
				return err // internal error
			}
		}
	}
	return nil
}

// Add n number of bottles and return their IDs. This is a multipart request
// and each part has field name 'bottle' and contains the encoded bottle info
// to be added.
func (s *storageSvc) MultiAdd(ctx context.Context, p []*storage.Bottle) (res []string, err error) {
	res = make([]string, 0, len(p))
	for _, bottle := range p {
		id, err := s.db.NewID("CELLAR")
		if err != nil {
			return nil, err // internal error
		}
		sb := storage.StoredBottle{
			ID:          id,
			Name:        bottle.Name,
			Winery:      bottle.Winery,
			Vintage:     bottle.Vintage,
			Composition: bottle.Composition,
			Description: bottle.Description,
			Rating:      bottle.Rating,
		}
		if err = s.db.Save("CELLAR", id, &sb); err != nil {
			return nil, err // internal error
		}
		res = append(res, id)
	}
	return res, nil
}

// Update bottles with the given IDs. This is a multipart request and each part
// has field name 'bottle' and contains the encoded bottle info to be updated.
// The IDs in the query parameter is mapped to each part in the request.
func (s *storageSvc) MultiUpdate(ctx context.Context, p *storage.MultiUpdatePayload) error {
	for _, id := range p.Ids {
		for _, bottle := range p.Bottles {
			sb := storage.StoredBottle{
				ID:          id,
				Name:        bottle.Name,
				Winery:      bottle.Winery,
				Vintage:     bottle.Vintage,
				Composition: bottle.Composition,
				Description: bottle.Description,
				Rating:      bottle.Rating,
			}
			if err := s.db.Save("CELLAR", id, &sb); err != nil {
				return err // internal error
			}
		}
	}
	s.logger.Print(fmt.Sprintf("Updated bottles: %s", strings.Join(p.Ids, ", ")))
	return nil
}
