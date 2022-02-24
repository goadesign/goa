package testdata

import . "goa.design/goa/v3/dsl"

var WithExampleDSL = func() {
	Service("WithExample", func() {
		Method("Method", func() {
			Payload(String, func() {
				Example("example")
			})
		})
	})
}

var WithArrayExampleDSL = func() {
	Service("WithArrayExample", func() {
		Method("Method", func() {
			Payload(ArrayOf(Int), func() {
				Example([]int{1, 2})
			})
		})
	})
}

var WithMapExampleDSL = func() {
	Service("WithMapExample", func() {
		Method("Method", func() {
			Payload(MapOf(String, Int), func() {
				Example(map[string]int{"name": 1, "value": 2})
			})
		})
	})
}

var WithMultipleExamplesDSL = func() {
	Service("WithMultipleExamples", func() {
		Method("Method", func() {
			Payload(Int, func() {
				Example(1)
				Example(100)
			})
		})
	})
}

var OverridingExampleDSL = func() {
	var AType = Type("AType", func() {
		Attribute("name", String, func() {
			Example("example")
		})
	})
	Service("OverridingExample", func() {
		Method("Method", func() {
			Payload(func() {
				Reference(AType)
				Attribute("name", String, func() {
					Example("overridden")
				})
			})
		})
	})
}

var WithExtendExampleDSL = func() {
	var AType = Type("AType", func() {
		Attribute("name", String, func() {
			Example("example")
		})
	})
	Service("OverridingExample", func() {
		Method("Method", func() {
			Payload(func() {
				Extend(AType)
			})
		})
	})
}

var InvalidExampleTypeDSL = func() {
	Service("InvalidExampleType", func() {
		Method("Method", func() {
			Payload(MapOf(String, Int), func() {
				Example(map[int]int{1: 1})
			})
		})
	})
}

var EmptyExampleDSL = func() {
	Service("EmptyExample", func() {
		Method("Method", func() {
			Payload(Int, func() {
				Example()
			})
		})
	})
}

var HidingExampleDSL = func() {
	Service("HidingExample", func() {
		Method("Method", func() {
			Payload(String, func() {
				Meta("openapi:example", "false")
			})
		})
	})
}

var OverridingHiddenExamplesDSL = func() {
	Service("OverridingHiddenExamples", func() {
		Meta("openapi:example", "false")
		Method("Method", func() {
			Payload(String, func() {
				Example("example")
			})
		})
	})
}
