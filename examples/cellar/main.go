package main

import (
	"fmt"
	"time"

	"github.com/raphael/goa"
)

func main() {
	app := goa.New("cellar")

	c := app.NewController("bottle")
	c.List(listBottles)
	c.Show(showBottle)
	c.Create(createBottle)
	c.Update(updateBottle)
	c.Delete(deleteBottle)
	c.Action("rate", rateBottle)
}

func listBottles(c *ListBottleContext) *goa.Response {
	var bottles []*db.Bottle
	if c.hasYear() {
		bottles = db.GetBottlesByYear(c.accountID(), c.year())
	} else {
		bottles = db.GetBottles(c.accountID())
	}
	return goa.Ok(BottleMediaType.Render(bottles))
}

func showBottle(c *ShowBottleContext) *oa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	return goa.Ok(BottleMediaType.Render(bottle))
}

func createBottle(c *CreateBottleContext) *goa.Response {
	bottle := db.NewBottle()
	payload = c.Payload()
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
	resp = goa.Created(BottleMediaType.Render(bottle))
	resp.Header.Set("Location", href(bottle))
	return resp
}

func updateBottle(c *UpdateBottleContext) *goa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	payload = c.Payload()
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
	return goa.NoContent()
}

func deleteBottle(c *DeleteBottleContext) *goa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	db.Delete(bottle)
	return goa.NoContent()
}

func rateBottle(c *RateBottleContext) *goa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	bottle.Ratings = c.Ratings()
	bottle.RatedAt = time.Now()
	db.Save(bottle)
	return goa.NoContent()
}

// href computes a bottle API href.
func href(bottle) string {
	return fmt.Sprintf("/bottles/%d", bottle.ID)
}
