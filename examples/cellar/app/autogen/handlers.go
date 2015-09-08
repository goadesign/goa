package autogen

import (
	"fmt"

	"github.com/raphael/goa"
)

func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"bottles", "list", "GET", "/accounts/:accountID/bottles", listBottlesHandler},
		&goa.HandlerFactory{"bottles", "show", "GET", "/accounts/:accountID/bottles/:ID", showBottlesHandler},
		&goa.HandlerFactory{"bottles", "create", "POST", "/accounts/:accountID/bottles", createBottlesHandler},
		&goa.HandlerFactory{"bottles", "update", "PUT", "/accounts/:accountID/bottles/:ID", updateBottlesHandler},
		&goa.HandlerFactory{"bottles", "delete", "DELETE", "/accounts/:accountID/bottles/:ID", deleteBottlesHandler},
		&goa.HandlerFactory{"bottles", "rate", "PATCH", "/accounts/:accountID/bottles/:ID", rateBottlesHandler},
	)
}

func listBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ListBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action list bottles, expected 'func(c *ListBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewListBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}

func showBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *ShowBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action show bottle, expected 'func(c *ShowBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewShowBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}

func createBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *CreateBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for create bottles, expected 'func(c *CreateBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewCreateBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}

func updateBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *UpdateBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for update bottles, expected 'func(c *UpdateBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewUpdateBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}

func rateBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *RateBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for rate bottles, expected 'func(c *RateBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewRateBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}

func deleteBottlesHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *DeleteBottleContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for delete bottles, expected 'func(c *DeleteBottleContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewDeleteBottleContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
