package controllers

import (
	"github.com/raphael/goa"
	"github.com/raphael/goa/examples/cellar/app"
)

// BottleController implements the bottle resource.
type BottleController struct {
	goa.Controller
	db *DB
}

// NewBottle creates a bottle controller.
func NewBottle(service goa.Service) *BottleController {
	return &BottleController{
		Controller: service.NewController("Bottle"),
		db:         NewDB(),
	}
}

// List lists all the bottles in the account optionally filtering by year.
func (b *BottleController) List(ctx *app.ListBottleContext) error {
	var bottles []*app.Bottle
	var err error
	if ctx.HasYears {
		bottles, err = b.db.GetBottlesByYears(ctx.AccountID, ctx.Years)
	} else {
		bottles, err = b.db.GetBottles(ctx.AccountID)
	}
	if err != nil {
		return ctx.NotFound()
	}
	return ctx.OK(bottles, "default")
}

// Show retrieves the bottle with the given id.
func (b *BottleController) Show(ctx *app.ShowBottleContext) error {
	bottle := b.db.GetBottle(ctx.AccountID, ctx.BottleID)
	if bottle == nil {
		return ctx.NotFound()
	}
	return ctx.OK(bottle, "default")
}

// Create records a new bottle.
func (b *BottleController) Create(ctx *app.CreateBottleContext) error {
	bottle := b.db.NewBottle(ctx.AccountID)
	payload := ctx.Payload
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
	ctx.Header().Set("Location", app.BottleHref(ctx.AccountID, bottle.ID))
	return ctx.Created()
}

// Update updates a bottle field(s).
func (b *BottleController) Update(ctx *app.UpdateBottleContext) error {
	bottle := b.db.GetBottle(ctx.AccountID, ctx.BottleID)
	if bottle == nil {
		return ctx.NotFound()
	}
	payload := ctx.Payload
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
	return ctx.NoContent()
}

// Delete removes a bottle from the database.
func (b *BottleController) Delete(ctx *app.DeleteBottleContext) error {
	bottle := b.db.GetBottle(ctx.AccountID, ctx.BottleID)
	if bottle == nil {
		return ctx.NotFound()
	}
	b.db.DeleteBottle(bottle)
	return ctx.NoContent()
}

// Rate rates a bottle.
func (b *BottleController) Rate(ctx *app.RateBottleContext) error {
	bottle := b.db.GetBottle(ctx.AccountID, ctx.BottleID)
	if bottle == nil {
		return ctx.NotFound()
	}
	bottle.Rating = ctx.Payload.Rating
	b.db.SaveBottle(bottle)
	return ctx.NoContent()
}
