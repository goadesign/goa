package main

import "github.com/raphael/goa"

func main() {
	app := goa.New("cellar")

	c := app.NewController("bottle")
	c.Action("list", listBottles)
	c.Action("show", showBottle)
	c.Action("create", createBottle)
	c.Action("update", updateBottle)
	c.Action("delete", deleteBottle)
}

func listBottles(c *ListBottleContext) *goa.Response {
	var bottles []*db.Bottle
	if c.hasYear() {
		bottles = db.GetBottlesByYear(c.accountID(), c.year())
	} else {
		bottles = db.GetBottles(c.accountID())
	}
	resp := goa.Ok()
	resp.Body = BottleMediaType.Render(bottles)
	return resp
}

func showBottle(c *ShowBottleContext) *goa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	resp := goa.Ok()
	resp.Body = BottleMediaType.Render(bottle)
	return resp
}

func createBottle(c *CreateBottleContext) *goa.Response {
	bottle := db.BottleFromPayload(c.Payload())
	resp = goa.Created()
	resp.Body = BottleMediaType.Render(bottle)
	resp.Header.Set("Location", bottle.Href)
	return resp
}

func updateBottle(c *UpdateBottleContext) *goa.Response {
	bottle := db.GetBottle(c.accountID(), c.id())
	if bottle == nil {
		return goa.NotFound()
	}
	db.UpdateBottle(bottle, c)
	return goa.NoContent()
}

func deleteBottle(c *DeleteBottleContext) *goa.Response {
	if db.Delete(c.accountID(), c.id()) {
		return goa.NoContent()
	}
	return goa.NotFound()
}
