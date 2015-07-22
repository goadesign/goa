package autogen

import "github.com/raphael/goa"

func init() {
	goa.RegisterHandlers(
		&goa.Handler{"bottles", "list", listBottlesHandler, "GET", "/accounts/:accountID/bottles"},
		&goa.Handler{"bottles", "show", showBottlesHandler, "GET", "/accounts/:accountID/bottles/:ID"},
		&goa.Handler{"bottles", "create", createBottlesHandler, "POST", "/accounts/:accountID/bottles"},
		&goa.Handler{"bottles", "update", updateBottlesHandler, "PUT", "/accounts/:accountID/bottles/:ID"},
		&goa.Handler{"bottles", "delete", deleteBottlesHandler, "DELETE", "/accounts/:accountID/bottles/:ID"},
		&goa.Handler{"bottles", "rate", rateBottlesHandler, "PATCH", "/accounts/:accountID/bottles/:ID"},
	)
}

func listBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewListBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *ListBottleContext) error'")
	}
	return h(ctx)
}

func showBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewShowBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *ShowBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *ShowBottleContext) error'")
	}
	return h(ctx)
}

func createBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewCreateBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *CreateBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *CreateBottleContext) error'")
	}
	return h(ctx)
}

func updateBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewUpdateBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *UpdateBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *UpdateBottleContext) error'")
	}
	return h(ctx)
}

func rateBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewRateBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *RateBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *RateBottleContext) error'")
	}
	return h(ctx)
}

func deleteBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx, err := NewDeleteBottleContext(c)
	if err != nil {
		return err
	}
	h, ok := userHandler.(func(c *DeleteBottleContext) error)
	if !ok {
		goa.Fatalf("invalid handler signature for '%s', expected 'func(c *DeleteBottleContext) error'")
	}
	return h(ctx)
}
