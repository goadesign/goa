package design

import (
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var BottlePayload = Type("BottlePayload", func() {
	Attribute("name", func() {
		MinLength(2)
	})
	Attribute("vineyard", func() {
		MinLength(2)
	})
	Attribute("varietal", func() {
		MinLength(4)
	})
	Attribute("vintage", Integer, func() {
		Minimum(1900)
		Maximum(2020)
	})
	Attribute("color", func() {
		Enum("red", "white", "rose", "yellow", "sparkling")
	})
	Attribute("sweetness", Integer, func() {
		Minimum(1)
		Maximum(5)
	})
	Attribute("country", func() {
		MinLength(2)
	})
	Attribute("region")
	Attribute("review", func() {
		MinLength(10)
		MaxLength(300)
	})
	Attribute("characteristics", func() {
		MinLength(10)
		MaxLength(300)
	})
})
