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

var ConstructorUnionHTTPReorderedDSL = func() {
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
			Payload(OneOf(JSONPayload, TextPayload))
			Result(OneOf(JSONPayload, TextPayload))
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

var ConstructorUnionRequiredBodyWithHeaderHTTPDSL = func() {
	var TextPayload = Type("RequiredBodyWithHeaderTextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("RequiredBodyWithHeaderJSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("ConstructorUnionRequiredBodyWithHeader", func() {
		Method("Show", func() {
			Payload(func() {
				Attribute("body", OneOf(TextPayload, JSONPayload))
				Attribute("token", String)
				Required("body", "token")
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

var ConstructorUnionNormalizedBranchNamesHTTPDSL = func() {
	var First = Type("FirstNormalizedBranch", func() {
		Meta("struct:type:name", "Foo Bar")
		Attribute("text", String)
		Required("text")
	})
	var Second = Type("SecondNormalizedBranch", func() {
		Meta("struct:type:name", "Foo-Bar")
		Attribute("count", Int)
		Required("count")
	})
	Service("ConstructorUnionNormalizedBranchNames", func() {
		Method("Show", func() {
			Payload(OneOf(First, Second))
			Result(OneOf(First, Second))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionRenamedTypesHTTPDSL = func() {
	var Alpha = Type("AlphaPayload", String, func() {
		TypeName("RenamedAlphaPayload")
	})
	var Beta = Type("BetaPayload", String, func() {
		TypeName("RenamedBetaPayload")
	})
	Service("ConstructorUnionRenamedTypes", func() {
		Method("Show", func() {
			Payload(OneOf(Alpha, Beta))
			Result(OneOf(Alpha, Beta))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionUnrelatedDeclarationOrderHTTPDSL = func() {
	var _ = Type("UnusedZeta", func() {
		Attribute("name", String)
	})
	var _ = Type("UnusedAlpha", func() {
		Attribute("count", Int)
	})
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("ConstructorUnionUnrelatedDeclarationOrder", func() {
		Method("Show", func() {
			Payload(OneOf(TextPayload, JSONPayload))
			Result(OneOf(TextPayload, JSONPayload))
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionUnrelatedDeclarationOrderReorderedHTTPDSL = func() {
	var _ = Type("UnusedAlpha", func() {
		Attribute("count", Int)
	})
	var _ = Type("UnusedZeta", func() {
		Attribute("name", String)
	})
	var JSONPayload = Type("JSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var TextPayload = Type("TextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	Service("ConstructorUnionUnrelatedDeclarationOrder", func() {
		Method("Show", func() {
			Payload(OneOf(TextPayload, JSONPayload))
			Result(OneOf(TextPayload, JSONPayload))
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

var ConstructorUnionAmbiguousUserExampleHTTPDSL = func() {
	var ObjectPayloadA = Type("ObjectPayloadA", func() {
		Attribute("message", String)
	})
	var ObjectPayloadB = Type("ObjectPayloadB", func() {
		Attribute("message", String)
	})
	var ExamplePayload = Type("AmbiguousExamplePayload", func() {
		Attribute("choice", OneOf(ObjectPayloadA, ObjectPayloadB), func() {
			Example(map[string]any{"message": "hello"})
		})
		Required("choice")
	})
	Service("ConstructorUnionAmbiguousUserExample", func() {
		Method("Show", func() {
			Payload(ExamplePayload)
			Result(ExamplePayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionExplicitWrappedUserExampleCustomKeysHTTPDSL = func() {
	var ObjectPayloadA = Type("ExplicitWrappedObjectPayloadA", func() {
		Attribute("message", String)
	})
	var ObjectPayloadB = Type("ExplicitWrappedObjectPayloadB", func() {
		Attribute("message", String)
	})
	var ExamplePayload = Type("ExplicitWrappedExamplePayload", func() {
		Attribute("choice", OneOf(ObjectPayloadA, ObjectPayloadB), func() {
			Meta("oneof:type:field", "kind")
			Meta("oneof:value:field", "data")
			Example(map[string]any{
				"kind": "ExplicitWrappedObjectPayloadB",
				"data": map[string]any{"message": "hello"},
			})
		})
		Required("choice")
	})
	Service("ConstructorUnionExplicitWrappedUserExampleCustomKeys", func() {
		Method("Show", func() {
			Payload(ExamplePayload)
			Result(ExamplePayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionAmbiguousUserExampleCustomKeysHTTPDSL = func() {
	var ObjectPayloadA = Type("AmbiguousCustomKeyObjectPayloadA", func() {
		Attribute("message", String)
	})
	var ObjectPayloadB = Type("AmbiguousCustomKeyObjectPayloadB", func() {
		Attribute("message", String)
	})
	var ExamplePayload = Type("AmbiguousCustomKeyExamplePayload", func() {
		Attribute("choice", OneOf(ObjectPayloadA, ObjectPayloadB), func() {
			Meta("oneof:type:field", "kind")
			Meta("oneof:value:field", "data")
			Example(map[string]any{"message": "hello"})
		})
		Required("choice")
	})
	Service("ConstructorUnionAmbiguousUserExampleCustomKeys", func() {
		Method("Show", func() {
			Payload(ExamplePayload)
			Result(ExamplePayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var ConstructorUnionMultipleExamplesHTTPDSL = func() {
	var TextPayload = Type("MultiExampleTextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("MultiExampleJSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var ExamplePayload = Type("MultiExamplePayload", func() {
		Attribute("choice", OneOf(TextPayload, JSONPayload), func() {
			Example("json", map[string]any{"message": "hello"})
			Example("text", map[string]any{"text": "world"})
		})
		Required("choice")
	})
	Service("ConstructorUnionMultipleExamples", func() {
		Method("Show", func() {
			Payload(ExamplePayload)
			Result(ExamplePayload)
			HTTP(func() {
				POST("/")
			})
		})
	})
}

var NestedMixedConstructorUnionExamplesCustomKeysHTTPDSL = func() {
	var InnerText = Type("NestedMixedInnerText", func() {
		Attribute("text", String)
		Required("text")
	})
	var InnerJSON = Type("NestedMixedInnerJSON", func() {
		Attribute("message", String)
		Required("message")
	})
	var OuterText = Type("NestedMixedOuterText", func() {
		Attribute("label", String)
		Required("label")
	})
	var OuterJSON = Type("NestedMixedOuterJSON", func() {
		Attribute("count", Int)
		Required("count")
	})
	var NestedPayload = Type("NestedMixedPayload", func() {
		Attribute("outer", func() {
			Attribute("choice", OneOf(OuterText, OuterJSON), func() {
				Example(map[string]any{"label": "from-user"})
			})
			Attribute("inner", OneOf(InnerText, InnerJSON), func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
			})
			Required("choice", "inner")
		})
		Required("outer")
	})
	Service("NestedMixedConstructorUnionExamplesCustomKeys", func() {
		Method("Show", func() {
			Payload(NestedPayload)
			Result(NestedPayload)
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

var ConstructorUnionClientValidatorReferenceHTTPDSL = func() {
	var All = Type("ClientValidatorReferenceAll", func() {
		Meta("name:original", "All")
		Meta("oneof:type:tag", "all")
	})
	var Single = Type("ClientValidatorReferenceSingle", func() {
		Meta("name:original", "Single")
		Meta("oneof:type:tag", "single")
		Attribute("task_id", String)
		Required("task_id")
	})
	var Batch = Type("ClientValidatorReferenceBatch", func() {
		Meta("name:original", "Batch")
		Meta("oneof:type:tag", "batch")
		Attribute("task_ids", ArrayOf(String), func() {
			MinLength(1)
		})
		Required("task_ids")
	})
	var PayloadType = Type("ClientValidatorReferencePayload", func() {
		Attribute("value", OneOf(All, Single, Batch), func() {
			Meta("oneof:typename", "ClientValidatorReferenceMode")
			Meta("oneof:type:field", "mode")
			Meta("oneof:value:field", "value")
		})
		Required("value")
	})
	Service("ClientValidatorReference", func() {
		Method("Show", func() {
			Payload(PayloadType)
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

var ConstructorUnionTraversalOrderHTTPDSL = func() {
	var AlphaText = Type("TraversalAlphaText", func() {
		Attribute("text", String)
		Required("text")
	})
	var AlphaJSON = Type("TraversalAlphaJSON", func() {
		Attribute("message", String)
		Required("message")
	})
	var BetaText = Type("TraversalBetaText", func() {
		Attribute("text", String)
		Required("text")
	})
	var BetaJSON = Type("TraversalBetaJSON", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("TraversalAlphaService", func() {
		Method("Alpha", func() {
			Payload(OneOf(AlphaText, AlphaJSON))
			Result(OneOf(AlphaText, AlphaJSON))
			HTTP(func() {
				POST("/alpha")
			})
		})
	})
	Service("TraversalBetaService", func() {
		Method("Beta", func() {
			Payload(OneOf(BetaText, BetaJSON))
			Result(OneOf(BetaText, BetaJSON))
			HTTP(func() {
				POST("/beta")
			})
		})
	})
}

var ConstructorUnionTraversalOrderReorderedHTTPDSL = func() {
	var AlphaText = Type("TraversalAlphaText", func() {
		Attribute("text", String)
		Required("text")
	})
	var AlphaJSON = Type("TraversalAlphaJSON", func() {
		Attribute("message", String)
		Required("message")
	})
	var BetaText = Type("TraversalBetaText", func() {
		Attribute("text", String)
		Required("text")
	})
	var BetaJSON = Type("TraversalBetaJSON", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("TraversalBetaService", func() {
		Method("Beta", func() {
			Payload(OneOf(BetaText, BetaJSON))
			Result(OneOf(BetaText, BetaJSON))
			HTTP(func() {
				POST("/beta")
			})
		})
	})
	Service("TraversalAlphaService", func() {
		Method("Alpha", func() {
			Payload(OneOf(AlphaText, AlphaJSON))
			Result(OneOf(AlphaText, AlphaJSON))
			HTTP(func() {
				POST("/alpha")
			})
		})
	})
}
