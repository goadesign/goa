package main

var _ = goa.RegisterHandlers(
	goa.Handler{"bottles", "list", listBottlesHandler},
	goa.Handler{"bottles", "show", showBottlesHandler},
	goa.Handler{"bottles", "create", createBottlesHandler},
	goa.Handler{"bottles", "update", updateBottlesHandler},
	goa.Handler{"bottles", "delete", deleteBottlesHandler},
)

func listBottlesHandler(c *goa.Context) *goa.Response {
	ctx := ListBottleContext{Context: c}
	return c.Action(&ctx)
}

func showBottlesHandler(c *goa.Context) *goa.Response {
	ctx := ShowBottleContext{Context: c}
	return c.Action(&ctx)
}

func createBottlesHandler(c *goa.Context) *goa.Response {
	ctx := CreateBottleContext{Context: c}
	return c.Action(&ctx)
}

func updateBottlesHandler(c *goa.Context) *goa.Response {
	ctx := UpdateBottleContext{Context: c}
	return c.Action(&ctx)
}

func deleteBottlesHandler(c *goa.Context) *goa.Response {
	ctx := DeleteBottleContext{Context: c}
	return c.Action(&ctx)
}
