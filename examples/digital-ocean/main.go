package main

import "github.com/labstack/echo"

func main() {
	e := echo.New()
	e.Get("/droplets/:id", getDroplet)
	e.Post("/droplets", createDroplet)
}

func (c *echo.Context) getDroplet() error {
}

func (c *echo.Context) createDroplet() error {
}
