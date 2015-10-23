package main

import "github.com/raphael/goa/examples/cellar/app"

// BottleController implements the bottle resource.
type BottleController struct {
	db *DB
}

// NewBottleController creates a bottle controller.
func NewBottleController() *BottleController {
	return &BottleController{db: NewDB()}
}

// List lists all the bottles in the account optionally filtering by year.
func (b *BottleController) List(c *app.ListBottleContext) error {
	var bottles []*app.Bottle
	var err error
	if c.HasYears {
		bottles, err = b.db.GetBottlesByYears(c.AccountID, c.Years)
	} else {
		bottles, err = b.db.GetBottles(c.AccountID)
	}
	if err != nil {
		return err
	}
	return c.OK(bottles, "default")
}

// Show retrieves the bottle with the given id.
func (b *BottleController) Show(c *app.ShowBottleContext) error {
	bottle := b.db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	return c.OK(bottle, "default")
}

// Create records a new bottle.
func (b *BottleController) Create(c *app.CreateBottleContext) error {
	bottle := b.db.NewBottle(c.AccountID)
	payload := c.Payload
	bottle.Name = payload.Name
	bottle.Vintage = payload.Vintage
	bottle.Vineyard = payload.Vineyard
	if payload.Varietal != "" {
		bottle.Varietal = payload.Varietal
	}
	if payload.Color != "" {
		bottle.Color = payload.Color
	}
	if payload.Sweetness != 0 {
		bottle.Sweetness = payload.Sweetness
	}
	if payload.Country != "" {
		bottle.Country = payload.Country
	}
	if payload.Region != "" {
		bottle.Region = payload.Region
	}
	if payload.Review != "" {
		bottle.Review = payload.Review
	}
	c.ResponseHeader().Set("Location", app.BottleHref(c.AccountID, bottle.ID))
	return c.Created()
}

// Update updates a bottle field(s).
func (b *BottleController) Update(c *app.UpdateBottleContext) error {
	bottle := b.db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	payload := c.Payload
	if payload.Name != "" {
		bottle.Name = payload.Name
	}
	if payload.Vintage != 0 {
		bottle.Vintage = payload.Vintage
	}
	if payload.Vineyard != "" {
		bottle.Vineyard = payload.Vineyard
	}
	if payload.Varietal != "" {
		bottle.Varietal = payload.Varietal
	}
	if payload.Color != "" {
		bottle.Color = payload.Color
	}
	if payload.Sweetness != 0 {
		bottle.Sweetness = payload.Sweetness
	}
	if payload.Country != "" {
		bottle.Country = payload.Country
	}
	if payload.Region != "" {
		bottle.Region = payload.Region
	}
	if payload.Review != "" {
		bottle.Review = payload.Review
	}
	b.db.SaveBottle(bottle)
	return c.NoContent()
}

// Delete removes a bottle from the database.
func (b *BottleController) Delete(c *app.DeleteBottleContext) error {
	bottle := b.db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	b.db.DeleteBottle(bottle)
	return c.NoContent()
}

// Rate rates a bottle.
func (b *BottleController) Rate(c *app.RateBottleContext) error {
	bottle := b.db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	bottle.Rating = c.Payload.Rating
	b.db.SaveBottle(bottle)
	return c.NoContent()
}
