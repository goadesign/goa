package main

import (
	"time"

	"github.com/raphael/goa/examples/cellar/app/autogen"
	"github.com/raphael/goa/examples/cellar/app/db"
)

// List all bottles in account optionally filtering by year
func ListBottles(c *autogen.ListBottleContext) error {
	var bottles []*autogen.Bottle
	var err error
	if c.HasYears() {
		bottles, err = db.GetBottlesByYears(c.AccountID(), c.Years())
	} else {
		bottles, err = db.GetBottles(c.AccountID())
	}
	if err != nil {
		return err
	}
	return c.OK(bottles)
}

// Retrieve bottle with given id
func ShowBottle(c *autogen.ShowBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		return c.NotFound()
	}
	return c.OK(bottle)
}

// Record new bottle
func CreateBottle(c *autogen.CreateBottleContext) error {
	bottle := db.NewBottle(c.AccountID())
	payload, err := c.Payload()
	if err != nil {
		return err
	}
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
	c.Header.Set("Location", bottle.ComputeHref())
	return c.Created()
}

func UpdateBottle(c *autogen.UpdateBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		return c.NotFound()
	}
	payload, err := c.Payload()
	if err != nil {
		return err
	}
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
	db.Save(bottle)
	return c.NoContent()
}

// Delete bottle
func DeleteBottle(c *autogen.DeleteBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		return c.NotFound()
	}
	db.Delete(bottle)
	return c.NoContent()
}

func RateBottle(c *autogen.RateBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		return c.NotFound()
	}
	bottle.Ratings = c.Ratings()
	bottle.RatedAt = time.Now()
	db.Save(bottle)
	return c.NoContent()
}
