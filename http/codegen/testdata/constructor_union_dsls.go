package testdata

import . "goa.design/goa/v3/dsl"

var ConstructorUnionHTTPDSL = func() {
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("ConstructorUnion", func() {
		Method("Show", func() {
			Payload(OneOf(TextPayload, JSONPayload))
			Result(OneOf(TextPayload, JSONPayload))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionOptionalBodyWithHeaderHTTPDSL = func() {
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("ConstructorUnionOptionalBodyWithHeader", func() {
		Method("Show", func() {
			Payload(func() {
				Attribute("body", OneOf(TextPayload, JSONPayload))
				Attribute("token", String)
				Required("token")
			})
			HTTP(func() {
				POST("/")
				Body("body")
				Header("token:Authorization")
			})
		})
	})
}

var NestedConstructorUnionHTTPDSL = func() {
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var NestedPayload = Type("NestedPayload", func() {
		Attribute("choice", OneOf(TextPayload, JSONPayload))
		Required("choice")
	})
	Service("NestedConstructorUnion", func() {
		Method("Show", func() {
			Payload(NestedPayload)
			Result(NestedPayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionCustomKeysHTTPDSL = func() {
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("ConstructorUnionCustomKeys", func() {
		Method("Show", func() {
			Payload(OneOf(TextPayload, JSONPayload), func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
			})
			Result(OneOf(TextPayload, JSONPayload), func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionUserExampleSecondBranchHTTPDSL = func() {
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var ExamplePayload = Type("ExamplePayload", func() {
		Attribute("choice", OneOf(TextPayload, JSONPayload), func() {
			Example(map[string]any{"message": "hello"})
		})
		Required("choice")
	})
	Service("ConstructorUnionUserExampleSecondBranch", func() {
		Method("Show", func() {
			Payload(ExamplePayload)
			Result(ExamplePayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var NestedTopLevelConstructorUnionHTTPDSL = func() {
	var InnerA = Type("InnerA", func() {
		Attribute("text", String)
		Required("text")
	})
	var InnerB = Type("InnerB", func() {
		Attribute("message", String)
		Required("message")
	})
	var OuterA = Type("OuterA", func() {
		Attribute("choice", OneOf(InnerA, InnerB))
		Required("choice")
	})
	var OuterB = Type("OuterB", func() {
		Attribute("count", Int)
		Required("count")
	})
	Service("NestedTopLevelConstructorUnion", func() {
		Method("Show", func() {
			Payload(OneOf(OuterA, OuterB))
			Result(OneOf(OuterA, OuterB))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var NestedTopLevelConstructorUnionCustomKeysHTTPDSL = func() {
	var InnerA = Type("InnerA", func() {
		Attribute("text", String)
		Required("text")
	})
	var InnerB = Type("InnerB", func() {
		Attribute("message", String)
		Required("message")
	})
	var OuterA = Type("OuterA", func() {
		Attribute("choice", OneOf(InnerA, InnerB), func() {
			Meta("oneof:type:field", "kind")
			Meta("oneof:value:field", "data")
		})
		Required("choice")
	})
	var OuterB = Type("OuterB", func() {
		Attribute("count", Int)
		Required("count")
	})
	Service("NestedTopLevelConstructorUnionCustomKeys", func() {
		Method("Show", func() {
			Payload(OneOf(OuterA, OuterB), func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
			})
			Result(OneOf(OuterA, OuterB), func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
			})
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var RecursiveConstructorUnionHTTPDSL = func() {
	var Leaf = Type("Leaf", func() {
		Attribute("value", String)
		Required("value")
	})
	var Node = Type("Node", func() {
		Attribute("next", OneOf("Leaf", "Node"))
		Required("next")
	})
	Service("RecursiveConstructorUnion", func() {
		Method("Show", func() {
			Payload(OneOf(Leaf, Node))
			Result(OneOf(Leaf, Node))
			HTTP(func() {
				POST("/")
			})
		})
	})
}
