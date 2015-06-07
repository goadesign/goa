package main

import (
	"fmt"
	"time"

	"github.com/raphael/goa/examples/cellar/app/autogen"
)

// List all bottles in account optionally filtering by year
func ListBottles(c *autogen.ListBottleContext) error {
	var bottles []*db.Bottle
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
	bottle, err := db.GetBottle(c.AccountID(), c.ID())
	if err != nil {
		return err
	}
	if bottle == nil {
		c.NotFound()
		return nil
	}
	return c.OK(bottle)
}

// Record new bottle
func CreateBottle(c *autogen.CreateBottleContext) error {
	bottle := db.NewBottle()
	payload, err := c.Payload()
	if err != nil {
		return err
	}
	bottle.Name = payload.Name
	bottle.Vintage = payload.Vintage
	bottle.Vineyard = payload.Vineyard
	if payload.HasVarietal() {
		bottle.Varietal = payload.Varietal
	}
	if payload.HasColor() {
		bottle.Color = payload.Color
	}
	if payload.HasSweet() {
		bottle.Sweet = payload.Sweet
	}
	if payload.HasCountry() {
		bottle.Country = payload.Country
	}
	if payload.HasRegion() {
		bottle.Region = payload.Region
	}
	if payload.HasReview() {
		bottle.Review = payload.Review
	}
	if payload.HasCharacteristics() {
		bottle.Characteristics = payload.Characteristics
	}
	c.Header.Set("Location", href(bottle))
	c.Created() // Make that optional (use first 2xx response as default)?
	return nil
}

func UpdateBottle(c *autogen.UpdateBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		c.NotFound()
		return nil
	}
	payload, err := c.Payload()
	if err != nil {
		return err
	}
	if payload.HasName() {
		bottle.Name = payload.Name
	}
	if payload.HasVintage() {
		bottle.Vintage = payload.Vintage
	}
	if payload.HasVineyard() {
		bottle.Vineyard = payload.Vineyard
	}
	if payload.HasVarietal() {
		bottle.Varietal = payload.Varietal
	}
	if payload.HasColor() {
		bottle.Color = payload.Color
	}
	if payload.HasSweet() {
		bottle.Sweet = payload.Sweet
	}
	if payload.HasCountry() {
		bottle.Country = payload.Country
	}
	if payload.HasRegion() {
		bottle.Region = payload.Region
	}
	if payload.HasReview() {
		bottle.Review = payload.Review
	}
	if payload.HasCharacteristics() {
		bottle.Characteristics = payload.Characteristics
	}
	db.Save(bottle)
	c.NoContent()
	return nil
}

// Delete bottle
func DeleteBottle(c *autogen.DeleteBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		c.NotFound()
		return nil
	}
	err := db.Delete(bottle)
	if err != nil {
		return err
	}
	c.NoContent()
	return nil
}

func RateBottle(c *autogen.RateBottleContext) error {
	bottle := db.GetBottle(c.AccountID(), c.ID())
	if bottle == nil {
		c.NotFound()
		return nil
	}
	bottle.Ratings = c.Ratings()
	bottle.RatedAt = time.Now()
	db.Save(bottle)
	c.NoContent()
	return nil
}

// href computes a bottle API href.
func href(bottle) string {
	return fmt.Sprintf("/bottles/%d", bottle.ID)
}
