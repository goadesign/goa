package autogen

import "github.com/raphael/goa"

var _ = goa.RegisterHandlers(
	&goa.Handler{"bottles", "list", listBottlesHandler},
	&goa.Handler{"bottles", "show", showBottlesHandler},
	&goa.Handler{"bottles", "create", createBottlesHandler},
	&goa.Handler{"bottles", "update", updateBottlesHandler},
	&goa.Handler{"bottles", "delete", deleteBottlesHandler},
)

func listBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx := ListBottleContext{Context: c}
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		fatalf("invalid handler signature for '%s', expected 'func(c *ListBottleContext) error'")
	}
	return h(&ctx)
}

func showBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx := ShowBottleContext{Context: c}
	h, ok := userHandler.(func(c *ShowBottleContext) error)
	if !ok {
		fatalf("invalid handler signature for '%s', expected 'func(c *ShowBottleContext) error'")
	}
	return h(&ctx)
}

func createBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx := CreateBottleContext{Context: c}
	h, ok := userHandler.(func(c *CreateBottleContext) error)
	if !ok {
		fatalf("invalid handler signature for '%s', expected 'func(c *CreateBottleContext) error'")
	}
	return h(&ctx)
}

func updateBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx := UpdateBottleContext{Context: c}
	h, ok := userHandler.(func(c *UpdateBottleContext) error)
	if !ok {
		fatalf("invalid handler signature for '%s', expected 'func(c *UpdateBottleContext) error'")
	}
	return h(&ctx)
}

func deleteBottlesHandler(userHandler interface{}, c *goa.Context) error {
	ctx := DeleteBottleContext{Context: c}
	h, ok := userHandler.(func(c *DeleteBottleContext) error)
	if !ok {
		fatalf("invalid handler signature for '%s', expected 'func(c *DeleteBottleContext) error'")
	}
	return h(&ctx)
}
