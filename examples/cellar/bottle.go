package main

import (
	"time"

	"github.com/raphael/goa/examples/cellar/app"
	"github.com/raphael/goa/examples/cellar/db"
)

// ListBottles lists all the bottles in the account optionally filtering by year.
func ListBottles(c *app.ListBottleContext) error {
	var bottles []*app.Bottle
	var err error
	if c.HasYears {
		bottles, err = db.GetBottlesByYears(c.AccountID, c.Years)
	} else {
		bottles, err = db.GetBottles(c.AccountID)
	}
	if err != nil {
		return err
	}
	return c.OK(bottles, "default")
}

// ShowBottle retrieves the bottle with the given id.
func ShowBottle(c *app.ShowBottleContext) error {
	bottle := db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	return c.OK(bottle)
}

// CreateBottle records a new bottle.
func CreateBottle(c *app.CreateBottleContext) error {
	bottle := db.NewBottle(c.AccountID)
	payload := c.Payload
	bottle.Name = payload.Name
	bottle.Vintage = payload.Vintage
	bottle.Vineyard = payload.Vineyard
	if payload.Varietal != nil {
		bottle.Varietal = *payload.Varietal
	}
	if payload.Color != nil {
		bottle.Color = *payload.Color
	}
	if payload.Sweet != nil {
		bottle.Sweet = *payload.Sweet
	}
	if payload.Country != nil {
		bottle.Country = *payload.Country
	}
	if payload.Region != nil {
		bottle.Region = *payload.Region
	}
	if payload.Review != nil {
		bottle.Review = *payload.Review
	}
	c.Header().Set("Location", bottle.ComputeHref())
	return c.Created(bottle)
}

// UpdateBottle updates a bottle field(s).
func UpdateBottle(c *app.UpdateBottleContext) error {
	bottle := db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	payload := c.Payload
	if payload.Name != nil {
		bottle.Name = *payload.Name
	}
	if payload.Vintage != nil {
		bottle.Vintage = *payload.Vintage
	}
	if payload.Vineyard != nil {
		bottle.Vineyard = *payload.Vineyard
	}
	if payload.Varietal != nil {
		bottle.Varietal = *payload.Varietal
	}
	if payload.Color != nil {
		bottle.Color = *payload.Color
	}
	if payload.Sweet != nil {
		bottle.Sweet = *payload.Sweet
	}
	if payload.Country != nil {
		bottle.Country = *payload.Country
	}
	if payload.Region != nil {
		bottle.Region = *payload.Region
	}
	if payload.Review != nil {
		bottle.Review = *payload.Review
	}
	db.Save(bottle)
	return c.NoContent()
}

// DeleteBottle removes a bottle from the database.
func DeleteBottle(c *app.DeleteBottleContext) error {
	bottle := db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	db.Delete(bottle)
	return c.NoContent()
}

// RateBottle rates a bottle.
func RateBottle(c *app.RateBottleContext) error {
	bottle := db.GetBottle(c.AccountID, c.ID)
	if bottle == nil {
		return c.NotFound()
	}
	bottle.Ratings = c.Payload.Ratings
	bottle.RatedAt = time.Now()
	db.Save(bottle)
	return c.NoContent()
}
